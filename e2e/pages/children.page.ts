import type { Locator, Page } from '@playwright/test'

export class ChildrenPage {
  readonly heading: Locator
  readonly addButton: Locator
  readonly dialog: Locator
  readonly inactive: Locator

  constructor(readonly page: Page) {
    this.heading = page.getByRole('heading', { name: 'Дети', level: 1 })
    this.addButton = page.getByRole('button', { name: 'Добавить ребёнка' })
    this.dialog = page.getByRole('dialog')
    this.inactive = page.locator('details')
  }

  async goto(): Promise<void> {
    await this.page.goto('/')
  }

  card(name: string): Locator {
    return this.page.getByRole('article').filter({ has: this.page.getByRole('button', { name, exact: true }) })
  }

  names(): Locator {
    return this.page.getByRole('article').getByTitle('Изменить ФИО')
  }
}
