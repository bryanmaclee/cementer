import { Controls } from "./controls.ts";
import { Readout } from "./readout.ts";
import { connectLive } from "./ws.ts";

const root = document.getElementById("app");
if (!root) throw new Error("missing #app");

const readout = new Readout(root);

// Job + recording controls (REST). They live in the strip the Readout reserves
// between its header and the view area. Recording is a marker over the always-on store
// and never gates the live stream/chart (axiom #1) — the two are wired independently.
// The active-job callback lets the Job History view load the right job's recorded chart.
new Controls(readout.controlsHost(), (id) => readout.setActiveJob(id));

connectLive(
  (reading) => readout.update(reading),
  (connected) => readout.setConnected(connected),
  (profile) => readout.applyProfile(profile),
);
