// Package api is the HTTP/JSON layer for Pump-Profile (and, later, Job +
// recording) CRUD. It is a thin shell over the store: handlers call store methods
// ONLY — they never open or touch a *sql.DB (axiom #4 / D2: the store is the single
// DB owner). The API is mounted on the existing net/http ServeMux in
// cmd/cementer/main.go using Go-1.22 method-pattern routes.
package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bryanmaclee/cementer/internal/store"
)

// API holds the dependencies the HTTP handlers need: the store (sole DB owner) and
// a resetVocab provider that supplies the active format's channel vocabulary for the
// reset escape hatch. main injects resetVocab so the api package stays
// format-agnostic (it does not import internal/daqformat).
type API struct {
	st         *store.Store
	resetVocab func() []store.SeedChannel
}

// New builds an API. resetVocab may be nil (then POST /api/profile/reset returns a
// 500 rather than reseeding); main always supplies it.
func New(st *store.Store, resetVocab func() []store.SeedChannel) *API {
	return &API{st: st, resetVocab: resetVocab}
}

// Register mounts the API routes on mux. Profile routes (Phase 3a) plus Job +
// recording routes (Phase 3b). All handlers call store methods only (axiom #4 / D2).
func (a *API) Register(mux *http.ServeMux) {
	// Pump Profile (3a).
	mux.HandleFunc("GET /api/profile", a.getProfile)
	mux.HandleFunc("PUT /api/profile", a.putProfile)
	mux.HandleFunc("POST /api/profile/reset", a.resetProfile)

	// Jobs (3b).
	mux.HandleFunc("GET /api/jobs", a.listJobs)
	mux.HandleFunc("POST /api/jobs", a.createJob)
	mux.HandleFunc("GET /api/jobs/{id}", a.getJob)
	mux.HandleFunc("PUT /api/jobs/{id}", a.updateJob)
	mux.HandleFunc("GET /api/job/active", a.getActiveJob)
	mux.HandleFunc("PUT /api/job/active", a.setActiveJob)

	// Recording segments — markers over the always-on store (axiom #1). These
	// routes ONLY insert/update marker rows; they never gate ingestion or the live
	// readout, and never reset stage volume (axiom #5).
	mux.HandleFunc("GET /api/recording/state", a.recordingState)
	mux.HandleFunc("POST /api/recording/start", a.startRecording)
	mux.HandleFunc("POST /api/recording/stop", a.stopRecording)
	mux.HandleFunc("GET /api/recording/segments", a.listSegments)
	mux.HandleFunc("PUT /api/recording/segments/{id}", a.adjustSegment)
}

// getProfile returns the active profile with ALL channels (enabled and disabled),
// so the editor can see and toggle disabled channels. The hello/profile WS frame, by
// contrast, sends enabled channels only.
func (a *API) getProfile(w http.ResponseWriter, _ *http.Request) {
	p, ok, err := a.st.ActiveEditorProfile()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "read profile", err)
		return
	}
	if !ok {
		writeJSONError(w, http.StatusNotFound, "no active profile")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// channelEditDTO is the JSON shape of one per-channel edit in a PUT body. Pointer
// fields are optional: an omitted field leaves that column unchanged. id is required.
type channelEditDTO struct {
	ID        string  `json:"id"`
	Enabled   *bool   `json:"enabled,omitempty"`
	Label     *string `json:"label,omitempty"`
	UoM       *string `json:"uom,omitempty"`
	Decimals  *int    `json:"decimals,omitempty"`
	SortOrder *int    `json:"sortOrder,omitempty"`
}

// putProfileDTO is the JSON body of PUT /api/profile. units is optional (omit or
// <= 0 to leave it unchanged); channels is the list of per-channel edits.
type putProfileDTO struct {
	Units    *int             `json:"units,omitempty"`
	Channels []channelEditDTO `json:"channels"`
}

// putProfile updates the active profile's units and per-channel
// enabled/label/uom/decimals/sortOrder, then returns the refreshed editor profile.
func (a *API) putProfile(w http.ResponseWriter, r *http.Request) {
	var body putProfileDTO
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}

	edits := make([]store.ChannelUpdate, 0, len(body.Channels))
	for _, c := range body.Channels {
		if c.ID == "" {
			writeJSONError(w, http.StatusBadRequest, "channel edit missing id")
			return
		}
		if c.Decimals != nil && (*c.Decimals < 0 || *c.Decimals > 10) {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("channel %q: decimals out of range (0..10)", c.ID))
			return
		}
		edits = append(edits, store.ChannelUpdate{
			ChannelID: c.ID,
			Enabled:   c.Enabled,
			Label:     c.Label,
			UoM:       c.UoM,
			Decimals:  c.Decimals,
			SortOrder: c.SortOrder,
		})
	}

	units := 0
	if body.Units != nil {
		if *body.Units < 1 {
			writeJSONError(w, http.StatusBadRequest, "units must be >= 1")
			return
		}
		units = *body.Units
	}

	if err := a.st.UpdateActiveProfile(units, edits); err != nil {
		// An unknown channel id is a client error; everything else is a server error.
		writeError(w, http.StatusBadRequest, "update profile", err)
		return
	}

	p, ok, err := a.st.ActiveEditorProfile()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reload profile", err)
		return
	}
	if !ok {
		writeJSONError(w, http.StatusNotFound, "no active profile")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// resetProfile reseeds the active profile's channels from the active format's vocab
// (operator escape hatch), then returns the refreshed editor profile.
func (a *API) resetProfile(w http.ResponseWriter, _ *http.Request) {
	if a.resetVocab == nil {
		writeJSONError(w, http.StatusInternalServerError, "reset vocabulary not configured")
		return
	}
	if err := a.st.ResetActiveProfileChannels(a.resetVocab()); err != nil {
		writeError(w, http.StatusInternalServerError, "reset profile", err)
		return
	}
	p, ok, err := a.st.ActiveEditorProfile()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reload profile", err)
		return
	}
	if !ok {
		writeJSONError(w, http.StatusNotFound, "no active profile")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// --- helpers ---------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("api: encode response: %v", err)
	}
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// writeError logs the wrapped server-side detail and returns a client-facing JSON
// error. The context string explains what failed.
func writeError(w http.ResponseWriter, status int, context string, err error) {
	log.Printf("api: %s: %v", context, err)
	writeJSONError(w, status, fmt.Sprintf("%s: %s", context, err.Error()))
}
