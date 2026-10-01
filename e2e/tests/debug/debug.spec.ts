import { test, expect } from '../../fixtures/backend.fixture'
import { anya, petya } from '../../data/children'
import { month, today, weekends } from '../../data/dates'
import { DebugPage } from '../../pages/debug.page'
import { Header } from '../../pages/header.page'
import { MonthPage } from '../../pages/month.page'

test.describe('Debug', () => {
  test.beforeEach(async ({ backend }) => {
    await backend.setNoSchoolDays(month, weekends(month))
    await backend.addChild(petya)
    await backend.addChild(anya)
  })

  test('is hidden without the debug key', async ({ page }) => {
    await page.goto('/debug')

    await expect(page.getByText('Страница не найдена')).toBeVisible()
    await page.goto('/')
    await new Header(page).openMenu()
    await expect(page.getByRole('link', { name: 'Месяц', exact: true })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Отладка' })).toHaveCount(0)
  })

  test('deleting the day message keeps explicit marks, sending again restores pending ones', async ({ page, backend }) => {
    const debugPage = new DebugPage(page)
    const monthPage = new MonthPage(page)
    await backend.publishDay(today)
    await backend.saveMarks(month, [{ date: today, child_id: 2, status: 'absent' }])
    await debugPage.goto()

    page.once('dialog', (dialog) => dialog.accept())
    await debugPage.dayCard.getByRole('button', { name: 'Удалить за 25.09.2026' }).click()
    await expect(page.getByText('Сообщение за 25.09.2026 удалено')).toBeVisible()

    await monthPage.goto(month)
    await expect(page.getByRole('columnheader').nth(25)).toHaveAttribute('title', 'Без публикации')
    await expect(monthPage.cell('Иванов', 25).getByTitle('Нет отметки')).toBeVisible()
    await expect(monthPage.cell('Козлова', 25).getByTitle('Не обедает')).toBeVisible()

    await debugPage.goto()
    await debugPage.dayCard.getByRole('button', { name: 'Отправить за 25.09.2026' }).click()
    await expect(page.getByText('Сообщение за 25.09.2026 отправлено')).toBeVisible()

    await monthPage.goto(month)
    await expect(monthPage.cell('Иванов', 25).getByTitle('Не отметился')).toBeVisible()
    await expect(monthPage.cell('Козлова', 25).getByTitle('Не обедает')).toBeVisible()
  })

  test('day actions follow the chosen day', async ({ page, backend }) => {
    const debugPage = new DebugPage(page)
    await backend.publishDay('2026-09-24')
    await backend.publishDay('2026-09-23')
    await backend.closeDay('2026-09-23')
    await debugPage.goto()

    await expect(debugPage.dayCard.getByRole('button', { name: 'Отправить за 25.09.2026' })).toBeVisible()
    await debugPage.day(debugPage.dayCard, 28).click()
    await expect(debugPage.dayCard).toContainText('Будущий день')
    await debugPage.day(debugPage.dayCard, 26).click()
    await expect(debugPage.dayCard).toContainText('Неучебный день')
    await debugPage.day(debugPage.dayCard, 24).click()
    await expect(debugPage.dayCard.getByRole('button', { name: 'Удалить за 24.09.2026' })).toBeVisible()

    await expect(debugPage.closeCard).toContainText('За этот день нет сообщения')
    await debugPage.day(debugPage.closeCard, 23).click()
    await expect(debugPage.closeCard).toContainText('День уже закрыт')
    await debugPage.day(debugPage.closeCard, 24).click()
    await debugPage.closeCard.getByRole('button', { name: 'Закрыть 24.09.2026' }).click()
    await expect(page.getByText('Отметка на 24.09.2026 закрыта')).toBeVisible()
    await expect(debugPage.closeCard).toContainText('День уже закрыт')
  })

  test('month summary is sent only with marks', async ({ page, backend }) => {
    const debugPage = new DebugPage(page)
    await debugPage.goto()

    await debugPage.monthCard.getByRole('button', { name: 'Отправить за сентябрь 2026' }).click()
    await expect(page.getByText('За месяц нет отметок')).toBeVisible()

    await backend.publishDay(today)
    await debugPage.monthCard.getByRole('button', { name: 'Отправить за сентябрь 2026' }).click()
    await expect(page.getByText('Итоги за сентябрь 2026 отправлены')).toBeVisible()
    await expect(debugPage.monthCard.getByRole('button', { name: 'Удалить за сентябрь 2026' })).toBeVisible()

    const monthPage = new MonthPage(page)
    await monthPage.goto(month)
    await monthPage.messageButton.click()
    await expect(monthPage.dialog).toContainText('Так сообщение выглядит в группе.')
  })
})
