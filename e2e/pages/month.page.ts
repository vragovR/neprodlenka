import type { Locator, Page } from '@playwright/test'

export class MonthPage {
  readonly heading: Locator
  readonly saveButton: Locator
  readonly messageButton: Locator
  readonly dialog: Locator

  constructor(readonly page: Page) {
    this.heading = page.getByRole('heading', { name: 'Месяц', level: 1 })
    this.saveButton = page.getByRole('button', { name: 'Сохранить' })
    this.messageButton = page.getByRole('button', { name: 'Сообщение месяца' })
    this.dialog = page.getByRole('dialog')
  }

  async goto(month: string, view: 'table' | 'calendar' = 'table'): Promise<void> {
    await this.page.addInitScript((v) => localStorage.setItem('view', v), view)
    await this.page.goto(`/month/${month}`)
  }

  async save(): Promise<void> {
    const reloaded = this.page.waitForResponse(
      (res) => res.request().method() === 'GET' && new URL(res.url()).pathname.startsWith('/api/attendance/'),
    )
    await this.saveButton.click()
    await reloaded
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

  paid(name: string): Locator {
    return this.pay(name).getByRole('button')
  }

  calendarPaid(name: string): Locator {
    return this.page.getByRole('article').filter({ hasText: name }).getByRole('button', { name: /оплачено/ })
  }
}
