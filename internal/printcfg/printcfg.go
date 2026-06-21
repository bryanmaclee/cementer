// Package printcfg holds the COMPANY print template (chart-config scope #2) and the
// merge logic for per-job overrides. This is the "what every job's printout looks
// like" standard (data-model.md § "Two chart-config scopes"): a change-controlled
// default bundled with the deploy (a Go literal here — NOT casually editable at
// runtime), which the cementer can tweak per job.
//
// Axiom #3: the Pi is a self-describing island — the company default ships with the
// binary, not from a central server. The PER-JOB overrides persist on the Pi with the
// job (the store owns that; this package only describes the template + merge).
//
// AXIS LAYOUT IS NOT A KNOB. The printed chart reuses the SAME automatic role/uom
// grouping the live and job charts use (one uPlot scale per role/uom); the template
// only governs which channels appear, the title block, the legend toggle, and the
// page size. That keeps axis assignment a property of the data, not a per-field
// setting (scope.md §4b).
package printcfg

// Page sizes the print view supports. Letter is the default (US oilfield context:
// psi / bbl / ppg); A4 is the override.
const (
	PageLetter = "letter"
	PageA4     = "a4"
)

// ValidPageSize reports whether s is a supported page size.
func ValidPageSize(s string) bool {
	return s == PageLetter || s == PageA4
}

// PrintConfig is the EFFECTIVE printed-chart template: the resolved set of values a
// job's report renders with (company default, merged with any per-job override). The
// JSON tags ARE the client contract (mirrored by hand in web/src/types.ts).
//
//   - Title:      the report title-block heading.
//   - PageSize:   "letter" | "a4" (drives the @media print page box on the client).
//   - ShowLegend: whether the legend prints.
//   - Channels:   the channel ids to include, in order. Empty means "all enabled
//     channels" (the client falls back to the profile's enabled, non-meta channels) —
//     so a fresh deploy prints the full chart without enumerating ids.
type PrintConfig struct {
	Title      string   `json:"title"`
	PageSize   string   `json:"pageSize"`
	ShowLegend bool     `json:"showLegend"`
	Channels   []string `json:"channels"`
}

// Override is the per-job tweak: ONLY the fields the cementer changed. A nil pointer
// means "leave the company default in place" for that field — so the stored override
// stays minimal (just the deltas) and a later change to the company default still
// flows through for any field the cementer didn't touch. The JSON tags mirror
// PrintConfig (camelCase); all fields are omitempty so an empty override marshals to
// "{}".
type Override struct {
	Title      *string   `json:"title,omitempty"`
	PageSize   *string   `json:"pageSize,omitempty"`
	ShowLegend *bool     `json:"showLegend,omitempty"`
	Channels   *[]string `json:"channels,omitempty"`
}

// CompanyDefault returns the bundled company print template (change-controlled — edit
// here and re-deploy; it is not runtime-editable). Channels is nil => "all enabled
// channels" so the default prints the full role-grouped chart on any pump profile.
func CompanyDefault() PrintConfig {
	return PrintConfig{
		Title:      "Cement Job Report",
		PageSize:   PageLetter,
		ShowLegend: true,
		Channels:   nil,
	}
}

// Merge returns the effective config: the company default with each non-nil override
// field applied. The default is taken by value, so Merge never mutates the caller's
// default. A nil Channels override leaves the default's Channels; a non-nil one
// (including an explicit empty slice) replaces it.
func Merge(def PrintConfig, ov Override) PrintConfig {
	eff := def
	if ov.Title != nil {
		eff.Title = *ov.Title
	}
	if ov.PageSize != nil {
		eff.PageSize = *ov.PageSize
	}
	if ov.ShowLegend != nil {
		eff.ShowLegend = *ov.ShowLegend
	}
	if ov.Channels != nil {
		eff.Channels = append([]string(nil), (*ov.Channels)...)
	}
	return eff
}
