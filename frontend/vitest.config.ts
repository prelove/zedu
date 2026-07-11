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
        include: ['src/**/*.ts'],
        exclude: ['src/**/*.d.ts', 'src/main.ts'],
        thresholds: {
          lines: 80,
          statements: 80,
          branches: 80,
          functions: 80,
        },
      },
    },
  })
)
