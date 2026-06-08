import type { Reading } from "./types.ts";

interface ChannelSpec {
  key: string;
  label: string;
  unit: string;
  decimals: number;
}

// Phase-1 channel set. Order here is display order. Unknown channels that arrive
// in a reading are ignored for now (the chart phase makes channels dynamic).
const CHANNELS: ChannelSpec[] = [
  { key: "pressure", label: "Pressure", unit: "psi", decimals: 0 },
  { key: "rate", label: "Rate", unit: "bbl/min", decimals: 2 },
  { key: "density", label: "Density", unit: "ppg", decimals: 2 },
  { key: "volume", label: "Volume", unit: "bbl", decimals: 1 },
];

const STALE_MS = 3000;

type El = HTMLElement;

function el(tag: string, className?: string, text?: string): El {
  const e = document.createElement(tag);
  if (className) e.className = className;
  if (text !== undefined) e.textContent = text;
  return e;
}

// Readout owns the live value screen: a status header and one big-number card per
// channel. It builds the DOM once and mutates text nodes on each reading.
export class Readout {
  private valueEls = new Map<string, El>();
  private statusDot: El;
  private statusText: El;
  private seqEl: El;
  private updatedEl: El;

  private connected = false;
  private lastReadingAt = 0;

  constructor(root: El) {
    root.replaceChildren();

    // --- header ---
    const header = el("header", "topbar");
    const brand = el("div", "brand");
    brand.append(el("span", "brand-mark", "●"), el("span", "brand-name", "cementer"));

    const status = el("div", "status");
    this.statusDot = el("span", "dot");
    this.statusText = el("span", "status-text", "connecting…");
    status.append(this.statusDot, this.statusText);

    header.append(brand, status);

    // --- value grid ---
    const grid = el("main", "grid");
    for (const ch of CHANNELS) {
      const card = el("section", "card");
      card.append(el("div", "card-label", ch.label));
      const valueRow = el("div", "card-valuerow");
      const value = el("span", "card-value", "—");
      const unit = el("span", "card-unit", ch.unit);
      valueRow.append(value, unit);
      card.append(valueRow);
      grid.append(card);
      this.valueEls.set(ch.key, value);
    }

    // --- footer meta ---
    const footer = el("footer", "meta");
    this.seqEl = el("span", "meta-item", "seq —");
    this.updatedEl = el("span", "meta-item", "no data yet");
    footer.append(this.seqEl, this.updatedEl);

    root.append(header, grid, footer);

    this.applyStatus();
    window.setInterval(() => this.applyStatus(), 1000);
  }

  setConnected(connected: boolean): void {
    this.connected = connected;
    this.applyStatus();
  }

  update(r: Reading): void {
    this.lastReadingAt = Date.now();
    for (const ch of CHANNELS) {
      const v = r.values[ch.key];
      const target = this.valueEls.get(ch.key);
      if (!target) continue;
      target.textContent =
        v === undefined || Number.isNaN(v)
          ? "—"
          : v.toLocaleString(undefined, {
              minimumFractionDigits: ch.decimals,
              maximumFractionDigits: ch.decimals,
            });
    }
    this.seqEl.textContent = `seq ${r.seq}`;
    this.applyStatus();
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
