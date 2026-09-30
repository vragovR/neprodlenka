import { test, expect } from '../../fixtures/api.fixture'
import { attendance } from '../../mocks/fixtures/attendance'
import { month } from '../../mocks/fixtures/dates'
import { MonthPage } from '../../pages/month.page'

test.describe('Month', () => {
  test.beforeEach(async ({ api }) => {
    await api.mockAttendance(attendance)
  })

  test('shows marks and totals', async ({ page }) => {
    const monthPage = new MonthPage(page)
    await monthPage.goto(month)

    await expect(monthPage.cell('Иванов', 1).getByTitle('Обедает')).toBeVisible()
    await expect(monthPage.cell('Иванов', 10).getByTitle('Не обедает')).toBeVisible()
    await expect(monthPage.cell('Иванов', 25).getByTitle('Не отметился')).toBeVisible()
    await expect(monthPage.cell('Иванов', 26)).toBeDisabled()
    await expect(monthPage.cell('Иванов', 28)).toBeDisabled()
    await expect(monthPage.pay('Иванов')).toContainText('2 210 ₽')
    await expect(monthPage.paid('Иванов')).toHaveText('не оплачено')
    await expect(monthPage.paid('Козлова')).toHaveText('✓ оплачено')
    await expect(monthPage.saveButton).toBeDisabled()
  })

  test('saves only changed marks', async ({ page, api }) => {
    const monthPage = new MonthPage(page)
    await api.mockPutAttendance(month)
    await monthPage.goto(month)

    await monthPage.cell('Иванов', 1).click()
    await monthPage.cell('Козлова', 2).click()
    await monthPage.cell('Козлова', 2).click()
    await monthPage.cell('Козлова', 2).click()
    await monthPage.cell('Иванов', 25).click()
    await expect(monthPage.cell('Иванов', 1).getByTitle('Не обедает')).toBeVisible()

    const request = page.waitForRequest(`/api/attendance/${month}`)
    await monthPage.saveButton.click()

    expect((await request).postDataJSON()).toEqual({
      marks: [
        { date: '2026-09-01', child_id: 1, status: 'absent' },
        { date: '2026-09-25', child_id: 1, status: 'present' },
      ],
      payments: [],
    })
    await expect(page.getByText('Отметки сохранены')).toBeVisible()
  })

  test('saves only changed payments', async ({ page, api }) => {
    const monthPage = new MonthPage(page)
    await api.mockPutAttendance(month)
    await monthPage.goto(month)

    await monthPage.paid('Иванов').click()
    await monthPage.paid('Козлова').click()
    await monthPage.paid('Козлова').click()
    await expect(monthPage.paid('Иванов')).toHaveText('✓ оплачено')

    const request = page.waitForRequest(`/api/attendance/${month}`)
    await monthPage.saveButton.click()

    expect((await request).postDataJSON()).toEqual({ marks: [], payments: [{ child_id: 1, paid: true }] })
  })

  test('shows the month message', async ({ page }) => {
    const monthPage = new MonthPage(page)
    await monthPage.goto(month)

    await page.getByRole('button', { name: 'Сообщение месяца' }).click()

    await expect(page.getByRole('dialog')).toContainText('Сообщение ещё не отправлено')
    await expect(page.getByRole('dialog')).toContainText('Питание за сентябрь 2026.')
  })
})
