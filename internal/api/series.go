package api

// Series HTTP handlers (Phase 4a — the chart's historical read). Thin shell over the
// store: each handler calls a store read method ONLY — never a *sql.DB (axiom #4 / D2).
// These are READS over the always-on samples store; they never gate or touch ingestion,
// the live stream, or recording (axiom #1: the chart is read-only). The JSON shapes are
// the client contract (mirrored by hand in web/src/types.ts — no codegen).

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/bryanmaclee/cementer/internal/store"
)

// defaultSeriesCap is the per-channel point cap when ?max= is omitted. ~4k points/
// channel is plenty for a chart yet keeps the payload and draw cheap.
const defaultSeriesCap = 4000

// samplesResponse is the GET /api/samples body: per-channel [ts_us, value] arrays.
type samplesResponse struct {
	Series map[string][]store.SeriesPoint `json:"series"`
}

// jobSeriesResponse is the GET /api/jobs/{id}/series body: the job's segments plus the
// in-segment per-channel series.
type jobSeriesResponse struct {
	Segments []store.Segment                `json:"segments"`
	Series   map[string][]store.SeriesPoint `json:"series"`
}

// getSamples serves GET /api/samples?from=<us>&to=<us>&channels=a,b,c[&max=N]. from/to
// are unix-microseconds; channels is a comma list (omit/empty => all channels in range).
// It validates from<=to and a sane cap, then returns the decimated series.
func (a *API) getSamples(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	from, ok := parseTimeParam(w, q.Get("from"), "from")
	if !ok {
		return
	}
	to, ok := parseTimeParam(w, q.Get("to"), "to")
	if !ok {
		return
	}
	if from > to {
		writeJSONError(w, http.StatusBadRequest, "from must be <= to")
		return
	}
	max, ok := parseMaxParam(w, q.Get("max"))
	if !ok {
		return
	}
	channels := parseChannels(q.Get("channels"))

	series, err := a.st.Series(from, to, channels, max)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "samples series", err)
		return
	}
	writeJSON(w, http.StatusOK, samplesResponse{Series: series})
}

// getJobSeries serves GET /api/jobs/{id}/series?channels=[&max=N]. It returns the job's
// segments and the in-segment series (recorded data only; gaps between segments stay
// gaps). 404 when the job does not exist.
func (a *API) getJobSeries(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	max, ok := parseMaxParam(w, q.Get("max"))
	if !ok {
		return
	}
	channels := parseChannels(q.Get("channels"))

	segs, series, found, err := a.st.JobSeries(id, channels, max)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "job series", err)
		return
	}
	if !found {
		writeJSONError(w, http.StatusNotFound, "no such job")
		return
	}
	writeJSON(w, http.StatusOK, jobSeriesResponse{Segments: segs, Series: series})
}

// --- param helpers ---------------------------------------------------------

// parseTimeParam parses a required unix-micros query param, writing a 400 and
// returning ok=false on a missing/invalid value.
func parseTimeParam(w http.ResponseWriter, raw, name string) (int64, bool) {
	if raw == "" {
		writeJSONError(w, http.StatusBadRequest, name+" is required (unix microseconds)")
		return 0, false
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, name+" must be an integer (unix microseconds)")
		return 0, false
	}
	return v, true
}

// parseMaxParam parses the optional ?max= per-channel cap. Empty => default. A negative
// value is rejected; the store clamps the upper bound.
func parseMaxParam(w http.ResponseWriter, raw string) (int, bool) {
	if raw == "" {
		return defaultSeriesCap, true
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		writeJSONError(w, http.StatusBadRequest, "max must be a non-negative integer")
		return 0, false
	}
	if v == 0 {
		return defaultSeriesCap, true
	}
	return v, true
}

// parseChannels splits a comma-separated channels param into a trimmed, non-empty list.
// An empty or whitespace-only param yields nil (=> all channels in range).
func parseChannels(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
