import { test, expect } from '../../fixtures/backend.fixture'
import { anya, petya } from '../../data/children'
import { isSchool, month, septemberDates, weekends, today } from '../../data/dates'
import { Header } from '../../pages/header.page'
import { MonthPage } from '../../pages/month.page'

test.describe('Month', () => {
  test.beforeEach(async ({ backend }) => {
    await backend.setNoSchoolDays(month, weekends(month))
    await backend.addChild(petya)
    await backend.addChild(anya)
    for (const date of septemberDates.filter((d) => isSchool(d) && d <= today)) {
      await backend.publishDay(date)
      if (date < today) await backend.closeDay(date)
    }
    await backend.saveMarks(month, [
      { date: '2026-09-10', child_id: 1, status: 'absent' },
      { date: today, child_id: 2, status: 'absent' },
    ])
    await backend.savePayment(month, 2)
  })

  test('shows marks and totals', async ({ page }) => {
    const monthPage = new MonthPage(page)
    await monthPage.goto(month)

    await expect(monthPage.cell('Иванов', 1).getByTitle('Обедает')).toBeVisible()
    await expect(monthPage.cell('Иванов', 10).getByTitle('Не обедает')).toBeVisible()
    await expect(monthPage.cell('Иванов', 25).getByTitle('Не отметился')).toBeVisible()
    await expect(monthPage.cell('Иванов', 26)).toBeDisabled()
    await expect(monthPage.cell('Иванов', 28)).toBeEnabled()
    await expect(monthPage.pay('Иванов')).toContainText('2 210 ₽')
    await expect(monthPage.paid('Иванов')).toHaveText('не оплачено')
    await expect(monthPage.paid('Козлова')).toHaveText('✓ оплачено')
    await expect(monthPage.saveButton).toBeDisabled()
  })

  test('saves only changed marks', async ({ page }) => {
    const monthPage = new MonthPage(page)
    await monthPage.goto(month)

    await monthPage.cell('Иванов', 1).click()
    await monthPage.cell('Козлова', 2).click()
    await monthPage.cell('Козлова', 2).click()
    await monthPage.cell('Козлова', 2).click()
    await monthPage.cell('Иванов', 25).click()
    await monthPage.cell('Козлова', 29).click()
    await monthPage.cell('Козлова', 30).click()
    await monthPage.cell('Козлова', 30).click()
    await expect(monthPage.cell('Иванов', 1).getByTitle('Не обедает')).toBeVisible()
    await expect(monthPage.cell('Козлова', 29).getByTitle('Не обедает')).toBeVisible()
    await monthPage.saveButton.click()
    await expect(page.getByText('Отметки сохранены')).toBeVisible()

    await page.reload()
    await expect(monthPage.cell('Иванов', 1).getByTitle('Не обедает')).toBeVisible()
    await expect(monthPage.cell('Иванов', 25).getByTitle('Обедает')).toBeVisible()
    await expect(monthPage.cell('Козлова', 2).getByTitle('Обедает')).toBeVisible()
    await expect(monthPage.cell('Козлова', 29).getByTitle('Не обедает')).toBeVisible()
    await expect(monthPage.cell('Козлова', 30).getByTitle('Нет отметки')).toBeVisible()
    await expect(monthPage.pay('Иванов')).toContainText('2 210 ₽')
  })

  test('saves only changed payments', async ({ page }) => {
    const monthPage = new MonthPage(page)
    await monthPage.goto(month)

    await monthPage.paid('Иванов').click()
    await monthPage.paid('Козлова').click()
    await monthPage.paid('Козлова').click()
    await expect(monthPage.paid('Иванов')).toHaveText('✓ оплачено')
    await monthPage.saveButton.click()
    await expect(monthPage.saveButton).toBeDisabled()

    await page.reload()
    await expect(monthPage.paid('Иванов')).toHaveText('✓ оплачено')
    await expect(monthPage.paid('Козлова')).toHaveText('✓ оплачено')
  })

  test('dims days the child does not attend', async ({ page, backend }) => {
    const monthPage = new MonthPage(page)
    await backend.updateChild(2, { weekdays: [2, 4] })
    await monthPage.goto(month)

    await expect(monthPage.cell('Козлова', 29).getByTitle('Нет отметки')).toBeVisible()
    await expect(monthPage.cell('Козлова', 30).locator('..')).toHaveAttribute('title', 'Не ходит в этот день')
    await expect(monthPage.cell('Козлова', 30).getByTitle('Нет отметки')).toBeHidden()
    await expect(monthPage.cell('Иванов', 30).getByTitle('Нет отметки')).toBeVisible()

    await monthPage.cell('Козлова', 30).click()
    await expect(monthPage.cell('Козлова', 30).getByTitle('Не обедает')).toBeVisible()
  })

  test('past month without marks', async ({ page, backend }) => {
    const monthPage = new MonthPage(page)
    await backend.setNoSchoolDays('2026-08', weekends('2026-08'))
    await monthPage.goto('2026-08')

    await expect(page.getByText('Отметок нет')).toBeVisible()
  })

  test('shows the month message', async ({ page }) => {
    const monthPage = new MonthPage(page)
    await monthPage.goto(month)

    await page.getByRole('button', { name: 'Сообщение месяца' }).click()

    await expect(page.getByRole('dialog')).toContainText('Сообщение ещё не отправлено')
    await expect(page.getByRole('dialog')).toContainText('Питание за сентябрь 2026.')
  })

  test('payments are shown in the month message', async ({ page }) => {
    const monthPage = new MonthPage(page)
    await monthPage.goto(month)

    await monthPage.messageButton.click()
    await expect(monthPage.dialog).toContainText('🔴 Иванов Петя')
    await expect(monthPage.dialog).toContainText('🟢 Козлова Аня')
    await page.keyboard.press('Escape')

    await monthPage.paid('Иванов').click()
    await monthPage.paid('Козлова').click()
    await monthPage.saveButton.click()
    await expect(page.getByText('Отметки сохранены')).toBeVisible()

    await monthPage.messageButton.click()
    await expect(monthPage.dialog).toContainText('🟢 Иванов Петя')
    await expect(monthPage.dialog).toContainText('🔴 Козлова Аня')
  })

  test('an unsaved payment counts as an unsaved change', async ({ page }) => {
    const monthPage = new MonthPage(page)
    await monthPage.goto(month)

    await monthPage.paid('Иванов').click()
    await expect(monthPage.saveButton).toBeEnabled()
    await monthPage.paid('Иванов').click()
    await expect(monthPage.saveButton).toBeDisabled()

    await monthPage.paid('Иванов').click()
    await monthPage.messageButton.click()
    await expect(monthPage.dialog).toContainText('Несохранённые правки в нём не учтены.')
    await expect(monthPage.dialog).toContainText('🔴 Иванов Петя')
    await page.keyboard.press('Escape')

    const messages: string[] = []
    page.once('dialog', async (dialog) => {
      messages.push(dialog.message())
      await dialog.dismiss()
    })
    await new Header(page).tab('Дети')
    expect(messages).toEqual(['Отметки не сохранены. Уйти?'])
    await expect(page).toHaveURL(`/month/${month}`)
    await expect(monthPage.paid('Иванов')).toHaveText('✓ оплачено')
  })

  test('saves marks and payments together', async ({ page }) => {
    const monthPage = new MonthPage(page)
    await monthPage.goto(month)

    await monthPage.cell('Иванов', 1).click()
    await monthPage.paid('Иванов').click()
    await monthPage.saveButton.click()
    await expect(page.getByText('Отметки сохранены')).toBeVisible()

    await page.reload()
    await expect(monthPage.cell('Иванов', 1).getByTitle('Не обедает')).toBeVisible()
    await expect(monthPage.pay('Иванов')).toContainText('2 080 ₽')
    await expect(monthPage.paid('Иванов')).toHaveText('✓ оплачено')
  })

  test('payment belongs to its month', async ({ page, backend }) => {
    const monthPage = new MonthPage(page)
    await backend.setNoSchoolDays('2026-08', weekends('2026-08'))
    await backend.saveMarks('2026-08', [{ date: '2026-08-31', child_id: 2, status: 'absent' }])
    await monthPage.goto('2026-08')

    await expect(monthPage.paid('Козлова')).toHaveText('не оплачено')
    await page.getByRole('button', { name: 'Следующий месяц' }).click()
    await expect(page).toHaveURL(`/month/${month}`)
    await expect(monthPage.paid('Козлова')).toHaveText('✓ оплачено')
  })

  test('toggles payment in the calendar view', async ({ page }) => {
    const monthPage = new MonthPage(page)
    await monthPage.goto(month, 'calendar')

    await expect(monthPage.calendarPaid('Козлова Аня')).toHaveText('✓ оплачено')
    await monthPage.calendarPaid('Иванов Петя').click()
    await expect(monthPage.calendarPaid('Иванов Петя')).toHaveText('✓ оплачено')
    await monthPage.saveButton.click()
    await expect(page.getByText('Отметки сохранены')).toBeVisible()

    await page.reload()
    await expect(monthPage.calendarPaid('Иванов Петя')).toHaveText('✓ оплачено')
  })
})
