import { test, expect } from '../../fixtures/backend.fixture'
import { petya } from '../../data/children'
import { CalendarPage } from '../../pages/calendar.page'
import { Header } from '../../pages/header.page'
import { MonthPage } from '../../pages/month.page'

const november = '2026-11'

test.describe('Calendar', () => {
  test.beforeEach(async ({ backend }) => {
    await backend.addChild(petya)
  })

  test('an unfilled month links to the calendar', async ({ page }) => {
    const monthPage = new MonthPage(page)
    await monthPage.goto(november)

    await expect(page.getByText('Календарь месяца не заполнен')).toBeVisible()
    await page.getByRole('button', { name: 'Открыть календарь' }).click()
    await expect(page).toHaveURL(`/calendar/${november}`)
  })

  test('saves no-school days and the month follows them', async ({ page }) => {
    const calendarPage = new CalendarPage(page)
    await calendarPage.goto(november)

    await expect(calendarPage.saveButton).toBeDisabled()
    await calendarPage.weekendsButton.click()
    await expect(calendarPage.day(1)).toHaveClass(/off/)
    await expect(calendarPage.day(7)).toHaveClass(/off/)
    await expect(calendarPage.day(2)).not.toHaveClass(/off/)
    await calendarPage.day(4).click()
    await expect(calendarPage.day(4)).toHaveClass(/off/)
    await calendarPage.saveButton.click()
    await expect(page.getByText('Календарь сохранён')).toBeVisible()
    await expect(calendarPage.saveButton).toBeDisabled()

    await page.reload()
    await expect(calendarPage.day(4)).toHaveClass(/off/)
    await expect(calendarPage.day(8)).toHaveClass(/off/)

    const monthPage = new MonthPage(page)
    await monthPage.goto(november)
    await expect(page.getByRole('columnheader').nth(4)).toHaveAttribute('title', 'Неучебный день')
    await expect(monthPage.cell('Иванов', 4)).toBeDisabled()
    await expect(monthPage.cell('Иванов', 5)).toBeEnabled()
  })

  test('asks before leaving with unsaved days', async ({ page }) => {
    const calendarPage = new CalendarPage(page)
    await calendarPage.goto(november)
    await calendarPage.day(4).click()

    const messages: string[] = []
    page.once('dialog', async (dialog) => {
      messages.push(dialog.message())
      await dialog.dismiss()
    })
    await new Header(page).tab('Дети')

    expect(messages).toEqual(['Изменения в календаре не сохранены. Уйти?'])
    await expect(page).toHaveURL(`/calendar/${november}`)
    await expect(calendarPage.day(4)).toHaveClass(/off/)
  })

  test('shows the month report link', async ({ page, backend }) => {
    const calendarPage = new CalendarPage(page)
    await calendarPage.goto(november)

    await expect(calendarPage.reportLink).toHaveValue(
      `http://localhost/report/${await backend.reportUID(november)}`,
    )
  })
})
