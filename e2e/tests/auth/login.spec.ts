import { test, expect } from '../../fixtures/api.fixture'
import { childrenResponse } from '../../mocks/fixtures/children'
import { invalidCredentials } from '../../mocks/fixtures/errors'
import { ChildrenPage } from '../../pages/children.page'
import { Header } from '../../pages/header.page'
import { LoginPage } from '../../pages/login.page'

test.describe('Login', () => {
  test('expired session shows login form with saved login', async ({ page, api }) => {
    const loginPage = new LoginPage(page)

    await api.mockUnauthorized()
    await new ChildrenPage(page).goto()

    await expect(loginPage.submitButton).toBeVisible()
    await expect(loginPage.loginInput).toHaveValue('admin')
  })

  test('invalid credentials shows error message', async ({ page, api }) => {
    const loginPage = new LoginPage(page)

    await api.mockUnauthorized()
    await api.mockLoginError(invalidCredentials)
    await new ChildrenPage(page).goto()
    await loginPage.login('admin', 'wrong')

    await expect(loginPage.errorMessage).toHaveText('Неверный логин или пароль')
  })

  test('successful login shows children', async ({ page, api }) => {
    const loginPage = new LoginPage(page)
    const childrenPage = new ChildrenPage(page)

    await api.mockUnauthorized()
    await api.mockLoginSuccess('teacher')
    await childrenPage.goto()
    await expect(loginPage.submitButton).toBeVisible()

    await api.mockChildren(childrenResponse)
    const request = page.waitForRequest('/api/login')
    await loginPage.login(' teacher ', 'secret')

    expect((await request).postDataJSON()).toEqual({ login: 'teacher', password: 'secret' })
    await expect(childrenPage.heading).toBeVisible()
    expect(await page.evaluate(() => localStorage.getItem('login'))).toBe('teacher')
  })
})

test.describe('Logout', () => {
  test('logout returns to login form', async ({ page, api }) => {
    await api.mockChildren(childrenResponse)
    await api.mockLogout()
    await new ChildrenPage(page).goto()

    await new Header(page).logout()

    await expect(new LoginPage(page).submitButton).toBeVisible()
    await expect(new LoginPage(page).loginInput).toHaveValue('')
  })
})
