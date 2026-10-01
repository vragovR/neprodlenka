import { test, expect } from '../../fixtures/backend.fixture'
import { anya, petya, vasya } from '../../data/children'
import { month, today, weekends } from '../../data/dates'
import { DayPage } from '../../pages/day.page'
import { MonthPage } from '../../pages/month.page'

const closedDays = ['2026-09-01', '2026-09-02', '2026-09-03']

test.beforeEach(async ({ backend }) => {
  await backend.setNoSchoolDays(month, weekends(month))
  await backend.addChild({ ...petya, clubs: [{ weekday: 5, start: '15:00', name: 'Шахматы' }] })
  await backend.addChild(anya)
})

test.describe('Prices', () => {
  test('a mark keeps the price it was set with', async ({ page, backend }) => {
    const monthPage = new MonthPage(page)
    for (const date of closedDays.slice(0, 2)) {
      await backend.publishDay(date)
      await backend.closeDay(date)
    }
    await backend.updateChild(1, { price: 15000 })
    await backend.publishDay(closedDays[2]!)
    await backend.closeDay(closedDays[2]!)
    await monthPage.goto(month)

    await expect(monthPage.pay('Иванов')).toContainText('410 ₽')
    await monthPage.messageButton.click()
    await expect(monthPage.dialog).toContainText('Иванов Петя: 3 дня = 410 ₽')
    await expect(monthPage.dialog).toContainText('Козлова Аня: 3 дня × 150 ₽ = 450 ₽')
    await page.keyboard.press('Escape')

    const uid = await backend.reportUID(month)
    await page.goto(`/report/${uid}/child/1`)
    await expect(page.getByRole('complementary')).toContainText('410 ₽')
    await expect(page.getByRole('complementary')).not.toContainText('за обед')
    await page.goto(`/report/${uid}/child/2`)
    await expect(page.getByRole('complementary')).toContainText('150 ₽за обед')
    await monthPage.goto(month)

    await monthPage.cell('Иванов', 1).click()
    await monthPage.save()
    await expect(page.getByText('Отметки сохранены')).toBeVisible()
    await monthPage.cell('Иванов', 1).click()
    await monthPage.cell('Иванов', 1).click()
    await monthPage.save()

    await page.reload()
    await expect(monthPage.cell('Иванов', 1).getByTitle('Обедает')).toBeVisible()
    await expect(monthPage.pay('Иванов')).toContainText('430 ₽')
  })
})

