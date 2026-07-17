import { defineConfig, mergeConfig } from 'vitest/config'
import viteConfig from './vite.config'

export default mergeConfig(
  viteConfig,
  defineConfig({
    test: {
      globals: true,
      environment: 'jsdom',
      coverage: {
        provider: 'v8',
        reporter: ['text', 'json', 'html'],
        include: ['src/**/*.ts', 'src/**/*.vue'],
        exclude: [
          'src/**/*.d.ts',
          'src/main.ts',
          'src/App.vue',
          'src/components/LocaleSwitcher.vue',
          'src/router/index.ts',
          'src/api/types.ts',
          'src/api/error-mapping.ts',
          'src/features/auth/HomeView.vue',
        ],
        thresholds: {
          lines: 80,
          statements: 80,
          branches: 75,
          functions: 80,
        },
      },
    },
  })
)
