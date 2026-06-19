package store

// Series reads: the Phase-4 chart's historical query. This is a READ over the
// always-on samples store (axiom #1: the chart never gates or touches ingestion, the
// live stream, or recording) on the SAME single connection the writer uses (axiom #4
// / D2: the store stays the sole DB owner — no second *sql.DB, no handler-side DB).
// Reads are infrequent (chart load / pan), so sharing the one serialized connection
// with the writer is fine; the writer keeps priority via the batch writeLoop.
//
// Decimation: a wide range can hold far more points than a chart (or a browser) wants
// to draw. When a channel would exceed maxPerChannel we bucket the range into time
// buckets and emit the MIN and MAX sample of each bucket — never an average — so
// transient spikes (a pressure peak matters) survive the reduction. The cap is
// approximate: a bucket emits up to 2 points (min, max), so the result is ~2x the
// bucket count, kept near maxPerChannel.

import (
	"fmt"
	"sort"
	"time"
)

// SeriesPoint is one [ts_us, value] pair, the wire/JSON shape the chart consumes
// (a 2-element array — uPlot-friendly).
type SeriesPoint = [2]float64

// maxSeriesCap bounds maxPerChannel so a caller cannot ask for an unbounded read.
const maxSeriesCap = 20000

// Series returns, per channel, the [ts_us, value] points in [fromUS, toUS] (inclusive),
// decimated to ~maxPerChannel points/channel by min/max-per-bucket reduction when a
// channel exceeds the cap. Points are time-ordered. A requested channel with no samples
// in range maps to an empty (non-nil) slice. When channels is empty, all distinct
// channels that have samples in range are used.
//
// fromUS must be <= toUS (the caller validates and returns a 4xx; this guards too).
// maxPerChannel <= 0 defaults to 4000; it is clamped to maxSeriesCap.
func (s *Store) Series(fromUS, toUS int64, channels []string, maxPerChannel int) (map[string][]SeriesPoint, error) {
	if fromUS > toUS {
		return nil, fmt.Errorf("series: from (%d) must be <= to (%d)", fromUS, toUS)
	}
	if maxPerChannel <= 0 {
		maxPerChannel = 4000
	}
	if maxPerChannel > maxSeriesCap {
		maxPerChannel = maxSeriesCap
	}

	// Resolve the channel set. An empty request means "all channels present in range".
	want := channels
	if len(want) == 0 {
		all, err := s.channelsInRange(fromUS, toUS)
		if err != nil {
			return nil, err
		}
		want = all
	}

	out := make(map[string][]SeriesPoint, len(want))
	for _, ch := range want {
		pts, err := s.channelSeries(fromUS, toUS, ch, maxPerChannel)
		if err != nil {
			return nil, err
		}
		out[ch] = pts
	}
	return out, nil
}

