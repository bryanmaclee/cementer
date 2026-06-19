package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/bryanmaclee/cementer/internal/store"
)

// doJSON is a small helper: perform method+path with an optional JSON body and decode
// the response into out (when out != nil). Returns the status code.
func doJSON(t *testing.T, method, url, body string, out any) int {
	t.Helper()
	var rdr *bytes.Buffer
	if body != "" {
		rdr = bytes.NewBufferString(body)
	} else {
		rdr = bytes.NewBufferString("")
	}
	req, err := http.NewRequest(method, url, rdr)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("decode %s %s: %v", method, url, err)
		}
	}
	return resp.StatusCode
}

// createActiveJob creates a job via the API and makes it active, returning its id.
func createActiveJob(t *testing.T, base, name string) int64 {
	t.Helper()
	var j store.Job
	if code := doJSON(t, http.MethodPost, base+"/api/jobs",
		fmt.Sprintf(`{"name":%q}`, name), &j); code != http.StatusCreated {
		t.Fatalf("create job status = %d", code)
	}
	if code := doJSON(t, http.MethodPut, base+"/api/job/active",
		fmt.Sprintf(`{"id":%d}`, j.ID), nil); code != http.StatusOK {
		t.Fatalf("set active status = %d", code)
	}
	return j.ID
}

func TestCreateAndGetJob(t *testing.T) {
	srv, _ := newTestServer(t)

	body := `{"name":"Smith 4-12H","company":"Acme","casingSize":"9-5/8\"","jobType":"surface"}`
	var created store.Job
	if code := doJSON(t, http.MethodPost, srv.URL+"/api/jobs", body, &created); code != http.StatusCreated {
		t.Fatalf("POST status = %d", code)
	}
	if created.ID == 0 || created.Name != "Smith 4-12H" || created.Company != "Acme" ||
		created.CasingSize != `9-5/8"` || created.JobType != "surface" {
		t.Fatalf("created job mismatch: %+v", created)
	}
	if created.IsActive {
		t.Fatal("created job should not be active")
	}

	var got store.Job
	if code := doJSON(t, http.MethodGet, fmt.Sprintf("%s/api/jobs/%d", srv.URL, created.ID), "", &got); code != http.StatusOK {
		t.Fatalf("GET status = %d", code)
	}
	if got.ID != created.ID || got.Name != created.Name {
		t.Fatalf("get job mismatch: %+v", got)
	}
}

func TestCreateJobNameRequiredIs400(t *testing.T) {
	srv, _ := newTestServer(t)
	if code := doJSON(t, http.MethodPost, srv.URL+"/api/jobs", `{"name":""}`, nil); code != http.StatusBadRequest {
		t.Fatalf("want 400 for empty name, got %d", code)
	}
}

func TestCreateJobUnknownFieldIs400(t *testing.T) {
	srv, _ := newTestServer(t)
	if code := doJSON(t, http.MethodPost, srv.URL+"/api/jobs", `{"name":"x","bogus":1}`, nil); code != http.StatusBadRequest {
		t.Fatalf("want 400 for unknown field, got %d", code)
	}
}

func TestGetJobNotFoundIs404(t *testing.T) {
	srv, _ := newTestServer(t)
	if code := doJSON(t, http.MethodGet, srv.URL+"/api/jobs/999", "", nil); code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", code)
	}
}

func TestGetJobBadIdIs400(t *testing.T) {
	srv, _ := newTestServer(t)
	if code := doJSON(t, http.MethodGet, srv.URL+"/api/jobs/abc", "", nil); code != http.StatusBadRequest {
		t.Fatalf("want 400 for non-numeric id, got %d", code)
	}
}

func TestUpdateJob(t *testing.T) {
	srv, _ := newTestServer(t)
	var created store.Job
	doJSON(t, http.MethodPost, srv.URL+"/api/jobs", `{"name":"orig"}`, &created)

	var updated store.Job
	body := `{"name":"renamed","notes":"edited"}`
	if code := doJSON(t, http.MethodPut, fmt.Sprintf("%s/api/jobs/%d", srv.URL, created.ID), body, &updated); code != http.StatusOK {
		t.Fatalf("PUT status = %d", code)
	}
	if updated.Name != "renamed" || updated.Notes != "edited" {
		t.Fatalf("update not applied: %+v", updated)
	}
}

func TestUpdateJobMissingIs404(t *testing.T) {
	srv, _ := newTestServer(t)
	if code := doJSON(t, http.MethodPut, srv.URL+"/api/jobs/999", `{"name":"x"}`, nil); code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", code)
	}
}

