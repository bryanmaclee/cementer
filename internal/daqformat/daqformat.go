// Package daqformat is the generic, config-driven mapping + compute engine that
// turns a raw ASCII line off a pump into a channel-keyed model.Reading.
//
// The whole point (project axiom #2) is that adapting to a new pump format is
// CONFIGURATION, not code: the engine is generic, a format is a plain DaqFormat
// value (data). A new pump format must require only a new DaqFormat — never an
// edit to this engine, the parser, the store, the hub, or the web client.
//
// A DaqFormat describes the wire: delimiter, optional header, where the timestamp
// lives (and how to interpret it), which columns feed which channels (with an
// optional scale/offset transform), and any channels that must be COMPUTED from
// other channels (for pumps that don't emit aggregates themselves).
//
// See docs/design/data-model.md (the normative design) and
// docs/changes/phase2-intellisense-daqformat/ for the Intellisense characterization.
package daqformat

// DaqFormat describes how one pump's serial format maps to channels. It is pure
// data — a preset (e.g. Intellisense) is just a DaqFormat literal.
type DaqFormat struct {
	// ID is a stable key for the format, e.g. "intellisense".
	ID string
	// Name is a human label, e.g. "Intellisense".
	Name string
	// Delimiter splits a line into fields. Empty is treated as ",".
	Delimiter string
	// HasHeader is true when the first non-blank, non-comment line is a column
	// header that must be skipped rather than parsed.
	HasHeader bool
	// ExpectedFields is the exact token count a valid steady-state line must have.
	// A line whose token count differs is skipped (torn fragments at power
	// interruption, partial reads). The raw log keeps the bytes regardless.
	ExpectedFields int
	// Timestamp says where the line's timestamp lives and how to interpret it.
	Timestamp TimestampSpec
	// Fields maps raw columns to channels (with an optional transform).
	Fields []FieldMap
	// Computed derives channels from already-mapped channels (sum/mean). It is a
	// no-op for pumps that emit their aggregates directly (those are field-mapped).
	Computed []ComputedChannel
}

// FieldMap maps one raw column to one channel, with an optional transform.
type FieldMap struct {
	// Column is the zero-based index of the field in the split line.
	Column int
	// ChannelID is the channel this column feeds, e.g. "unit1.pressure".
	ChannelID string
	// Transform optionally rescales the raw value (nil = identity).
	Transform *Transform
}

// Transform is a no-code linear rescale applied to a value: out = in*Scale + Offset.
// A nil *Transform is the identity.
type Transform struct {
	Scale  float64
	Offset float64
}

// Apply returns t applied to v. A nil receiver is the identity, so callers can
// hold an optional *Transform and call it unconditionally.
func (t *Transform) Apply(v float64) float64 {
	if t == nil {
		return v
	}
	return v*t.Scale + t.Offset
}

// ComputedChannel derives a channel from other channels' values within the same
// frame. Op is "sum" or "mean" over Inputs; the result is then run through
// Transform. This covers pumps that do NOT emit an aggregate themselves.
type ComputedChannel struct {
	// ChannelID is the derived channel, e.g. "agg.rate".
	ChannelID string
	// Op is the reduction over Inputs: "sum" or "mean".
	Op string
	// Inputs are the channel ids whose values feed Op.
	Inputs []string
	// Transform optionally rescales the computed value (nil = identity).
	Transform *Transform
}

// TimestampKind enumerates how a line's timestamp column is interpreted. Only the
// kinds the shipped formats need are implemented today (ServerStamp); the rest are
// reserved so a new format can add one without redesigning the engine.
type TimestampKind int

const (
	// ServerStamp ignores any embedded value and uses the ingest server clock.
	// This is correct when the wire carries no absolute date — e.g. Intellisense,
	// whose column 0 is an HH:MM:SS uptime counter, not a wall-clock date.
	ServerStamp TimestampKind = iota
	// HMSUptime marks an HH:MM:SS uptime column. It is recognized as the timestamp
	// position (so it is not field-mapped) but, being uptime not a date, the
	// Reading.TS is still the server stamp. Reserved for future use; today the
	// Intellisense preset uses ServerStamp directly and leaves column 0 unmapped.
	HMSUptime
	// ExcelSerial is an Excel serial day-number (epoch 1899-12-30). Reserved; not
	// implemented yet (no shipped format needs it — the live Intellisense wire has
	// no Excel-serial value; that was a property of the file EXPORTS only).
	ExcelSerial
	// Unix is whole/fractional Unix seconds. Reserved; not implemented yet.
	Unix
	// RFC3339 is an RFC 3339 timestamp string. Reserved; not implemented yet.
	RFC3339
)

// TimestampSpec locates and interprets a line's timestamp.
type TimestampSpec struct {
	// Column is the zero-based index of the timestamp field. It is recognized as
	// the timestamp position so it is never accidentally field-mapped.
	Column int
	// Kind is how the column is interpreted. For ServerStamp/HMSUptime the
	// Reading.TS is the server stamp regardless of the column's contents.
	Kind TimestampKind
}

// Channel is the minimal bundled metadata the engine (and, later, the client)
// needs to display a channel: id/role/scope/uom/decimals/label. This is NOT the
// Phase-3 Pump Profile CRUD — just a default vocabulary a preset keys against.
type Channel struct {
	// ID is the stable key, e.g. "unit1.pressure".
	ID string
	// Role is the physical quantity: pressure | rate | density | volume | meta | ...
	Role string
	// Scope is the topology: "unit", "aggregate", "stage", "job", or "" (none).
	Scope string
	// UnitIndex is the 1-based pumping-unit index when Scope == "unit"; 0 otherwise.
	UnitIndex int
	// UoM is the unit of measure, e.g. "psi", "bbl/min", "ppg", "bbl".
	UoM string
	// Label is the display label, e.g. "Unit 1 Pressure".
	Label string
	// Decimals is the display precision.
	Decimals int
}
