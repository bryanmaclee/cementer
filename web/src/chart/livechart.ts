import uPlot from "uplot";
import "uplot/dist/uPlot.min.css";
import type { Channel, Profile, Reading } from "../types.ts";
import { colorFor, orderScales, scaleKey, uomLabel } from "./roles.ts";
import { loadLiveConfig, setHidden, type LiveConfig } from "./config.ts";

// LiveChart: the rolling real-time chart that REPLACES the value readout as the
// default view (Phase 4a). X-axis = time. Traces = all ENABLED profile channels,
// grouped by role/uom into one uPlot scale per uom, distinct color per channel. A
// custom legend keeps each channel's LATEST value glanceable (the readout's
// value-at-a-glance utility must survive). NO framework — plain TS + direct DOM + the
// uPlot charting library (ratified, not a framework).
//
// AXIOM #1: this is a passive consumer of the live stream. It never gates ingestion,
// the live stream, or recording — push() just appends to an in-memory ring buffer.

// uPlot's time scale (scales.x.time = true) expects Unix timestamps in SECONDS, so the
// chart's x axis and all window/trim math are kept in seconds end-to-end. (Reading.ts on
// the wire stays RFC3339; we convert to epoch seconds here, at the chart boundary.)
const DEFAULT_WINDOW_SEC = 5 * 60; // 5 minutes
const MAX_POINTS = 4000; // ring cap, independent of the time window

function el(tag: string, className?: string, text?: string): HTMLElement {
  const e = document.createElement(tag);
  if (className) e.className = className;
  if (text !== undefined) e.textContent = text;
  return e;
}

function cssVar(name: string, fallback: string): string {
  const v = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  return v || fallback;
}

interface SeriesMeta {
  channel: Channel;
  color: string;
  colIndex: number; // index into the aligned-data columns (1-based; col 0 is x)
}

export class LiveChart {
  private host: HTMLElement;
  private chartHost: HTMLElement;
  private legendHost: HTMLElement;
  private plot: uPlot | null = null;

  // Aligned ring buffers: x (epoch seconds) + one y-array per series, all index-aligned.
  private xs: number[] = [];
  private ys: number[][] = [];
  private metas: SeriesMeta[] = [];
  private channelCol = new Map<string, number>(); // channel id -> column index

  private windowSec = DEFAULT_WINDOW_SEC;
  private cfg: LiveConfig = {};

  // Custom legend rows (latest value at a glance).
  private legendRows = new Map<string, { wrap: HTMLElement; value: HTMLElement }>();

  // Debounced redraw so a burst of readings coalesces into one paint.
  private dirty = false;

  constructor(host: HTMLElement) {
    this.host = host;
    this.host.replaceChildren();
    this.legendHost = el("div", "chart-legend");
    this.chartHost = el("div", "chart-canvas");
    const empty = el("div", "chart-empty", "waiting for profile…");
    this.chartHost.append(empty);
    this.host.append(this.legendHost, this.chartHost);

    this.cfg = loadLiveConfig();
    if (this.cfg.windowSec && this.cfg.windowSec > 0) this.windowSec = this.cfg.windowSec;

    window.addEventListener("resize", () => this.resize());
    // Paint loop: coalesce pushes into ~animation-frame cadence.
    const tick = () => {
      if (this.dirty && this.plot) {
        this.dirty = false;
        this.trim();
        this.plot.setData([this.xs, ...this.ys]);
        this.updateLegendValues();
      }
      requestAnimationFrame(tick);
    };
    requestAnimationFrame(tick);
  }

  // setWindowSec changes the rolling window in seconds (personal config; persisted by
  // the caller).
  setWindowSec(sec: number): void {
    if (sec > 0) this.windowSec = sec;
  }

