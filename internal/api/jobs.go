package api

// Job + recording HTTP handlers (Phase 3b). Thin shell over the store: every handler
// calls store methods ONLY — never a *sql.DB (axiom #4 / D2). The recording handlers
// insert/update marker rows only; they never gate ingestion or the live readout, and
// never reset stage volume (axioms #1 & #5). Typed store conditions map to clean HTTP
// status codes (404 unknown, 400 bad input/no-active-job, 409 record-state conflict).

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/bryanmaclee/cementer/internal/store"
)

// --- jobs ------------------------------------------------------------------

// jobDTO is the JSON body of POST/PUT /api/jobs (the descriptive D8 fields). The
// server owns id/isActive/timestamps; they are ignored on input.
type jobDTO struct {
	Name       string `json:"name"`
	Company    string `json:"company"`
	Well       string `json:"well"`
	CasingSize string `json:"casingSize"`
	JobType    string `json:"jobType"`
	Location   string `json:"location"`
	Cementer   string `json:"cementer"`
	Notes      string `json:"notes"`
}

func (d jobDTO) toJob() store.Job {
	return store.Job{
		Name:       d.Name,
		Company:    d.Company,
		Well:       d.Well,
		CasingSize: d.CasingSize,
		JobType:    d.JobType,
		Location:   d.Location,
		Cementer:   d.Cementer,
		Notes:      d.Notes,
	}
}

func (a *API) listJobs(w http.ResponseWriter, _ *http.Request) {
	jobs, err := a.st.ListJobs()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list jobs", err)
		return
	}
	writeJSON(w, http.StatusOK, jobs)
}

func (a *API) createJob(w http.ResponseWriter, r *http.Request) {
	var body jobDTO
	if err := decodeStrict(r, &body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}
	if body.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "job name is required")
		return
	}
	id, err := a.st.CreateJob(body.toJob())
	if err != nil {
		if errors.Is(err, store.ErrJobNameRequired) {
			writeJSONError(w, http.StatusBadRequest, "job name is required")
			return
		}
		writeError(w, http.StatusInternalServerError, "create job", err)
		return
	}
	j, _, err := a.st.GetJob(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reload job", err)
		return
	}
	writeJSON(w, http.StatusCreated, j)
}

func (a *API) getJob(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	j, found, err := a.st.GetJob(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get job", err)
		return
	}
	if !found {
		writeJSONError(w, http.StatusNotFound, "no such job")
		return
	}
	writeJSON(w, http.StatusOK, j)
}

func (a *API) updateJob(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body jobDTO
	if err := decodeStrict(r, &body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}
	if body.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "job name is required")
		return
	}
	if err := a.st.UpdateJob(id, body.toJob()); err != nil {
		if errors.Is(err, store.ErrNoSuchJob) {
			writeJSONError(w, http.StatusNotFound, "no such job")
			return
		}
		if errors.Is(err, store.ErrJobNameRequired) {
			writeJSONError(w, http.StatusBadRequest, "job name is required")
			return
		}
		writeError(w, http.StatusInternalServerError, "update job", err)
		return
	}
	j, _, err := a.st.GetJob(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reload job", err)
		return
	}
	writeJSON(w, http.StatusOK, j)
}

// getActiveJob returns the active job or {"active":null} when none is active.
func (a *API) getActiveJob(w http.ResponseWriter, _ *http.Request) {
	j, ok, err := a.st.ActiveJob()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "active job", err)
		return
	}
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{"active": nil})
		return
	}
	writeJSON(w, http.StatusOK, j)
}

// setActiveJobDTO is the body of PUT /api/job/active.
type setActiveJobDTO struct {
	ID int64 `json:"id"`
}

// setActiveJob makes a job active. It returns 409 when a DIFFERENT job is currently
// recording (the open segment must stay bound to its job — stop recording first).
func (a *API) setActiveJob(w http.ResponseWriter, r *http.Request) {
	var body setActiveJobDTO
	if err := decodeStrict(r, &body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}
	if body.ID <= 0 {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}
	if err := a.st.SetActiveJob(body.ID); err != nil {
		switch {
		case errors.Is(err, store.ErrNoSuchJob):
			writeJSONError(w, http.StatusNotFound, "no such job")
		case errors.Is(err, store.ErrRecording):
			writeJSONError(w, http.StatusConflict, "stop recording before changing the active job")
		default:
			writeError(w, http.StatusInternalServerError, "set active job", err)
		}
		return
	}
	j, _, err := a.st.ActiveJob()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reload active job", err)
		return
	}
	writeJSON(w, http.StatusOK, j)
}

