import { test, expect } from '../../fixtures/api.fixture'
import { childReport, report, reportUID } from '../../mocks/fixtures/report'
import { ReportPage } from '../../pages/report.page'

test.describe('Report', () => {
  test('opens without login and links to the child page', async ({ page, api }) => {
    const reportPage = new ReportPage(page)
    await api.mockReport(reportUID, report)
    await api.mockChildReport(reportUID, childReport)
    await reportPage.goto(reportUID)

    await expect(reportPage.heading).toHaveText('Отчёт за сентябрь 2026')
    await expect(page.getByRole('row', { name: /Иванов П\./ })).toContainText('390 ₽')

    await page.getByRole('link', { name: 'Иванов П.' }).click()

    await expect(page).toHaveURL(`/report/${reportUID}/child/1`)
    await expect(reportPage.heading).toHaveText('Иванов Петя')
    await expect(page.getByRole('complementary')).toContainText(/3\s*обеда в месяце/)
    await expect(page.getByText('к оплате за сентябрь 2026')).toBeVisible()
  })

  test('unknown report', async ({ page }) => {
    await new ReportPage(page).goto('missing')

    await expect(page.getByText('Отчёт не найден')).toBeVisible()
  })
})