func TestActiveJobNullWhenNone(t *testing.T) {
	srv, _ := newTestServer(t)
	var resp map[string]any
	if code := doJSON(t, http.MethodGet, srv.URL+"/api/job/active", "", &resp); code != http.StatusOK {
		t.Fatalf("GET status = %d", code)
	}
	if v, ok := resp["active"]; !ok || v != nil {
		t.Fatalf("want {active:null}, got %+v", resp)
	}
}

func TestSetActiveJobThenGet(t *testing.T) {
	srv, _ := newTestServer(t)
	id := createActiveJob(t, srv.URL, "Job 1")

	var active store.Job
	if code := doJSON(t, http.MethodGet, srv.URL+"/api/job/active", "", &active); code != http.StatusOK {
		t.Fatalf("GET active status = %d", code)
	}
	if active.ID != id || !active.IsActive {
		t.Fatalf("active job mismatch: %+v", active)
	}
}

func TestSetActiveJobMissingIs404(t *testing.T) {
	srv, _ := newTestServer(t)
	if code := doJSON(t, http.MethodPut, srv.URL+"/api/job/active", `{"id":999}`, nil); code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", code)
	}
}

// TestSetActiveJobWhileRecordingIs409 proves the 409 conflict: a different job cannot
// be made active while a segment is open.
func TestSetActiveJobWhileRecordingIs409(t *testing.T) {
	srv, _ := newTestServer(t)
	createActiveJob(t, srv.URL, "Job 1")

	// Create a second job (not active).
	var other store.Job
	doJSON(t, http.MethodPost, srv.URL+"/api/jobs", `{"name":"Job 2"}`, &other)

	// Start recording on Job 1.
	if code := doJSON(t, http.MethodPost, srv.URL+"/api/recording/start", "", nil); code != http.StatusCreated {
		t.Fatalf("start status = %d", code)
	}
	// Switching to Job 2 mid-recording -> 409.
	if code := doJSON(t, http.MethodPut, srv.URL+"/api/job/active", fmt.Sprintf(`{"id":%d}`, other.ID), nil); code != http.StatusConflict {
		t.Fatalf("want 409 switching jobs mid-recording, got %d", code)
	}
}

// --- recording -------------------------------------------------------------

func TestRecordingLifecycleAPI(t *testing.T) {
	srv, _ := newTestServer(t)
	jobID := createActiveJob(t, srv.URL, "Job 1")

	// Not recording.
	var state recordingStateDTO
	if code := doJSON(t, http.MethodGet, srv.URL+"/api/recording/state", "", &state); code != http.StatusOK {
		t.Fatalf("state status = %d", code)
	}
	if state.Recording {
		t.Fatal("should not be recording initially")
	}

	// Start.
	var seg store.Segment
	if code := doJSON(t, http.MethodPost, srv.URL+"/api/recording/start", "", &seg); code != http.StatusCreated {
		t.Fatalf("start status = %d", code)
	}
	if seg.JobID != jobID || seg.StoppedAtUS != nil {
		t.Fatalf("open segment mismatch: %+v", seg)
	}

	// State reflects recording.
	if code := doJSON(t, http.MethodGet, srv.URL+"/api/recording/state", "", &state); code != http.StatusOK {
		t.Fatalf("state status = %d", code)
	}
	if !state.Recording || state.OpenSegmentID != seg.ID || state.JobID != jobID {
		t.Fatalf("state mismatch: %+v", state)
	}

	// Stop.
	var closed store.Segment
	if code := doJSON(t, http.MethodPost, srv.URL+"/api/recording/stop", "", &closed); code != http.StatusOK {
		t.Fatalf("stop status = %d", code)
	}
	if closed.StoppedAtUS == nil {
		t.Fatal("stopped segment should have stoppedAtUs")
	}

	// Segments list.
	var segs []store.Segment
	if code := doJSON(t, http.MethodGet, fmt.Sprintf("%s/api/recording/segments?job_id=%d", srv.URL, jobID), "", &segs); code != http.StatusOK {
		t.Fatalf("segments status = %d", code)
	}
	if len(segs) != 1 || segs[0].ID != seg.ID {
		t.Fatalf("segments mismatch: %+v", segs)
	}
}

func TestStartRecordingNoActiveJobIs400(t *testing.T) {
	srv, _ := newTestServer(t)
	// No active job created.
	if code := doJSON(t, http.MethodPost, srv.URL+"/api/recording/start", "", nil); code != http.StatusBadRequest {
		t.Fatalf("want 400 with no active job, got %d", code)
	}
}

