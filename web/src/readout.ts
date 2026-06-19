import type { Channel, Profile, Reading } from "./types.ts";
import { initTheme, toggleTheme } from "./theme.ts";

const STALE_MS = 3000;

type El = HTMLElement;

function el(tag: string, className?: string, text?: string): El {
  const e = document.createElement(tag);
  if (className) e.className = className;
  if (text !== undefined) e.textContent = text;
  return e;
}

// --- scope grouping --------------------------------------------------------
// The Pi sends a Pump Profile describing exactly the channels THIS rig has, in
// display order, with real labels/units/decimals (axiom #3). The client is a thin
// renderer of that — no id inference. Cards are grouped by scope:
//   Unit 1, Unit 2, … (by unitIndex)  →  Aggregate  →  Stage  →  Job
// `meta`-scoped channels are hidden by default. A streamed channel id absent from
// the (enabled) profile gets a minimal defensive card in a trailing "Other" group.

// scopeRank orders the non-unit scope groups after the per-unit groups.
const SCOPE_RANK: Record<string, number> = {
  aggregate: 1000,
  stage: 2000,
  job: 3000,
};
const SCOPE_TITLE: Record<string, string> = {
  aggregate: "Aggregate",
  stage: "Stage",
  job: "Job",
};
const OTHER_KEY = "__other__";
const OTHER_RANK = 9000;

// groupKey/rank/title bucket a channel into its display group. Per-unit groups sort
// by unitIndex (Unit 1 before Unit 2); the rest follow SCOPE_RANK.
function groupKey(c: Channel): string {
  if (c.scope === "unit") return `unit:${c.unitIndex || 0}`;
  return c.scope;
}
function groupRank(c: Channel): number {
  if (c.scope === "unit") return c.unitIndex || 0; // units first
  return SCOPE_RANK[c.scope] ?? OTHER_RANK;
}
function groupTitle(c: Channel): string {
  if (c.scope === "unit") return `Unit ${c.unitIndex || ""}`.trim();
  return SCOPE_TITLE[c.scope] ?? cap(c.scope);
}

function cap(s: string): string {
  return s.length ? s[0].toUpperCase() + s.slice(1) : s;
}

// --- readout ---------------------------------------------------------------

interface Card {
  card: El;
  value: El;
  decimals: number;
}

interface Group {
  section: El;
  body: El;
  rank: number;
}

// Readout renders the live value screen. The displayed channels come from the
// PumpProfile the Pi sends (enabled channels only), grouped by scope. Channels not
// in the profile never get a card; an unexpected streamed id lands in "Other".
export class Readout {
  private content: El; // holds the group sections
  private placeholder: El;
  private controls: El; // host for the job/record control strip (filled by Controls)

  private cards = new Map<string, Card>();
  private groups = new Map<string, Group>();
  private profileChannels = new Map<string, Channel>(); // id -> channel (enabled)

  private statusDot: El;
  private statusText: El;
  private seqEl: El;
  private updatedEl: El;

  private connected = false;
  private lastReadingAt = 0;

  constructor(root: El) {
    initTheme();
    root.replaceChildren();

    // --- header ---
    const header = el("header", "topbar");
    const brand = el("div", "brand");
    brand.append(el("span", "brand-mark", "●"), el("span", "brand-name", "cementer"));

    const right = el("div", "topbar-right");
    const status = el("div", "status");
    this.statusDot = el("span", "dot");
    this.statusText = el("span", "status-text", "connecting…");
    status.append(this.statusDot, this.statusText);

    const themeBtn = el("button", "theme-btn", "◐");
    themeBtn.title = "Toggle dark / light";
    themeBtn.addEventListener("click", () => toggleTheme());

    right.append(status, themeBtn);
    header.append(brand, right);

    // --- control strip host (job selector + record button live here; the Controls
    // module populates it). It sits between the header and the live values. Recording
    // is a marker over the always-on store and never gates the readout (axiom #1). ---
    this.controls = el("div", "controls-host");

    // --- grouped value area ---
    this.content = el("main", "content");
    this.placeholder = el("div", "placeholder", "waiting for data…");
    this.content.append(this.placeholder);

    // --- footer meta ---
    const footer = el("footer", "meta");
    this.seqEl = el("span", "meta-item", "seq —");
    this.updatedEl = el("span", "meta-item", "no data yet");
    footer.append(this.seqEl, this.updatedEl);

    root.append(header, this.controls, this.content, footer);

    this.applyStatus();
    window.setInterval(() => this.applyStatus(), 1000);
  }

  setConnected(connected: boolean): void {
    this.connected = connected;
    this.applyStatus();
  }

  // controlsHost is the empty container (between header and live values) where the
  // job/record controls mount. The Readout owns the layout; Controls owns the strip.
  controlsHost(): El {
    return this.controls;
  }