  // applyConfig re-reads personal config (line on/off, window, colors) and rebuilds.
  applyConfig(cfg: LiveConfig): void {
    this.cfg = cfg;
    if (cfg.windowSec && cfg.windowSec > 0) this.windowSec = cfg.windowSec;
    if (this.lastProfile) this.applyProfile(this.lastProfile);
  }

  private lastProfile: Profile | null = null;

  // applyProfile (re)builds the chart from the enabled, non-meta channels: one scale
  // per uom, distinct color per channel, axes auto-assigned. Existing ring data for
  // surviving channels is preserved across a rebuild (reconnect with an edited profile).
  applyProfile(p: Profile): void {
    this.lastProfile = p;

    // Preserve existing column data by channel id across the rebuild.
    const prevXs = this.xs;
    const prevByChannel = new Map<string, number[]>();
    for (const [id, col] of this.channelCol) prevByChannel.set(id, this.ys[col - 1] ?? []);

    // Exclude meta channels by EITHER scope or role: e.g. job.number is
    // role:"meta", scope:"job", so a scope-only filter would chart it as a flat-0
    // trace. vol.job (role:"volume") still charts.
    const channels = p.channels.filter((c) => c.scope !== "meta" && c.role !== "meta");
    // Honor personal line on/off (default on).
    const visible = channels.filter((c) => this.cfg.hidden?.[c.id] !== true);

    this.metas = [];
    this.channelCol.clear();
    this.ys = [];
    this.xs = prevXs.slice();

    visible.forEach((c, i) => {
      const col = i + 1;
      const color = this.cfg.colors?.[c.id] ?? colorFor(i);
      this.metas.push({ channel: c, color, colIndex: col });
      this.channelCol.set(c.id, col);
      // Re-seed preserved data aligned to the (possibly shorter) shared x; pad/truncate.
      const prev = prevByChannel.get(c.id) ?? [];
      const seed = this.xs.map((_, idx) => prev[idx] ?? NaN);
      this.ys.push(seed);
    });

    this.buildPlot(p);
    this.buildLegend();
    this.dirty = true;
  }

  // push appends one reading to the ring buffers. Channels absent from the reading get
  // NaN for this column (uPlot draws a gap). The actual repaint happens on the next
  // animation frame (coalesced). Reading.ts is RFC3339 — convert to epoch SECONDS,
  // which is the unit uPlot's time scale expects.
  push(r: Reading): void {
    if (this.metas.length === 0) return; // no profile yet
    const tMs = Date.parse(r.ts);
    if (Number.isNaN(tMs)) return;
    const tSec = tMs / 1000;

    this.xs.push(tSec);
    for (const m of this.metas) {
      const v = r.values[m.channel.id];
      this.ys[m.colIndex - 1].push(v === undefined ? NaN : v);
    }
    this.dirty = true;
  }

  // trim drops points older than the rolling window and enforces the hard cap.
  private trim(): void {
    const n = this.xs.length;
    if (n === 0) return;
    const cutoff = this.xs[n - 1] - this.windowSec; // xs and windowSec are both in seconds
    let start = 0;
    while (start < n && this.xs[start] < cutoff) start++;
    if (n - start > MAX_POINTS) start = n - MAX_POINTS;
    if (start > 0) {
      this.xs = this.xs.slice(start);
      this.ys = this.ys.map((col) => col.slice(start));
    }
  }

