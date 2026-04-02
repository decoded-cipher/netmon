import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

// Dev-only: forward /api/* to the Go server (default :8080). The Vite app alone
// does not implement these routes — run `make dev` from the repo root, or start
// `go run ./cmd/netmon` in another terminal before `npm run dev`.
const apiTarget = process.env.VITE_PROXY_TARGET ?? 'http://127.0.0.1:8080'

export default defineConfig({
  plugins: [tailwindcss(), vue()],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks: { apexcharts: ['apexcharts'] },
      },
    },
  },
  server: {
    proxy: {
      '/api': {
        target: apiTarget,
        changeOrigin: true,
      },
    },
  },
})
