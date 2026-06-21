package printcfg

import (
	"encoding/json"
	"testing"
)

func strptr(s string) *string { return &s }
func boolptr(b bool) *bool    { return &b }

func TestCompanyDefault(t *testing.T) {
	d := CompanyDefault()
	if d.PageSize != PageLetter {
		t.Fatalf("default page size should be letter, got %q", d.PageSize)
	}
	if !d.ShowLegend {
		t.Fatal("default should show the legend")
	}
	if d.Channels != nil {
		t.Fatalf("default Channels should be nil (= all enabled), got %v", d.Channels)
	}
	if d.Title == "" {
		t.Fatal("default should have a non-empty title")
	}
}

func TestMergeEmptyOverrideIsDefault(t *testing.T) {
	d := CompanyDefault()
	eff := Merge(d, Override{})
	if eff.Title != d.Title || eff.PageSize != d.PageSize || eff.ShowLegend != d.ShowLegend {
		t.Fatalf("empty override must equal default: %+v vs %+v", eff, d)
	}
}

func TestMergeAppliesOnlySetFields(t *testing.T) {
	d := CompanyDefault()
	ch := []string{"agg.pressure", "agg.rate"}
	eff := Merge(d, Override{
		Title:    strptr("Smith 4-12H Surface"),
		PageSize: strptr(PageA4),
		Channels: &ch,
	})
	if eff.Title != "Smith 4-12H Surface" {
		t.Fatalf("title not overridden: %q", eff.Title)
	}
	if eff.PageSize != PageA4 {
		t.Fatalf("page size not overridden: %q", eff.PageSize)
	}
	// ShowLegend was NOT overridden -> keeps the default.
	if eff.ShowLegend != d.ShowLegend {
		t.Fatalf("ShowLegend should be untouched: %v", eff.ShowLegend)
	}
	if len(eff.Channels) != 2 || eff.Channels[0] != "agg.pressure" {
		t.Fatalf("channels not overridden: %v", eff.Channels)
	}
}

func TestMergeDoesNotMutateDefault(t *testing.T) {
	d := CompanyDefault()
	ch := []string{"x"}
	_ = Merge(d, Override{ShowLegend: boolptr(false), Channels: &ch})
	if !d.ShowLegend || d.Channels != nil {
		t.Fatalf("Merge mutated the default: %+v", d)
	}
}

func TestEmptyOverrideMarshalsToEmptyObject(t *testing.T) {
	b, err := json.Marshal(Override{})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(b) != "{}" {
		t.Fatalf("empty override should marshal to {}, got %s", b)
	}
}

func TestValidPageSize(t *testing.T) {
	for _, ok := range []string{PageLetter, PageA4} {
		if !ValidPageSize(ok) {
			t.Fatalf("%q should be valid", ok)
		}
	}
	for _, bad := range []string{"", "legal", "A4", "Letter"} {
		if ValidPageSize(bad) {
			t.Fatalf("%q should be invalid", bad)
		}
	}
}
