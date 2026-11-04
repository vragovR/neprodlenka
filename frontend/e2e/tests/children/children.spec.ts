import { test, expect } from '../../fixtures/backend.fixture'
import { anya, kolya, members, petya } from '../../data/children'
import { ChildrenPage } from '../../pages/children.page'

test.describe('Children', () => {
  test.use({ members: [members, { scope: 'test' }] })

  test.beforeEach(async ({ backend }) => {
    for (const child of [anya, kolya, petya]) await backend.addChild(child)
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

  test('adds a child', async ({ page }) => {
    const childrenPage = new ChildrenPage(page)
    await childrenPage.goto()

    await childrenPage.addButton.click()
    await childrenPage.dialog.getByRole('button', { name: 'Добавить' }).click()
    await expect(childrenPage.dialog).toContainText('Укажите фамилию')
    await expect(childrenPage.dialog).toContainText('Укажите имя')

    await childrenPage.dialog.getByLabel('Фамилия').fill(' Ёлкин ')
    await childrenPage.dialog.getByLabel('Имя').fill('Олег')
    await childrenPage.dialog.getByRole('button', { name: 'Добавить' }).click()

    await expect(childrenPage.dialog).toBeHidden()
    await expect(page.getByText('Ёлкин Олег добавлен(а)')).toBeVisible()
    await expect(childrenPage.names().first()).toHaveText('Ёлкин Олег')

    await page.reload()
    await expect(childrenPage.names()).toHaveText([
      'Ёлкин Олег',
      'Иванов Петя',
      'Козлова Аня Олеговна',
    ])
    await expect(childrenPage.card('Ёлкин Олег').getByTitle('Изменить цену')).toHaveText(
      '130 ₽ / день',
    )
  })

  test('saves price and a new club by the card button', async ({ page }) => {
    const childrenPage = new ChildrenPage(page)
    await childrenPage.goto()
    const card = childrenPage.card('Иванов Петя')

    await card.getByTitle('Изменить цену').click()
    await card.getByRole('textbox').fill('150')
    await card.getByRole('textbox').press('Tab')
    await expect(card.getByTitle('Изменить цену')).toHaveText('150 ₽ / день')
    await card.getByRole('button', { name: '+ Кружок' }).click()
    await card.getByLabel('Название кружка').fill('Робототехника')
    await card
      .getByRole('group', { name: 'Дни недели' })
      .getByRole('button', { name: 'Пт' })
      .click()
    await card.getByLabel('Время начала').fill('14:30')
    await card.getByRole('button', { name: 'Добавить' }).click()
    await expect(card).toContainText('Робототехника · Пт 14:30')

    await card.getByRole('button', { name: 'Сохранить' }).click()
    await expect(page.getByText('Сохранено')).toBeVisible()
    await expect(card.getByRole('button', { name: 'Сохранить' })).toBeHidden()

    await page.reload()
    await expect(card.getByTitle('Изменить цену')).toHaveText('150 ₽ / день')
    await expect(card).toContainText('Шахматы · Пн, Ср 15:00')
    await expect(card).toContainText('Робототехника · Пт 14:30')
  })

  test('sets attendance weekdays', async ({ page }) => {
    const childrenPage = new ChildrenPage(page)
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
    await save.click()
    await expect(page.getByText('Сохранено')).toBeVisible()
    await expect(save).toBeHidden()

    await page.reload()
    for (const day of ['Пн', 'Ср', 'Пт']) {
      await expect(days.getByRole('button', { name: day })).toHaveAttribute('aria-pressed', 'false')
    }
    for (const day of ['Вт', 'Чт']) {
      await expect(days.getByRole('button', { name: day })).toHaveAttribute('aria-pressed', 'true')
    }
  })

  test('links a parent by the card button', async ({ page }) => {
    const childrenPage = new ChildrenPage(page)
    await childrenPage.goto()
    const card = childrenPage.card('Иванов Петя')

    await card.getByRole('combobox').selectOption('Папа Ани')
    await expect(card).toContainText('Папа Ани')
    await expect(card.getByRole('combobox')).toBeHidden()

    await card.getByRole('button', { name: 'Сохранить' }).click()
    await expect(page.getByText('Сохранено')).toBeVisible()
    await expect(card.getByRole('button', { name: 'Сохранить' })).toBeHidden()

    await page.reload()
    await expect(card).toContainText('Мама Пети')
    await expect(card).toContainText('Папа Ани')
  })

  test('unlinks the last parent by the card button', async ({ page }) => {
    const childrenPage = new ChildrenPage(page)
    await childrenPage.goto()
    const card = childrenPage.card('Иванов Петя')

    await card.getByRole('button', { name: 'Отвязать Мама Пети' }).click()
    await expect(card).toContainText('Родители не привязаны')

    await card.getByRole('button', { name: 'Сохранить' }).click()
    await expect(page.getByText('Сохранено')).toBeVisible()
    await expect(card.getByRole('button', { name: 'Сохранить' })).toBeHidden()

    await page.reload()
    await expect(card).toContainText('Родители не привязаны')
  })

  test('a parent can be linked to several children', async ({ page }) => {
    const childrenPage = new ChildrenPage(page)
    await childrenPage.goto()
    const card = childrenPage.card('Козлова Аня Олеговна')

    await card.getByRole('combobox').selectOption('Мама Пети')
    await card.getByRole('button', { name: 'Сохранить' }).click()
    await expect(page.getByText('Сохранено')).toBeVisible()

    await page.reload()
    await expect(card).toContainText('Мама Пети')
    await expect(childrenPage.card('Иванов Петя')).toContainText('Мама Пети')
  })

  test('a parent who left the group is shown by id and can be unlinked', async ({
    page,
    backend,
  }) => {
    const childrenPage = new ChildrenPage(page)
    await backend.updateChild(1, { parent_ids: [999] })
    await childrenPage.goto()
    const card = childrenPage.card('Козлова Аня Олеговна')

    await expect(card).toContainText('не в группе (id 999)')
    await card.getByRole('button', { name: 'Отвязать не в группе (id 999)' }).click()
    await card.getByRole('button', { name: 'Сохранить' }).click()
    await expect(page.getByText('Сохранено')).toBeVisible()

    await page.reload()
    await expect(card).toContainText('Родители не привязаны')
  })

  test('deactivates and activates by the card button', async ({ page }) => {
    const childrenPage = new ChildrenPage(page)
    await childrenPage.goto()

    await childrenPage.card('Иванов Петя').getByRole('button', { name: 'Деактивировать' }).click()
    await expect(childrenPage.names()).toHaveText(['Козлова Аня Олеговна'])
    await childrenPage.inactive.getByText('Деактивированные · 2').click()
    await childrenPage.card('Иванов Петя').getByRole('button', { name: 'Активировать' }).click()

    await page.reload()
    await expect(childrenPage.names()).toHaveText(['Иванов Петя', 'Козлова Аня Олеговна'])
    await expect(childrenPage.inactive).toContainText('Деактивированные · 1')
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
