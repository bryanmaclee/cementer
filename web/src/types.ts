// Mirrors the Go contracts in internal/model and the wsEnvelope in cmd/cementer.

export interface Reading {
  seq: number;
  ts: string; // RFC3339 timestamp
  values: Record<string, number>;
}

export interface WSEnvelope {
  type: string;
  reading?: Reading;
}
