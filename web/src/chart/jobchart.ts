import uPlot from "uplot";
import "uplot/dist/uPlot.min.css";
import type { Channel, Profile, Segment } from "../types.ts";
import { colorFor, orderScales, scaleKey, uomLabel } from "./roles.ts";

// JobChart: the historical per-job chart over recorded segments (Phase 4a). Fetches
// GET /api/jobs/{id}/series, renders vs time with the SAME role-grouped axes as the
// live chart, and SHADES each recording segment band. The default chart shows only
// recorded data (data-model.md) — the server already returns in-segment samples, so
// gaps between segments are real gaps. Pan/zoom enabled. NO framework; uPlot only.
//
// AXIOM #1: this is a READ-ONLY view over the store (GET only) — it never touches
// ingestion, the live stream, or recording.

type Point = [number, number]; // [tsUs, value]

interface JobSeriesResponse {
  segments: Segment[];
  series: Record<string, Point[]>;
}

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

export class JobChart {
  private host: HTMLElement;
  private chartHost: HTMLElement;
  private statusEl: HTMLElement;
  private plot: uPlot | null = null;
  private profile: Profile | null = null;

  constructor(host: HTMLElement) {
    this.host = host;
    this.host.replaceChildren();
    this.statusEl = el("div", "jobchart-status", "select a job to view its recorded chart");
    this.chartHost = el("div", "chart-canvas");
    this.host.append(this.statusEl, this.chartHost);
    window.addEventListener("resize", () => this.resize());
  }

  setProfile(p: Profile): void {
    this.profile = p;
  }

  // load fetches and renders a job's recorded series. jobId<=0 clears the chart.
  async load(jobId: number): Promise<void> {
    if (this.plot) {
      this.plot.destroy();
      this.plot = null;
    }
    this.chartHost.replaceChildren();
    if (jobId <= 0) {
      this.statusEl.textContent = "no active job — select or create one to record, then view it here";
      return;
    }

    this.statusEl.textContent = "loading recorded data…";
    let data: JobSeriesResponse;
    try {
      const r = await fetch(`/api/jobs/${jobId}/series`, { headers: { Accept: "application/json" } });
      if (r.status === 404) {
        this.statusEl.textContent = "no such job";
        return;
      }
      if (!r.ok) throw new Error(`HTTP ${r.status}`);
      data = (await r.json()) as JobSeriesResponse;
    } catch (e) {
      this.statusEl.textContent = `could not load job series (${(e as Error).message})`;
      return;
    }

    const channelIds = Object.keys(data.series);
    const hasData = channelIds.some((id) => data.series[id].length > 0);
    if (!hasData) {
      this.statusEl.textContent =
        data.segments.length === 0
          ? "this job has no recording segments yet — press Record on the live view"
          : "no samples recorded in this job's segments yet";
      return;
    }

    const segCount = data.segments.length;
    this.statusEl.textContent = `${segCount} recording segment${segCount === 1 ? "" : "s"}`;
    this.render(data);
  }

