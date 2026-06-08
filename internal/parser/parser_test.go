package parser

import (
	"testing"
	"time"
)

func TestParsePermissive(t *testing.T) {
	p := New(DefaultConfig())
	ts := time.Unix(0, 0)

	tests := []struct {
		name    string
		line    string
		wantOK  bool
		wantLen int
		check   map[string]float64
	}{
		{
			name:    "full line",
			line:    "5377,5.08,13.77,10.5",
			wantOK:  true,
			wantLen: 4,
			check:   map[string]float64{"pressure": 5377, "rate": 5.08, "density": 13.77, "volume": 10.5},
		},
		{name: "blank line", line: "", wantOK: false},
		{name: "whitespace only", line: "   ", wantOK: false},
		{name: "comment", line: "# stage marker", wantOK: false},
		{
			name:    "short line (fewer fields)",
			line:    "5377,5.08",
			wantOK:  true,
			wantLen: 2,
			check:   map[string]float64{"pressure": 5377, "rate": 5.08},
		},
		{
			name:    "garbage field tolerated",
			line:    "5377,abc,13.77,10.5",
			wantOK:  true,
			wantLen: 3, // rate skipped, others kept
			check:   map[string]float64{"pressure": 5377, "density": 13.77, "volume": 10.5},
		},
		{
			name:    "extra trailing fields ignored",
			line:    "5377,5.08,13.77,10.5,999,888",
			wantOK:  true,
			wantLen: 4,
		},
		{name: "all garbage", line: "x,y,z", wantOK: false},
		{
			name:    "whitespace around values",
			line:    " 5377 , 5.08 , 13.77 , 10.5 ",
			wantOK:  true,
			wantLen: 4,
			check:   map[string]float64{"pressure": 5377},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, ok := p.Parse([]byte(tc.line), ts)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if len(r.Values) != tc.wantLen {
				t.Fatalf("len(values) = %d, want %d (%v)", len(r.Values), tc.wantLen, r.Values)
			}
			for k, want := range tc.check {
				if got := r.Values[k]; got != want {
					t.Errorf("values[%q] = %v, want %v", k, got, want)
				}
			}
		})
	}
}

func TestParseSeqIncrementsOnlyOnReadings(t *testing.T) {
	p := New(DefaultConfig())
	ts := time.Unix(0, 0)

	if _, ok := p.Parse([]byte("# comment"), ts); ok {
		t.Fatal("comment should not parse")
	}
	r1, ok := p.Parse([]byte("100,1,1,1"), ts)
	if !ok || r1.Seq != 1 {
		t.Fatalf("first reading seq = %d (ok=%v), want 1", r1.Seq, ok)
	}
	if _, ok := p.Parse([]byte(""), ts); ok {
		t.Fatal("blank should not parse")
	}
	r2, _ := p.Parse([]byte("200,2,2,2"), ts)
	if r2.Seq != 2 {
		t.Fatalf("second reading seq = %d, want 2 (skipped lines must not advance seq)", r2.Seq)
	}
}
