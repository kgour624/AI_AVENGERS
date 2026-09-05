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