test.describe('Open day', () => {
  test('an absence set before publishing survives publishing and closing', async ({ page, backend }) => {
    const dayPage = new DayPage(page)
    await backend.saveMarks(month, [{ date: today, child_id: 2, status: 'absent' }])
    await backend.publishDay(today)
    await dayPage.goto(today)

    await expect(dayPage.card('Козлова Аня Олеговна')).toContainText('Не обедает')
    await expect(dayPage.card('Иванов Петя')).toContainText('Не отметился')

    await backend.closeDay(today)
    const monthPage = new MonthPage(page)
    await monthPage.goto(month)
    await expect(monthPage.cell('Козлова', 25).getByTitle('Не обедает')).toBeVisible()
    await expect(monthPage.cell('Иванов', 25).getByTitle('Обедает')).toBeVisible()
    await expect(monthPage.pay('Козлова')).toContainText(/^0\s*₽/)
  })

  test('a cleared mark on the open day becomes a meal when it closes', async ({ page, backend }) => {
    const monthPage = new MonthPage(page)
    await backend.publishDay(today)
    await backend.saveMarks(month, [{ date: today, child_id: 1, status: 'present' }])
    await monthPage.goto(month)

    await monthPage.cell('Иванов', 25).click()
    await monthPage.cell('Иванов', 25).click()
    await expect(monthPage.cell('Иванов', 25).getByTitle('Нет отметки')).toBeVisible()
    await monthPage.saveButton.click()
    await expect(page.getByText('Отметки сохранены')).toBeVisible()

    await backend.closeDay(today)
    await page.reload()
    await expect(monthPage.cell('Иванов', 25).getByTitle('Обедает')).toBeVisible()
    await expect(monthPage.pay('Иванов')).toContainText('130 ₽')
  })

  test('a pending cell cycles back to pending without changes', async ({ page, backend }) => {
    const monthPage = new MonthPage(page)
    await backend.publishDay(today)
    await monthPage.goto(month)

    await monthPage.cell('Иванов', 25).click()
    await expect(monthPage.cell('Иванов', 25).getByTitle('Обедает')).toBeVisible()
    await monthPage.cell('Иванов', 25).click()
    await expect(monthPage.cell('Иванов', 25).getByTitle('Не обедает')).toBeVisible()
    await monthPage.cell('Иванов', 25).click()
    await expect(monthPage.cell('Иванов', 25).getByTitle('Не отметился')).toBeVisible()
    await expect(monthPage.saveButton).toBeDisabled()
  })

  test('a child added after publishing is marked by the admin and closing keeps the mark', async ({ page, backend }) => {
    const monthPage = new MonthPage(page)
    const dayPage = new DayPage(page)
    await backend.publishDay(today)
    await backend.addChild(vasya)
    await monthPage.goto(month)

    await monthPage.cell('Новиков', 25).click()
    await monthPage.cell('Новиков', 25).click()
    await expect(monthPage.cell('Новиков', 25).getByTitle('Не обедает')).toBeVisible()
    await monthPage.saveButton.click()
    await expect(page.getByText('Отметки сохранены')).toBeVisible()

    await dayPage.goto(today)
    await expect(dayPage.card('Новиков Вася')).toContainText('Не обедает')
    await dayPage.messageButton.click()
    await expect(dayPage.dialog).toContainText(/Не обедают \(1\):\s*Новиков Вася/)

    await backend.closeDay(today)
    await monthPage.goto(month)
    await expect(monthPage.cell('Новиков', 25).getByTitle('Не обедает')).toBeVisible()
    await expect(monthPage.pay('Новиков')).toContainText(/^0\s*₽/)
  })

  test('a cleared mark keeps the child on the day only with day clubs', async ({ page, backend }) => {
    const dayPage = new DayPage(page)
    await backend.publishDay(today)
    await backend.saveMarks(month, [{ date: today, child_id: 1, status: null }])
    await dayPage.goto(today)

    await expect(dayPage.card('Иванов Петя')).toContainText('Нет отметки')
    await expect(dayPage.card('Иванов Петя')).toContainText('15:00 Шахматы')
    await dayPage.messageButton.click()
    await expect(dayPage.dialog).toContainText('15:00 Шахматы: Иванов Петя')
  })

  test('a day removed from attendance days after publishing loses its pending mark on close', async ({ page, backend }) => {
    await backend.publishDay(today)
    await backend.updateChild(2, { weekdays: [2, 4] })
    await backend.closeDay(today)

    const monthPage = new MonthPage(page)
    await monthPage.goto(month)
    await expect(monthPage.cell('Козлова', 25).locator('..')).toHaveAttribute('title', 'Не ходит в этот день')
    await expect(monthPage.cell('Козлова', 25).getByTitle(/обедает|Не отметился/)).toHaveCount(0)

    const dayPage = new DayPage(page)
    await dayPage.goto(today)
    await expect(dayPage.card('Иванов Петя')).toContainText('Обедает')
    await expect(dayPage.card('Козлова Аня Олеговна')).toHaveCount(0)
    await expect(page.getByText('Отметка закрыта.')).not.toContainText('не отметились')
  })
})

test.describe('Attendance days', () => {
  test.beforeEach(async ({ backend }) => {
    await backend.updateChild(2, { weekdays: [2, 4] })
  })

  test('a meal on a day the child does not attend is kept and charged', async ({ page, backend }) => {
    const monthPage = new MonthPage(page)
    await backend.publishDay('2026-09-23')
    await monthPage.goto(month)

    await monthPage.cell('Козлова', 23).click()
    await expect(monthPage.cell('Козлова', 23).getByTitle('Обедает')).toBeVisible()
    await monthPage.saveButton.click()
    await expect(page.getByText('Отметки сохранены')).toBeVisible()

    await backend.closeDay('2026-09-23')
    await page.reload()
    await expect(monthPage.cell('Козлова', 23).getByTitle('Обедает')).toBeVisible()
    await expect(monthPage.pay('Козлова')).toContainText('150 ₽')

    const dayPage = new DayPage(page)
    await dayPage.goto('2026-09-23')
    await expect(dayPage.card('Козлова Аня Олеговна')).toContainText('Обедает')
  })
})

