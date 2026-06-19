package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/bryanmaclee/cementer/internal/store"
)

func newTestServer(t *testing.T) (*httptest.Server, *store.Store) {
	t.Helper()
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "api.db"), 50*time.Millisecond, nil)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	vocab := func() []store.SeedChannel {
		return []store.SeedChannel{
			{ID: "unit1.pressure", Role: "pressure", Scope: "unit", UnitIndex: 1, UoM: "psi", Label: "Unit 1 Pressure", Decimals: 0},
			{ID: "unit2.pressure", Role: "pressure", Scope: "unit", UnitIndex: 2, UoM: "psi", Label: "Unit 2 Pressure", Decimals: 0},
			{ID: "agg.rate", Role: "rate", Scope: "aggregate", UoM: "bbl/min", Label: "Rate (total)", Decimals: 2},
		}
	}
	if err := st.SeedActiveProfile("Test Pump", 2, "intellisense", vocab()); err != nil {
		t.Fatalf("seed: %v", err)
	}

	mux := http.NewServeMux()
	New(st, vocab).Register(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, st
}

func TestGetProfileReturnsAllChannels(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, err := http.Get(srv.URL + "/api/profile")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	var p store.EditorProfile
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if p.Name != "Test Pump" || p.Units != 2 || p.FormatID != "intellisense" {
		t.Fatalf("header mismatch: %+v", p)
	}
	if len(p.Channels) != 3 {
		t.Fatalf("want 3 channels, got %d", len(p.Channels))
	}
	for _, c := range p.Channels {
		if !c.Enabled {
			t.Fatalf("seeded channel %q should be enabled", c.ID)
		}
	}
}

func TestPutProfileDisablesAndRelabels(t *testing.T) {
	srv, _ := newTestServer(t)

	body := `{"units":3,"channels":[
		{"id":"unit2.pressure","enabled":false},
		{"id":"agg.rate","label":"Total Rate","uom":"m3/min","decimals":3,"sortOrder":5}
	]}`
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/api/profile", bytes.NewBufferString(body))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("PUT status = %d", resp.StatusCode)
	}

	var p store.EditorProfile
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if p.Units != 3 {
		t.Fatalf("units not updated: %d", p.Units)
	}
	var sawDisabled, sawEdited bool
	for _, c := range p.Channels {
		if c.ID == "unit2.pressure" {
			sawDisabled = true
			if c.Enabled {
				t.Fatal("unit2.pressure should be disabled")
			}
		}
		if c.ID == "agg.rate" {
			sawEdited = true
			if c.Label != "Total Rate" || c.UoM != "m3/min" || c.Decimals != 3 || c.SortOrder != 5 {
				t.Fatalf("agg.rate edit not applied: %+v", c)
			}
		}
	}
	if !sawDisabled || !sawEdited {
		t.Fatalf("missing channels (disabled=%v edited=%v)", sawDisabled, sawEdited)
	}
}

func TestPutProfileUnknownChannelIs400(t *testing.T) {
	srv, _ := newTestServer(t)
	body := `{"channels":[{"id":"nope.nope","enabled":true}]}`
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/api/profile", bytes.NewBufferString(body))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400 for unknown channel, got %d", resp.StatusCode)
	}
}

func TestPutProfileBadJSONIs400(t *testing.T) {
	srv, _ := newTestServer(t)
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/api/profile", bytes.NewBufferString(`{not json`))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400 for bad JSON, got %d", resp.StatusCode)
	}
}

func TestPutProfileMissingChannelIdIs400(t *testing.T) {
	srv, _ := newTestServer(t)
	body := `{"channels":[{"enabled":true}]}`
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/api/profile", bytes.NewBufferString(body))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400 for missing id, got %d", resp.StatusCode)
	}
}

func TestResetRestoresVocab(t *testing.T) {
	srv, _ := newTestServer(t)

	// Disable a channel.
	body := `{"channels":[{"id":"unit2.pressure","enabled":false}]}`
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/api/profile", bytes.NewBufferString(body))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT: %v", err)
	}
	resp.Body.Close()

	// Reset.
	rresp, err := http.Post(srv.URL+"/api/profile/reset", "application/json", nil)
	if err != nil {
		t.Fatalf("POST reset: %v", err)
	}
	defer rresp.Body.Close()
	if rresp.StatusCode != http.StatusOK {
		t.Fatalf("reset status = %d", rresp.StatusCode)
	}
	var p store.EditorProfile
	if err := json.NewDecoder(rresp.Body).Decode(&p); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, c := range p.Channels {
		if !c.Enabled {
			t.Fatalf("after reset, %q should be enabled", c.ID)
		}
	}
}

// TestGetProfileNotFound covers the no-active-profile path (404).
func TestGetProfileNotFound(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "empty.db"), 50*time.Millisecond, nil)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()
	mux := http.NewServeMux()
	New(st, nil).Register(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/profile")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("want 404 on empty store, got %d", resp.StatusCode)
	}
}
