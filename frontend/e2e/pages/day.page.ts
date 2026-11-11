import type { Locator, Page } from '@playwright/test'

export interface ClubChangeInput {
  child?: string
  name?: string
  start?: string
  canceled?: boolean
}

export class DayPage {
  readonly heading: Locator
  readonly messageButton: Locator
  readonly dialog: Locator
  readonly clubChanges: Locator
  readonly addClubChangeButton: Locator

  constructor(readonly page: Page) {
    this.heading = page.getByRole('heading', { name: 'День', level: 1 })
    this.messageButton = page.getByRole('button', { name: 'Сообщение дня' })
    this.dialog = page.getByRole('dialog')
    this.clubChanges = page
      .locator('section')
      .filter({ has: page.getByRole('heading', { name: 'Изменения кружков' }) })
    this.addClubChangeButton = this.clubChanges.getByRole('button', { name: 'Добавить изменение' })
  }

  async goto(date: string): Promise<void> {
    await this.page.goto(`/day/${date}`)
  }

  card(name: string): Locator {
    return this.page.getByRole('article').filter({ hasText: name })
  }

  clubChange(text: string): Locator {
    return this.clubChanges.getByRole('listitem').filter({ hasText: text })
  }

  async addClubChange(change: ClubChangeInput): Promise<void> {
    await this.addClubChangeButton.click()
    await this.saveClubChange(change)
  }

  async editClubChange(text: string, change: ClubChangeInput): Promise<void> {
    await this.clubChange(text).getByRole('button', { name: 'Изменить' }).click()
    await this.saveClubChange(change)
  }

  async deleteClubChange(text: string): Promise<void> {
    await this.clubChange(text).getByRole('button', { name: 'Удалить' }).click()
  }

  private async saveClubChange({ child, name, start, canceled }: ClubChangeInput): Promise<void> {
    if (child !== undefined) await this.dialog.getByLabel('Ребёнок').selectOption({ label: child })
    if (name !== undefined) await this.dialog.getByLabel('Кружок', { exact: true }).fill(name)
    if (canceled !== undefined) await this.dialog.getByLabel('Кружок отменён').setChecked(canceled)
    if (start !== undefined) await this.dialog.getByLabel('Время').fill(start)
    await this.dialog.getByRole('button', { name: 'Сохранить' }).click()
  }
}
