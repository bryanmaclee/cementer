// Personal live-view config (chart-config scope #1 in data-model.md): which lines are
// on/off, the rolling-window length, and optional per-channel color overrides. This is
// a PERSONAL preference stored per-laptop in localStorage (axiom #3 — pump/job data
// lives on the Pi, never here). NO state library — plain read/write of a JSON blob.

export interface LiveConfig {
  // hidden[channelId] === true => that trace is off in the live chart.
  hidden?: Record<string, boolean>;
  // colors[channelId] => a CSS color override for that trace.
  colors?: Record<string, string>;
  // rolling window length, in SECONDS (uPlot's time-scale unit; kept consistent
  // across the chart, the persisted config, and the window selector).
  windowSec?: number;
}

const KEY = "cementer.liveview";

export function loadLiveConfig(): LiveConfig {
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return {};
    const parsed = JSON.parse(raw) as LiveConfig;
    return parsed && typeof parsed === "object" ? parsed : {};
  } catch {
    return {};
  }
}

export function saveLiveConfig(cfg: LiveConfig): void {
  try {
    localStorage.setItem(KEY, JSON.stringify(cfg));
  } catch {
    // localStorage may be unavailable (private mode); the chart still works in-memory.
  }
}

// setHidden persists a single trace's on/off state and returns the updated config.
export function setHidden(channelId: string, hidden: boolean): LiveConfig {
  const cfg = loadLiveConfig();
  cfg.hidden = cfg.hidden ?? {};
  if (hidden) cfg.hidden[channelId] = true;
  else delete cfg.hidden[channelId];
  saveLiveConfig(cfg);
  return cfg;
}

// setWindowSec persists the rolling-window length (seconds) and returns the updated config.
export function setWindowSec(sec: number): LiveConfig {
  const cfg = loadLiveConfig();
  cfg.windowSec = sec;
  saveLiveConfig(cfg);
  return cfg;
}
