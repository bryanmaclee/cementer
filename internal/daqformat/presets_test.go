package daqformat

import (
	"testing"
	"time"
)

// TestIntellisenseRealLine maps a REAL captured idle line through the Intellisense
// preset and asserts the full 13-channel set and a few load-bearing values.
// Source line (idle): captures/capture-2026-06-16T150318-19200-8N1.bin
//
//	14:53:41,0.04,0,0.00,42.5,0,0,0.00,0.00,0.0,0.00,0.0,42.5,0
func TestIntellisenseRealLine(t *testing.T) {
	e := New(Intellisense())
	serverTS := time.Unix(1700000000, 0)
	line := "14:53:41,0.04,0,0.00,42.5,0,0,0.00,0.00,0.0,0.00,0.0,42.5,0"

	r, ok := e.Apply([]byte(line), serverTS)
	if !ok {
		t.Fatal("expected the real idle line to parse")
	}

	// TS is the server stamp (col-0 HH:MM:SS uptime is a hint, not a date).
	if !r.TS.Equal(serverTS) {
		t.Errorf("TS = %v, want server stamp %v", r.TS, serverTS)
	}

	// Exactly the 13 mapped channels (col-0 timestamp not mapped).
	want := map[string]float64{
		"density.1":       0.04,
		"agg.pressure":    0,
		"agg.rate":        0.00,
		"vol.job":         42.5,
		"unit1.pressure":  0,
		"unit2.pressure":  0,
		"unit1.rate":      0.00,
		"unit2.rate":      0.00,
		"water.rate":      0.0,
		"density.2":       0.00,
		"vol.water.stage": 0.0,
		"vol.stage":       42.5,
		"job.number":      0,
	}
	if len(r.Values) != len(want) {
		t.Fatalf("len(values) = %d, want %d (%v)", len(r.Values), len(want), r.Values)
	}
	for ch, wv := range want {
		got, present := r.Values[ch]
		if !present {
			t.Errorf("channel %q missing", ch)
			continue
		}
		if got != wv {
			t.Errorf("values[%q] = %v, want %v", ch, got, wv)
		}
	}
	// Spot the brief's named assertions explicitly.
	if r.Values["density.1"] != 0.04 {
		t.Errorf("density.1 = %v, want 0.04", r.Values["density.1"])
	}
	if r.Values["vol.job"] != 42.5 {
		t.Errorf("vol.job = %v, want 42.5", r.Values["vol.job"])
	}
	if r.Values["vol.stage"] != 42.5 {
		t.Errorf("vol.stage = %v, want 42.5", r.Values["vol.stage"])
	}
}

// TestIntellisensePressureSum confirms the field-mapped aggregate relationship from
// the live wire: with only unit 1 pressurized, agg.pressure == unit1.pressure and
// unit2.pressure == 0. Line from the pressure capture (valve closing):
//
//	16:04:xx ... cols 2 & 5 both 1306. Use a representative pressurized line.
func TestIntellisensePressureSum(t *testing.T) {
	e := New(Intellisense())
	// col2 (agg.pressure) and col5 (unit1.pressure) track together; col6 (unit2) is 0.
	line := "16:04:10,0.00,1306,0.00,2.5,1306,0,0.00,0.00,0.0,0.00,0.0,2.5,0"
	r, ok := e.Apply([]byte(line), time.Unix(0, 0))
	if !ok {
		t.Fatal("expected ok")
	}
	if r.Values["agg.pressure"] != 1306 {
		t.Errorf("agg.pressure = %v, want 1306", r.Values["agg.pressure"])
	}
	if r.Values["unit1.pressure"] != 1306 {
		t.Errorf("unit1.pressure = %v, want 1306", r.Values["unit1.pressure"])
	}
	if r.Values["unit2.pressure"] != 0 {
		t.Errorf("unit2.pressure = %v, want 0", r.Values["unit2.pressure"])
	}
	// agg.pressure == unit1.pressure + unit2.pressure (single-unit rig).
	if r.Values["agg.pressure"] != r.Values["unit1.pressure"]+r.Values["unit2.pressure"] {
		t.Errorf("agg.pressure (%v) != unit1.pressure (%v) + unit2.pressure (%v)",
			r.Values["agg.pressure"], r.Values["unit1.pressure"], r.Values["unit2.pressure"])
	}
}

// TestIntellisenseTornLineSkipped confirms the power-interruption fragment is
// dropped by the field-count guard (not 14 fields).
func TestIntellisenseTornLineSkipped(t *testing.T) {
	e := New(Intellisense())
	if _, ok := e.Apply([]byte("?,,,,,,,,,,,,,00:00:00,extra,more"), time.Unix(0, 0)); ok {
		t.Error("torn fragment should be skipped by the 14-field guard")
	}
}

// TestSyntheticPreset keeps the Phase-1 4-channel replay layout green.
func TestSyntheticPreset(t *testing.T) {
	e := New(Synthetic())
	r, ok := e.Apply([]byte("5377,5.08,13.77,10.5"), time.Unix(0, 0))
	if !ok {
		t.Fatal("expected ok")
	}
	want := map[string]float64{"pressure": 5377, "rate": 5.08, "density": 13.77, "volume": 10.5}
	if len(r.Values) != len(want) {
		t.Fatalf("len(values) = %d, want %d (%v)", len(r.Values), len(want), r.Values)
	}
	for ch, wv := range want {
		if r.Values[ch] != wv {
			t.Errorf("values[%q] = %v, want %v", ch, r.Values[ch], wv)
		}
	}
	// Comment lines are skipped (sample-stream.txt has "#" headers).
	if _, ok := e.Apply([]byte("# stage 1"), time.Unix(0, 0)); ok {
		t.Error("comment line should be skipped")
	}
}

// TestIntellisenseChannelsCoverPreset asserts the channel vocabulary has metadata
// for every field-mapped channel id (no orphan mappings, no missing channels).
func TestIntellisenseChannelsCoverPreset(t *testing.T) {
	f := Intellisense()
	chans := IntellisenseChannels()
	byID := make(map[string]Channel, len(chans))
	for _, c := range chans {
		byID[c.ID] = c
	}
	if len(chans) != len(f.Fields) {
		t.Errorf("channel count %d != field-map count %d", len(chans), len(f.Fields))
	}
	for _, fm := range f.Fields {
		if _, ok := byID[fm.ChannelID]; !ok {
			t.Errorf("field-mapped channel %q has no Channel metadata", fm.ChannelID)
		}
	}
}
