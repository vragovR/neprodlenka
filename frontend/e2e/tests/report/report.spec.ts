import { test, expect } from '../../fixtures/backend.fixture'
import { anya, petya } from '../../data/children'
import { month, weekends } from '../../data/dates'
import { ReportPage } from '../../pages/report.page'

test.describe('Report', () => {
  test.use({ signedIn: false })

  test('opens without login and links to the child page', async ({ page, backend }) => {
    const reportPage = new ReportPage(page)
    await backend.setNoSchoolDays(month, weekends(month))
    const child = await backend.addChild(petya)
    for (const date of ['2026-09-01', '2026-09-02', '2026-09-03']) {
      await backend.publishDay(date)
      await backend.closeDay(date)
    }
    const uid = await backend.reportUID(month)
    await reportPage.goto(uid)

    await expect(reportPage.heading).toHaveText('Отчёт за сентябрь 2026')
    await expect(page.getByRole('row', { name: /Иванов П\./ })).toContainText('390 ₽')

    await page.getByRole('link', { name: 'Иванов П.' }).click()

    await expect(page).toHaveURL(`/report/${uid}/child/${child.id}`)
    await expect(reportPage.heading).toHaveText('Иванов Петя')
    await expect(page.getByRole('complementary')).toContainText(/3\s*обеда в месяце/)
    await expect(page.getByText('к оплате за сентябрь 2026')).toBeVisible()
  })

  test('a child without marks opens with zeros and the current price', async ({
    page,
    backend,
  }) => {
    const reportPage = new ReportPage(page)
    await backend.setNoSchoolDays(month, weekends(month))
    await backend.addChild(petya)
    await backend.publishDay('2026-09-01')
    await backend.closeDay('2026-09-01')
    const child = await backend.addChild(anya)
    const uid = await backend.reportUID(month)

    await reportPage.goto(uid)
    await expect(page.getByRole('row', { name: /Иванов П\./ })).toBeVisible()
    await expect(page.getByRole('row', { name: /Козлова/ })).toHaveCount(0)

    await page.goto(`/report/${uid}/child/${child.id}`)
    await expect(reportPage.heading).toHaveText('Козлова Аня Олеговна')
    await expect(page.getByRole('complementary')).toContainText(/0\s*обедов в месяце/)
    await expect(page.getByRole('complementary')).toContainText(/150\s*₽\s*за обед/)
    await expect(page.getByRole('complementary')).toContainText(/0\s*₽\s*к оплате за сентябрь 2026/)
  })

  test('the child page shows its month without a month switcher', async ({ page, backend }) => {
    const reportPage = new ReportPage(page)
    await backend.setNoSchoolDays(month, weekends(month))
    const child = await backend.addChild(petya)
    const uid = await backend.reportUID(month)

    await page.goto(`/report/${uid}/child/${child.id}`)
    await expect(reportPage.heading).toHaveText('Иванов Петя')
    await expect(page.getByText('Календарь школьного питания за сентябрь 2026')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Предыдущий месяц' })).toHaveCount(0)
    await expect(page.getByRole('button', { name: 'Следующий месяц' })).toHaveCount(0)
  })

  test('the child page of an unfilled month', async ({ page, backend }) => {
    const child = await backend.addChild(petya)
    const uid = await backend.reportUID(month)

    await page.goto(`/report/${uid}/child/${child.id}`)
    await expect(page.getByText('Календарь месяца не заполнен')).toBeVisible()
    await expect(page.getByRole('complementary')).toContainText('Пн, 15:00Шахматы')
  })

  test('the child page shows clubs and upcoming club changes', async ({ page, backend }) => {
    await backend.setNoSchoolDays(month, weekends(month))
    const child = await backend.addChild(petya)
    await backend.setClubChanges('2026-09-28', child.id, [{ name: 'Шахматы', start: null }])
    await backend.setClubChanges('2026-09-30', child.id, [{ name: 'Шахматы', start: '16:00' }])
    const uid = await backend.reportUID(month)

    await page.goto(`/report/${uid}/child/${child.id}`)
    const aside = page.getByRole('complementary')
    await expect(aside.getByRole('listitem')).toHaveText([
      'Пн, 15:00Шахматы',
      'Ср, 15:00Шахматы',
      'Понедельник, 28 сентября Шахматы — отменён',
      'Среда, 30 сентября Шахматы — в 16:00',
    ])
  })

  test('unknown child', async ({ page, backend }) => {
    await backend.setNoSchoolDays(month, weekends(month))
    const uid = await backend.reportUID(month)

    await page.goto(`/report/${uid}/child/999`)
    await expect(page.getByText('Страница не найдена')).toBeVisible()
  })

  test('unknown report', async ({ page }) => {
    await new ReportPage(page).goto('missing')

    await expect(page.getByText('Отчёт не найден')).toBeVisible()
  })
})
