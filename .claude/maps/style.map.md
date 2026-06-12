# style.map.md
# project: cementer
# updated: 2026-06-12T09:02:13-06:00  commit: ee446c3

Hand-written CSS custom properties. No Tailwind, no CSS-in-JS, no design-token build step, no component library.

## Token Source
File: web/src/styles.css
Format: CSS custom properties on :root (dark, default) and :root[data-theme="light"] (light override).
Theme switch: web/src/theme.ts toggles document.documentElement.dataset.theme; preference persisted in localStorage key "cementer.theme" (dark default). index.html sets meta color-scheme=dark.

## Color Tokens (dark default → light override)
--bg          page background
--surface     card / top-bar background
--surface-2   secondary surface (theme button, etc.)
--border      hairline borders
--text        primary text
--text-dim    secondary / muted text (labels, units, meta)
--value       large readout value text
--accent      links / brand mark / focus (blue)
--live        status: live (green)
--stalled     status: connected-but-no-data (amber)
--offline     status: disconnected (red)

## Typography Tokens
--mono  ui-monospace stack (SF Mono / JetBrains Mono / Fira Code / Menlo / Consolas) — used for readout values
--sans  system-ui stack — body default

## Breakpoints
None defined as named tokens (grep found no media-query token system). The value grid uses CSS grid auto-fit (responsive without explicit breakpoints).

## Component Library
Source: none (no shadcn/MUI/Chakra). UI is built imperatively by web/src/readout.ts creating raw DOM elements with class names that styles.css targets.
Class vocabulary: topbar, brand, brand-mark, brand-name, status, dot (data-state=live|stalled|offline), status-text, theme-btn, grid, placeholder, card, card-label, card-valuerow, card-value, card-unit, meta, meta-item.

## Channel display inference (not styling, but display config)  [web/src/readout.ts]
ROLE_INFO maps role → { uom, decimals, order }: pressure(psi,0), rate(bbl/min,2), density(ppg,2), volume(bbl,1), temperature(°F,0). Channel id parts are titleized for labels (e.g. "unit1.pressure" → "Unit 1 Pressure"). This is a stopgap until the pump profile (real labels/units) arrives over the wire.

## Tags
#cementer #map #style #css-variables #dark-mode #theming #vanilla-ts

## Links
- [primary.map.md](./primary.map.md)
- [state.map.md](./state.map.md)
- [structure.map.md](./structure.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
