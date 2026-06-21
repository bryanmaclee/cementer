# docs/pa — PA scaffolding layout (multi-operator)

cementer is run by **two co-equal PA operators** (Bryan + Peter) since Session 6. This maps which
PA docs are **shared** vs **per-operator**, and where live coordination lives. Background +
rationale: [`../deep-dives/multi-party-pa-orchestration-2026-06-21.md`](../deep-dives/multi-party-pa-orchestration-2026-06-21.md).

## Shared — on `main` (PR-gated); one canonical truth both operators maintain

| Doc | Role |
|---|---|
| `status.md` | live SoT (done / in-flight / left). Per-operator **in-flight** is a *section* within it (each operator edits only their block); the phase board is shared. |
| `changelog.md` | cross-session audit trail (both append). |
| `pa.md` / `pa-base.md` | the PA contract. |
| `design-insights.md` | scoped design insights. |
| `anti-patterns.md`, `briefs/` | shared references / archived dispatch briefs. |

## Per-operator — each operator owns + rewrites ONLY their own (CODEOWNERS enforces)

| Bryan | Peter |
|---|---|
| `hand-off.md` (Bryan's live baton) | `hand-off-peter.md` |
| `user-voice.md` (Bryan's directive ledger) | `user-voice-peter.md` |

> **NOTE (S6, deferred):** Bryan's per-operator docs keep their original un-suffixed names for
> now. The symmetric rename (`hand-off-bryan.md` / `user-voice-bryan.md`) **plus** the `pa.md`
> path-fills + topology/§10 threading are part of the **pending `pa.md` topology rewrite** —
> they're coupled, so they land together once that rewrite is blessed. Until then `pa.md` validly
> points at the un-suffixed files (= Bryan's).

## Live coordination — the `coord` BRANCH (not `main`; push directly, no PR)

`make coord` → checks the unprotected orphan `coord` branch out at `.coord/`:

- `ledger.md` — append-only session log (one block per open + close); ids are operator-prefixed **`B<n>` / `P<n>`**.
- `claims/<op>.md` — advisory, optimistic claim (not a lock).
- `inbox/<op>/` — create-only cross-operator notices; ack → `read/`.

See `.coord/README.md` for the open→claim→land→close handshake. Coordination is kept OFF `main`
so it stays low-latency (no PR gate).