// JobSeries returns a job's recording segments plus the sample series WITHIN those
// segments, per channel, for the chart's historical per-job view. The default chart
// shows only recorded data (data-model.md), so this reads samples inside segment
// windows only — the union span [earliest start, latest stop] bounds the read, but a
// sample is kept only if it falls inside SOME segment, so gaps between segments stay
// gaps. An open segment (stopped_at NULL) is treated as running up to now.
//
// ok is false (no error) when the job does not exist (the handler maps that to 404).
// channels empty => all channels with samples in the job's span. The per-channel cap
// applies to the in-segment points (decimated by the same min/max rule). This is a
// read on the single store connection (axioms #1 & #4: never gates ingestion).
func (s *Store) JobSeries(jobID int64, channels []string, maxPerChannel int) (segs []Segment, series map[string][]SeriesPoint, ok bool, err error) {
	// Confirm the job exists (404 vs empty-but-real job are different).
	var exists int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM jobs WHERE id = ?`, jobID).Scan(&exists); err != nil {
		return nil, nil, false, fmt.Errorf("job series job lookup: %w", err)
	}
	if exists == 0 {
		return nil, nil, false, nil
	}

	segs, err = s.ListSegments(jobID)
	if err != nil {
		return nil, nil, false, err
	}
	// No segments yet => a real job with empty series.
	if len(segs) == 0 {
		return segs, map[string][]SeriesPoint{}, true, nil
	}

	// Union span across segments (open segments extend to now).
	now := time.Now().UnixMicro()
	fromUS := segs[0].StartedAtUS
	var toUS int64
	for _, sg := range segs {
		if sg.StartedAtUS < fromUS {
			fromUS = sg.StartedAtUS
		}
		end := now
		if sg.StoppedAtUS != nil {
			end = *sg.StoppedAtUS
		}
		if end > toUS {
			toUS = end
		}
	}
	if toUS < fromUS {
		toUS = fromUS
	}

	// Read the union span, then keep only points that fall inside some segment so the
	// gaps between segments remain gaps. Decimation runs over the union read; the
	// in-segment filter trims it further, so the result stays within the cap.
	raw, err := s.Series(fromUS, toUS, channels, maxPerChannel)
	if err != nil {
		return nil, nil, false, err
	}
	series = make(map[string][]SeriesPoint, len(raw))
	for ch, pts := range raw {
		kept := make([]SeriesPoint, 0, len(pts))
		for _, p := range pts {
			if inAnySegment(int64(p[0]), segs, now) {
				kept = append(kept, p)
			}
		}
		series[ch] = kept
	}
	return segs, series, true, nil
}

// inAnySegment reports whether ts (unix-micros) falls inside any segment window
// [started, stopped] (an open segment's window ends at now). Segments are few, so a
// linear scan is fine.
func inAnySegment(ts int64, segs []Segment, now int64) bool {
	for _, sg := range segs {
		end := now
		if sg.StoppedAtUS != nil {
			end = *sg.StoppedAtUS
		}
		if ts >= sg.StartedAtUS && ts <= end {
			return true
		}
	}
	return false
}

// channelsInRange returns the distinct channel ids that have at least one sample in
// [fromUS, toUS]. Used when the caller does not name channels explicitly.
func (s *Store) channelsInRange(fromUS, toUS int64) ([]string, error) {
	rows, err := s.db.Query(
		`SELECT DISTINCT channel FROM samples WHERE ts_us BETWEEN ? AND ? ORDER BY channel`,
		fromUS, toUS,
	)
	if err != nil {
		return nil, fmt.Errorf("series channels: %w", err)
	}
	defer rows.Close()

	var chans []string
	for rows.Next() {
		var ch string
		if err := rows.Scan(&ch); err != nil {
			return nil, fmt.Errorf("scan series channel: %w", err)
		}
		chans = append(chans, ch)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate series channels: %w", err)
	}
	return chans, nil
}

// channelSeries reads one channel's points in range, time-ordered, applying min/max
// decimation when the row count exceeds maxPerChannel. It returns a non-nil slice
// (possibly empty) so the JSON shape is a [] not null.
func (s *Store) channelSeries(fromUS, toUS int64, channel string, maxPerChannel int) ([]SeriesPoint, error) {
	// Count first: a cheap COUNT(*) on the composite index decides whether to decimate.
	var n int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM samples WHERE channel = ? AND ts_us BETWEEN ? AND ?`,
		channel, fromUS, toUS,
	).Scan(&n); err != nil {
		return nil, fmt.Errorf("series count %q: %w", channel, err)
	}
	if n == 0 {
		return []SeriesPoint{}, nil
	}
	if n <= maxPerChannel {
		return s.rawChannelSeries(fromUS, toUS, channel, n)
	}
	return s.decimatedChannelSeries(fromUS, toUS, channel, maxPerChannel)
}

// rawChannelSeries reads every point for a channel in range, time-ordered. capHint
// pre-sizes the slice.
func (s *Store) rawChannelSeries(fromUS, toUS int64, channel string, capHint int) ([]SeriesPoint, error) {
	rows, err := s.db.Query(
		`SELECT ts_us, value FROM samples
		   WHERE channel = ? AND ts_us BETWEEN ? AND ?
		   ORDER BY ts_us, id`,
		channel, fromUS, toUS,
	)
	if err != nil {
		return nil, fmt.Errorf("series rows %q: %w", channel, err)
	}
	defer rows.Close()

	pts := make([]SeriesPoint, 0, capHint)
	for rows.Next() {
		var ts int64
		var v float64
		if err := rows.Scan(&ts, &v); err != nil {
			return nil, fmt.Errorf("scan series point %q: %w", channel, err)
		}
		pts = append(pts, SeriesPoint{float64(ts), v})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate series points %q: %w", channel, err)
	}
	return pts, nil
}

