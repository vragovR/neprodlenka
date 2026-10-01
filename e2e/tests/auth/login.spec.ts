import { test, expect } from '../../fixtures/backend.fixture'
import { ChildrenPage } from '../../pages/children.page'
import { Header } from '../../pages/header.page'
import { LoginPage } from '../../pages/login.page'

test.describe('Login', () => {
  test.use({ signedIn: false })

  test('expired session shows login form with saved login', async ({ page }) => {
    const loginPage = new LoginPage(page)
    await page.addInitScript(() => localStorage.setItem('login', 'admin'))

    await new ChildrenPage(page).goto()

    await expect(loginPage.submitButton).toBeVisible()
    await expect(loginPage.loginInput).toHaveValue('admin')
  })

  test('invalid credentials shows error message', async ({ page }) => {
    const loginPage = new LoginPage(page)

    await new ChildrenPage(page).goto()
    await loginPage.login('admin', 'wrong-password')

    await expect(loginPage.errorMessage).toHaveText('Неверный логин или пароль')
  })

  test('locks the login after five failed attempts', async ({ page }) => {
    const loginPage = new LoginPage(page)
    await new ChildrenPage(page).goto()

    for (let i = 0; i < 5; i++) {
      await loginPage.login('admin', 'wrong-password')
      await expect(loginPage.errorMessage).toHaveText('Неверный логин или пароль')
      await loginPage.errorMessage.evaluate((el) => (el.textContent = ''))
    }
    await loginPage.login('admin', 'admin-password')

    await expect(loginPage.errorMessage).toHaveText('Слишком много попыток, попробуйте позже')
  })

  test('successful login shows children', async ({ page }) => {
    const loginPage = new LoginPage(page)
    const childrenPage = new ChildrenPage(page)

    await childrenPage.goto()
    await loginPage.login(' admin ', 'admin-password')

    await expect(childrenPage.heading).toBeVisible()
    expect(await page.evaluate(() => localStorage.getItem('login'))).toBe('admin')

    await page.reload()
    await expect(childrenPage.heading).toBeVisible()
  })
})

test.describe('Logout', () => {
  test('logout returns to login form', async ({ page }) => {
    const loginPage = new LoginPage(page)
    await new ChildrenPage(page).goto()

    await new Header(page).logout()

    await expect(loginPage.submitButton).toBeVisible()
    await expect(loginPage.loginInput).toHaveValue('')

    await page.goto('/')
    await expect(loginPage.submitButton).toBeVisible()
  })
})
