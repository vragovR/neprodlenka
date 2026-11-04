import process from 'node:process'
import { defineConfig, devices } from '@playwright/test'

const port = 5174

export default defineConfig({
  testDir: './e2e/tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  reporter: process.env.CI ? 'list' : 'html',
  use: {
    baseURL: `http://localhost:${port}`,
    locale: 'ru-RU',
    timezoneId: 'Europe/Moscow',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
    },
  ],
  globalSetup: './e2e/global-setup.ts',
  webServer: {
    command: `npm run build-only && npm run preview -- --port ${port} --strictPort`,
    url: `http://localhost:${port}`,
  },
})
