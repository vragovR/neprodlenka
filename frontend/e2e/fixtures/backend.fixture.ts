import { spawn } from 'node:child_process'
import { createServer } from 'node:net'
import { fileURLToPath } from 'node:url'
import { test as base, type APIRequestContext } from '@playwright/test'
import type { Child, ChildPatch, Member, Status } from '@/api/types'
import { today } from '../data/dates'

export type NewChild = Pick<Child, 'last_name' | 'first_name'> & ChildPatch

const admin = { login: 'admin', password: 'admin-password' }

export class Backend {
  constructor(readonly api: APIRequestContext) {}

  private async call(method: string, path: string, data: unknown = {}): Promise<unknown> {
    const res = await this.api.fetch(`/api${path}`, { method, data })
    if (!res.ok()) throw new Error(`${method} ${path}: ${res.status()} ${await res.text()}`)
    return res.status() === 204 ? undefined : res.json()
  }

  async addChild({
    last_name,
    first_name,
    middle_name,
    price,
    ...patch
  }: NewChild): Promise<Child> {
    const child = (await this.call('POST', '/children', {
      last_name,
      first_name,
      middle_name,
      price,
    })) as Child
    if (Object.keys(patch).length === 0) return child
    return this.updateChild(child.id, patch)
  }

  async updateChild(id: number, patch: ChildPatch): Promise<Child> {
    return (await this.call('PATCH', `/children/${id}`, patch)) as Child
  }

  async setNoSchoolDays(month: string, days: string[]): Promise<void> {
    await this.call('PUT', `/months/${month}`, { no_school_days: days })
  }

  async publishDay(date: string): Promise<void> {
    await this.call('POST', `/debug/day/${date}`)
  }

  async closeDay(date: string): Promise<void> {
    await this.call('POST', `/debug/close/${date}`)
  }

  async saveMarks(
    month: string,
    marks: { date: string; child_id: number; status: Status | null }[],
  ): Promise<void> {
    await this.call('PUT', `/attendance/${month}`, { marks, payments: [] })
  }

  async savePayment(month: string, childId: number): Promise<void> {
    await this.call('PUT', `/attendance/${month}`, {
      marks: [],
      payments: [{ child_id: childId, paid: true }],
    })
  }

  async setClubChanges(
    date: string,
    childId: number,
    changes: { name: string; start: string | null }[],
  ): Promise<void> {
    await this.call('PUT', `/days/${date}/children/${childId}/club-changes`, { changes })
  }

  async reportUID(month: string): Promise<string> {
    const { report_url } = (await this.call('GET', `/months/${month}`)) as { report_url: string }
    return report_url.split('/').pop() ?? ''
  }
}

interface BackendOptions {
  now: string
  members: Member[]
  signedIn: boolean
}

interface BackendFixtures {
  backendURL: string
  backend: Backend
}

const freePort = () =>
  new Promise<number>((resolve, reject) => {
    const srv = createServer()
    srv.on('error', reject)
    srv.listen(0, '127.0.0.1', () => {
      const address = srv.address()
      srv.close(() =>
        typeof address === 'object' && address
          ? resolve(address.port)
          : reject(new Error('no port')),
      )
    })
  })

export const test = base.extend<BackendOptions & BackendFixtures>({
  now: [`${today}T10:00:00+03:00`, { option: true }],
  members: [[], { option: true }],
  signedIn: [true, { option: true }],

  backendURL: async ({ now, members }, use, testInfo) => {
    const addr = `127.0.0.1:${await freePort()}`
    const proc = spawn(
      fileURLToPath(new URL('../../../backend/bin/e2eserver', import.meta.url)),
      [
        '--addr',
        addr,
        '--db',
        testInfo.outputPath('e2e.db'),
        '--now',
        now,
        '--members',
        JSON.stringify(members),
      ],
      { env: { ...process.env, TZ: 'Europe/Moscow' }, stdio: ['ignore', 'pipe', 'inherit'] },
    )
    await new Promise<void>((resolve, reject) => {
      proc.on('exit', (code) => reject(new Error(`e2eserver exited with ${code}`)))
      proc.stdout.on('data', (chunk: Buffer) => chunk.toString().includes('ready') && resolve())
    })
    await use(`http://${addr}`)
    proc.kill()
  },

  backend: async ({ playwright, backendURL }, use) => {
    const api = await playwright.request.newContext({ baseURL: backendURL })
    await api.post('/api/login', { data: admin })
    await use(new Backend(api))
    await api.dispose()
  },

  page: async ({ page, now, signedIn, backendURL }, use) => {
    await page.clock.setFixedTime(new Date(now))
    await page.route('/api/**', async (route) => {
      const url = new URL(route.request().url())
      await route.fulfill({
        response: await route.fetch({ url: backendURL + url.pathname + url.search }),
      })
    })
    if (signedIn) {
      await page.addInitScript((login) => localStorage.setItem('login', login), admin.login)
      await page.request.post(`${backendURL}/api/login`, { data: admin })
    }
    await use(page)
  },
})

export { expect } from '@playwright/test'
