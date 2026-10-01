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

  test('a child without marks opens with zeros and the current price', async ({ page, backend }) => {
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

  test('the child page switches only to months with a report link', async ({ page, backend }) => {
    const reportPage = new ReportPage(page)
    await backend.setNoSchoolDays('2026-08', weekends('2026-08'))
    await backend.setNoSchoolDays(month, weekends(month))
    const child = await backend.addChild(petya)
    const august = await backend.reportUID('2026-08')
    const september = await backend.reportUID(month)

    await page.goto(`/report/${september}/child/${child.id}`)
    await expect(reportPage.heading).toHaveText('Иванов Петя')
    await expect(page.getByRole('button', { name: 'Следующий месяц' })).toBeDisabled()
    await page.getByRole('button', { name: 'Предыдущий месяц' }).click()

    await expect(page).toHaveURL(`/report/${august}/child/${child.id}`)
    await expect(page.getByText('к оплате за август 2026')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Предыдущий месяц' })).toBeDisabled()
    await expect(page.getByRole('button', { name: 'Следующий месяц' })).toBeEnabled()
  })

  test('the child page of an unfilled month', async ({ page, backend }) => {
    const child = await backend.addChild(petya)
    const uid = await backend.reportUID(month)

    await page.goto(`/report/${uid}/child/${child.id}`)
    await expect(page.getByText('Календарь месяца не заполнен')).toBeVisible()
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
