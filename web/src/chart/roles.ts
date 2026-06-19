// Role/uom grouping + colors shared by the live and job charts. The user decision
// (Phase 4a): traces are ALL enabled profile channels, auto-grouped by ROLE — one
// uPlot scale per role/uom (pressure->psi, rate->bbl/min, density->ppg, volume->bbl),
// axes auto-assigned, a distinct color per channel. This module is the single source
// of that mapping so both charts agree. NO framework — plain functions + data.

import type { Channel } from "../types.ts";

// A scale key is the channel's uom (so all psi channels share the psi scale, etc.).
// Empty uom falls back to the role so e.g. meta-ish numeric channels still get a scale.
export function scaleKey(c: Channel): string {
  return c.uom?.trim() ? c.uom.trim() : c.role || "value";
}

// Distinct, legible line colors (works on dark + light). Assigned per channel in
// profile order; wraps if there are more channels than colors.
const PALETTE = [
  "#4aa3ff", // blue
  "#3fb950", // green
  "#f0883e", // orange
  "#bc8cff", // purple
  "#f85149", // red
  "#e3b341", // amber
  "#39c5cf", // cyan
  "#db61a2", // pink
  "#a5d6ff", // light blue
  "#7ee787", // light green
  "#ffa657", // light orange
  "#d2a8ff", // light purple
];

export function colorFor(index: number): string {
  return PALETTE[index % PALETTE.length];
}

// roleOrder gives a stable axis-side preference: pressure + rate get the prominent
// sides (psi is what the cementer watches), then density, then volume.
const ROLE_RANK: Record<string, number> = {
  pressure: 0,
  rate: 1,
  density: 2,
  volume: 3,
};

export function roleRank(role: string): number {
  return ROLE_RANK[role] ?? 50;
}

// uomLabel is the axis/legend unit label.
export function uomLabel(uom: string): string {
  return uom?.trim() ? uom.trim() : "";
}

// orderScales returns the distinct scale keys present in the channels, ordered by the
// dominant role's rank (so axes are assigned predictably: psi, bbl/min, ppg, bbl).
// Each scale key carries its uom (for the axis label) and the first role that used it.
export interface ScaleInfo {
  key: string; // the uom (scale key)
  uom: string;
  role: string;
}

export function orderScales(channels: Channel[]): ScaleInfo[] {
  const seen = new Map<string, ScaleInfo>();
  for (const c of channels) {
    const key = scaleKey(c);
    if (!seen.has(key)) {
      seen.set(key, { key, uom: c.uom ?? "", role: c.role });
    }
  }
  return [...seen.values()].sort((a, b) => roleRank(a.role) - roleRank(b.role));
}
