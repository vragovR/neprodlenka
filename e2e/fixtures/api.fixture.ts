import { test as base } from '@playwright/test'
import { ApiMock } from '../mocks'

interface ApiFixtures {
  api: ApiMock
}

export const test = base.extend<ApiFixtures>({
  api: [
    async ({ page }, use) => {
      const api = new ApiMock(page)
      await api.setupAll()
      await use(api)
    },
    { auto: true },
  ],
})

export { expect } from '@playwright/test'
