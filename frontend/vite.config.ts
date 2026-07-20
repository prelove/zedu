import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'node:path'

const spaRoutes = [
  '/',
  '/login',
  '/dashboard',
  '/onboarding',
  '/students',
  '/teachers',
  '/courses',
  '/finance/config',
  '/finance/payments',
  '/lessons',
  '/notifications',
  '/enrollments',
] as const

/**
 * API paths intentionally share names with client routes. A document
 * navigation must therefore reach Vite's SPA fallback, while fetch requests
 * (whose Accept normally permits any media type or requests JSON) must proxy.
 */
export function isSpaNavigationRequest(url: string | undefined, accept: string | undefined): boolean {
  if (!accept?.includes('text/html')) {
    return false
  }

  const pathname = (url ?? '/').split('?', 1)[0]
  return spaRoutes.some((route) => {
    if (route === '/') {
      return pathname === route
    }
    return pathname === route || pathname.startsWith(`${route}/`)
  })
}

export default defineConfig({
  plugins: [
    vue(),
    {
      name: 'zedu-spa-navigation-before-api-proxy',
      configureServer(server) {
        server.middlewares.use((req, _res, next) => {
          if (isSpaNavigationRequest(req.url, req.headers.accept)) {
            req.url = '/index.html'
          }
          next()
        })
      },
    },
  ],
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
      '/lessons': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/notifications': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/dashboard': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
