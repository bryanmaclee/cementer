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

// Job mirrors store.Job — the REST shape of /api/jobs* (NOT a WS frame). id,
// isActive and the timestamps are server-owned; a create/update body sends only the
// descriptive fields.
export interface Job {
  id: number;
  name: string;
  company: string;
  well: string;
  casingSize: string;
  jobType: string;
  location: string;
  cementer: string;
  notes: string;
  isActive: boolean;
  createdAtUs: number;
  updatedAtUs: number;
}

// JobInput is the body of POST/PUT /api/jobs (descriptive fields only). Only name is
// required; the rest default to "".
export interface JobInput {
  name: string;
  company?: string;
  well?: string;
  casingSize?: string;
  jobType?: string;
  location?: string;
  cementer?: string;
  notes?: string;
}

// Segment mirrors store.Segment — one recording marker over the always-on samples
// store. stoppedAtUs is null while the segment is open (recording in progress).
// Times are unix microseconds, the same timeline as the sample stream.
export interface Segment {
  id: number;
  jobId: number;
  startedAtUs: number;
  stoppedAtUs: number | null;
  createdAtUs: number;
}

// RecordingState mirrors the GET /api/recording/state response. openSegmentId/jobId
// are present only while recording.
export interface RecordingState {
  recording: boolean;
  openSegmentId?: number;
  jobId?: number;
}
