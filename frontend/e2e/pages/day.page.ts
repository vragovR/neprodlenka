import type { Locator, Page } from '@playwright/test'

export class DayPage {
  readonly heading: Locator
  readonly messageButton: Locator
  readonly dialog: Locator

  constructor(readonly page: Page) {
    this.heading = page.getByRole('heading', { name: 'День', level: 1 })
    this.messageButton = page.getByRole('button', { name: 'Сообщение дня' })
    this.dialog = page.getByRole('dialog')
  }

  async goto(date: string): Promise<void> {
    await this.page.goto(`/day/${date}`)
  }

  card(name: string): Locator {
    return this.page.getByRole('article').filter({ hasText: name })
  }
}
