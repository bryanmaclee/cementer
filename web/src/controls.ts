import type { Job, JobInput, RecordingState, Segment } from "./types.ts";

// Minimal job + recording controls (vanilla TS, no framework — anti-patterns Part B).
// A control strip with an active-job <select> (plus an inline "new job" form) and a
// Record/Stop button with an open-segment elapsed timer. State is plain module-level
// fields + DOM; no state library. All persistence is on the Pi (axiom #3) — nothing
// here goes to localStorage.
//
// AXIOM #1: these controls only call /api/recording/* (marker insert/update). They
// never gate ingestion or the live readout — the readout keeps streaming regardless
// of record state. The Record button is purely a marker over the always-on store.

const STATE_POLL_MS = 3000;

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

interface SendResult<T> {
  ok: boolean;
  status: number;
  body: T | null;
}

async function sendJSON<T>(method: string, url: string, body?: unknown): Promise<SendResult<T>> {
  const init: RequestInit = { method, headers: { Accept: "application/json" } };
  if (body !== undefined) {
    init.headers = { ...init.headers, "Content-Type": "application/json" };
    init.body = JSON.stringify(body);
  }
  const r = await fetch(url, init);
  let parsed: T | null = null;
  try {
    parsed = (await r.json()) as T;
  } catch {
    parsed = null;
  }
  return { ok: r.ok, status: r.status, body: parsed };
}

// fmtElapsed renders milliseconds as mm:ss (or h:mm:ss past an hour).
function fmtElapsed(ms: number): string {
  const total = Math.max(0, Math.floor(ms / 1000));
  const s = total % 60;
  const m = Math.floor(total / 60) % 60;
  const h = Math.floor(total / 3600);
  const pad = (n: number) => n.toString().padStart(2, "0");
  return h > 0 ? `${h}:${pad(m)}:${pad(s)}` : `${pad(m)}:${pad(s)}`;
}

// ActiveJobHandler is notified whenever the active job changes (or is first known), so
// the shell's Job History view can load that job's recorded chart.
export type ActiveJobHandler = (id: number | null) => void;

// Controls mounts the strip into a host element and manages its own polling timers.
export class Controls {
  private jobSelect: HTMLSelectElement;
  private recordBtn: HTMLButtonElement;
  private recordDot: HTMLElement;
  private stateLine: HTMLElement;
  private newJobBtn: HTMLButtonElement;
  private formWrap: HTMLElement;

  private jobs: Job[] = [];
  private activeJobId: number | null = null;
  private recording = false;
  private openStartedAtUs: number | null = null;
  // Clock offset: server unix-micros minus client Date.now()*1000, so the elapsed
  // timer is correct even if the laptop clock differs from the Pi's.
  private clockSkewUs = 0;

  private onActiveJob: ActiveJobHandler;

  constructor(host: HTMLElement, onActiveJob: ActiveJobHandler = () => {}) {
    this.onActiveJob = onActiveJob;
    const strip = el("div", "controls");

    // Active-job selector.
    const jobField = el("label", "control-field");
    jobField.append(el("span", "control-label", "Job"));
    this.jobSelect = el("select", "job-select") as HTMLSelectElement;
    this.jobSelect.addEventListener("change", () => this.onJobSelect());
    jobField.append(this.jobSelect);

    // New-job toggle.
    this.newJobBtn = el("button", "ghost-btn", "+ New job…") as HTMLButtonElement;
    this.newJobBtn.type = "button";
    this.newJobBtn.addEventListener("click", () => this.toggleForm());

    // Record button.
    this.recordBtn = el("button", "record-btn") as HTMLButtonElement;
    this.recordBtn.type = "button";
    this.recordDot = el("span", "record-dot");
    const recordLabel = el("span", "record-label", "Record");
    this.recordBtn.append(this.recordDot, recordLabel);
    this.recordBtn.addEventListener("click", () => this.onRecordClick());

    // State line.
    this.stateLine = el("span", "control-state", "");

    strip.append(jobField, this.newJobBtn, this.recordBtn, this.stateLine);

    // Inline new-job form (hidden until toggled).
    this.formWrap = this.buildForm();
    this.formWrap.hidden = true;

    host.append(strip, this.formWrap);

    void this.init();
  }

  private async init(): Promise<void> {
    await this.refreshJobs();
    await this.refreshActiveJob();
    await this.refreshState();
    // Poll record state so multiple clients converge; refresh after each action too.
    window.setInterval(() => void this.refreshState(), STATE_POLL_MS);
    // Tick the elapsed timer once a second when recording.
    window.setInterval(() => this.renderRecordButton(), 1000);
  }