  private buildPlot(p: Profile): void {
    if (this.plot) {
      this.plot.destroy();
      this.plot = null;
    }
    this.chartHost.replaceChildren();

    if (this.metas.length === 0) {
      this.chartHost.append(el("div", "chart-empty", "no chartable channels"));
      return;
    }

    const scales = orderScales(p.channels.filter((c) => c.scope !== "meta" && c.role !== "meta"));
    const axisColor = cssVar("--text-dim", "#8b97a3");
    const gridColor = cssVar("--border", "#262d35");

    // One axis per scale, alternating left/right so up to 4 roles stay legible.
    const axes: uPlot.Axis[] = [
      {
        scale: "x",
        stroke: axisColor,
        grid: { stroke: gridColor, width: 1 },
        ticks: { stroke: gridColor, width: 1 },
      },
    ];
    scales.forEach((s, i) => {
      axes.push({
        scale: s.key,
        side: i % 2 === 0 ? 3 : 1, // 3 = left, 1 = right
        stroke: axisColor,
        grid: { show: i === 0, stroke: gridColor, width: 1 },
        ticks: { stroke: gridColor, width: 1 },
        label: uomLabel(s.uom),
        labelSize: 14,
        size: 52,
      });
    });

    const series: uPlot.Series[] = [
      { label: "time" },
      ...this.metas.map((m) => ({
        label: m.channel.label,
        scale: scaleKey(m.channel),
        stroke: m.color,
        width: 1.75,
        points: { show: false },
        spanGaps: false,
      })),
    ];

    const opts: uPlot.Options = {
      width: this.chartHost.clientWidth || 800,
      height: this.chartHost.clientHeight || 420,
      scales: { x: { time: true } },
      axes,
      series,
      legend: { show: false }, // we render a custom always-on latest-value legend
      cursor: { focus: { prox: 30 } },
    };

    this.plot = new uPlot(opts, [this.xs, ...this.ys], this.chartHost);
  }

  // buildLegend renders an always-on legend with a swatch, label, and live LATEST value
  // per channel (the readout's at-a-glance utility). A click toggles the trace.
  private buildLegend(): void {
    this.legendHost.replaceChildren();
    this.legendRows.clear();
    this.metas.forEach((m, i) => {
      const row = el("button", "legend-row");
      (row as HTMLButtonElement).type = "button";
      const sw = el("span", "legend-swatch");
      sw.style.background = m.color;
      const name = el("span", "legend-name", m.channel.label);
      const value = el("span", "legend-value", "—");
      const unit = el("span", "legend-unit", m.channel.uom);
      row.append(sw, name, value, unit);
      row.addEventListener("click", () => this.toggleSeries(i));
      this.legendHost.append(row);
      this.legendRows.set(m.channel.id, { wrap: row, value });
    });
  }

  private toggleSeries(metaIdx: number): void {
    if (!this.plot) return;
    const seriesIdx = metaIdx + 1; // series 0 is x
    const s = this.plot.series[seriesIdx];
    const show = !s.show;
    this.plot.setSeries(seriesIdx, { show });
    const meta = this.metas[metaIdx];
    const row = this.legendRows.get(meta.channel.id);
    if (row) row.wrap.classList.toggle("off", !show);
    // Persist the on/off as a personal preference (scope #1, localStorage).
    this.cfg = setHidden(meta.channel.id, !show);
  }

  // updateLegendValues writes each channel's most recent non-NaN value into its legend
  // row, formatted to the channel's decimals.
  private updateLegendValues(): void {
    for (const m of this.metas) {
      const col = this.ys[m.colIndex - 1];
      let v = NaN;
      for (let i = col.length - 1; i >= 0; i--) {
        if (!Number.isNaN(col[i])) {
          v = col[i];
          break;
        }
      }
      const row = this.legendRows.get(m.channel.id);
      if (!row) continue;
      row.value.textContent = Number.isNaN(v)
        ? "—"
        : v.toLocaleString(undefined, {
            minimumFractionDigits: m.channel.decimals,
            maximumFractionDigits: m.channel.decimals,
          });
    }
  }

  private resize(): void {
    if (!this.plot) return;
    this.plot.setSize({
      width: this.chartHost.clientWidth || 800,
      height: this.chartHost.clientHeight || 420,
    });
  }

  // onShow must be called when the view becomes visible so uPlot sizes to the now-laid-
  // out container (it measures 0 while display:none).
  onShow(): void {
    this.resize();
  }

  destroy(): void {
    this.plot?.destroy();
    this.plot = null;
  }
}
