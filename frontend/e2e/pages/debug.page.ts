import type { Locator, Page } from '@playwright/test'

export class DebugPage {
  readonly dayCard: Locator
  readonly monthCard: Locator
  readonly closeCard: Locator

  constructor(readonly page: Page) {
    this.dayCard = this.card('Сообщение за день')
    this.monthCard = this.card('Итоги месяца')
    this.closeCard = this.card('Закрыть день')
  }

  async goto(): Promise<void> {
    await this.page.addInitScript(() => localStorage.setItem('debug', '1'))
    await this.page.goto('/debug')
  }

  private card(title: string): Locator {
    return this.page
      .locator('section')
      .filter({ has: this.page.getByRole('heading', { name: title }) })
  }

  day(card: Locator, day: number): Locator {
    return card.getByRole('button', { name: String(day), exact: true })
  }
}
