import type { Page, Route } from '@playwright/test'
import type { Attendance, Child, ChildReport, ChildrenResponse, Day, Report } from '@/api/types'
import { today } from './fixtures/dates'
import { type ApiError, notFound } from './fixtures/errors'

const json = (route: Route, body: unknown, status = 200) =>
  route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) })

const noContent = (route: Route) => route.fulfill({ status: 204 })

export class ApiMock {
  constructor(private page: Page) {}

  async setupAll(): Promise<void> {
    await this.page.clock.setFixedTime(new Date(`${today}T10:00:00+03:00`))
    await this.page.addInitScript(() => localStorage.setItem('login', 'admin'))
    await this.page.route('/api/**', (route) => json(route, notFound, 404))
  }

  private async on(method: string, path: string, handle: (route: Route) => Promise<void>): Promise<void> {
    await this.page.route(`/api${path}`, (route) =>
      route.request().method() === method ? handle(route) : route.fallback(),
    )
  }

  async mockUnauthorized(): Promise<void> {
    await this.page.route('/api/**', (route) =>
      json(route, { error: 'unauthorized', message: 'Требуется вход' }, 401),
    )
  }

  async mockLoginSuccess(login = 'admin'): Promise<void> {
    await this.on('POST', '/login', (route) => json(route, { login }))
  }

  async mockLoginError(error: ApiError, status = 401): Promise<void> {
    await this.on('POST', '/login', (route) => json(route, error, status))
  }

  async mockLogout(): Promise<void> {
    await this.on('DELETE', '/logout', noContent)
  }

  async mockChildren(response: ChildrenResponse): Promise<void> {
    await this.on('GET', '/children', (route) => json(route, response))
  }

  async mockAddChild(id: number): Promise<void> {
    await this.on('POST', '/children', (route) =>
      json(route, { ...route.request().postDataJSON(), id, active: true, parent_ids: [], clubs: [] }, 201),
    )
  }

  async mockUpdateChild(child: Child): Promise<void> {
    await this.on('PATCH', `/children/${child.id}`, (route) => json(route, { ...child, ...route.request().postDataJSON() }))
  }

  async mockDay(day: Day): Promise<void> {
    await this.on('GET', `/days/${day.date}`, (route) => json(route, day))
  }

  async mockPutDayClubs(date: string, childId: number): Promise<void> {
    await this.on('PUT', `/days/${date}/children/${childId}/clubs`, noContent)
  }

  async mockAttendance(attendance: Attendance): Promise<void> {
    await this.on('GET', `/attendance/${attendance.month}`, (route) => json(route, attendance))
  }

  async mockPutAttendance(month: string): Promise<void> {
    await this.on('PUT', `/attendance/${month}`, noContent)
  }

  async mockReport(uid: string, report: Report): Promise<void> {
    await this.on('GET', `/reports/${uid}`, (route) => json(route, report))
  }

  async mockChildReport(uid: string, report: ChildReport): Promise<void> {
    await this.on('GET', `/reports/${uid}/children/${report.child.id}`, (route) => json(route, report))
  }
}
