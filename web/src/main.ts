import { Controls } from "./controls.ts";
import { Readout } from "./readout.ts";
import { connectLive } from "./ws.ts";

const root = document.getElementById("app");
if (!root) throw new Error("missing #app");

const readout = new Readout(root);

// Job + recording controls (REST). They live in the strip the Readout reserves
// between its header and the live values. Recording is a marker over the always-on
// store and never gates the live readout (axiom #1) — the two are wired independently.
new Controls(readout.controlsHost());

connectLive(
  (reading) => readout.update(reading),
  (connected) => readout.setConnected(connected),
  (profile) => readout.applyProfile(profile),
);
