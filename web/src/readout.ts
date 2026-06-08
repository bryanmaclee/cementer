import type { Reading } from "./types.ts";
import { initTheme, toggleTheme } from "./theme.ts";

const STALE_MS = 3000;

type El = HTMLElement;

function el(tag: string, className?: string, text?: string): El {
  const e = document.createElement(tag);
  if (className) e.className = className;
  if (text !== undefined) e.textContent = text;
  return e;
}

// --- channel description ---------------------------------------------------
// Until the pump profile (with real labels/units) arrives over the wire, infer a
// reasonable label, unit, and precision from the channel id. Channel ids are flat
// today (e.g. "pressure") and will become scoped later (e.g. "unit1.pressure").

interface ChannelSpec {
  label: string;
  uom: string;
  decimals: number;
  order: number; // role priority for display ordering
}

const ROLE_INFO: Record<string, { uom: string; decimals: number; order: number }> = {
  pressure: { uom: "psi", decimals: 0, order: 0 },
  rate: { uom: "bbl/min", decimals: 2, order: 1 },
  density: { uom: "ppg", decimals: 2, order: 2 },
  volume: { uom: "bbl", decimals: 1, order: 3 },
  temperature: { uom: "°F", decimals: 0, order: 4 },
};

const PART_LABEL: Record<string, string> = {
  agg: "Aggregate",
  vol: "Volume",
  stage: "Stage",
  job: "Job",
};

function titleize(part: string): string {
  if (PART_LABEL[part]) return PART_LABEL[part];
  // "unit1" -> "Unit 1"
  const m = /^([a-z]+)(\d+)$/i.exec(part);
  if (m) return `${cap(m[1])} ${m[2]}`;
  return cap(part);
}

function cap(s: string): string {
  return s.length ? s[0].toUpperCase() + s.slice(1) : s;
}

function describeChannel(id: string): ChannelSpec {
  const parts = id.split(".");
  let role = "";
  for (const p of parts) {
    const key = p === "vol" ? "volume" : p;
    if (ROLE_INFO[key]) role = key;
  }
  const info = ROLE_INFO[role] ?? { uom: "", decimals: 2, order: 99 };
  return {
    label: parts.map(titleize).join(" "),
    uom: info.uom,
    decimals: info.decimals,
    order: info.order,
  };
}

// --- readout ---------------------------------------------------------------

interface Card {
  card: El;
  value: El;
  spec: ChannelSpec;
}

// Readout renders the live value screen. Channels are DYNAMIC: a card appears for
// whatever channel ids arrive in the stream, so the display aligns to the pump.
export class Readout {
  private grid: El;
  private placeholder: El;
  private cards = new Map<string, Card>();

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

    // --- value grid ---
    this.grid = el("main", "grid");
    this.placeholder = el("div", "placeholder", "waiting for data…");
    this.grid.append(this.placeholder);

    // --- footer meta ---
    const footer = el("footer", "meta");
    this.seqEl = el("span", "meta-item", "seq —");
    this.updatedEl = el("span", "meta-item", "no data yet");
    footer.append(this.seqEl, this.updatedEl);

    root.append(header, this.grid, footer);

    this.applyStatus();
    window.setInterval(() => this.applyStatus(), 1000);
  }

  setConnected(connected: boolean): void {
    this.connected = connected;
    this.applyStatus();
  }

  update(r: Reading): void {
    this.lastReadingAt = Date.now();

    let added = false;
    for (const [id, v] of Object.entries(r.values)) {
      let card = this.cards.get(id);
      if (!card) {
        card = this.createCard(id);
        this.cards.set(id, card);
        added = true;
      }
      card.value.textContent =
        v === undefined || Number.isNaN(v)
          ? "—"
          : v.toLocaleString(undefined, {
              minimumFractionDigits: card.spec.decimals,
              maximumFractionDigits: card.spec.decimals,
            });
    }
    if (added) this.reorder();

    this.seqEl.textContent = `seq ${r.seq}`;
    this.applyStatus();
  }

  private createCard(id: string): Card {
    if (this.placeholder.isConnected) this.placeholder.remove();
    const spec = describeChannel(id);
    const card = el("section", "card");
    card.append(el("div", "card-label", spec.label));
    const valueRow = el("div", "card-valuerow");
    const value = el("span", "card-value", "—");
    valueRow.append(value);
    if (spec.uom) valueRow.append(el("span", "card-unit", spec.uom));
    card.append(valueRow);
    this.grid.append(card);
    return { card, value, spec };
  }

  // Re-order cards by role priority then id when a new channel shows up.
  private reorder(): void {
    const sorted = [...this.cards.entries()].sort((a, b) => {
      const o = a[1].spec.order - b[1].spec.order;
      return o !== 0 ? o : a[0].localeCompare(b[0]);
    });
    for (const [, card] of sorted) this.grid.append(card.card);
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
