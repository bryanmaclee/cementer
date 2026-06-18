package daqformat

import (
	"testing"
	"time"
)

// crafted4 is a tiny 4-field format for exercising the engine mechanics without
// depending on a preset: col0 timestamp (server-stamp), col1->a, col2->b, col3->c.
func crafted4() DaqFormat {
	return DaqFormat{
		ID:             "crafted4",
		Name:           "Crafted4",
		Delimiter:      ",",
		HasHeader:      false,
		ExpectedFields: 4,
		Timestamp:      TimestampSpec{Column: 0, Kind: ServerStamp},
		Fields: []FieldMap{
			{Column: 1, ChannelID: "a"},
			{Column: 2, ChannelID: "b"},
			{Column: 3, ChannelID: "c"},
		},
	}
}

func TestApplyMechanics(t *testing.T) {
	ts := time.Unix(1700000000, 0)

	tests := []struct {
		name    string
		fmtFn   func() DaqFormat
		line    string
		wantOK  bool
		wantLen int
		check   map[string]float64
	}{
		{
			name:    "full steady-state line",
			fmtFn:   crafted4,
			line:    "00:00:01,1.5,2.5,3.5",
			wantOK:  true,
			wantLen: 3,
			check:   map[string]float64{"a": 1.5, "b": 2.5, "c": 3.5},
		},
		{name: "blank line", fmtFn: crafted4, line: "", wantOK: false},
		{name: "whitespace only", fmtFn: crafted4, line: "   ", wantOK: false},
		{name: "comment line", fmtFn: crafted4, line: "# stage marker", wantOK: false},
		{
			name:   "torn fragment (wrong field count) is skipped",
			fmtFn:  crafted4,
			line:   "?,,,,,,,,,,,,,00:00:00",
			wantOK: false,
		},
		{
			name:   "too-few-fields line is skipped (guard)",
			fmtFn:  crafted4,
			line:   "00:00:01,1.5",
			wantOK: false,
		},
		{
			name:    "individual unparseable field omitted, rest kept",
			fmtFn:   crafted4,
			line:    "00:00:01,1.5,abc,3.5",
			wantOK:  true,
			wantLen: 2, // b skipped
			check:   map[string]float64{"a": 1.5, "c": 3.5},
		},
		{
			name:    "warmup negative passes through unfiltered",
			fmtFn:   crafted4,
			line:    "00:00:01,-4.989,-71,3.5",
			wantOK:  true,
			wantLen: 3,
			check:   map[string]float64{"a": -4.989, "b": -71, "c": 3.5},
		},
		{
			name:    "whitespace around values tolerated",
			fmtFn:   crafted4,
			line:    "00:00:01, 1.5 , 2.5 , 3.5 ",
			wantOK:  true,
			wantLen: 3,
			check:   map[string]float64{"a": 1.5, "b": 2.5, "c": 3.5},
		},
		{
			name:   "all-garbage mapped fields -> no values",
			fmtFn:  crafted4,
			line:   "00:00:01,x,y,z",
			wantOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := New(tc.fmtFn())
			r, ok := e.Apply([]byte(tc.line), ts)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if len(r.Values) != tc.wantLen {
				t.Fatalf("len(values) = %d, want %d (%v)", len(r.Values), tc.wantLen, r.Values)
			}
			if !r.TS.Equal(ts) {
				t.Errorf("TS = %v, want server stamp %v", r.TS, ts)
			}
			for k, want := range tc.check {
				if got := r.Values[k]; got != want {
					t.Errorf("values[%q] = %v, want %v", k, got, want)
				}
			}
		})
	}
}

func TestTimestampColumnNotFieldMapped(t *testing.T) {
	// Column 0 is the timestamp position and must never appear as a channel even
	// though it holds non-numeric "HH:MM:SS" text.
	e := New(crafted4())
	r, ok := e.Apply([]byte("12:34:56,1,2,3"), time.Unix(0, 0))
	if !ok {
		t.Fatal("expected ok")
	}
	if _, present := r.Values["0"]; present {
		t.Error("timestamp column leaked into values")
	}
	if len(r.Values) != 3 {
		t.Fatalf("len(values) = %d, want 3 (%v)", len(r.Values), r.Values)
	}
}

