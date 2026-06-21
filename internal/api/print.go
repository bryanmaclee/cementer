package api

// Print-config HTTP handlers (Phase 4b — chart-config scope #2). The cementer's
// per-job print OVERRIDE persists with the job on the Pi (axiom #3) via the store
// (axiom #4 / D2: handlers call store methods only — never a *sql.DB). The COMPANY
// DEFAULT template is bundled (internal/printcfg.CompanyDefault) and the effective
// config = default merged with the override. The store persists only the raw override
// JSON; this layer owns the default + the merge.
//
// These routes are READ/CONFIG only over the always-on store — they never gate or
// touch ingestion, the live stream, or recording (axiom #1).

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bryanmaclee/cementer/internal/printcfg"
	"github.com/bryanmaclee/cementer/internal/store"
)

// printConfigResponse is the GET/PUT body: the effective (rendered) config, the raw
// per-job override (just the cementer's deltas), and the company default — so the
// client can show the editor pre-filled and reset to default. The JSON tags ARE the
// client contract (mirrored by hand in web/src/types.ts).
type printConfigResponse struct {
	Effective printcfg.PrintConfig `json:"effective"`
	Override  printcfg.Override    `json:"override"`
	Default   printcfg.PrintConfig `json:"default"`
}

// buildPrintConfigResponse loads a job's stored override, parses it, and assembles the
// effective/override/default triple. found is false when the job does not exist.
func (a *API) buildPrintConfigResponse(id int64) (printConfigResponse, bool, error) {
	raw, found, err := a.st.JobPrintConfig(id)
	if err != nil || !found {
		return printConfigResponse{}, found, err
	}
	var ov printcfg.Override
	if raw != "" {
		// A stored blob is already canonical (we marshal it on PUT), so a parse error
		// here means hand-corruption; surface it as a server error.
		if err := json.Unmarshal([]byte(raw), &ov); err != nil {
			return printConfigResponse{}, true, err
		}
	}
	def := printcfg.CompanyDefault()
	return printConfigResponse{
		Effective: printcfg.Merge(def, ov),
		Override:  ov,
		Default:   def,
	}, true, nil
}

// getPrintConfig serves GET /api/jobs/{id}/print-config -> { effective, override,
// default }. 404 when the job does not exist.
func (a *API) getPrintConfig(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	resp, found, err := a.buildPrintConfigResponse(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "print config", err)
		return
	}
	if !found {
		writeJSONError(w, http.StatusNotFound, "no such job")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// putPrintConfig serves PUT /api/jobs/{id}/print-config. The body is a printcfg.Override
// (only the changed fields; DisallowUnknownFields). It validates pageSize (when set),
// canonicalizes the override to JSON, stores it on the job, and returns the refreshed
// { effective, override, default }. 404 unknown job; 400 bad body / bad page size.
func (a *API) putPrintConfig(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var ov printcfg.Override
	if err := decodeStrict(r, &ov); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}
	if ov.PageSize != nil && !printcfg.ValidPageSize(*ov.PageSize) {
		writeJSONError(w, http.StatusBadRequest, "pageSize must be \"letter\" or \"a4\"")
		return
	}

	// Canonicalize: store only what the cementer set (omitempty drops nil fields), so
	// the persisted blob stays minimal and a later company-default change still flows
	// through for fields the cementer never touched.
	raw, err := json.Marshal(ov)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "encode print config", err)
		return
	}
	if err := a.st.SetJobPrintConfig(id, string(raw)); err != nil {
		if errors.Is(err, store.ErrNoSuchJob) {
			writeJSONError(w, http.StatusNotFound, "no such job")
			return
		}
		writeError(w, http.StatusInternalServerError, "save print config", err)
		return
	}

	resp, found, err := a.buildPrintConfigResponse(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reload print config", err)
		return
	}
	if !found {
		writeJSONError(w, http.StatusNotFound, "no such job")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