  // --- data ----------------------------------------------------------------

  private async refreshJobs(): Promise<void> {
    try {
      this.jobs = await getJSON<Job[]>("/api/jobs");
    } catch {
      this.jobs = [];
    }
    this.renderJobOptions();
  }

  private async refreshActiveJob(): Promise<void> {
    let next: number | null = null;
    try {
      const r = await fetch("/api/job/active", { headers: { Accept: "application/json" } });
      if (r.ok) {
        const data = (await r.json()) as Job | { active: null };
        next = "id" in data ? data.id : null;
      }
    } catch {
      next = null;
    }
    this.setActiveJobId(next);
    this.renderJobOptions();
  }

  // setActiveJobId records the active job and notifies the shell only on a real change.
  private setActiveJobId(id: number | null): void {
    if (this.activeJobId === id) return;
    this.activeJobId = id;
    this.onActiveJob(id);
  }

  private async refreshState(): Promise<void> {
    let st: RecordingState;
    try {
      st = await getJSON<RecordingState>("/api/recording/state");
    } catch {
      return; // leave the last known state; the readout is unaffected (axiom #1)
    }
    this.recording = st.recording;
    if (st.recording && st.openSegmentId) {
      // Fetch the open segment's started_at for the elapsed timer (and clock skew).
      if (st.jobId) await this.syncOpenSegment(st.jobId, st.openSegmentId);
    } else {
      this.openStartedAtUs = null;
    }
    this.renderRecordButton();
    this.renderState();
  }

  // syncOpenSegment loads the open segment to read its started_at_us and calibrate
  // the client→server clock skew so elapsed time is accurate.
  private async syncOpenSegment(jobId: number, segId: number): Promise<void> {
    try {
      const segs = await getJSON<Segment[]>(`/api/recording/segments?job_id=${jobId}`);
      const open = segs.find((s) => s.id === segId);
      if (open) {
        this.openStartedAtUs = open.startedAtUs;
        this.clockSkewUs = open.startedAtUs - Date.now() * 1000;
      }
    } catch {
      // Non-fatal: the button still shows "Stop" without an exact elapsed.
    }
  }

  // --- actions -------------------------------------------------------------

  private async onJobSelect(): Promise<void> {
    const val = this.jobSelect.value;
    if (val === "__new__") {
      this.jobSelect.value = this.activeJobId != null ? String(this.activeJobId) : "";
      this.showForm(true);
      return;
    }
    const id = Number(val);
    if (!id || id === this.activeJobId) return;
    const res = await sendJSON<unknown>("PUT", "/api/job/active", { id });
    if (!res.ok) {
      this.flashState(
        res.status === 409
          ? "Stop recording before switching jobs"
          : "Could not switch job",
      );
      // Revert the select to the actual active job.
      this.jobSelect.value = this.activeJobId != null ? String(this.activeJobId) : "";
      return;
    }
    this.setActiveJobId(id);
    this.renderState();
  }

  private async onRecordClick(): Promise<void> {
    if (this.recording) {
      const res = await sendJSON<Segment>("POST", "/api/recording/stop");
      if (!res.ok && res.status !== 409) {
        this.flashState("Stop failed");
        return;
      }
    } else {
      const res = await sendJSON<Segment>("POST", "/api/recording/start");
      if (!res.ok) {
        this.flashState(
          res.status === 400 ? "Select a job before recording" : "Start failed",
        );
        return;
      }
    }
    await this.refreshState();
  }

  // --- new-job form --------------------------------------------------------

  private nameInput!: HTMLInputElement;
  private fieldInputs: Record<string, HTMLInputElement> = {};

