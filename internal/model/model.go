// Package model holds the shared data contracts that flow through the cementer
// pipeline: a single parsed frame (Reading) and its long-form storage unit (Sample).
package model

import "time"

// Reading is one parsed frame from a single ASCII line off the pump: a timestamp
// plus the channel values present on that line (e.g. pressure, rate, density,
// volume). It is the unit that gets broadcast live to clients.
type Reading struct {
	// Seq is a monotonic per-process sequence number assigned at ingest. It lets
	// clients detect gaps and is handy for debugging. It is not persisted as a key.
	Seq int64 `json:"seq"`
	// TS is the ingest timestamp (server clock). When the pump line carries its own
	// timestamp we can switch to that in the parser; for now the server stamps it.
	TS time.Time `json:"ts"`
	// Values maps channel name -> numeric value for this frame.
	Values map[string]float64 `json:"values"`
}

// Samples expands a Reading into its long-form storage rows, one per channel.
func (r Reading) Samples() []Sample {
	out := make([]Sample, 0, len(r.Values))
	for ch, v := range r.Values {
		out = append(out, Sample{TS: r.TS, Channel: ch, Value: v})
	}
	return out
}

// Sample is the long-form storage unit: one channel's value at one instant.
type Sample struct {
	TS      time.Time `json:"ts"`
	Channel string    `json:"channel"`
	Value   float64   `json:"value"`
}
