package store

import (
	"testing"
)

// insertSample writes one sample row synchronously (bypassing the async writeLoop) so
// series tests have deterministic, fully-committed data. Same-package access to s.db.
func insertSample(t *testing.T, s *Store, tsUS int64, channel string, value float64) {
	t.Helper()
	if _, err := s.db.Exec(
		`INSERT INTO samples (ts_us, channel, value) VALUES (?, ?, ?)`, tsUS, channel, value,
	); err != nil {
		t.Fatalf("insert sample: %v", err)
	}
}

func TestSeriesRangeBoundariesInclusive(t *testing.T) {
	s := openTestStore(t)
	for _, ts := range []int64{10, 20, 30, 40, 50} {
		insertSample(t, s, ts, "p", float64(ts))
	}

	got, err := s.Series(20, 40, []string{"p"}, 0)
	if err != nil {
		t.Fatalf("Series: %v", err)
	}
	pts := got["p"]
	// Inclusive [20,40] => 20,30,40.
	if len(pts) != 3 {
		t.Fatalf("want 3 points, got %d: %v", len(pts), pts)
	}
	if pts[0][0] != 20 || pts[2][0] != 40 {
		t.Fatalf("boundary mismatch: %v", pts)
	}
	// Values track timestamps.
	for _, p := range pts {
		if p[0] != p[1] {
			t.Fatalf("value/ts mismatch: %v", p)
		}
	}
}

func TestSeriesChannelFilter(t *testing.T) {
	s := openTestStore(t)
	insertSample(t, s, 1, "a", 1)
	insertSample(t, s, 2, "b", 2)
	insertSample(t, s, 3, "c", 3)

	got, err := s.Series(0, 100, []string{"a", "c"}, 0)
	if err != nil {
		t.Fatalf("Series: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 channels, got %d", len(got))
	}
	if _, ok := got["b"]; ok {
		t.Fatal("channel b should not be returned (not requested)")
	}
	if len(got["a"]) != 1 || len(got["c"]) != 1 {
		t.Fatalf("expected one point each: %v", got)
	}
}

