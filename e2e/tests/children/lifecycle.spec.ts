import { test, expect } from '../../fixtures/backend.fixture'
import { anya, petya, vasya } from '../../data/children'
import { month, today, weekends } from '../../data/dates'
import { DayPage } from '../../pages/day.page'
import { MonthPage } from '../../pages/month.page'

const nextMonth = '2026-10'
const previousMonth = '2026-08'

test.beforeEach(async ({ backend }) => {
  for (const m of [previousMonth, month, nextMonth]) await backend.setNoSchoolDays(m, weekends(m))
  await backend.addChild(petya)
  await backend.addChild(anya)
  for (const date of ['2026-09-01', '2026-09-02', '2026-09-03']) {
    await backend.publishDay(date)
    await backend.closeDay(date)
  }
  await backend.publishDay(today)
})

test.describe('Added child', () => {
  test.beforeEach(async ({ backend }) => {
    await backend.addChild(vasya)
  })

  test('appears in the current month without marks', async ({ page }) => {
    const monthPage = new MonthPage(page)
    await monthPage.goto(month)

    await expect(monthPage.cell('Новиков', 2).getByTitle('Нет отметки')).toBeVisible()
    await expect(monthPage.cell('Новиков', 25).getByTitle('Нет отметки')).toBeVisible()
    await expect(monthPage.cell('Новиков', 28)).toBeEnabled()
    await expect(monthPage.pay('Новиков')).toContainText(/^0\s*₽/)

    await monthPage.cell('Новиков', 2).click()
    await monthPage.saveButton.click()
    await expect(page.getByText('Отметки сохранены')).toBeVisible()

    await page.reload()
    await expect(monthPage.cell('Новиков', 2).getByTitle('Обедает')).toBeVisible()
    await expect(monthPage.pay('Новиков')).toContainText('130 ₽')
  })

  test('appears in the next month, only absence can be marked', async ({ page }) => {
    const monthPage = new MonthPage(page)
    await monthPage.goto(nextMonth)

    await expect(page.getByRole('rowheader', { name: /Новиков/ })).toBeVisible()
    await monthPage.cell('Новиков', 1).click()
    await expect(monthPage.cell('Новиков', 1).getByTitle('Не обедает')).toBeVisible()
    await monthPage.cell('Новиков', 1).click()
    await expect(monthPage.cell('Новиков', 1).getByTitle('Нет отметки')).toBeVisible()
    await monthPage.cell('Новиков', 1).click()
    await monthPage.saveButton.click()
    await expect(page.getByText('Отметки сохранены')).toBeVisible()

    await page.reload()
    await expect(monthPage.cell('Новиков', 1).getByTitle('Не обедает')).toBeVisible()
    await expect(monthPage.pay('Новиков')).toContainText(/^0\s*₽/)
  })

  test('is not on the open day and gets a meal when it closes', async ({ page, backend }) => {
    const dayPage = new DayPage(page)
    await dayPage.goto(today)

    await expect(dayPage.card('Иванов Петя')).toContainText('Не отметился')
    await expect(dayPage.card('Новиков Вася')).toHaveCount(0)

    await backend.closeDay(today)
    const monthPage = new MonthPage(page)
    await monthPage.goto(month)
    await expect(monthPage.cell('Новиков', 25).getByTitle('Обедает')).toBeVisible()
    await expect(monthPage.pay('Новиков')).toContainText('130 ₽')
  })
})

test.describe('Deactivated child', () => {
  test.beforeEach(async ({ backend }) => {
    await backend.updateChild(1, { active: false })
  })

  test('stays in the month with its marks and stays editable', async ({ page }) => {
    const monthPage = new MonthPage(page)
    await monthPage.goto(month)

    await expect(page.getByRole('rowheader', { name: /Иванов/ })).toContainText('деактивирован')
    await expect(monthPage.cell('Иванов', 1).getByTitle('Обедает')).toBeVisible()
    await expect(monthPage.pay('Иванов')).toContainText('390 ₽')

    await monthPage.cell('Иванов', 1).click()
    await monthPage.paid('Иванов').click()
    await monthPage.saveButton.click()
    await expect(page.getByText('Отметки сохранены')).toBeVisible()

    await page.reload()
    await expect(monthPage.pay('Иванов')).toContainText('260 ₽')
    await expect(monthPage.paid('Иванов')).toHaveText('✓ оплачено')
  })

  test('disappears from months without its marks', async ({ page }) => {
    const monthPage = new MonthPage(page)

    await monthPage.goto(nextMonth)
    await expect(page.getByRole('rowheader', { name: /Козлова/ })).toBeVisible()
    await expect(page.getByRole('rowheader', { name: /Иванов/ })).toHaveCount(0)

    await monthPage.goto(previousMonth)
    await expect(page.getByText('Отметок нет')).toBeVisible()
  })

  test('is not charged for the open day', async ({ page, backend }) => {
    const dayPage = new DayPage(page)
    await dayPage.goto(today)

    await expect(dayPage.card('Иванов Петя')).toContainText('деактивирован')
    await expect(dayPage.card('Иванов Петя')).toContainText('Не отметился')

    await backend.closeDay(today)
    const monthPage = new MonthPage(page)
    await monthPage.goto(month)
    await expect(monthPage.cell('Иванов', 25).getByTitle('Нет отметки')).toBeVisible()
    await expect(monthPage.cell('Козлова', 25).getByTitle('Обедает')).toBeVisible()
    await expect(monthPage.pay('Иванов')).toContainText('390 ₽')
  })

  test('is not in days published after deactivation', async ({ page, backend }) => {
    const dayPage = new DayPage(page)
    await backend.publishDay('2026-09-24')
    await dayPage.goto('2026-09-24')

    await expect(dayPage.card('Козлова Аня Олеговна')).toBeVisible()
    await expect(dayPage.card('Иванов Петя')).toHaveCount(0)

    await dayPage.goto('2026-09-02')
    await expect(dayPage.card('Иванов Петя')).toContainText('Обедает')
  })
})

test.describe('Child with attendance days', () => {
  test.beforeEach(async ({ backend }) => {
    await backend.updateChild(2, { weekdays: [2, 4] })
    for (const date of ['2026-09-22', '2026-09-23']) {
      await backend.publishDay(date)
      await backend.closeDay(date)
    }
  })

  test('gets marks only on its days', async ({ page }) => {
    const monthPage = new MonthPage(page)
    await monthPage.goto(month)

    await expect(monthPage.cell('Козлова', 22).getByTitle('Обедает')).toBeVisible()
    await expect(monthPage.cell('Козлова', 23).locator('..')).toHaveAttribute('title', 'Не ходит в этот день')
    await expect(monthPage.cell('Козлова', 23).getByTitle('Обедает')).toHaveCount(0)
    await expect(monthPage.cell('Иванов', 23).getByTitle('Обедает')).toBeVisible()
    await expect(monthPage.pay('Козлова')).toContainText('600 ₽')
  })

  test('is not on the day it does not attend', async ({ page }) => {
    const dayPage = new DayPage(page)

    await dayPage.goto('2026-09-23')
    await expect(dayPage.card('Иванов Петя')).toBeVisible()
    await expect(dayPage.card('Козлова Аня Олеговна')).toHaveCount(0)

    await dayPage.goto(today)
    await expect(dayPage.card('Козлова Аня Олеговна')).toContainText('Не отметился')
  })
})
