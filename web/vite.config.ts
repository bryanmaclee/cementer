import { defineConfig } from "vite";

// The build output goes to web/dist, which the Go binary embeds. During `npm run
// dev`, requests to the Go server (WebSocket + debug) are proxied so the client can
// run on Vite's dev server with hot reload while the Go backend runs on :8080.
export default defineConfig({
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
  server: {
    proxy: {
      "/ws": { target: "ws://localhost:8080", ws: true },
      "/debug": "http://localhost:8080",
    },
  },
});
