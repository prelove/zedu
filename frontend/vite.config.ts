import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'node:path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/healthz': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/auth': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/onboarding': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/students': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/teachers': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/course-domains': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/tracks': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/levels': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/capability-tags': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/enrollments': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/assignments': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/system': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/finance': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
