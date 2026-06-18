package daqformat

import (
	"strconv"
	"strings"
	"time"

	"github.com/bryanmaclee/cementer/internal/model"
)

// Engine applies a DaqFormat to raw lines, producing channel-keyed Readings. It
// owns a monotonic per-process sequence counter and is safe for use from a single
// ingest goroutine (the same contract as the Phase-1 parser).
type Engine struct {
	fmt DaqFormat
	seq int64
	// headerSkipped tracks whether the leading header line has been consumed
	// (only relevant when fmt.HasHeader is true).
	headerSkipped bool
}

// New builds an Engine for the given format. An empty Delimiter defaults to ",".
func New(f DaqFormat) *Engine {
	if f.Delimiter == "" {
		f.Delimiter = ","
	}
	return &Engine{fmt: f}
}

// Format returns the DaqFormat this engine applies.
func (e *Engine) Format() DaqFormat { return e.fmt }

// Apply converts one raw line into a channel-keyed Reading stamped per the
// format's TimestampSpec. ok is false for lines that carry no usable data: blank,
// comment ("#"), the header (when HasHeader), a torn line whose token count is not
// ExpectedFields, or a line that maps to zero values. Skipped lines are already
// safe in the raw log regardless; this only governs structured storage/broadcast.
//
// serverTS is the ingest timestamp; it is used as Reading.TS for ServerStamp and
// HMSUptime formats (the only kinds implemented today), since those carry no
// absolute date on the wire.
func (e *Engine) Apply(line []byte, serverTS time.Time) (model.Reading, bool) {
	s := strings.TrimSpace(string(line))
	if s == "" || strings.HasPrefix(s, "#") {
		return model.Reading{}, false
	}

	fields := strings.Split(s, e.fmt.Delimiter)

	// Header skip: the first non-blank, non-comment line is the column header.
	if e.fmt.HasHeader && !e.headerSkipped {
		e.headerSkipped = true
		return model.Reading{}, false
	}

	// Field-count guard: only steady-state lines of the exact expected shape are
	// accepted. Torn fragments (e.g. "?,,,,,,,,,,,,,00:00:00,...") are dropped from
	// structured storage; the raw log already kept the bytes. ExpectedFields <= 0
	// disables the guard (no shape known).
	if e.fmt.ExpectedFields > 0 && len(fields) != e.fmt.ExpectedFields {
		return model.Reading{}, false
	}

	ts := e.timestamp(fields, serverTS)

	values := make(map[string]float64, len(e.fmt.Fields)+len(e.fmt.Computed))
	for _, fm := range e.fmt.Fields {
		if fm.Column < 0 || fm.Column >= len(fields) {
			continue
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(fields[fm.Column]), 64)
		if err != nil {
			// Tolerate an individual unparseable field; omit just that channel.
			continue
		}
		// Warmup negatives (density/pressure < 0) are real, not errors — pass them
		// through unfiltered. Any clamping is a display concern (Phase 4).
		values[fm.ChannelID] = fm.Transform.Apply(v)
	}

	// Compute pass: derive channels from already-mapped values. No-op when the
	// format field-maps its aggregates (e.g. Intellisense).
	for _, cc := range e.fmt.Computed {
		v, ok := compute(cc, values)
		if !ok {
			continue
		}
		values[cc.ChannelID] = cc.Transform.Apply(v)
	}

	if len(values) == 0 {
		return model.Reading{}, false
	}

	e.seq++
	return model.Reading{Seq: e.seq, TS: ts, Values: values}, true
}

// timestamp resolves the Reading.TS for a line. ServerStamp and HMSUptime both
// use the server clock (the wire carries no absolute date). Other kinds are
// reserved and not yet implemented; they fall back to the server stamp so an
// unfinished spec degrades safely rather than producing a wrong date.
func (e *Engine) timestamp(_ []string, serverTS time.Time) time.Time {
	switch e.fmt.Timestamp.Kind {
	case ServerStamp, HMSUptime:
		return serverTS
	default:
		return serverTS
	}
}

// compute reduces a ComputedChannel's inputs to a single value. It returns ok =
// false when no input value is present in the frame (nothing to compute from).
// Missing individual inputs are skipped; a "mean" averages only present inputs.
func compute(cc ComputedChannel, values map[string]float64) (float64, bool) {
	var sum float64
	var n int
	for _, id := range cc.Inputs {
		v, present := values[id]
		if !present {
			continue
		}
		sum += v
		n++
	}
	if n == 0 {
		return 0, false
	}
	switch cc.Op {
	case "sum":
		return sum, true
	case "mean":
		return sum / float64(n), true
	default:
		// Unknown op: don't guess — skip the computed channel.
		return 0, false
	}
}
