import { fileURLToPath, URL } from 'node:url'

import { configDefaults, defineConfig, lazyPlugins } from 'vite-plus'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

const apiTarget = process.env.API_PROXY_TARGET ?? 'http://localhost:8080'

export default defineConfig({
  fmt: {
    singleQuote: true,
    semi: false,
    overrides: [
      {
        files: ['**/*.vue'],
        options: {
          htmlWhitespaceSensitivity: 'strict',
          bracketSameLine: true,
          printWidth: 140,
        },
      },
    ],
  },
  lint: {
    options: {
      typeAware: true,
      typeCheck: false,
    },
    ignorePatterns: [
      'dist/',
      'dist-ssr/',
      'coverage/',
      'playwright-report/',
      'test-results/',
      'playwright/.cache/',
    ],
    plugins: ['eslint', 'typescript', 'unicorn', 'oxc', 'vue'],
    env: {
      browser: true,
    },
    categories: {
      correctness: 'error',
    },
    jsPlugins: [
      {
        name: 'vite-plus',
        specifier: 'vite-plus/oxlint-plugin',
      },
    ],
    rules: {
      'no-array-constructor': 'error',
      'no-var': 'error',
      'prefer-const': 'error',
      'prefer-rest-params': 'error',
      'prefer-spread': 'error',
      'typescript/ban-ts-comment': [
        'error',
        {
          'ts-expect-error': true,
          'ts-ignore': true,
          'ts-nocheck': true,
        },
      ],
      'typescript/no-empty-object-type': 'error',
      'typescript/no-explicit-any': 'error',
      'typescript/no-namespace': 'error',
      'typescript/no-require-imports': 'error',
      'typescript/no-unnecessary-type-constraint': 'error',
      'typescript/no-unsafe-function-type': 'error',
      'vite-plus/prefer-vite-plus-imports': 'error',
    },
    overrides: [
      {
        files: ['src/**/__tests__/*'],
        plugins: ['vitest'],
        rules: {
          'vitest/expect-expect': 'error',
          'vitest/no-commented-out-tests': 'error',
          'vitest/no-conditional-expect': 'error',
          'vitest/no-disabled-tests': 'warn',
          'vitest/no-focused-tests': 'error',
          'vitest/no-identical-title': 'error',
          'vitest/no-import-node-test': 'error',
          'vitest/no-interpolation-in-snapshots': 'error',
          'vitest/no-mocks-import': 'error',
          'vitest/no-standalone-expect': 'error',
          'vitest/no-unneeded-async-expect-function': 'error',
          'vitest/prefer-called-exactly-once-with': 'error',
          'vitest/require-local-test-context-for-concurrent-snapshots': 'error',
          'vitest/valid-describe-callback': 'error',
          'vitest/valid-expect': 'error',
          'vitest/valid-expect-in-promise': 'error',
          'vitest/valid-title': 'error',
        },
      },
    ],
  },
  plugins: lazyPlugins(() => [vue(), vueDevTools()]),
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    host: true,
    port: 5173,
    allowedHosts: ['.local', '.test'],
    proxy: {
      '/api': { target: apiTarget, changeOrigin: true },
    },
  },
  test: {
    exclude: [...configDefaults.exclude, 'e2e/**'],
    root: fileURLToPath(new URL('./', import.meta.url)),
  },
})
