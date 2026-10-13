import { test, expect } from '../../fixtures/api.fixture'
import { childrenResponse, petya } from '../../mocks/fixtures/children'
import { ChildrenPage } from '../../pages/children.page'

test.describe('Children', () => {
  test.beforeEach(async ({ api }) => {
    await api.mockChildren(childrenResponse)
  })

  test('lists active children by name, inactive under a spoiler', async ({ page }) => {
    const childrenPage = new ChildrenPage(page)
    await childrenPage.goto()

    await expect(childrenPage.names()).toHaveText(['Иванов Петя', 'Козлова Аня Олеговна'])
    await expect(childrenPage.card('Сидоров Коля')).toBeHidden()
    await childrenPage.inactive.getByText('Деактивированные · 1').click()
    await expect(childrenPage.card('Сидоров Коля')).toBeVisible()
    await expect(childrenPage.card('Иванов Петя')).toContainText('Мама Пети')
    await expect(childrenPage.card('Иванов Петя')).toContainText('Шахматы · Пн, Ср 15:00')
  })

  test('adds a child', async ({ page, api }) => {
    const childrenPage = new ChildrenPage(page)
    await api.mockAddChild(10)
    await childrenPage.goto()

    await childrenPage.addButton.click()
    await childrenPage.dialog.getByRole('button', { name: 'Добавить' }).click()
    await expect(childrenPage.dialog).toContainText('Укажите фамилию')
    await expect(childrenPage.dialog).toContainText('Укажите имя')

    await childrenPage.dialog.getByLabel('Фамилия').fill(' Ёлкин ')
    await childrenPage.dialog.getByLabel('Имя').fill('Олег')
    const request = page.waitForRequest('/api/children')
    await childrenPage.dialog.getByRole('button', { name: 'Добавить' }).click()

    expect((await request).postDataJSON()).toEqual({ last_name: 'Ёлкин', first_name: 'Олег', middle_name: '', price: 13000 })
    await expect(childrenPage.dialog).toBeHidden()
    await expect(page.getByText('Ёлкин Олег добавлен(а)')).toBeVisible()
    await expect(childrenPage.names().first()).toHaveText('Ёлкин Олег')
  })

  test('saves price and a new club by the card button', async ({ page, api }) => {
    const childrenPage = new ChildrenPage(page)
    await api.mockUpdateChild(petya)
    await childrenPage.goto()
    const card = childrenPage.card('Иванов Петя')

    await card.getByTitle('Изменить цену').click()
    await card.getByRole('textbox').fill('150')
    await card.getByRole('textbox').press('Tab')
    await expect(card.getByTitle('Изменить цену')).toHaveText('150 ₽ / день')
    await card.getByRole('button', { name: '+ Кружок' }).click()
    await card.getByLabel('Название кружка').fill('Робототехника')
    await card.getByRole('group', { name: 'Дни недели' }).getByRole('button', { name: 'Пт' }).click()
    await card.getByLabel('Время начала').fill('14:30')
    await card.getByRole('button', { name: 'Добавить' }).click()
    await expect(card).toContainText('Робототехника · Пт 14:30')

    const request = page.waitForRequest('/api/children/1')
    await card.getByRole('button', { name: 'Сохранить' }).click()

    const body = (await request).postDataJSON()
    expect(body.price).toBe(15000)
    expect(body.clubs).toEqual([...petya.clubs, { name: 'Робототехника', start: '14:30', weekday: 5 }])
    await expect(page.getByText('Сохранено')).toBeVisible()
    await expect(card.getByRole('button', { name: 'Сохранить' })).toBeHidden()
  })

  test('sets attendance weekdays', async ({ page, api }) => {
    const childrenPage = new ChildrenPage(page)
    await api.mockUpdateChild(petya)
    await childrenPage.goto()
    const card = childrenPage.card('Иванов Петя')
    const days = card.getByRole('group', { name: 'Дни посещения' })
    const save = card.getByRole('button', { name: 'Сохранить' })

    await expect(days.getByRole('button', { name: 'Сб' })).toHaveAttribute('aria-pressed', 'false')
    for (const day of ['Пн', 'Вт', 'Ср', 'Чт', 'Пт']) {
      await days.getByRole('button', { name: day }).click()
    }
    await expect(save).toBeDisabled()
    await expect(card).toContainText('Выберите хотя бы один день посещения')

    await days.getByRole('button', { name: 'Вт' }).click()
    await days.getByRole('button', { name: 'Чт' }).click()
    await expect(days.getByRole('button', { name: 'Чт' })).toHaveAttribute('aria-pressed', 'true')

    const request = page.waitForRequest('/api/children/1')
    await save.click()

    expect((await request).postDataJSON().weekdays).toEqual([2, 4])
    await expect(page.getByText('Сохранено')).toBeVisible()
    await expect(save).toBeHidden()
  })

  test('links a parent by the card button', async ({ page, api }) => {
    const childrenPage = new ChildrenPage(page)
    await api.mockUpdateChild(petya)
    await childrenPage.goto()
    const card = childrenPage.card('Иванов Петя')

    await card.getByRole('combobox').selectOption('Папа Ани')
    await expect(card).toContainText('Папа Ани')
    await expect(card.getByRole('combobox')).toBeHidden()

    const request = page.waitForRequest('/api/children/1')
    await card.getByRole('button', { name: 'Сохранить' }).click()

    expect((await request).postDataJSON().parent_ids).toEqual([101, 102])
    await expect(page.getByText('Сохранено')).toBeVisible()
    await expect(card.getByRole('button', { name: 'Сохранить' })).toBeHidden()
    await expect(card).toContainText('Мама Пети')
    await expect(card).toContainText('Папа Ани')
  })

  test('unlinks the last parent by the card button', async ({ page, api }) => {
    const childrenPage = new ChildrenPage(page)
    await api.mockUpdateChild(petya)
    await childrenPage.goto()
    const card = childrenPage.card('Иванов Петя')

    await card.getByRole('button', { name: 'Отвязать Мама Пети' }).click()
    await expect(card).toContainText('Родители не привязаны')

    const request = page.waitForRequest('/api/children/1')
    await card.getByRole('button', { name: 'Сохранить' }).click()

    expect((await request).postDataJSON().parent_ids).toEqual([])
    await expect(page.getByText('Сохранено')).toBeVisible()
    await expect(card.getByRole('button', { name: 'Сохранить' })).toBeHidden()
    await expect(card).toContainText('Родители не привязаны')
  })

  test('cancel restores the card', async ({ page }) => {
    const childrenPage = new ChildrenPage(page)
    await childrenPage.goto()
    const card = childrenPage.card('Иванов Петя')

    await card.getByRole('button', { name: 'Удалить Шахматы · Пн, Ср 15:00' }).click()
    await expect(card).not.toContainText('Шахматы')
    await card.getByRole('button', { name: 'Отмена' }).click()

    await expect(card).toContainText('Шахматы · Пн, Ср 15:00')
    await expect(card.getByRole('button', { name: 'Сохранить' })).toBeHidden()
  })
})
