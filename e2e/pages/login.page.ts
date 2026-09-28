import type { Locator, Page } from '@playwright/test'

export class LoginPage {
  readonly loginInput: Locator
  readonly passwordInput: Locator
  readonly submitButton: Locator
  readonly errorMessage: Locator

  constructor(readonly page: Page) {
    this.loginInput = page.getByLabel('Логин')
    this.passwordInput = page.getByLabel('Пароль')
    this.submitButton = page.getByRole('button', { name: 'Войти' })
    this.errorMessage = page.getByRole('alert')
  }

  async login(login: string, password: string): Promise<void> {
    await this.loginInput.fill(login)
    await this.passwordInput.fill(password)
    await this.submitButton.click()
  }
}
