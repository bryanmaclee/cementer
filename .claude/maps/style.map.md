# style.map.md
# project: cementer
# updated: 2026-06-19T23:05:55Z  commit: 1465bd9

Hand-written CSS custom properties. No Tailwind, no CSS-in-JS, no design-token build step, no external component library.

## Token Source
File: `web/src/styles.css`
Format: CSS custom properties on `:root` (dark, default) and `:root[data-theme="light"]` (light override).
Theme switch: `web/src/theme.ts` toggles `document.documentElement.dataset.theme`; preference persisted in `localStorage["cementer.theme"]` (dark default). `index.html` sets `meta color-scheme=dark`.

## Color Tokens
| Token        | Purpose                                         |
|--------------|-------------------------------------------------|
| --bg         | page background                                 |
| --surface    | top-bar + controls-host background              |
| --surface-2  | secondary surface (buttons, inputs, form fields)|
| --border     | hairline borders + select outlines              |
| --text       | primary text                                    |
| --text-dim   | secondary / muted text (labels, units, meta)    |
| --value      | large readout / legend value text               |
| --accent     | brand mark + active tab + focus + primary btn   |
| --live       | status: live (green) + ready state line         |
| --stalled    | status: connected-but-no-data (amber) + warn    |
| --offline    | status: disconnected (red) + record dot         |

## Typography Tokens
| Token  | Stack                                                     |
|--------|-----------------------------------------------------------|
| --mono | ui-monospace, SF Mono, JetBrains Mono, Fira Code, Menlo, Consolas |
| --sans | system-ui, -apple-system, Segoe UI, Roboto, Helvetica, Arial |

## Breakpoints
None defined as named tokens. The new-job form uses CSS grid `repeat(auto-fit, minmax(220px, 1fr))` for responsiveness without explicit breakpoints.

## Component Library
Source: none (no shadcn/MUI/Chakra). UI is built imperatively in vanilla TS with direct DOM creation. CSS targets class names applied in code.

## CSS Class Vocabulary
**Header + status:** `topbar`, `brand`, `brand-mark`, `brand-name`, `topbar-right`, `status`, `dot` (data-state=live|stalled|offline), `status-text`, `theme-btn`, `view-tabs`, `view-tab`, `window-select`.
**View area:** `content-chart`, `view`, `view-live`, `view-job`.
**Chart + legend:** `chart-legend`, `legend-row` (.off), `legend-swatch`, `legend-name`, `legend-value`, `legend-unit`, `chart-canvas`, `chart-empty`, `jobchart-status`.
**uPlot overrides:** `.uplot`, `.u-wrap`, `.u-legend`, `.u-legend th`, `.u-legend .u-value`.
**Controls strip:** `controls-host`, `controls`, `control-field`, `control-label`, `job-select`, `ghost-btn`, `record-btn` (.recording), `record-dot`, `control-state` (data-kind=recording|ready|warn|idle).
**New-job form:** `newjob-form`, `newjob-field`, `newjob-label`, `newjob-input`, `newjob-actions`, `primary-btn`.
**Footer:** `meta`, `meta-item`.

## uPlot chart colors  [web/src/chart/roles.ts]
12-color palette assigned per channel in profile order (wraps). Palette: #4aa3ff (blue), #3fb950 (green), #f0883e (orange), #bc8cff (purple), #f85149 (red), #e3b341 (amber), #39c5cf (cyan), #db61a2 (pink), + 4 lighter variants. Axis/grid stroke colors read from CSS vars `--text-dim` and `--border` at chart build time.

## Tags
#cementer #map #style #css-variables #dark-mode #theming #vanilla-ts #uplot #chart

## Links
- [primary.map.md](./primary.map.md)
- [state.map.md](./state.map.md)
- [structure.map.md](./structure.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
