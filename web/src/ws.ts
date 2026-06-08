import type { Reading, WSEnvelope } from "./types.ts";

export type ReadingHandler = (r: Reading) => void;
export type StatusHandler = (connected: boolean) => void;

// connectLive opens the /ws/live WebSocket and auto-reconnects with capped
// exponential backoff. Returns a function that closes the connection for good.
export function connectLive(
  onReading: ReadingHandler,
  onStatus: StatusHandler,
): () => void {
  const proto = location.protocol === "https:" ? "wss://" : "ws://";
  const url = `${proto}${location.host}/ws/live`;

  let ws: WebSocket | null = null;
  let retry = 0;
  let stopped = false;
  let reconnectTimer: number | undefined;

  const open = () => {
    ws = new WebSocket(url);

    ws.onopen = () => {
      retry = 0;
      onStatus(true);
    };

    ws.onmessage = (ev) => {
      try {
        const env = JSON.parse(ev.data as string) as WSEnvelope;
        if (env.type === "reading" && env.reading) onReading(env.reading);
      } catch {
        // Ignore malformed frames.
      }
    };

    ws.onclose = () => {
      onStatus(false);
      if (!stopped) scheduleReconnect();
    };

    ws.onerror = () => {
      ws?.close();
    };
  };

  const scheduleReconnect = () => {
    const delay = Math.min(1000 * 2 ** retry, 10000);
    retry += 1;
    reconnectTimer = window.setTimeout(open, delay);
  };

  open();

  return () => {
    stopped = true;
    if (reconnectTimer) window.clearTimeout(reconnectTimer);
    ws?.close();
  };
}
