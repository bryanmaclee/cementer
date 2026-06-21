import type { Channel, Job, PrintConfigResponse, PrintOverride, Profile } from "./types.ts";
import { JobChart } from "./chart/jobchart.ts";

// ReportView: the printable per-job report (Phase 4b — chart-config scope #2). It
// renders the 3b job header block (company / well / casing / job type / location /
// cementer / date) + the job's RECORDED chart (reusing JobChart over
// /api/jobs/{id}/series, segment-aware) + a MINIMAL override editor (channels on/off,
// report title, page size) + a "Print / Save as PDF" button that calls window.print().
//
// PDF = browser Save-as-PDF ONLY (D-pdf): no server render, no Pi-side archival, no new
// deps — @media print CSS hides the app chrome and prints only the report at the chosen
// page size. The company DEFAULT template is bundled server-side; per-job overrides
// persist on the Pi (axiom #3) via GET/PUT /api/jobs/{id}/print-config.
//
// AXIOM #1: this is a READ/CONFIG view over the always-on store — it never gates or
// touches ingestion, the live stream, or recording.
//
// NO framework — plain TS modules + direct DOM + the uPlot library (via JobChart).

function el(tag: string, className?: string, text?: string): HTMLElement {
  const e = document.createElement(tag);
  if (className) e.className = className;
  if (text !== undefined) e.textContent = text;
  return e;
}

async function getJSON<T>(url: string): Promise<T> {
  const r = await fetch(url, { headers: { Accept: "application/json" } });
  if (!r.ok) throw new Error(`${url}: HTTP ${r.status}`);
  return (await r.json()) as T;
}

