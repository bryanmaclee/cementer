// Mirrors the Go contracts in internal/model, internal/store (Profile/Channel), and
// the wsEnvelope in cmd/cementer. Kept in sync BY HAND (no codegen) — see
// anti-patterns Part B.

export interface Reading {
  seq: number;
  ts: string; // RFC3339 timestamp
  values: Record<string, number>;
}

// Channel mirrors store.Channel (the enabled-only display metadata sent in the
// hello/profile frame). No `enabled` field: the frame already carries enabled
// channels only.
export interface Channel {
  id: string;
  role: string;
  scope: string; // unit | aggregate | stage | job | meta
  unitIndex: number; // 1-based when scope === "unit"; 0 otherwise
  label: string;
  uom: string;
  decimals: number;
}

// Profile mirrors store.Profile: the active pump profile the Pi sends on connect.
// channels are the ENABLED channels only, already in display (sort) order.
export interface Profile {
  name: string;
  units: number;
  formatId: string;
  channels: Channel[];
}

export interface WSEnvelope {
  type: string; // "reading" | "profile"
  reading?: Reading;
  profile?: Profile;
}
