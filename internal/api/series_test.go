package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/bryanmaclee/cementer/internal/model"
	"github.com/bryanmaclee/cementer/internal/store"
)

// newSeriesServer opens a store with a fast batch interval, submits the given readings,
// waits for them to commit, then mounts the API. It returns the server and store. The
// chart routes are reads only (axiom #1) so no profile seed is needed here.
func newSeriesServer(t *testing.T, readings []model.Reading) (*httptest.Server, *store.Store) {
	t.Helper()
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "series.db"), 20*time.Millisecond, nil)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	want := 0
	for _, r := range readings {
		st.Submit(r)
		want += len(r.Values)
	}
	// Wait for the async writeLoop to commit everything.
	waitRows(t, st, int64(want))

	mux := http.NewServeMux()
	New(st, nil).Register(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, st
}

func waitRows(t *testing.T, st *store.Store, want int64) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		s, err := st.Stats()
		if err != nil {
			t.Fatalf("stats: %v", err)
		}
		if s.Rows >= want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d rows", want)
}

func reading(tsUS int64, vals map[string]float64) model.Reading {
	return model.Reading{TS: time.UnixMicro(tsUS), Values: vals}
}

func TestGetSamplesReturnsSeries(t *testing.T) {
	readings := []model.Reading{
		reading(1_000, map[string]float64{"p": 10, "r": 1}),
		reading(2_000, map[string]float64{"p": 20, "r": 2}),
		reading(3_000, map[string]float64{"p": 30, "r": 3}),
	}
	srv, _ := newSeriesServer(t, readings)

	resp, err := http.Get(srv.URL + "/api/samples?from=0&to=10000&channels=p,r")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body struct {
		Series map[string][][2]float64 `json:"series"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Series["p"]) != 3 || len(body.Series["r"]) != 3 {
		t.Fatalf("expected 3 points each: %+v", body.Series)
	}
	// First p point should be [1000,10].
	if body.Series["p"][0][0] != 1000 || body.Series["p"][0][1] != 10 {
		t.Fatalf("p[0] mismatch: %v", body.Series["p"][0])
	}
}

func TestGetSamplesEmptyChannelsAllInRange(t *testing.T) {
	srv, _ := newSeriesServer(t, []model.Reading{
		reading(1_000, map[string]float64{"a": 1, "b": 2}),
	})
	resp, err := http.Get(srv.URL + "/api/samples?from=0&to=10000")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	var body struct {
		Series map[string][][2]float64 `json:"series"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Series) != 2 {
		t.Fatalf("want 2 channels, got %d", len(body.Series))
	}
}

func TestGetSamplesValidation(t *testing.T) {
	srv, _ := newSeriesServer(t, nil)
	cases := []struct {
		url  string
		want int
	}{
		{"/api/samples", http.StatusBadRequest},                      // missing from/to
		{"/api/samples?from=0", http.StatusBadRequest},               // missing to
		{"/api/samples?from=abc&to=10", http.StatusBadRequest},       // bad from
		{"/api/samples?from=100&to=50", http.StatusBadRequest},       // from > to
		{"/api/samples?from=0&to=10&max=-5", http.StatusBadRequest},  // bad max
		{"/api/samples?from=0&to=10&max=xyz", http.StatusBadRequest}, // bad max
		{"/api/samples?from=0&to=10", http.StatusOK},                 // valid, empty
	}
	for _, c := range cases {
		resp, err := http.Get(srv.URL + c.url)
		if err != nil {
			t.Fatalf("GET %s: %v", c.url, err)
		}
		resp.Body.Close()
		if resp.StatusCode != c.want {
			t.Fatalf("%s: status %d, want %d", c.url, resp.StatusCode, c.want)
		}
	}
}

func TestGetJobSeries404(t *testing.T) {
	srv, _ := newSeriesServer(t, nil)
	resp, err := http.Get(srv.URL + "/api/jobs/999/series")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestGetJobSeriesInSegment(t *testing.T) {
	// Samples across three eras; a segment over only the middle era.
	var readings []model.Reading
	for ts := int64(100); ts <= 110; ts++ {
		readings = append(readings, reading(ts, map[string]float64{"p": float64(ts)}))
	}
	for ts := int64(200); ts <= 210; ts++ {
		readings = append(readings, reading(ts, map[string]float64{"p": float64(ts)}))
	}
	srv, st := newSeriesServer(t, readings)

	// Create + activate a job and a segment over [200,210].
	jobID, err := st.CreateJob(store.Job{Name: "J"})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if err := st.SetActiveJob(jobID); err != nil {
		t.Fatalf("SetActiveJob: %v", err)
	}
	// Use the adjust path: start then adjust endpoints onto [200,210].
	seg, err := st.StartRecording()
	if err != nil {
		t.Fatalf("StartRecording: %v", err)
	}
	start := int64(200)
	stop := int64(210)
	if err := st.AdjustSegment(seg.ID, &start, &stop); err != nil {
		t.Fatalf("AdjustSegment: %v", err)
	}

	resp, err := http.Get(srv.URL + "/api/jobs/" + strconv.FormatInt(jobID, 10) + "/series?channels=p")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body struct {
		Segments []store.Segment         `json:"segments"`
		Series   map[string][][2]float64 `json:"series"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Segments) != 1 {
		t.Fatalf("want 1 segment, got %d", len(body.Segments))
	}
	pts := body.Series["p"]
	if len(pts) != 11 {
		t.Fatalf("want 11 in-segment points, got %d: %v", len(pts), pts)
	}
	for _, p := range pts {
		if p[0] < 200 || p[0] > 210 {
			t.Fatalf("point outside segment leaked: %v", p)
		}
	}
}
