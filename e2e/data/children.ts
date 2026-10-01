import type { Member } from '@/api/types'
import type { NewChild } from '../fixtures/backend.fixture'

export const members: Member[] = [
  { user_id: 101, name: 'Мама Пети' },
  { user_id: 102, name: 'Папа Ани' },
]

export const petya: NewChild = {
  last_name: 'Иванов',
  first_name: 'Петя',
  parent_ids: [101],
  clubs: [
    { weekday: 1, start: '15:00', name: 'Шахматы' },
    { weekday: 3, start: '15:00', name: 'Шахматы' },
  ],
}

export const anya: NewChild = {
  last_name: 'Козлова',
  first_name: 'Аня',
  middle_name: 'Олеговна',
  price: 15000,
}

export const kolya: NewChild = {
  last_name: 'Сидоров',
  first_name: 'Коля',
  active: false,
}

export const vasya: NewChild = {
  last_name: 'Новиков',
  first_name: 'Вася',
}
