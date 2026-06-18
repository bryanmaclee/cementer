package daqformat

// This file holds the bundled format presets and their channel vocabularies. A
// preset is PURE DATA — adding a new pump format means adding a value here (or
// loading one from config), never editing the engine. That is project axiom #2.

// Intellisense returns the DaqFormat for the Intellisense cement-pump DAQ,
// characterized from a live-wire capture (2026-06-16, 19200 8N1):
//
//   - 14 comma-separated fields, NO header.
//   - Column 0 is an HH:MM:SS uptime counter (resets on power-up), NOT a wall-clock
//     date — so the Reading.TS is the server stamp (decision D2). Column 0 is the
//     timestamp position and is therefore not field-mapped.
//   - Columns 1..13 feed the 13 channels below. The aggregates the DAQ emits
//     (agg.pressure, agg.rate, water.rate) are FIELD-MAPPED, not computed, because
//     the pump provides them directly. No computed channels.
//
// See docs/changes/phase2-intellisense-daqformat/intellisense-wire-capture-2026-06-16.md.
func Intellisense() DaqFormat {
	return DaqFormat{
		ID:             "intellisense",
		Name:           "Intellisense",
		Delimiter:      ",",
		HasHeader:      false,
		ExpectedFields: 14,
		Timestamp:      TimestampSpec{Column: 0, Kind: ServerStamp},
		Fields: []FieldMap{
			{Column: 1, ChannelID: "density.1"},
			{Column: 2, ChannelID: "agg.pressure"}, // DAQ-emitted sum of unit pressures
			{Column: 3, ChannelID: "agg.rate"},
			{Column: 4, ChannelID: "vol.job"},
			{Column: 5, ChannelID: "unit1.pressure"},
			{Column: 6, ChannelID: "unit2.pressure"},
			{Column: 7, ChannelID: "unit1.rate"},
			{Column: 8, ChannelID: "unit2.rate"},
			{Column: 9, ChannelID: "water.rate"},
			{Column: 10, ChannelID: "density.2"},
			{Column: 11, ChannelID: "vol.water.stage"},
			{Column: 12, ChannelID: "vol.stage"},
			{Column: 13, ChannelID: "job.number"},
		},
		// Computed is nil: Intellisense field-maps the aggregates it emits.
	}
}

// IntellisenseChannels returns the default channel vocabulary the Intellisense
// preset keys against: id/role/scope/uom/decimals/label. This is bundled metadata,
// NOT the Phase-3 Pump Profile CRUD. Units of measure are project-attested (the
// synthetic stream header + the rig: density 8.21 ppg confirmed at the interface).
func IntellisenseChannels() []Channel {
	return []Channel{
		{ID: "density.1", Role: "density", Scope: "unit", UnitIndex: 1, UoM: "ppg", Label: "Density", Decimals: 2},
		{ID: "agg.pressure", Role: "pressure", Scope: "aggregate", UoM: "psi", Label: "Pressure (total)", Decimals: 0},
		{ID: "agg.rate", Role: "rate", Scope: "aggregate", UoM: "bbl/min", Label: "Rate (total)", Decimals: 2},
		{ID: "vol.job", Role: "volume", Scope: "job", UoM: "bbl", Label: "Job Volume", Decimals: 1},
		{ID: "unit1.pressure", Role: "pressure", Scope: "unit", UnitIndex: 1, UoM: "psi", Label: "Unit 1 Pressure", Decimals: 0},
		{ID: "unit2.pressure", Role: "pressure", Scope: "unit", UnitIndex: 2, UoM: "psi", Label: "Unit 2 Pressure", Decimals: 0},
		{ID: "unit1.rate", Role: "rate", Scope: "unit", UnitIndex: 1, UoM: "bbl/min", Label: "Unit 1 Rate", Decimals: 2},
		{ID: "unit2.rate", Role: "rate", Scope: "unit", UnitIndex: 2, UoM: "bbl/min", Label: "Unit 2 Rate", Decimals: 2},
		{ID: "water.rate", Role: "rate", Scope: "aggregate", UoM: "bbl/min", Label: "Water Rate", Decimals: 2},
		{ID: "density.2", Role: "density", Scope: "unit", UnitIndex: 1, UoM: "ppg", Label: "Density (backup)", Decimals: 2},
		{ID: "vol.water.stage", Role: "volume", Scope: "stage", UoM: "bbl", Label: "Water Stage Volume", Decimals: 1},
		{ID: "vol.stage", Role: "volume", Scope: "stage", UoM: "bbl", Label: "Stage Volume", Decimals: 1},
		{ID: "job.number", Role: "meta", Scope: "job", UoM: "", Label: "Job Number", Decimals: 0},
	}
}

// Synthetic returns the Phase-1 development format: a 4-channel comma-separated
// stream (pressure,rate,density,volume) matching testdata/sample-stream.txt. It
// keeps the synthetic replay path working under the daqformat engine. There is no
// timestamp column, so the Reading.TS is the server stamp; comment ("#") lines are
// skipped by Apply.
func Synthetic() DaqFormat {
	return DaqFormat{
		ID:             "synthetic",
		Name:           "Synthetic (dev)",
		Delimiter:      ",",
		HasHeader:      false,
		ExpectedFields: 4,
		Timestamp:      TimestampSpec{Kind: ServerStamp},
		Fields: []FieldMap{
			{Column: 0, ChannelID: "pressure"},
			{Column: 1, ChannelID: "rate"},
			{Column: 2, ChannelID: "density"},
			{Column: 3, ChannelID: "volume"},
		},
	}
}

// SyntheticChannels returns the default channel vocabulary for the synthetic dev
// format (uom values per testdata/sample-stream.txt's header).
func SyntheticChannels() []Channel {
	return []Channel{
		{ID: "pressure", Role: "pressure", Scope: "aggregate", UoM: "psi", Label: "Pressure", Decimals: 0},
		{ID: "rate", Role: "rate", Scope: "aggregate", UoM: "bbl/min", Label: "Rate", Decimals: 2},
		{ID: "density", Role: "density", Scope: "unit", UnitIndex: 1, UoM: "ppg", Label: "Density", Decimals: 2},
		{ID: "volume", Role: "volume", Scope: "job", UoM: "bbl", Label: "Volume", Decimals: 1},
	}
}
