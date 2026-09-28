import type { Locator, Page } from '@playwright/test'

export class MonthPage {
  readonly heading: Locator
  readonly saveButton: Locator

  constructor(readonly page: Page) {
    this.heading = page.getByRole('heading', { name: 'Месяц', level: 1 })
    this.saveButton = page.getByRole('button', { name: 'Сохранить' })
  }

  async goto(month: string): Promise<void> {
    await this.page.addInitScript(() => localStorage.setItem('view', 'table'))
    await this.page.goto(`/month/${month}`)
  }

  cell(name: string, day: number): Locator {
    return this.page
      .getByRole('row')
      .filter({ has: this.page.getByRole('rowheader', { name }) })
      .getByRole('cell')
      .nth(day - 1)
      .getByRole('button')
  }

  pay(name: string): Locator {
    return this.page
      .getByRole('row')
      .filter({ has: this.page.getByRole('rowheader', { name }) })
      .getByRole('cell')
      .last()
  }
}