// fmtDate renders a unix-micros timestamp as a local date (the report's "date").
function fmtDate(us: number): string {
  if (!us) return "—";
  return new Date(us / 1000).toLocaleString(undefined, {
    year: "numeric",
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export class ReportView {
  private host: HTMLElement;
  // Print sheet: everything inside prints; the app chrome is hidden by @media print.
  private sheet: HTMLElement;
  private titleEl: HTMLElement;
  private metaGrid: HTMLElement;
  private chartHost: HTMLElement;
  private chart: JobChart;

  // Editor (NOT printed — hidden by @media print).
  private editor: HTMLElement;
  private titleInput: HTMLInputElement;
  private pageSelect: HTMLSelectElement;
  private legendCheck: HTMLInputElement;
  private channelList: HTMLElement;
  private statusLine: HTMLElement;

  private profile: Profile | null = null;
  private jobId = 0;
  private job: Job | null = null;
  private cfg: PrintConfigResponse | null = null;
  private channelChecks = new Map<string, HTMLInputElement>();
  // A managed <style> whose @page size we rewrite at print time. An @page at-rule
  // can't be conditioned on a selector, so the chosen page size lives here.
  private pageStyle: HTMLStyleElement;

  constructor(host: HTMLElement) {
    this.host = host;
    this.host.replaceChildren();

    // --- editor (screen-only) ---
    this.editor = el("div", "report-editor");
    const editorTitle = el("div", "report-editor-title", "Report settings");

    const titleField = el("label", "report-field");
    titleField.append(el("span", "report-field-label", "Report title"));
    this.titleInput = el("input", "report-input") as HTMLInputElement;
    this.titleInput.type = "text";
    titleField.append(this.titleInput);

    const pageField = el("label", "report-field");
    pageField.append(el("span", "report-field-label", "Page size"));
    this.pageSelect = el("select", "report-select") as HTMLSelectElement;
    for (const [val, label] of [
      ["letter", "Letter"],
      ["a4", "A4"],
    ] as Array<[string, string]>) {
      const opt = el("option", undefined, label) as HTMLOptionElement;
      opt.value = val;
      this.pageSelect.append(opt);
    }
    pageField.append(this.pageSelect);

    const legendField = el("label", "report-field report-field-inline");
    this.legendCheck = el("input") as HTMLInputElement;
    this.legendCheck.type = "checkbox";
    legendField.append(this.legendCheck, el("span", "report-field-label", "Show legend"));

    const channelsWrap = el("div", "report-field");
    channelsWrap.append(el("span", "report-field-label", "Channels"));
    this.channelList = el("div", "report-channels");
    channelsWrap.append(this.channelList);

    const actions = el("div", "report-actions");
    const saveBtn = el("button", "primary-btn", "Save settings") as HTMLButtonElement;
    saveBtn.type = "button";
    saveBtn.addEventListener("click", () => void this.save());
    const resetBtn = el("button", "ghost-btn", "Reset to company default") as HTMLButtonElement;
    resetBtn.type = "button";
    resetBtn.addEventListener("click", () => void this.reset());
    const printBtn = el("button", "primary-btn", "Print / Save as PDF") as HTMLButtonElement;
    printBtn.type = "button";
    printBtn.addEventListener("click", () => this.print());
    this.statusLine = el("span", "report-status", "");
    actions.append(saveBtn, resetBtn, printBtn, this.statusLine);

    this.editor.append(editorTitle, titleField, pageField, legendField, channelsWrap, actions);

    // --- printable sheet ---
    this.sheet = el("section", "report-sheet");
    this.titleEl = el("h1", "report-title", "Cement Job Report");
    this.metaGrid = el("div", "report-meta");
    const chartFrame = el("div", "report-chart");
    this.chartHost = el("div", "report-chart-host");
    chartFrame.append(this.chartHost);
    this.sheet.append(this.titleEl, this.metaGrid, chartFrame);

    this.host.append(this.editor, this.sheet);

    // Managed @page size element (defaults to letter). Lives in <head>.
    this.pageStyle = document.createElement("style");
    this.pageStyle.textContent = "@media print { @page { size: letter; } }";
    document.head.append(this.pageStyle);

    this.chart = new JobChart(this.chartHost);

    // uPlot print sizing: size the chart for the print page on print start, restore
    // for the screen after. matchMedia covers Save-as-PDF in browsers that fire it.
    window.addEventListener("beforeprint", () => this.sizeForPrint());
    window.addEventListener("afterprint", () => this.chart.onShow());
    const mql = window.matchMedia("print");
    mql.addEventListener("change", (e) => {
      if (e.matches) this.sizeForPrint();
      else this.chart.onShow();
    });
  }

  setProfile(p: Profile): void {
    this.profile = p;
    this.chart.setProfile(p);
  }

  // setActiveJob targets the report at a job. 0/null clears it.
  setActiveJob(id: number | null): void {
    this.jobId = id ?? 0;
  }

  // load fetches the job header + print config and renders the report. Called when the
  // Report tab is shown.
  async load(): Promise<void> {
    if (this.jobId <= 0) {
      this.titleEl.textContent = "Cement Job Report";
      this.metaGrid.replaceChildren(el("div", "report-meta-empty", "no active job — select or create one, record a segment, then print"));
      this.chart.setChannelFilter([]);
      void this.chart.load(0);
      this.setEditorEnabled(false);
      return;
    }

    this.setStatus("loading…");
    try {
      this.job = await getJSON<Job>(`/api/jobs/${this.jobId}`);
      this.cfg = await getJSON<PrintConfigResponse>(`/api/jobs/${this.jobId}/print-config`);
    } catch (e) {
      this.setStatus(`could not load report (${(e as Error).message})`);
      return;
    }
    this.setStatus("");
    this.setEditorEnabled(true);
    this.fillEditor();
    this.renderHeader();
    this.applyConfigToChart();
    await this.chart.load(this.jobId);
  }

  onShow(): void {
    void this.load();
  }

  // --- rendering -----------------------------------------------------------

  private renderHeader(): void {
    const j = this.job;
    const cfg = this.cfg;
    this.titleEl.textContent = cfg?.effective.title || "Cement Job Report";
    this.metaGrid.replaceChildren();
    if (!j) return;
    const rows: Array<[string, string]> = [
      ["Company", j.company],
      ["Well", j.well],
      ["Casing size", j.casingSize],
      ["Job type", j.jobType],
      ["Location", j.location],
      ["Cementer", j.cementer],
      ["Job", j.name],
      ["Date", fmtDate(j.createdAtUs)],
    ];
    for (const [label, value] of rows) {
      const cell = el("div", "report-meta-cell");
      cell.append(el("span", "report-meta-label", label), el("span", "report-meta-value", value || "—"));
      this.metaGrid.append(cell);
    }
  }

  // chartableChannels are the profile's enabled, non-meta channels (job.number etc.
  // excluded — same rule as the live + job charts).
  private chartableChannels(): Channel[] {
    if (!this.profile) return [];
    return this.profile.channels.filter((c) => c.scope !== "meta" && c.role !== "meta");
  }

  // effectiveChannelIds resolves the effective config's channel list: explicit ids when
  // set, else all chartable channels (the company default's "all enabled").
  private effectiveChannelIds(): string[] {
    const explicit = this.cfg?.effective.channels;
    if (explicit && explicit.length > 0) return explicit;
    return this.chartableChannels().map((c) => c.id);
  }

  private applyConfigToChart(): void {
    this.chart.setChannelFilter(this.effectiveChannelIds());
    this.chart.setLegendVisible(this.cfg?.effective.showLegend ?? true);
    this.applyPageSize(this.cfg?.effective.pageSize ?? "letter");
  }

  private applyPageSize(size: string): void {
    const a4 = size === "a4";
    this.sheet.dataset.page = a4 ? "a4" : "letter";
    this.pageStyle.textContent = `@media print { @page { size: ${a4 ? "a4" : "letter"}; } }`;
  }

  // --- editor --------------------------------------------------------------

  private fillEditor(): void {
    const eff = this.cfg?.effective;
    this.titleInput.value = eff?.title ?? "";
    this.pageSelect.value = eff?.pageSize ?? "letter";
    this.legendCheck.checked = eff?.showLegend ?? true;

    // Channel on/off list: a checkbox per chartable channel, checked when included in
    // the effective set.
    const included = new Set(this.effectiveChannelIds());
    this.channelList.replaceChildren();
    this.channelChecks.clear();
    for (const c of this.chartableChannels()) {
      const row = el("label", "report-channel-row");
      const cb = el("input") as HTMLInputElement;
      cb.type = "checkbox";
      cb.checked = included.has(c.id);
      row.append(cb, el("span", "report-channel-name", c.label || c.id));
      this.channelList.append(row);
      this.channelChecks.set(c.id, cb);
    }
  }

  private setEditorEnabled(on: boolean): void {
    this.editor.classList.toggle("disabled", !on);
    for (const ctl of [this.titleInput, this.pageSelect, this.legendCheck]) {
      (ctl as HTMLInputElement | HTMLSelectElement).disabled = !on;
    }
  }

  // collectOverride builds a MINIMAL override: only fields that differ from the company
  // default land in the body (so the stored blob stays just the cementer's deltas and a
  // later default change still flows through untouched fields).
  private collectOverride(): PrintOverride {
    const def = this.cfg?.default;
    const ov: PrintOverride = {};

    const title = this.titleInput.value.trim();
    if (def && title !== def.title) ov.title = title;

    const page = (this.pageSelect.value === "a4" ? "a4" : "letter") as "letter" | "a4";
    if (def && page !== def.pageSize) ov.pageSize = page;

    const legend = this.legendCheck.checked;
    if (def && legend !== def.showLegend) ov.showLegend = legend;

    // Channels: compare the chosen set to the default's effective set. The default's
    // "all enabled" is represented by an empty/absent channels list.
    const chosen = this.chartableChannels()
      .map((c) => c.id)
      .filter((id) => this.channelChecks.get(id)?.checked);
    const all = this.chartableChannels().map((c) => c.id);
    const defaultIsAll = !def?.channels || def.channels.length === 0;
    const chosenIsAll = chosen.length === all.length;
    if (defaultIsAll) {
      if (!chosenIsAll) ov.channels = chosen;
    } else {
      const sameAsDefault =
        def!.channels!.length === chosen.length && def!.channels!.every((id, i) => chosen[i] === id);
      if (!sameAsDefault) ov.channels = chosen;
    }
    return ov;
  }

  private async save(): Promise<void> {
    if (this.jobId <= 0) return;
    const ov = this.collectOverride();
    this.setStatus("saving…");
    try {
      const r = await fetch(`/api/jobs/${this.jobId}/print-config`, {
        method: "PUT",
        headers: { "Content-Type": "application/json", Accept: "application/json" },
        body: JSON.stringify(ov),
      });
      if (!r.ok) throw new Error(`HTTP ${r.status}`);
      this.cfg = (await r.json()) as PrintConfigResponse;
    } catch (e) {
      this.setStatus(`save failed (${(e as Error).message})`);
      return;
    }
    this.setStatus("saved");
    this.fillEditor();
    this.renderHeader();
    this.applyConfigToChart();
    await this.chart.load(this.jobId);
  }

  // reset writes an empty override ({}) so the job falls back to the company default.
  private async reset(): Promise<void> {
    if (this.jobId <= 0) return;
    this.setStatus("resetting…");
    try {
      const r = await fetch(`/api/jobs/${this.jobId}/print-config`, {
        method: "PUT",
        headers: { "Content-Type": "application/json", Accept: "application/json" },
        body: "{}",
      });
      if (!r.ok) throw new Error(`HTTP ${r.status}`);
      this.cfg = (await r.json()) as PrintConfigResponse;
    } catch (e) {
      this.setStatus(`reset failed (${(e as Error).message})`);
      return;
    }
    this.setStatus("reset to company default");
    this.fillEditor();
    this.renderHeader();
    this.applyConfigToChart();
    await this.chart.load(this.jobId);
  }

  // --- print ---------------------------------------------------------------

  // sizeForPrint sizes the report chart to the printable page width BEFORE the print
  // dialog reads the layout, so the printed chart isn't blank or clipped. uPlot reads 0
  // while display:none, so we compute the width from the chosen page size.
  private sizeForPrint(): void {
    const a4 = this.sheet.dataset.page === "a4";
    // Printable inner width at ~0.5in margins: Letter 7.5in, A4 ~7.27in. 96 px/in.
    const widthIn = a4 ? 7.27 : 7.5;
    const width = Math.round(widthIn * 96);
    const height = Math.round(width * 0.5); // a calm 2:1 report aspect
    this.chart.setSize(width, height);
  }

  private print(): void {
    // Ensure the chart is sized for the page even if the browser doesn't fire
    // beforeprint synchronously before measuring.
    this.sizeForPrint();
    window.print();
  }

  private setStatus(msg: string): void {
    this.statusLine.textContent = msg;
  }
}
