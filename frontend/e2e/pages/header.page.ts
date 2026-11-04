import type { Locator, Page } from '@playwright/test'

export class Header {
  readonly burger: Locator
  readonly logoutButton: Locator

  constructor(readonly page: Page) {
    this.burger = page.getByRole('button', { name: 'Меню' })
    this.logoutButton = page.getByRole('button', { name: 'Выйти' })
  }

  async openMenu(): Promise<void> {
    if (await this.burger.isVisible()) {
      await this.burger.click()
    }
  }

  async tab(name: string): Promise<void> {
    await this.openMenu()
    await this.page.getByRole('link', { name, exact: true }).click()
  }

  async logout(): Promise<void> {
    await this.openMenu()
    await this.logoutButton.click()
  }
}
