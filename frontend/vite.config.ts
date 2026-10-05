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
    host: '0.0.0.0',
    // 2026-09-09 fix (round 12): this dev server runs inside Docker
    // with a host:container port REMAP (docker-compose.yml maps
    // 3002:3000 for the frontend service) - the browser connects to
    // 192.168.1.7:3002, but the container itself listens on 3000.
    // Vite's HMR client, by default, tells the browser to open its
    // WebSocket back to the SAME port the dev server is configured
    // for (3000) - not the host-remapped port (3002) the browser
    // actually used to load the page. That mismatch is exactly
    // "[vite] failed to connect to websocket (ws://<host>:3002)" -
    // the browser tries 3002 (correct, matches its own URL) but the
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
      // FIX 2026-10-04: MUST be http://api:8080 inside Docker
      // BEFORE it was http://localhost:8080 which inside frontend container = itself -> ECONNREFUSED
      // api service is reachable as http://api:8080 on avengers-network
      '/api': {
        target: 'http://api:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    sourcemap: true,
  },
})

