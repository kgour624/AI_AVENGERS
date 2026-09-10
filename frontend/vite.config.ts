import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'node:path'

// WHY Vite: SPA, no SSR needed, fast HMR (FRONTEND_SYSTEM_DESIGN.md section 2)
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3000,
    // 2026-09-09 fix (round 12): this dev server runs inside Docker
    // with a host:container port REMAP (docker-compose.yml maps
    // 3001:3000 for the frontend service) - the browser connects to
    // 192.168.1.7:3001, but the container itself listens on 3000.
    // Vite's HMR client, by default, tells the browser to open its
    // WebSocket back to the SAME port the dev server is configured
    // for (3000) - not the host-remapped port (3001) the browser
    // actually used to load the page. That mismatch is exactly
    // "[vite] failed to connect to websocket (ws://<host>:3001)" -
    // the browser tries 3001 (correct, matches its own URL) but the
    // injected HMR client script was built assuming 3000. Setting
    // clientPort explicitly (overridable via env for non-Docker/
    // direct-port setups, where host port == container port and this
    // env var is simply unset) tells Vite to inject the REAL
    // browser-facing port into the HMR client script.
    hmr: {
      clientPort: process.env.VITE_HMR_CLIENT_PORT
        ? Number(process.env.VITE_HMR_CLIENT_PORT)
        : undefined,
    },
    proxy: {
      // Dev-only convenience proxy; production nginx handles this (see /nginx)
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    sourcemap: true,
  },
})
