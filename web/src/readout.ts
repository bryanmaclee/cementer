import type { Profile, Reading } from "./types.ts";
import { initTheme, toggleTheme } from "./theme.ts";
import { LiveChart } from "./chart/livechart.ts";
import { JobChart } from "./chart/jobchart.ts";
import { loadLiveConfig, setWindowSec } from "./chart/config.ts";

const STALE_MS = 3000;

type El = HTMLElement;
type View = "live" | "job";

function el(tag: string, className?: string, text?: string): El {
  const e = document.createElement(tag);
  if (className) e.className = className;
  if (text !== undefined) e.textContent = text;
  return e;
}

// Readout is the app shell (Phase 4a). It keeps the header (brand + connection/stale
// status + theme toggle), the job/record control strip host, the footer meta, and a
// VIEW AREA with a Live | Job History toggle. The default Live view is now a rolling
// real-time CHART (replacing the old value grid); the chart's legend keeps each
// channel's latest value glanceable. Job History shows the recorded per-job chart.
// NO framework — plain modules + direct DOM + the uPlot library.
//
// AXIOM #1: nothing here gates the live stream or recording. The chart is a passive
// consumer; record controls are markers; switching views is a client-only concern.
export class Readout {
  private controls: El; // host for the job/record control strip (filled by Controls)

  private statusDot: El;
  private statusText: El;
  private seqEl: El;
  private updatedEl: El;

  private liveHost: El;
  private jobHost: El;
  private liveTab: HTMLButtonElement;
  private jobTab: HTMLButtonElement;
  private windowSelect: HTMLSelectElement;

  private liveChart: LiveChart;
  private jobChart: JobChart;
  private view: View = "live";

  private connected = false;
  private lastReadingAt = 0;
  private activeJobId: number | null = null;

  constructor(root: El) {
    initTheme();
    root.replaceChildren();

    // --- header ---
    const header = el("header", "topbar");
    const brand = el("div", "brand");
    brand.append(el("span", "brand-mark", "●"), el("span", "brand-name", "cementer"));

    const right = el("div", "topbar-right");

    // View toggle: Live | Job History.
    const tabs = el("div", "view-tabs");
    this.liveTab = el("button", "view-tab active", "Live") as HTMLButtonElement;
    this.liveTab.type = "button";
    this.jobTab = el("button", "view-tab", "Job History") as HTMLButtonElement;
    this.jobTab.type = "button";
    this.liveTab.addEventListener("click", () => this.setView("live"));
    this.jobTab.addEventListener("click", () => this.setView("job"));
    tabs.append(this.liveTab, this.jobTab);

    // Rolling-window selector (personal config). Values are SECONDS (uPlot's time unit).
    this.windowSelect = el("select", "window-select") as HTMLSelectElement;
    for (const [label, mins] of [
      ["1 min", 1],
      ["5 min", 5],
      ["15 min", 15],
      ["30 min", 30],
      ["60 min", 60],
    ] as Array<[string, number]>) {
      const opt = el("option", undefined, label) as HTMLOptionElement;
      opt.value = String(mins * 60);
      this.windowSelect.append(opt);
    }
    const cfg = loadLiveConfig();
    this.windowSelect.value = String(cfg.windowSec && cfg.windowSec > 0 ? cfg.windowSec : 5 * 60);
    this.windowSelect.title = "Live rolling window";
    this.windowSelect.addEventListener("change", () => {
      const sec = Number(this.windowSelect.value);
      this.liveChart.setWindowSec(sec);
      setWindowSec(sec);
    });

    const status = el("div", "status");
    this.statusDot = el("span", "dot");
    this.statusText = el("span", "status-text", "connecting…");
    status.append(this.statusDot, this.statusText);

    const themeBtn = el("button", "theme-btn", "◐");
    themeBtn.title = "Toggle dark / light";
    themeBtn.addEventListener("click", () => toggleTheme());

    right.append(tabs, this.windowSelect, status, themeBtn);
    header.append(brand, right);

    // --- control strip host (job selector + record button live here; Controls fills
    // it). Recording is a marker over the always-on store and never gates the live
    // stream/chart (axiom #1). ---
    this.controls = el("div", "controls-host");

    // --- view area: Live chart (default) + Job chart (hidden until toggled) ---
    const content = el("main", "content content-chart");
    this.liveHost = el("section", "view view-live");
    this.jobHost = el("section", "view view-job");
    this.jobHost.hidden = true;
    content.append(this.liveHost, this.jobHost);

    // --- footer meta ---
    const footer = el("footer", "meta");
    this.seqEl = el("span", "meta-item", "seq —");
    this.updatedEl = el("span", "meta-item", "no data yet");
    footer.append(this.seqEl, this.updatedEl);

    root.append(header, this.controls, content, footer);

    this.liveChart = new LiveChart(this.liveHost);
    this.jobChart = new JobChart(this.jobHost);

    this.applyStatus();
    window.setInterval(() => this.applyStatus(), 1000);
  }

  setConnected(connected: boolean): void {
    this.connected = connected;
    this.applyStatus();
  }

  // controlsHost is the empty container (between header and the view area) where the
  // job/record controls mount. The Readout owns the layout; Controls owns the strip.
  controlsHost(): El {
    return this.controls;
  }

  // setActiveJob lets the Controls strip tell the shell which job the Job History view
  // should load. Called by main's wiring when the active job changes.
  setActiveJob(id: number | null): void {
    this.activeJobId = id;
    if (this.view === "job") void this.jobChart.load(id ?? 0);
  }

  applyProfile(p: Profile): void {
    this.liveChart.applyProfile(p);
    this.jobChart.setProfile(p);
  }

  update(r: Reading): void {
    this.lastReadingAt = Date.now();
    this.liveChart.push(r);
    this.seqEl.textContent = `seq ${r.seq}`;
    this.applyStatus();
  }

  private setView(v: View): void {
    if (this.view === v) return;
    this.view = v;
    const live = v === "live";
    this.liveHost.hidden = !live;
    this.jobHost.hidden = live;
    this.liveTab.classList.toggle("active", live);
    this.jobTab.classList.toggle("active", !live);
    this.windowSelect.hidden = !live;
    if (live) {
      this.liveChart.onShow();
    } else {
      void this.jobChart.load(this.activeJobId ?? 0);
      this.jobChart.onShow();
    }
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