func TestDoubleStartIs409WithOpenSegment(t *testing.T) {
	srv, _ := newTestServer(t)
	createActiveJob(t, srv.URL, "Job 1")

	var first store.Segment
	doJSON(t, http.MethodPost, srv.URL+"/api/recording/start", "", &first)

	var again store.Segment
	if code := doJSON(t, http.MethodPost, srv.URL+"/api/recording/start", "", &again); code != http.StatusConflict {
		t.Fatalf("want 409 on double-start, got %d", code)
	}
	if again.ID != first.ID {
		t.Fatalf("409 body should be the open segment: got %+v want id %d", again, first.ID)
	}
}

func TestStopNotRecordingIs409(t *testing.T) {
	srv, _ := newTestServer(t)
	createActiveJob(t, srv.URL, "Job 1")
	if code := doJSON(t, http.MethodPost, srv.URL+"/api/recording/stop", "", nil); code != http.StatusConflict {
		t.Fatalf("want 409 stopping when not recording, got %d", code)
	}
}

func TestListSegmentsMissingJobIdIs400(t *testing.T) {
	srv, _ := newTestServer(t)
	if code := doJSON(t, http.MethodGet, srv.URL+"/api/recording/segments", "", nil); code != http.StatusBadRequest {
		t.Fatalf("want 400 without job_id, got %d", code)
	}
}

func TestAdjustSegmentMovesStartAPI(t *testing.T) {
	srv, _ := newTestServer(t)
	jobID := createActiveJob(t, srv.URL, "Job 1")

	var seg store.Segment
	doJSON(t, http.MethodPost, srv.URL+"/api/recording/start", "", &seg)
	doJSON(t, http.MethodPost, srv.URL+"/api/recording/stop", "", nil)

	earlier := seg.StartedAtUS - 5_000_000
	var adjusted store.Segment
	body := fmt.Sprintf(`{"startedAtUs":%d}`, earlier)
	if code := doJSON(t, http.MethodPut, fmt.Sprintf("%s/api/recording/segments/%d", srv.URL, seg.ID), body, &adjusted); code != http.StatusOK {
		t.Fatalf("adjust status = %d", code)
	}
	if adjusted.StartedAtUS != earlier {
		t.Fatalf("start not moved: %+v", adjusted)
	}

	// Re-GET confirms persistence.
	var segs []store.Segment
	doJSON(t, http.MethodGet, fmt.Sprintf("%s/api/recording/segments?job_id=%d", srv.URL, jobID), "", &segs)
	if len(segs) != 1 || segs[0].StartedAtUS != earlier {
		t.Fatalf("adjust not persisted: %+v", segs)
	}
}

func TestAdjustSegmentBadOrderingIs400(t *testing.T) {
	srv, _ := newTestServer(t)
	createActiveJob(t, srv.URL, "Job 1")
	var seg store.Segment
	doJSON(t, http.MethodPost, srv.URL+"/api/recording/start", "", &seg)
	doJSON(t, http.MethodPost, srv.URL+"/api/recording/stop", "", nil)

	// Move start far past stop -> 400.
	body := fmt.Sprintf(`{"startedAtUs":%d}`, seg.StartedAtUS+1_000_000_000)
	if code := doJSON(t, http.MethodPut, fmt.Sprintf("%s/api/recording/segments/%d", srv.URL, seg.ID), body, nil); code != http.StatusBadRequest {
		t.Fatalf("want 400 for bad ordering, got %d", code)
	}
}

func TestAdjustSegmentMissingIs404(t *testing.T) {
	srv, _ := newTestServer(t)
	if code := doJSON(t, http.MethodPut, srv.URL+"/api/recording/segments/999", `{"startedAtUs":1}`, nil); code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", code)
	}
}

// TestSegmentTimelineMatchesSamples is a light sanity check that segment timestamps
// are microsecond-scale (same basis as samples.ts_us) — i.e. not seconds/millis.
func TestSegmentTimelineMatchesSamples(t *testing.T) {
	srv, _ := newTestServer(t)
	createActiveJob(t, srv.URL, "Job 1")
	var seg store.Segment
	doJSON(t, http.MethodPost, srv.URL+"/api/recording/start", "", &seg)
	// Unix micros for 2026 are ~1.7e15; reject a value that looks like seconds/millis.
	if seg.StartedAtUS < 1_000_000_000_000_000 {
		t.Fatalf("startedAtUs not in unix-micros scale: %d", seg.StartedAtUS)
	}
}
