import eslintPluginVue from 'eslint-plugin-vue'
import { createConfig } from '@vue/eslint-config-typescript'

export default [
  {
    ignores: ['dist/**', 'coverage/**', 'node_modules/**'],
  },
  ...eslintPluginVue.configs['flat/recommended'],
  ...createConfig({
    extends: ['recommended'],
  }),
  {
    rules: {
      'no-console': ['warn', { allow: ['warn', 'error'] }],
      '@typescript-eslint/no-explicit-any': 'error',
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
    },
  },
]