test.describe('Past days', () => {
  test('a cleared mark removes the child from the published day and its message', async ({ page, backend }) => {
    const monthPage = new MonthPage(page)
    const dayPage = new DayPage(page)
    await backend.publishDay(closedDays[0]!)
    await backend.closeDay(closedDays[0]!)
    await monthPage.goto(month)

    await monthPage.cell('Иванов', 1).click()
    await monthPage.cell('Иванов', 1).click()
    await monthPage.saveButton.click()
    await expect(page.getByText('Отметки сохранены')).toBeVisible()

    await dayPage.goto(closedDays[0]!)
    await expect(dayPage.card('Козлова Аня Олеговна')).toContainText('Обедает')
    await expect(dayPage.card('Иванов Петя')).toHaveCount(0)
    await dayPage.messageButton.click()
    await expect(dayPage.dialog).toContainText('Козлова Аня')
    await expect(dayPage.dialog).not.toContainText('Иванов Петя')
  })

  test('an unpublished day shows only marked children', async ({ page, backend }) => {
    const dayPage = new DayPage(page)
    await backend.saveMarks(month, [{ date: '2026-09-18', child_id: 2, status: 'absent' }])
    await dayPage.goto('2026-09-18')

    await expect(dayPage.card('Козлова Аня Олеговна')).toContainText('Не обедает')
    await expect(dayPage.card('Иванов Петя')).toHaveCount(0)
    await dayPage.messageButton.click()
    await expect(dayPage.dialog).toContainText('Козлова Аня')
    await expect(dayPage.dialog).not.toContainText('Иванов Петя')
    await expect(dayPage.dialog).not.toContainText('Шахматы')
  })

  test('a future absence is shown read-only on the day', async ({ page }) => {
    const monthPage = new MonthPage(page)
    await monthPage.goto(month)
    await monthPage.cell('Козлова', 29).click()
    await monthPage.saveButton.click()
    await expect(page.getByText('Отметки сохранены')).toBeVisible()

    const dayPage = new DayPage(page)
    await dayPage.goto('2026-09-29')
    await expect(page.getByText('Сообщения за этот день не было')).toBeVisible()
    await expect(dayPage.card('Козлова Аня Олеговна')).toContainText('Не обедает')
    await expect(dayPage.card('Козлова Аня Олеговна').getByRole('button').first()).toBeDisabled()
    await expect(dayPage.card('Иванов Петя')).toContainText('Нет отметки')
    await dayPage.messageButton.click()
    await expect(dayPage.dialog).toContainText(/Не обедают \(1\):\s*Козлова Аня/)
  })
})

test.describe('Month view', () => {
  test('day headers show the publication state', async ({ page, backend }) => {
    const monthPage = new MonthPage(page)
    await backend.publishDay(closedDays[0]!)
    await backend.closeDay(closedDays[0]!)
    await backend.publishDay(today)
    await monthPage.goto(month)

    const header = (day: number) => page.getByRole('columnheader').nth(day)
    await expect(header(1)).toHaveAttribute('title', 'Отметка закрыта')
    await expect(header(2)).toHaveAttribute('title', 'Без публикации')
    await expect(header(25)).toHaveAttribute('title', 'Идёт отметка')
    await expect(header(26)).toHaveAttribute('title', 'Неучебный день')
    await expect(header(28)).toHaveAttribute('title', 'Ещё не наступил')
  })

  test('the month message and the report list only children with marks', async ({ page, backend }) => {
    const monthPage = new MonthPage(page)
    await backend.publishDay(today)
    await backend.addChild(vasya)
    const uid = await backend.reportUID(month)
    await monthPage.goto(month)

    await expect(page.getByRole('rowheader', { name: /Новиков/ })).toBeVisible()
    await monthPage.messageButton.click()
    await expect(monthPage.dialog).toContainText('Иванов Петя: 0 дней × 130 ₽ = 0 ₽')
    await expect(monthPage.dialog).not.toContainText('Новиков')
    await page.keyboard.press('Escape')

    await page.goto(`/report/${uid}`)
    await expect(page.getByRole('row', { name: /Иванов П\./ }).getByTitle('Не отметился')).toBeVisible()
    await expect(page.getByRole('row', { name: /Новиков/ })).toHaveCount(0)

    await monthPage.goto(month)
    await monthPage.cell('Новиков', 29).click()
    await monthPage.saveButton.click()
    await expect(page.getByText('Отметки сохранены')).toBeVisible()
    await monthPage.messageButton.click()
    await expect(monthPage.dialog).toContainText('Новиков Вася: 0 дней × 130 ₽ = 0 ₽')
  })

  test('no month message without marks, though active children are listed', async ({ page, backend }) => {
    const monthPage = new MonthPage(page)
    await backend.setNoSchoolDays('2026-10', weekends('2026-10'))
    await monthPage.goto('2026-10')

    await expect(page.getByRole('rowheader', { name: /Иванов/ })).toBeVisible()
    await expect(monthPage.messageButton).toBeHidden()
  })
})
