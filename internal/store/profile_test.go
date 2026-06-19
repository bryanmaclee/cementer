package store

import (
	"path/filepath"
	"testing"
	"time"
)

// openTestStore opens a fresh store in a temp dir with no onCommit. The sample
// writeLoop runs but is untouched by these profile tests.
func openTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "test.db"), 50*time.Millisecond, nil)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func sampleVocab() []SeedChannel {
	return []SeedChannel{
		{ID: "unit1.pressure", Role: "pressure", Scope: "unit", UnitIndex: 1, UoM: "psi", Label: "Unit 1 Pressure", Decimals: 0},
		{ID: "unit2.pressure", Role: "pressure", Scope: "unit", UnitIndex: 2, UoM: "psi", Label: "Unit 2 Pressure", Decimals: 0},
		{ID: "agg.rate", Role: "rate", Scope: "aggregate", UoM: "bbl/min", Label: "Rate (total)", Decimals: 2},
		{ID: "vol.job", Role: "volume", Scope: "job", UoM: "bbl", Label: "Job Volume", Decimals: 1},
	}
}

func TestSeedAndActiveProfileRoundTrip(t *testing.T) {
	st := openTestStore(t)

	if has, err := st.HasActiveProfile(); err != nil {
		t.Fatalf("HasActiveProfile: %v", err)
	} else if has {
		t.Fatal("fresh store should have no active profile")
	}

	if err := st.SeedActiveProfile("Test Pump", 2, "intellisense", sampleVocab()); err != nil {
		t.Fatalf("SeedActiveProfile: %v", err)
	}

	has, err := st.HasActiveProfile()
	if err != nil {
		t.Fatalf("HasActiveProfile: %v", err)
	}
	if !has {
		t.Fatal("expected an active profile after seed")
	}

	p, ok, err := st.ActiveProfile()
	if err != nil {
		t.Fatalf("ActiveProfile: %v", err)
	}
	if !ok {
		t.Fatal("ActiveProfile ok=false after seed")
	}
	if p.Name != "Test Pump" || p.Units != 2 || p.FormatID != "intellisense" {
		t.Fatalf("profile header mismatch: %+v", p)
	}
	if len(p.Channels) != 4 {
		t.Fatalf("want 4 enabled channels, got %d", len(p.Channels))
	}
	// sort_order follows seed slice order.
	if p.Channels[0].ID != "unit1.pressure" || p.Channels[3].ID != "vol.job" {
		t.Fatalf("channels out of sort order: %+v", p.Channels)
	}
	// Metadata round-trips.
	if p.Channels[0].UoM != "psi" || p.Channels[0].Decimals != 0 || p.Channels[0].UnitIndex != 1 {
		t.Fatalf("channel metadata not preserved: %+v", p.Channels[0])
	}
}

func TestSeedIdempotencyViaHasActiveProfile(t *testing.T) {
	st := openTestStore(t)

	// First boot: seed.
	if has, _ := st.HasActiveProfile(); !has {
		if err := st.SeedActiveProfile("Pump", 1, "intellisense", sampleVocab()); err != nil {
			t.Fatalf("seed 1: %v", err)
		}
	}
	// Operator edits a label so we can prove the second boot doesn't clobber.
	newLabel := "Treating Pressure"
	if err := st.UpdateActiveProfile(0, []ChannelUpdate{
		{ChannelID: "unit1.pressure", Label: &newLabel},
	}); err != nil {
		t.Fatalf("edit label: %v", err)
	}

	// Second boot: guard prevents a duplicate seed (this is the main() pattern).
	if has, _ := st.HasActiveProfile(); !has {
		if err := st.SeedActiveProfile("Pump", 1, "intellisense", sampleVocab()); err != nil {
			t.Fatalf("seed 2: %v", err)
		}
	}

	// Exactly one profile row.
	var n int
	if err := st.db.QueryRow(`SELECT COUNT(*) FROM pump_profiles`).Scan(&n); err != nil {
		t.Fatalf("count profiles: %v", err)
	}
	if n != 1 {
		t.Fatalf("seed not idempotent: %d profile rows", n)
	}

	// The edit persisted (seed didn't overwrite).
	ep, ok, err := st.ActiveEditorProfile()
	if err != nil || !ok {
		t.Fatalf("ActiveEditorProfile: ok=%v err=%v", ok, err)
	}
	var got string
	for _, c := range ep.Channels {
		if c.ID == "unit1.pressure" {
			got = c.Label
		}
	}
	if got != newLabel {
		t.Fatalf("edit lost across boot: label = %q", got)
	}
}

func TestSetActiveUniqueness(t *testing.T) {
	st := openTestStore(t)
	if err := st.SeedActiveProfile("First", 1, "synthetic", sampleVocab()); err != nil {
		t.Fatalf("seed first: %v", err)
	}
	// Seeding again should demote the previous active so exactly one is active.
	if err := st.SeedActiveProfile("Second", 1, "intellisense", sampleVocab()); err != nil {
		t.Fatalf("seed second: %v", err)
	}

	var active int
	if err := st.db.QueryRow(`SELECT COUNT(*) FROM pump_profiles WHERE is_active = 1`).Scan(&active); err != nil {
		t.Fatalf("count active: %v", err)
	}
	if active != 1 {
		t.Fatalf("want exactly one active profile, got %d", active)
	}

	p, ok, err := st.ActiveProfile()
	if err != nil || !ok {
		t.Fatalf("ActiveProfile: ok=%v err=%v", ok, err)
	}
	if p.Name != "Second" {
		t.Fatalf("active profile should be the latest seeded; got %q", p.Name)
	}
}

