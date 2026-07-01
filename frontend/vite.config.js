import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0', // expose inside Docker
    port: 5173,
    proxy: {
      // Forward API calls to the backend service.
      // Uses the docker-compose service name; override with VITE_API_TARGET locally.
      '/api': {
        target: process.env.VITE_API_TARGET || 'http://backend:8080',
        changeOrigin: true,
      },
    },
  },
})