  // render builds the aligned data (union x timeline, gaps as null) + the plot.
  private render(data: JobSeriesResponse): void {
    // Order channels by the profile (label/uom/color), falling back to the raw ids.
    const channels = this.orderedChannels(Object.keys(data.series));

    // Union x timeline (epoch SECONDS — uPlot's time-scale unit) across all channels'
    // decimated points. The store/wire stay in microseconds; we convert here.
    const xsSet = new Set<number>();
    for (const c of channels) {
      for (const p of data.series[c.id] ?? []) xsSet.add(p[0] / 1_000_000); // us -> s
    }
    const xs = [...xsSet].sort((a, b) => a - b);
    const xIndex = new Map<number, number>();
    xs.forEach((x, i) => xIndex.set(x, i));

    // Per-channel y aligned to the union x; missing => null (uPlot gap).
    const ys: (number | null)[][] = channels.map(() => new Array(xs.length).fill(null));
    channels.forEach((c, ci) => {
      for (const p of data.series[c.id] ?? []) {
        const i = xIndex.get(p[0] / 1_000_000);
        if (i !== undefined) ys[ci][i] = p[1];
      }
    });

    const scales = orderScales(channels);
    const axisColor = cssVar("--text-dim", "#8b97a3");
    const gridColor = cssVar("--border", "#262d35");

    const axes: uPlot.Axis[] = [
      { scale: "x", stroke: axisColor, grid: { stroke: gridColor, width: 1 }, ticks: { stroke: gridColor, width: 1 } },
    ];
    scales.forEach((s, i) => {
      axes.push({
        scale: s.key,
        side: i % 2 === 0 ? 3 : 1,
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
      ...channels.map((c, i) => ({
        label: c.label,
        scale: scaleKey(c),
        stroke: colorFor(i),
        width: 1.75,
        points: { show: false },
        spanGaps: false,
      })),
    ];

    const shade = this.segmentShadePlugin(data.segments);

    const opts: uPlot.Options = {
      width: this.chartHost.clientWidth || 800,
      height: this.chartHost.clientHeight || 460,
      scales: { x: { time: true } },
      axes,
      series,
      legend: { show: true, live: true },
      cursor: { drag: { x: true, y: false }, focus: { prox: 30 } },
      plugins: [shade],
    };

    this.plot = new uPlot(opts, [xs, ...ys], this.chartHost);
  }

  // segmentShadePlugin shades each recording segment band [started, stopped] behind the
  // series. Open segments (stoppedAtUs null) shade up to the latest x. Uses the drawClear
  // hook so the band paints before the series lines.
  private segmentShadePlugin(segments: Segment[]): uPlot.Plugin {
    return {
      hooks: {
        drawClear: (u: uPlot) => {
          const { ctx } = u;
          const { left, top, width, height } = u.bbox;
          const fill = cssVar("--accent", "#4aa3ff");
          ctx.save();
          ctx.globalAlpha = 0.08;
          ctx.fillStyle = fill;
          for (const seg of segments) {
            // Segment bounds are microseconds on the wire; the x scale is in seconds.
            const startSec = seg.startedAtUs / 1_000_000;
            const stopSec = (seg.stoppedAtUs ?? Date.now() * 1000) / 1_000_000;
            let x0 = u.valToPos(startSec, "x", true);
            let x1 = u.valToPos(stopSec, "x", true);
            // Clip to the plotting area.
            x0 = Math.max(left, Math.min(left + width, x0));
            x1 = Math.max(left, Math.min(left + width, x1));
            const w = Math.max(0, x1 - x0);
            if (w > 0) ctx.fillRect(x0, top, w, height);
          }
          ctx.restore();
        },
      },
    };
  }

  // orderedChannels maps the returned channel ids onto profile channels (for label/uom),
  // ordered by the profile's sort order; unknown ids fall back to a minimal channel.
  private orderedChannels(ids: string[]): Channel[] {
    const byId = new Map<string, Channel>();
    if (this.profile) for (const c of this.profile.channels) byId.set(c.id, c);
    const known: Channel[] = [];
    const unknown: Channel[] = [];
    // Preserve profile order for known channels.
    if (this.profile) {
      for (const c of this.profile.channels) {
        if (ids.includes(c.id) && c.scope !== "meta") known.push(c);
      }
    }
    for (const id of ids) {
      if (!byId.has(id)) {
        unknown.push({ id, role: "value", scope: "aggregate", unitIndex: 0, label: id, uom: "", decimals: 2 });
      }
    }
    return [...known, ...unknown];
  }

  private resize(): void {
    if (!this.plot) return;
    this.plot.setSize({
      width: this.chartHost.clientWidth || 800,
      height: this.chartHost.clientHeight || 460,
    });
  }

  // onShow re-measures the container (uPlot reads 0 while display:none).
  onShow(): void {
    this.resize();
  }
}
