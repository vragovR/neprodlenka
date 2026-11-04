import { test, expect } from '../../fixtures/backend.fixture'
import { anya, petya } from '../../data/children'
import { month, weekends, today } from '../../data/dates'
import { DayPage } from '../../pages/day.page'

test.describe('Day', () => {
  test.beforeEach(async ({ backend }) => {
    await backend.setNoSchoolDays(month, weekends(month))
    await backend.addChild({ ...petya, clubs: [{ weekday: 5, start: '15:00', name: 'Шахматы' }] })
    await backend.addChild(anya)
  })

  test.describe('published', () => {
    test.beforeEach(async ({ backend }) => {
      await backend.publishDay(today)
      await backend.saveMarks(month, [
        { date: today, child_id: 1, status: 'present' },
        { date: today, child_id: 2, status: 'absent' },
      ])
    })

    test('shows marks, clubs and summary', async ({ page }) => {
      const dayPage = new DayPage(page)
      await dayPage.goto(today)

      await expect(page.getByText('Отметка идёт')).toContainText('Обедают — 1, не обедают — 1.')
      await expect(dayPage.card('Иванов Петя')).toContainText('Обедает')
      await expect(dayPage.card('Иванов Петя')).toContainText('15:00 Шахматы')
      await expect(dayPage.card('Козлова Аня Олеговна')).toContainText('Не обедает')
    })

    test('saves a mark from the card', async ({ page }) => {
      const dayPage = new DayPage(page)
      await dayPage.goto(today)
      const card = dayPage.card('Иванов Петя')

      await card.getByTitle('Обедает').click()
      await expect(card).toContainText('Не обедает')
      await card.getByRole('button', { name: 'Сохранить' }).click()
      await expect(page.getByText('Сохранено')).toBeVisible()

      await page.reload()
      await expect(card.getByTitle('Не обедает')).toBeVisible()
      await expect(page.getByText('Отметка идёт')).toContainText('не обедают — 2.')
    })

    test('saves day clubs from the card', async ({ page }) => {
      const dayPage = new DayPage(page)
      await dayPage.goto(today)
      const card = dayPage.card('Козлова Аня Олеговна')

      await card.getByRole('button', { name: '+ Кружок' }).click()
      await card.getByLabel('Название кружка').fill('Танцы')
      await card.getByLabel('Время начала').fill('12:30')
      await card.getByRole('button', { name: 'Добавить' }).click()
      await expect(card).toContainText('12:30 Танцы')
      await card.getByRole('button', { name: 'Сохранить' }).click()
      await expect(page.getByText('Сохранено')).toBeVisible()

      await page.reload()
      await expect(card).toContainText('12:30 Танцы')
    })

    test('day clubs are a snapshot taken on publishing', async ({ page, backend }) => {
      const dayPage = new DayPage(page)
      await backend.updateChild(1, { clubs: [{ weekday: 5, start: '16:00', name: 'Футбол' }] })
      await dayPage.goto(today)

      await expect(dayPage.card('Иванов Петя')).toContainText('15:00 Шахматы')
      await expect(dayPage.card('Иванов Петя')).not.toContainText('Футбол')

      await dayPage.goto('2026-10-02')
      await dayPage.messageButton.click()
      await expect(dayPage.dialog).toContainText('16:00 Футбол: Иванов Петя')
    })

    test('clubs of an absent child are left out of the message', async ({ page }) => {
      const dayPage = new DayPage(page)
      await dayPage.goto(today)
      const card = dayPage.card('Иванов Петя')

      await card.getByTitle('Обедает').click()
      await card.getByRole('button', { name: 'Сохранить' }).click()
      await expect(page.getByText('Сохранено')).toBeVisible()

      await expect(card).toContainText('15:00 Шахматы')
      await dayPage.messageButton.click()
      await expect(dayPage.dialog).not.toContainText('Шахматы')
    })

    test('the message changes when the day closes', async ({ page, backend }) => {
      const dayPage = new DayPage(page)
      await backend.saveMarks(month, [{ date: today, child_id: 2, status: null }])
      await dayPage.goto(today)

      await dayPage.messageButton.click()
      await expect(dayPage.dialog).toContainText('Доброе утро!')
      await expect(dayPage.dialog).toContainText(/Обедают все 1 человек:\s*Иванов Петя/)
      await expect(dayPage.dialog).toContainText('❌ Не обедает')
      await page.keyboard.press('Escape')

      await backend.closeDay(today)
      await page.reload()
      await expect(page.getByText('Отметка закрыта.')).toBeVisible()
      await dayPage.messageButton.click()
      await expect(dayPage.dialog).toContainText(
        'Сбор информации по обеду на 25 сентября (пятница) завершён.',
      )
      await expect(dayPage.dialog).not.toContainText('❌ Не обедает')
    })

    test('asks before leaving with an unsaved card', async ({ page }) => {
      const dayPage = new DayPage(page)
      await dayPage.goto(today)
      await dayPage.card('Иванов Петя').getByTitle('Обедает').click()

      const messages: string[] = []
      page.once('dialog', async (dialog) => {
        messages.push(dialog.message())
        await dialog.dismiss()
      })
      await page.getByRole('button', { name: 'Предыдущий день' }).click()

      expect(messages).toEqual(['Изменения в карточках детей не сохранены. Уйти?'])
      await expect(page).toHaveURL(`/day/${today}`)
      await expect(dayPage.card('Иванов Петя')).toContainText('Не обедает')
    })

    test('shows the day message', async ({ page }) => {
      const dayPage = new DayPage(page)
      await dayPage.goto(today)

      await dayPage.messageButton.click()

      await expect(dayPage.dialog).toContainText('Так сообщение выглядит в группе.')
      await expect(dayPage.dialog).toContainText('15:00 Шахматы')
      await expect(dayPage.dialog).toContainText('Не обедает')
    })
  })

  test('without a message clubs are read-only', async ({ page, backend }) => {
    const dayPage = new DayPage(page)
    await backend.saveMarks(month, [{ date: '2026-09-18', child_id: 1, status: 'absent' }])
    await dayPage.goto('2026-09-18')

    await expect(page.getByText('Сообщения за этот день не было')).toBeVisible()
    await expect(dayPage.card('Иванов Петя')).toContainText('15:00 Шахматы')
    await expect(page.getByRole('button', { name: '+ Кружок' })).toHaveCount(0)
  })

  test('today before the message is sent', async ({ page }) => {
    await new DayPage(page).goto(today)

    await expect(
      page.getByText('Сообщение за сегодня ещё не отправлено — отметки появятся после отправки.'),
    ).toBeVisible()
  })

  test('past day without marks links to the month', async ({ page }) => {
    const dayPage = new DayPage(page)
    await dayPage.goto('2026-09-24')

    await expect(page.getByText('Отметок нет')).toBeVisible()
    await page.getByRole('button', { name: 'Открыть месяц' }).click()
    await expect(page).toHaveURL(`/month/${month}`)
  })

  test('upcoming day message lists children attending that weekday', async ({ page, backend }) => {
    const dayPage = new DayPage(page)
    await backend.updateChild(2, { weekdays: [2, 4] })

    await dayPage.goto('2026-09-29')
    await expect(page.getByText('Этот день ещё не наступил.')).toBeVisible()
    await dayPage.messageButton.click()
    await expect(dayPage.dialog).toContainText('Иванов Петя')
    await expect(dayPage.dialog).toContainText('Козлова Аня')

    await dayPage.goto('2026-09-30')
    await dayPage.messageButton.click()
    await expect(dayPage.dialog).toContainText('Иванов Петя')
    await expect(dayPage.dialog).not.toContainText('Козлова Аня')
  })

  test('no school day', async ({ page }) => {
    const dayPage = new DayPage(page)
    await dayPage.goto('2026-09-26')

    await expect(page.getByText('Неучебный день')).toBeVisible()
    await expect(dayPage.messageButton).toBeHidden()
  })

  test('switches to the previous day', async ({ page }) => {
    const dayPage = new DayPage(page)
    await dayPage.goto('2026-09-28')

    await page.getByRole('button', { name: 'Предыдущий день' }).click()

    await expect(page).toHaveURL('/day/2026-09-27')
    await expect(page.getByText('Неучебный день')).toBeVisible()
  })
})
