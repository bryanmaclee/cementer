import { Readout } from "./readout.ts";
import { connectLive } from "./ws.ts";

const root = document.getElementById("app");
if (!root) throw new Error("missing #app");

const readout = new Readout(root);

connectLive(
  (reading) => readout.update(reading),
  (connected) => readout.setConnected(connected),
);