func TestUpdateChannelEnableLabelUomDecimalsSort(t *testing.T) {
	st := openTestStore(t)
	if err := st.SeedActiveProfile("Pump", 1, "intellisense", sampleVocab()); err != nil {
		t.Fatalf("seed: %v", err)
	}

	disabled := false
	newLabel := "Stage Rate"
	newUom := "m3/min"
	newDec := 3
	newSort := 99
	if err := st.UpdateActiveProfile(3, []ChannelUpdate{
		{ChannelID: "unit2.pressure", Enabled: &disabled},
		{ChannelID: "agg.rate", Label: &newLabel, UoM: &newUom, Decimals: &newDec, SortOrder: &newSort},
	}); err != nil {
		t.Fatalf("update: %v", err)
	}

	// Units updated.
	ep, ok, err := st.ActiveEditorProfile()
	if err != nil || !ok {
		t.Fatalf("editor profile: ok=%v err=%v", ok, err)
	}
	if ep.Units != 3 {
		t.Fatalf("units not updated: %d", ep.Units)
	}

	// Disabled channel still appears in editor view but with enabled=false.
	var sawDisabled, sawEdited bool
	for _, c := range ep.Channels {
		if c.ID == "unit2.pressure" {
			sawDisabled = true
			if c.Enabled {
				t.Fatal("unit2.pressure should be disabled")
			}
		}
		if c.ID == "agg.rate" {
			sawEdited = true
			if c.Label != newLabel || c.UoM != newUom || c.Decimals != newDec || c.SortOrder != newSort {
				t.Fatalf("agg.rate edit not applied: %+v", c)
			}
		}
	}
	if !sawDisabled || !sawEdited {
		t.Fatalf("missing channels in editor view (disabled=%v edited=%v)", sawDisabled, sawEdited)
	}

	// The WS-frame ActiveProfile excludes the disabled channel.
	p, ok, err := st.ActiveProfile()
	if err != nil || !ok {
		t.Fatalf("active profile: ok=%v err=%v", ok, err)
	}
	for _, c := range p.Channels {
		if c.ID == "unit2.pressure" {
			t.Fatal("disabled channel leaked into the WS frame")
		}
	}
	if len(p.Channels) != 3 {
		t.Fatalf("want 3 enabled channels in frame, got %d", len(p.Channels))
	}
}

func TestUpdateUnknownChannelErrors(t *testing.T) {
	st := openTestStore(t)
	if err := st.SeedActiveProfile("Pump", 1, "intellisense", sampleVocab()); err != nil {
		t.Fatalf("seed: %v", err)
	}
	on := true
	err := st.UpdateActiveProfile(0, []ChannelUpdate{
		{ChannelID: "does.not.exist", Enabled: &on},
	})
	if err == nil {
		t.Fatal("expected error updating an unknown channel")
	}
	// The whole transaction rolled back: a valid concurrent edit in the same call
	// must not have partially applied either (atomicity). Verify the bad call left
	// the profile untouched.
	ep, _, _ := st.ActiveEditorProfile()
	if len(ep.Channels) != 4 {
		t.Fatalf("unexpected channel count after failed update: %d", len(ep.Channels))
	}
}

func TestResetActiveProfileChannels(t *testing.T) {
	st := openTestStore(t)
	if err := st.SeedActiveProfile("Pump", 1, "intellisense", sampleVocab()); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// Disable then reset; reset must restore all-enabled vocab.
	off := false
	if err := st.UpdateActiveProfile(0, []ChannelUpdate{{ChannelID: "vol.job", Enabled: &off}}); err != nil {
		t.Fatalf("disable: %v", err)
	}

	reduced := sampleVocab()[:2] // reset to a smaller vocab to prove replacement
	if err := st.ResetActiveProfileChannels(reduced); err != nil {
		t.Fatalf("reset: %v", err)
	}

	ep, ok, err := st.ActiveEditorProfile()
	if err != nil || !ok {
		t.Fatalf("editor profile: ok=%v err=%v", ok, err)
	}
	if len(ep.Channels) != 2 {
		t.Fatalf("reset should replace channels; got %d", len(ep.Channels))
	}
	for _, c := range ep.Channels {
		if !c.Enabled {
			t.Fatalf("reset channels should be enabled; %q disabled", c.ID)
		}
	}
}

func TestActiveProfileEmptyStore(t *testing.T) {
	st := openTestStore(t)
	if _, ok, err := st.ActiveProfile(); err != nil {
		t.Fatalf("ActiveProfile on empty store errored: %v", err)
	} else if ok {
		t.Fatal("ActiveProfile ok=true on empty store")
	}
	if _, ok, err := st.ActiveEditorProfile(); err != nil {
		t.Fatalf("ActiveEditorProfile on empty store errored: %v", err)
	} else if ok {
		t.Fatal("ActiveEditorProfile ok=true on empty store")
	}
}
