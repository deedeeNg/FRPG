import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Local dev hits the backend on localhost; docker-compose overrides this with
// VITE_API_TARGET=http://backend:8080.
const target = process.env.VITE_API_TARGET || 'http://localhost:8080'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0', // expose inside Docker
    port: 5173,
    proxy: {
      // Forward backend calls so they're same-origin (no CORS needed in dev).
      '/api': { target, changeOrigin: true },
      '/auth': { target, changeOrigin: true },
      '/signup': { target, changeOrigin: true },
    },
  },
})
