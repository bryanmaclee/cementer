package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bryanmaclee/cementer/internal/printcfg"
)

func TestGetPrintConfigDefaults(t *testing.T) {
	srv, _ := newTestServer(t)
	id := createActiveJob(t, srv.URL, "Job 1")

	var resp printConfigResponse
	url := fmt.Sprintf("%s/api/jobs/%d/print-config", srv.URL, id)
	if code := doJSON(t, http.MethodGet, url, "", &resp); code != http.StatusOK {
		t.Fatalf("GET status = %d", code)
	}
	// No override yet -> effective equals the company default.
	if resp.Effective.Title != resp.Default.Title ||
		resp.Effective.PageSize != resp.Default.PageSize ||
		resp.Effective.ShowLegend != resp.Default.ShowLegend {
		t.Fatalf("effective should equal default with no override: %+v", resp)
	}
	if resp.Default.PageSize != printcfg.PageLetter {
		t.Fatalf("default page size should be letter, got %q", resp.Default.PageSize)
	}
}

func TestPutPrintConfigOverrideRoundTrips(t *testing.T) {
	srv, _ := newTestServer(t)
	id := createActiveJob(t, srv.URL, "Job 1")
	url := fmt.Sprintf("%s/api/jobs/%d/print-config", srv.URL, id)

	// Override the title + page size + channel set; leave showLegend unset.
	body := `{"title":"Smith Surface","pageSize":"a4","channels":["agg.pressure","agg.rate"]}`
	var put printConfigResponse
	if code := doJSON(t, http.MethodPut, url, body, &put); code != http.StatusOK {
		t.Fatalf("PUT status = %d", code)
	}
	if put.Effective.Title != "Smith Surface" || put.Effective.PageSize != "a4" {
		t.Fatalf("effective not reflecting override: %+v", put.Effective)
	}
	if len(put.Effective.Channels) != 2 || put.Effective.Channels[0] != "agg.pressure" {
		t.Fatalf("effective channels wrong: %v", put.Effective.Channels)
	}
	// showLegend was NOT overridden -> matches default; the override should not carry it.
	if put.Override.ShowLegend != nil {
		t.Fatal("override should not include the unset showLegend field")
	}
	if put.Effective.ShowLegend != put.Default.ShowLegend {
		t.Fatalf("showLegend should fall back to default: %v", put.Effective.ShowLegend)
	}

	// Re-GET persists.
	var got printConfigResponse
	if code := doJSON(t, http.MethodGet, url, "", &got); code != http.StatusOK {
		t.Fatalf("re-GET status = %d", code)
	}
	if got.Effective.Title != "Smith Surface" || got.Effective.PageSize != "a4" {
		t.Fatalf("override not persisted: %+v", got.Effective)
	}
}

func TestPutPrintConfigBadPageSizeIs400(t *testing.T) {
	srv, _ := newTestServer(t)
	id := createActiveJob(t, srv.URL, "Job 1")
	url := fmt.Sprintf("%s/api/jobs/%d/print-config", srv.URL, id)
	if code := doJSON(t, http.MethodPut, url, `{"pageSize":"legal"}`, nil); code != http.StatusBadRequest {
		t.Fatalf("want 400 for bad page size, got %d", code)
	}
}

func TestPutPrintConfigUnknownFieldIs400(t *testing.T) {
	srv, _ := newTestServer(t)
	id := createActiveJob(t, srv.URL, "Job 1")
	url := fmt.Sprintf("%s/api/jobs/%d/print-config", srv.URL, id)
	if code := doJSON(t, http.MethodPut, url, `{"bogus":1}`, nil); code != http.StatusBadRequest {
		t.Fatalf("want 400 for unknown field, got %d", code)
	}
}

func TestPrintConfigMissingJobIs404(t *testing.T) {
	srv, _ := newTestServer(t)
	if code := doJSON(t, http.MethodGet, srv.URL+"/api/jobs/999/print-config", "", nil); code != http.StatusNotFound {
		t.Fatalf("want 404 GET, got %d", code)
	}
	if code := doJSON(t, http.MethodPut, srv.URL+"/api/jobs/999/print-config", `{"title":"x"}`, nil); code != http.StatusNotFound {
		t.Fatalf("want 404 PUT, got %d", code)
	}
}

// TestPutPrintConfigEmptyResetsToDefault proves an empty override ({}) clears any prior
// per-job tweak so the report falls back to the company default.
func TestPutPrintConfigEmptyResetsToDefault(t *testing.T) {
	srv, _ := newTestServer(t)
	id := createActiveJob(t, srv.URL, "Job 1")
	url := fmt.Sprintf("%s/api/jobs/%d/print-config", srv.URL, id)

	doJSON(t, http.MethodPut, url, `{"title":"Custom"}`, nil)
	var reset printConfigResponse
	if code := doJSON(t, http.MethodPut, url, `{}`, &reset); code != http.StatusOK {
		t.Fatalf("reset PUT status = %d", code)
	}
	if reset.Effective.Title != reset.Default.Title {
		t.Fatalf("empty override should reset to default title: %+v", reset.Effective)
	}
	if reset.Override.Title != nil {
		t.Fatal("override.title should be nil after reset")
	}
}