// --- recording -------------------------------------------------------------

// recordingStateDTO is the GET /api/recording/state response. The id fields are
// omitted when not recording.
type recordingStateDTO struct {
	Recording     bool  `json:"recording"`
	OpenSegmentID int64 `json:"openSegmentId,omitempty"`
	JobID         int64 `json:"jobId,omitempty"`
}

func (a *API) recordingState(w http.ResponseWriter, _ *http.Request) {
	rec, openID, jobID, err := a.st.RecordingState()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "recording state", err)
		return
	}
	writeJSON(w, http.StatusOK, recordingStateDTO{Recording: rec, OpenSegmentID: openID, JobID: jobID})
}

// startRecording opens a segment under the active job (axiom #1: a marker insert
// only — ingestion and the live readout are untouched). 400 when no active job; 409
// (with the already-open segment in the body) when already recording.
func (a *API) startRecording(w http.ResponseWriter, _ *http.Request) {
	seg, err := a.st.StartRecording()
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNoActiveJob):
			writeJSONError(w, http.StatusBadRequest, "no active job — select a job before recording")
		case errors.Is(err, store.ErrRecording):
			// Already recording: 409 and return the OPEN segment so the client can sync.
			writeJSON(w, http.StatusConflict, seg)
		default:
			writeError(w, http.StatusInternalServerError, "start recording", err)
		}
		return
	}
	writeJSON(w, http.StatusCreated, seg)
}

// stopRecording closes the open segment (axiom #1: a marker update only). 409 when
// not recording.
func (a *API) stopRecording(w http.ResponseWriter, _ *http.Request) {
	seg, err := a.st.StopRecording()
	if err != nil {
		if errors.Is(err, store.ErrNotRecording) {
			writeJSONError(w, http.StatusConflict, "not recording")
			return
		}
		writeError(w, http.StatusInternalServerError, "stop recording", err)
		return
	}
	writeJSON(w, http.StatusOK, seg)
}

// listSegments returns the segments for ?job_id=N.
func (a *API) listSegments(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("job_id")
	if raw == "" {
		writeJSONError(w, http.StatusBadRequest, "job_id query parameter is required")
		return
	}
	jobID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || jobID <= 0 {
		writeJSONError(w, http.StatusBadRequest, "job_id must be a positive integer")
		return
	}
	segs, err := a.st.ListSegments(jobID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list segments", err)
		return
	}
	writeJSON(w, http.StatusOK, segs)
}

// adjustSegmentDTO carries the optional endpoint nudges (axiom #5 after-the-fact).
// Pointer fields: omit a field to leave it unchanged.
type adjustSegmentDTO struct {
	StartedAtUS *int64 `json:"startedAtUs,omitempty"`
	StoppedAtUS *int64 `json:"stoppedAtUs,omitempty"`
}

// adjustSegment nudges a segment's endpoints. 404 unknown id; 400 bad ordering.
func (a *API) adjustSegment(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body adjustSegmentDTO
	if err := decodeStrict(r, &body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}
	if err := a.st.AdjustSegment(id, body.StartedAtUS, body.StoppedAtUS); err != nil {
		switch {
		case errors.Is(err, store.ErrNoSuchSegment):
			writeJSONError(w, http.StatusNotFound, "no such segment")
		case errors.Is(err, store.ErrBadSegmentRange):
			writeJSONError(w, http.StatusBadRequest, "started_at must be <= stopped_at")
		default:
			writeError(w, http.StatusInternalServerError, "adjust segment", err)
		}
		return
	}
	// Return the refreshed segment so the client can re-render.
	seg, found, err := a.st.GetSegment(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reload segment", err)
		return
	}
	if !found {
		writeJSONError(w, http.StatusNotFound, "no such segment")
		return
	}
	writeJSON(w, http.StatusOK, seg)
}

// --- helpers ---------------------------------------------------------------

// decodeStrict decodes a JSON request body, rejecting unknown fields (the same
// discipline the profile PUT uses).
func decodeStrict(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// pathID parses the {id} path value as a positive int64, writing a 400 and returning
// ok=false on failure.
func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	raw := r.PathValue("id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "id must be a positive integer")
		return 0, false
	}
	return id, true
}
