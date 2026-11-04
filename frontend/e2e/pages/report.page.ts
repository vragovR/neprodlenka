import type { Locator, Page } from '@playwright/test'

export class ReportPage {
  readonly heading: Locator

  constructor(readonly page: Page) {
    this.heading = page.getByRole('heading', { level: 1 })
  }

  async goto(uid: string): Promise<void> {
    await this.page.addInitScript(() => localStorage.setItem('view', 'table'))
    await this.page.goto(`/report/${uid}`)
  }
}