func TestSeriesEmptyChannelsUsesAllInRange(t *testing.T) {
	s := openTestStore(t)
	insertSample(t, s, 5, "a", 1)
	insertSample(t, s, 6, "b", 2)
	insertSample(t, s, 7, "c", 3)
	// Out of range — must NOT appear.
	insertSample(t, s, 200, "d", 4)

	got, err := s.Series(0, 100, nil, 0)
	if err != nil {
		t.Fatalf("Series: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("want 3 channels in range, got %d (%v)", len(got), keys(got))
	}
	if _, ok := got["d"]; ok {
		t.Fatal("channel d is out of range; should not be selected")
	}
}

func TestSeriesEmptyRangeYieldsEmptySlices(t *testing.T) {
	s := openTestStore(t)
	insertSample(t, s, 10, "p", 1)

	// Requested channel with no samples in range => empty (non-nil) slice.
	got, err := s.Series(1000, 2000, []string{"p", "q"}, 0)
	if err != nil {
		t.Fatalf("Series: %v", err)
	}
	for _, ch := range []string{"p", "q"} {
		v, ok := got[ch]
		if !ok {
			t.Fatalf("channel %q missing from result", ch)
		}
		if v == nil {
			t.Fatalf("channel %q should be empty non-nil slice, got nil", ch)
		}
		if len(v) != 0 {
			t.Fatalf("channel %q should be empty, got %v", ch, v)
		}
	}
}

func TestSeriesFromAfterToErrors(t *testing.T) {
	s := openTestStore(t)
	if _, err := s.Series(100, 50, nil, 0); err == nil {
		t.Fatal("want error when from > to")
	}
}

// TestSeriesDecimationCapAndSpike checks the cap is honored AND a spike is preserved.
// We insert a dense ramp plus one extreme spike; with a small cap the result must be
// near the cap (not the full count) and must still contain the spike value.
func TestSeriesDecimationCapAndSpike(t *testing.T) {
	s := openTestStore(t)
	const n = 4000
	spikeTS := int64(2500)
	const spikeVal = 99999.0
	for i := 0; i < n; i++ {
		ts := int64(i)
		v := float64(i % 100) // small oscillation, max 99
		if ts == spikeTS {
			v = spikeVal // the lone spike to preserve
		}
		insertSample(t, s, ts, "p", v)
	}

	const cap = 200
	got, err := s.Series(0, n-1, []string{"p"}, cap)
	if err != nil {
		t.Fatalf("Series: %v", err)
	}
	pts := got["p"]

	// Cap: each bucket emits up to 2 points; result must be well under the raw count
	// and within ~2x the cap.
	if len(pts) >= n {
		t.Fatalf("decimation did not reduce: %d points (raw %d)", len(pts), n)
	}
	if len(pts) > cap+2 {
		t.Fatalf("decimation exceeded cap: %d > %d", len(pts), cap+2)
	}

	// Spike preserved: the extreme max value must survive (min/max-per-bucket, not avg).
	foundSpike := false
	for _, p := range pts {
		if p[1] == spikeVal {
			foundSpike = true
			if int64(p[0]) != spikeTS {
				t.Fatalf("spike kept wrong timestamp: %v want ts %d", p, spikeTS)
			}
		}
	}
	if !foundSpike {
		t.Fatalf("spike value %v was averaged away by decimation", spikeVal)
	}

	// Points must be time-ordered (uPlot requires monotonic x).
	for i := 1; i < len(pts); i++ {
		if pts[i][0] < pts[i-1][0] {
			t.Fatalf("points not time-ordered at %d: %v", i, pts)
		}
	}
}

// TestSeriesNoDecimationUnderCap confirms a channel under the cap returns every point.
func TestSeriesNoDecimationUnderCap(t *testing.T) {
	s := openTestStore(t)
	for i := int64(0); i < 50; i++ {
		insertSample(t, s, i, "p", float64(i))
	}
	got, err := s.Series(0, 49, []string{"p"}, 4000)
	if err != nil {
		t.Fatalf("Series: %v", err)
	}
	if len(got["p"]) != 50 {
		t.Fatalf("want all 50 points (under cap), got %d", len(got["p"]))
	}
}

func TestJobSeriesUnknownJobNotOK(t *testing.T) {
	s := openTestStore(t)
	_, _, ok, err := s.JobSeries(999, nil, 0)
	if err != nil {
		t.Fatalf("JobSeries: %v", err)
	}
	if ok {
		t.Fatal("unknown job should yield ok=false")
	}
}

func TestJobSeriesOnlyInSegmentSamples(t *testing.T) {
	s := openTestStore(t)
	jobID := activeJob(t, s, "Job A")

	// Three eras of samples: before segment, inside segment, after segment (a gap).
	for ts := int64(100); ts <= 110; ts++ {
		insertSample(t, s, ts, "p", float64(ts))
	}
	for ts := int64(200); ts <= 210; ts++ {
		insertSample(t, s, ts, "p", float64(ts))
	}
	for ts := int64(300); ts <= 310; ts++ {
		insertSample(t, s, ts, "p", float64(ts))
	}

	// A single closed segment over [200,210] — only those samples should be returned.
	if _, err := s.db.Exec(
		`INSERT INTO recording_segments (job_id, started_at_us, stopped_at_us, created_at_us)
		 VALUES (?, 200, 210, 200)`, jobID,
	); err != nil {
		t.Fatalf("insert segment: %v", err)
	}

	segs, series, ok, err := s.JobSeries(jobID, []string{"p"}, 0)
	if err != nil {
		t.Fatalf("JobSeries: %v", err)
	}
	if !ok {
		t.Fatal("known job should yield ok=true")
	}
	if len(segs) != 1 {
		t.Fatalf("want 1 segment, got %d", len(segs))
	}
	pts := series["p"]
	if len(pts) != 11 { // ts 200..210 inclusive
		t.Fatalf("want 11 in-segment points, got %d: %v", len(pts), pts)
	}
	for _, p := range pts {
		if p[0] < 200 || p[0] > 210 {
			t.Fatalf("point outside segment leaked: %v", p)
		}
	}
}

func TestJobSeriesGapBetweenSegments(t *testing.T) {
	s := openTestStore(t)
	jobID := activeJob(t, s, "Job B")

	// Samples across a wide span; two segments with a gap between them.
	for ts := int64(0); ts <= 1000; ts++ {
		insertSample(t, s, ts, "p", float64(ts))
	}
	// Segment 1: [100,200], Segment 2: [500,600]; the 201..499 gap must be excluded.
	for _, seg := range [][2]int64{{100, 200}, {500, 600}} {
		if _, err := s.db.Exec(
			`INSERT INTO recording_segments (job_id, started_at_us, stopped_at_us, created_at_us)
			 VALUES (?, ?, ?, ?)`, jobID, seg[0], seg[1], seg[0],
		); err != nil {
			t.Fatalf("insert segment: %v", err)
		}
	}

	_, series, _, err := s.JobSeries(jobID, []string{"p"}, 0)
	if err != nil {
		t.Fatalf("JobSeries: %v", err)
	}
	for _, p := range series["p"] {
		ts := int64(p[0])
		inSeg := (ts >= 100 && ts <= 200) || (ts >= 500 && ts <= 600)
		if !inSeg {
			t.Fatalf("gap sample leaked into job series: ts=%d", ts)
		}
	}
	// Should have 101 + 101 = 202 points.
	if got := len(series["p"]); got != 202 {
		t.Fatalf("want 202 in-segment points, got %d", got)
	}
}

func keys(m map[string][]SeriesPoint) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