func TestSeqAdvancesOnlyOnReadings(t *testing.T) {
	e := New(crafted4())
	ts := time.Unix(0, 0)

	if _, ok := e.Apply([]byte("# comment"), ts); ok {
		t.Fatal("comment should not parse")
	}
	r1, ok := e.Apply([]byte("00:00:01,1,2,3"), ts)
	if !ok || r1.Seq != 1 {
		t.Fatalf("first reading seq = %d (ok=%v), want 1", r1.Seq, ok)
	}
	if _, ok := e.Apply([]byte("torn,line"), ts); ok {
		t.Fatal("torn line should not parse")
	}
	if _, ok := e.Apply([]byte(""), ts); ok {
		t.Fatal("blank should not parse")
	}
	r2, _ := e.Apply([]byte("00:00:02,4,5,6"), ts)
	if r2.Seq != 2 {
		t.Fatalf("second reading seq = %d, want 2 (skipped lines must not advance seq)", r2.Seq)
	}
}

func TestHeaderSkip(t *testing.T) {
	f := crafted4()
	f.HasHeader = true
	e := New(f)
	ts := time.Unix(0, 0)

	// First non-blank line is the header -> skipped.
	if _, ok := e.Apply([]byte("ts,a,b,c"), ts); ok {
		t.Fatal("header line should be skipped")
	}
	// Next line parses normally; seq must start at 1 (header didn't advance it).
	r, ok := e.Apply([]byte("00:00:01,1,2,3"), ts)
	if !ok {
		t.Fatal("data line after header should parse")
	}
	if r.Seq != 1 {
		t.Errorf("seq = %d, want 1 (header must not advance seq)", r.Seq)
	}
}

func TestTransformScaleOffset(t *testing.T) {
	f := DaqFormat{
		ID:             "xform",
		Delimiter:      ",",
		ExpectedFields: 2,
		Timestamp:      TimestampSpec{Column: 0, Kind: ServerStamp},
		Fields: []FieldMap{
			{Column: 1, ChannelID: "scaled", Transform: &Transform{Scale: 2, Offset: 10}},
		},
	}
	e := New(f)
	r, ok := e.Apply([]byte("00:00:01,5"), time.Unix(0, 0))
	if !ok {
		t.Fatal("expected ok")
	}
	if got := r.Values["scaled"]; got != 20 { // 5*2 + 10
		t.Errorf("transform: got %v, want 20", got)
	}
}

func TestComputePass(t *testing.T) {
	// A format that does NOT emit an aggregate: agg.rate = sum(unit1.rate, unit2.rate).
	f := DaqFormat{
		ID:             "compute",
		Delimiter:      ",",
		ExpectedFields: 3,
		Timestamp:      TimestampSpec{Column: 0, Kind: ServerStamp},
		Fields: []FieldMap{
			{Column: 1, ChannelID: "unit1.rate"},
			{Column: 2, ChannelID: "unit2.rate"},
		},
		Computed: []ComputedChannel{
			{ChannelID: "agg.rate", Op: "sum", Inputs: []string{"unit1.rate", "unit2.rate"}},
			{ChannelID: "avg.rate", Op: "mean", Inputs: []string{"unit1.rate", "unit2.rate"}},
		},
	}
	e := New(f)
	r, ok := e.Apply([]byte("00:00:01,2.0,1.5"), time.Unix(0, 0))
	if !ok {
		t.Fatal("expected ok")
	}
	if got := r.Values["agg.rate"]; got != 3.5 {
		t.Errorf("agg.rate (sum) = %v, want 3.5", got)
	}
	if got := r.Values["avg.rate"]; got != 1.75 {
		t.Errorf("avg.rate (mean) = %v, want 1.75", got)
	}
}

func TestComputeSkippedWhenInputsMissing(t *testing.T) {
	f := DaqFormat{
		ID:             "compute-missing",
		Delimiter:      ",",
		ExpectedFields: 3,
		Timestamp:      TimestampSpec{Column: 0, Kind: ServerStamp},
		Fields: []FieldMap{
			{Column: 1, ChannelID: "unit1.rate"},
			{Column: 2, ChannelID: "unit2.rate"},
		},
		Computed: []ComputedChannel{
			{ChannelID: "agg.rate", Op: "sum", Inputs: []string{"unit1.rate", "unit2.rate"}},
		},
	}
	e := New(f)
	// unit2.rate is garbage -> omitted; sum computes from the one present input.
	r, ok := e.Apply([]byte("00:00:01,2.0,xyz"), time.Unix(0, 0))
	if !ok {
		t.Fatal("expected ok")
	}
	if got := r.Values["agg.rate"]; got != 2.0 {
		t.Errorf("agg.rate from one input = %v, want 2.0", got)
	}
}