// decimatedChannelSeries reduces a channel's points to ~maxPerChannel by bucketing the
// time range and emitting the MIN and MAX sample per bucket (preserving spikes). Bucket
// count is maxPerChannel/2 (each bucket yields up to 2 points). Each emitted point keeps
// its REAL timestamp (the ts at which the min/max actually occurred), and the two points
// per bucket are ordered by time so the line draws correctly. A spike (a lone high
// value) becomes that bucket's max and is therefore never averaged away.
func (s *Store) decimatedChannelSeries(fromUS, toUS int64, channel string, maxPerChannel int) ([]SeriesPoint, error) {
	buckets := maxPerChannel / 2
	if buckets < 1 {
		buckets = 1
	}

	// Bucket over the channel's ACTUAL data extent within the requested range, not the
	// raw [from,to] span. Otherwise a query whose range dwarfs the data (e.g. from=0)
	// makes huge buckets that collapse all points into one — the decimation degenerates.
	// Bucketing over [dataMin, dataMax] keeps the reduction uniform regardless of how
	// wide the caller's range is. Each emitted point still keeps its real timestamp.
	var dataMin, dataMax int64
	if err := s.db.QueryRow(
		`SELECT MIN(ts_us), MAX(ts_us) FROM samples WHERE channel = ? AND ts_us BETWEEN ? AND ?`,
		channel, fromUS, toUS,
	).Scan(&dataMin, &dataMax); err != nil {
		return nil, fmt.Errorf("series extent %q: %w", channel, err)
	}
	span := dataMax - dataMin
	if span < 1 {
		span = 1
	}
	// bucketWidth >= 1 so integer division places points deterministically.
	bucketWidth := span / int64(buckets)
	if bucketWidth < 1 {
		bucketWidth = 1
	}

	rows, err := s.db.Query(
		`SELECT ts_us, value FROM samples
		   WHERE channel = ? AND ts_us BETWEEN ? AND ?
		   ORDER BY ts_us, id`,
		channel, fromUS, toUS,
	)
	if err != nil {
		return nil, fmt.Errorf("series rows %q: %w", channel, err)
	}
	defer rows.Close()

	// One accumulator per non-empty bucket, in first-seen (time) order.
	type acc struct {
		minTS, maxTS int64
		minV, maxV   float64
	}
	order := make([]int64, 0, buckets)
	accs := make(map[int64]*acc, buckets)

	for rows.Next() {
		var ts int64
		var v float64
		if err := rows.Scan(&ts, &v); err != nil {
			return nil, fmt.Errorf("scan decimated point %q: %w", channel, err)
		}
		b := (ts - dataMin) / bucketWidth
		// A point exactly at dataMax (or rounding at the tail) can land one past the last
		// bucket; fold it into the final bucket so the bucket count stays == buckets and
		// the emitted-point cap holds.
		if b >= int64(buckets) {
			b = int64(buckets) - 1
		}
		a := accs[b]
		if a == nil {
			a = &acc{minTS: ts, maxTS: ts, minV: v, maxV: v}
			accs[b] = a
			order = append(order, b)
			continue
		}
		if v < a.minV {
			a.minV, a.minTS = v, ts
		}
		if v > a.maxV {
			a.maxV, a.maxTS = v, ts
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate decimated points %q: %w", channel, err)
	}

	pts := make([]SeriesPoint, 0, len(order)*2)
	for _, b := range order {
		a := accs[b]
		// Emit min and max ordered by their real timestamps so the trace is monotonic
		// in time within the bucket. When min==max (or same ts) emit a single point.
		if a.minTS == a.maxTS {
			pts = append(pts, SeriesPoint{float64(a.minTS), a.minV})
			continue
		}
		if a.minTS < a.maxTS {
			pts = append(pts, SeriesPoint{float64(a.minTS), a.minV}, SeriesPoint{float64(a.maxTS), a.maxV})
		} else {
			pts = append(pts, SeriesPoint{float64(a.maxTS), a.maxV}, SeriesPoint{float64(a.minTS), a.minV})
		}
	}
	// order is already time-ascending (buckets discovered in scan order) but the
	// within-bucket min/max ordering above keeps it globally sorted; a defensive sort
	// guarantees monotonic time for uPlot regardless of bucket boundaries.
	sort.Slice(pts, func(i, j int) bool { return pts[i][0] < pts[j][0] })
	return pts, nil
}
