# coord — multi-operator PA coordination

The autonomous handshake substrate for cementer's two **co-equal** PA operators sharing one
GitHub repo under **PR-flow**. Ratified S6 in
[`multi-party-pa-orchestration-2026-06-21`](../../deep-dives/multi-party-pa-orchestration-2026-06-21.md).

Operators: **`bryan`** (bryanmaclee) · **`peter`** (Peter Oliver). Per-operator session ids:
**`B<n>` / `P<n>`** (no shared global counter).

## Why this exists

Two co-equal PAs both touching `main` and the single-writer meta-docs (hand-off / status /
changelog / user-voice) collide. This layer lets each PA **announce what it's doing** and
**learn what the peer did** — with no central server and **no real-time locks** (work is
mostly sequential/async, so coordination is *optimistic*: claims are advisory, verified at
push, never a mutex).

## Conflict-free-by-construction invariant

Every file here is EITHER **append-only** (`ledger.md`) OR **single-operator-owned**
(`claims/<op>.md`, `inbox/<op>/`). Two PAs never write the same bytes, so the coordination
layer itself can never merge-conflict — the one thing a coordination layer must not do.

## Files

| Path | Owner | Write mode | Purpose |
|---|---|---|---|
| `ledger.md` | shared | **append-only** (never edit prior blocks) | one block per session open + close: operator / branch / tip / arcs / push state |
| `claims/<op>.md` | that operator | **overwrite own only** | the operator's current advisory claim (arc + branch + push-intent SHA) |
| `inbox/<op>/` | written by the OTHER op (**create-only**); read + acked by `<op>` | create-only | cross-operator notices ("landed X — rebase before your next push") |

## The handshake (each PA runs this)

**Session OPEN**
1. `git fetch`; read the tail of `ledger.md` + the peer's `claims/<peer>.md` + your unread `inbox/<you>/`.
2. Surface to your operator: what the peer did since your last session, any arc the peer currently claims, any inbox notices.
3. Append an `open` block to `ledger.md`.

**CLAIM (before starting an arc)**
4. Overwrite your `claims/<you>.md`: arc, branch, push-intent SHA. Optimistic — if the peer already claims that arc, surface the overlap and pick another (rare under async work).

**LAND / PUSH**
5. Verify-before-push (fetch + rebase + the `pa.md` §7 coherence checks). Under PR-flow, work lands via a **`<op>/<arc>` branch → PR → protected `main`** (don't push straight to `main`).
6. If your merge touches the peer's in-flight arc, drop a create-only notice into `inbox/<peer>/`.

**Session CLOSE**
7. Append a `close` block to `ledger.md` (final tip, arcs, push state); reset your `claims/<you>.md` to idle.
8. Ack inbox notices you've handled: move `inbox/<you>/<msg>.md` → `inbox/<you>/read/`.

## Inbox message format (base §10 envelope)

Filename `<YYYY-MM-DD>-<slug>.md`; frontmatter `from / to / date / subject / needs / status`;
then a body. Create-only into the peer's inbox; the peer acks by moving it to `read/`.

## Status

Seeded S6 (B6). **Not yet wired into `pa.md`** (§4 session lifecycle + §10 cross-repo graph) —
that contract rewrite is the follow-up. Until then, a PA follows THIS doc directly. `pa.md`'s
"standalone single-operator" topology note is stale pending that rewrite.
