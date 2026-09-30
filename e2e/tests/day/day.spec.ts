import { test, expect } from '../../fixtures/api.fixture'
import { today } from '../../mocks/fixtures/dates'
import { noSchoolDay, pastUnmarkedDay, publishedDay, unpublishedDay } from '../../mocks/fixtures/day'
import { DayPage } from '../../pages/day.page'

test.describe('Day', () => {
  test('shows marks, clubs and summary', async ({ page, api }) => {
    const dayPage = new DayPage(page)
    await api.mockDay(publishedDay)
    await dayPage.goto(today)

    await expect(page.getByText('Отметка идёт')).toContainText('Обедают — 1, не обедают — 1.')
    await expect(dayPage.card('Иванов Петя')).toContainText('Обедает')
    await expect(dayPage.card('Иванов Петя')).toContainText('15:00 Шахматы')
    await expect(dayPage.card('Козлова Аня Олеговна')).toContainText('Не обедает')
  })

  test('saves a mark from the card', async ({ page, api }) => {
    const dayPage = new DayPage(page)
    await api.mockDay(publishedDay)
    await api.mockPutAttendance('2026-09')
    await dayPage.goto(today)
    const card = dayPage.card('Иванов Петя')

    await card.getByTitle('Обедает').click()
    await expect(card).toContainText('Не обедает')
    const request = page.waitForRequest('/api/attendance/2026-09')
    await card.getByRole('button', { name: 'Сохранить' }).click()

    expect((await request).postDataJSON()).toEqual({ marks: [{ date: today, child_id: 1, status: 'absent' }] })
    await expect(page.getByText('Сохранено')).toBeVisible()
  })

  test('saves day clubs from the card', async ({ page, api }) => {
    const dayPage = new DayPage(page)
    await api.mockDay(publishedDay)
    await api.mockPutDayClubs(today, 2)
    await dayPage.goto(today)
    const card = dayPage.card('Козлова Аня Олеговна')

    await card.getByRole('button', { name: '+ Кружок' }).click()
    await card.getByLabel('Название кружка').fill('Танцы')
    await card.getByLabel('Время начала').fill('12:30')
    await card.getByRole('button', { name: 'Добавить' }).click()
    await expect(card).toContainText('12:30 Танцы')

    const request = page.waitForRequest(`/api/days/${today}/children/2/clubs`)
    await card.getByRole('button', { name: 'Сохранить' }).click()

    expect((await request).postDataJSON()).toEqual({ clubs: [{ start: '12:30', name: 'Танцы' }] })
    await expect(page.getByText('Сохранено')).toBeVisible()
  })

  test('without a message clubs are read-only', async ({ page, api }) => {
    const dayPage = new DayPage(page)
    await api.mockDay(unpublishedDay)
    await dayPage.goto(today)

    await expect(page.getByText('Сообщения за этот день не было')).toBeVisible()
    await expect(dayPage.card('Иванов Петя')).toContainText('15:00 Шахматы')
    await expect(page.getByRole('button', { name: '+ Кружок' })).toHaveCount(0)
  })

  test('past day without marks links to the month', async ({ page, api }) => {
    const dayPage = new DayPage(page)
    await api.mockDay(pastUnmarkedDay)
    await dayPage.goto(pastUnmarkedDay.date)

    await expect(page.getByText('Отметок нет')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Открыть месяц' })).toBeVisible()
  })

  test('no school day', async ({ page, api }) => {
    const dayPage = new DayPage(page)
    await api.mockDay(noSchoolDay)
    await dayPage.goto(noSchoolDay.date)

    await expect(page.getByText('Неучебный день')).toBeVisible()
    await expect(dayPage.messageButton).toBeHidden()
  })

  test('shows the day message', async ({ page, api }) => {
    const dayPage = new DayPage(page)
    await api.mockDay(publishedDay)
    await dayPage.goto(today)

    await dayPage.messageButton.click()

    await expect(dayPage.dialog).toContainText('Так сообщение выглядит в группе.')
    await expect(dayPage.dialog.locator('b')).toHaveText('15:00 Шахматы')
    await expect(dayPage.dialog).toContainText('Не обедает')
  })

  test('switches to the previous day', async ({ page, api }) => {
    const dayPage = new DayPage(page)
    await api.mockDay(publishedDay)
    await api.mockDay({ ...noSchoolDay, date: '2026-09-24' })
    await dayPage.goto(today)

    await page.getByRole('button', { name: 'Предыдущий день' }).click()

    await expect(page).toHaveURL('/day/2026-09-24')
    await expect(page.getByText('Неучебный день')).toBeVisible()
  })
})