  // applyProfile (re)builds the display from the pump profile. It rebuilds groups
  // and cards from scratch so a reconnect with an edited profile (e.g. channels
  // disabled) renders cleanly. Existing live values for surviving channels are
  // preserved.
  applyProfile(p: Profile): void {
    const prevValues = new Map<string, string>();
    for (const [id, card] of this.cards) prevValues.set(id, card.value.textContent ?? "—");

    this.cards.clear();
    this.groups.clear();
    this.profileChannels.clear();
    this.content.replaceChildren();

    // meta scope is hidden by default; everything else gets a card.
    const visible = p.channels.filter((c) => c.scope !== "meta");
    for (const c of visible) {
      this.profileChannels.set(c.id, c);
      const card = this.ensureCard(c.id, c.label, c.uom, c.decimals, groupKey(c), groupTitle(c), groupRank(c));
      const prev = prevValues.get(c.id);
      if (prev) card.value.textContent = prev;
    }

    // Track enabled-but-meta ids too, so update() knows they are "in profile" and
    // won't fabricate an Other card for them.
    for (const c of p.channels) {
      if (!this.profileChannels.has(c.id)) this.profileChannels.set(c.id, c);
    }

    if (this.content.childElementCount === 0) {
      this.content.append(this.placeholder);
    }
    this.reorderGroups();
  }

  update(r: Reading): void {
    this.lastReadingAt = Date.now();

    for (const [id, v] of Object.entries(r.values)) {
      const ch = this.profileChannels.get(id);
      if (ch && ch.scope === "meta") continue; // meta hidden

      let card = this.cards.get(id);
      if (!card) {
        if (ch) {
          // Enabled, non-meta channel that somehow lacked a card — create it.
          card = this.ensureCard(id, ch.label, ch.uom, ch.decimals, groupKey(ch), groupTitle(ch), groupRank(ch));
        } else {
          // Defensive: a streamed id absent from the profile. Render minimally in
          // the trailing "Other" group rather than dropping it silently.
          card = this.ensureCard(id, id, "", 2, OTHER_KEY, "Other", OTHER_RANK);
        }
        this.reorderGroups();
      }
      card.value.textContent =
        v === undefined || Number.isNaN(v)
          ? "—"
          : v.toLocaleString(undefined, {
              minimumFractionDigits: card.decimals,
              maximumFractionDigits: card.decimals,
            });
    }

    this.seqEl.textContent = `seq ${r.seq}`;
    this.applyStatus();
  }

  // ensureCard creates (or returns) the card for a channel, creating its group
  // section on demand. Cards append to their group body in arrival order, which for
  // a profile is the profile's sort order.
  private ensureCard(
    id: string,
    label: string,
    uom: string,
    decimals: number,
    gKey: string,
    gTitle: string,
    gRank: number,
  ): Card {
    if (this.placeholder.isConnected) this.placeholder.remove();

    const existing = this.cards.get(id);
    if (existing) return existing;

    const group = this.ensureGroup(gKey, gTitle, gRank);

    const card = el("section", "card");
    card.append(el("div", "card-label", label));
    const valueRow = el("div", "card-valuerow");
    const value = el("span", "card-value", "—");
    valueRow.append(value);
    if (uom) valueRow.append(el("span", "card-unit", uom));
    card.append(valueRow);
    group.body.append(card);

    const c: Card = { card, value, decimals };
    this.cards.set(id, c);
    return c;
  }

  private ensureGroup(key: string, title: string, rank: number): Group {
    const existing = this.groups.get(key);
    if (existing) return existing;

    const section = el("section", "group");
    section.append(el("h2", "group-title", title));
    const body = el("div", "group-grid");
    section.append(body);

    const g: Group = { section, body, rank };
    this.groups.set(key, g);
    this.content.append(section);
    return g;
  }

  // reorderGroups sorts the group sections by rank (units ascending, then
  // aggregate/stage/job, then Other) and reattaches them in order.
  private reorderGroups(): void {
    const sorted = [...this.groups.values()].sort((a, b) => a.rank - b.rank);
    for (const g of sorted) this.content.append(g.section);
  }

  private applyStatus(): void {
    const sinceReading = Date.now() - this.lastReadingAt;
    let state: "offline" | "stalled" | "live";
    let label: string;

    if (!this.connected) {
      state = "offline";
      label = "offline";
    } else if (this.lastReadingAt === 0 || sinceReading > STALE_MS) {
      state = "stalled";
      label = this.lastReadingAt === 0 ? "connected — waiting for data" : "connected — no data";
    } else {
      state = "live";
      label = "live";
    }

    this.statusDot.dataset.state = state;
    this.statusText.textContent = label;

    if (this.lastReadingAt === 0) {
      this.updatedEl.textContent = "no data yet";
    } else {
      const secs = Math.max(0, Math.round(sinceReading / 1000));
      this.updatedEl.textContent = secs === 0 ? "updated just now" : `updated ${secs}s ago`;
    }
  }
}