  private buildForm(): HTMLElement {
    const wrap = el("form", "newjob-form");
    const fields: Array<[string, string]> = [
      ["name", "Name *"],
      ["company", "Company"],
      ["well", "Well"],
      ["casingSize", "Casing size"],
      ["jobType", "Job type"],
      ["location", "Location"],
      ["cementer", "Cementer"],
      ["notes", "Notes"],
    ];
    for (const [key, label] of fields) {
      const f = el("label", "newjob-field");
      f.append(el("span", "newjob-label", label));
      const input = el("input", "newjob-input") as HTMLInputElement;
      input.type = "text";
      input.name = key;
      f.append(input);
      this.fieldInputs[key] = input;
      if (key === "name") this.nameInput = input;
      wrap.append(f);
    }

    const actions = el("div", "newjob-actions");
    const save = el("button", "primary-btn", "Create + activate") as HTMLButtonElement;
    save.type = "submit";
    const cancel = el("button", "ghost-btn", "Cancel") as HTMLButtonElement;
    cancel.type = "button";
    cancel.addEventListener("click", () => this.showForm(false));
    actions.append(save, cancel);
    wrap.append(actions);

    wrap.addEventListener("submit", (e) => {
      e.preventDefault();
      void this.onCreateJob();
    });
    return wrap;
  }

  private toggleForm(): void {
    this.showForm(this.formWrap.hidden === true);
  }

  private showForm(show: boolean): void {
    this.formWrap.hidden = !show;
    if (show) this.nameInput.focus();
  }

  private async onCreateJob(): Promise<void> {
    const name = this.nameInput.value.trim();
    if (!name) {
      this.flashState("Job name is required");
      this.nameInput.focus();
      return;
    }
    const body: JobInput = { name };
    const extra = body as unknown as Record<string, string>;
    for (const key of ["company", "well", "casingSize", "jobType", "location", "cementer", "notes"]) {
      const v = this.fieldInputs[key]?.value.trim();
      if (v) extra[key] = v;
    }
    const res = await sendJSON<Job>("POST", "/api/jobs", body);
    if (!res.ok || !res.body) {
      this.flashState("Could not create job");
      return;
    }
    const created = res.body;
    // Make it active.
    await sendJSON<unknown>("PUT", "/api/job/active", { id: created.id });
    this.setActiveJobId(created.id);
    // Reset + hide the form, refresh the list.
    for (const input of Object.values(this.fieldInputs)) input.value = "";
    this.showForm(false);
    await this.refreshJobs();
    this.renderJobOptions();
    this.renderState();
  }

  // --- render --------------------------------------------------------------

  private renderJobOptions(): void {
    const prev = this.jobSelect.value;
    this.jobSelect.replaceChildren();

    if (this.jobs.length === 0) {
      const opt = el("option", undefined, "No jobs yet") as HTMLOptionElement;
      opt.value = "";
      opt.disabled = true;
      this.jobSelect.append(opt);
    }
    for (const j of this.jobs) {
      const opt = el("option", undefined, j.name) as HTMLOptionElement;
      opt.value = String(j.id);
      this.jobSelect.append(opt);
    }
    const newOpt = el("option", undefined, "+ New job…") as HTMLOptionElement;
    newOpt.value = "__new__";
    this.jobSelect.append(newOpt);

    if (this.activeJobId != null && this.jobs.some((j) => j.id === this.activeJobId)) {
      this.jobSelect.value = String(this.activeJobId);
    } else if (prev && this.jobs.some((j) => String(j.id) === prev)) {
      this.jobSelect.value = prev;
    }
  }

  private renderRecordButton(): void {
    const label = this.recordBtn.querySelector(".record-label");
    if (this.recording) {
      this.recordBtn.classList.add("recording");
      this.recordDot.textContent = "■";
      let text = "Stop";
      if (this.openStartedAtUs != null) {
        const nowUs = Date.now() * 1000 + this.clockSkewUs;
        text = `Stop (${fmtElapsed((nowUs - this.openStartedAtUs) / 1000)})`;
      }
      if (label) label.textContent = text;
    } else {
      this.recordBtn.classList.remove("recording");
      this.recordDot.textContent = "●";
      if (label) label.textContent = "Record";
    }
  }

  private renderState(): void {
    const active = this.jobs.find((j) => j.id === this.activeJobId);
    if (this.recording) {
      this.stateLine.textContent = active ? `Recording — ${active.name}` : "Recording";
      this.stateLine.dataset.kind = "recording";
    } else if (active) {
      this.stateLine.textContent = `Ready — ${active.name}`;
      this.stateLine.dataset.kind = "ready";
    } else {
      this.stateLine.textContent = "Select or create a job to record";
      this.stateLine.dataset.kind = "idle";
    }
  }

  private flashTimer: number | undefined;
  private flashState(msg: string): void {
    this.stateLine.textContent = msg;
    this.stateLine.dataset.kind = "warn";
    if (this.flashTimer) window.clearTimeout(this.flashTimer);
    this.flashTimer = window.setTimeout(() => this.renderState(), 4000);
  }
}
