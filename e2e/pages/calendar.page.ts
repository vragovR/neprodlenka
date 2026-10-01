import type { Locator, Page } from '@playwright/test'

export class CalendarPage {
  readonly saveButton: Locator
  readonly weekendsButton: Locator
  readonly reportLink: Locator

  constructor(readonly page: Page) {
    this.saveButton = page.getByRole('button', { name: 'Сохранить' })
    this.weekendsButton = page.getByRole('button', { name: 'Отметить все сб и вс' })
    this.reportLink = page.getByRole('textbox')
  }

  async goto(month: string): Promise<void> {
    await this.page.goto(`/calendar/${month}`)
  }

  day(day: number): Locator {
    return this.page.getByRole('button', { name: String(day), exact: true })
  }
}
