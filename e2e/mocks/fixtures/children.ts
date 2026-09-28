import type { Child, ChildrenResponse } from '@/api/types'

export const petya: Child = {
  id: 1,
  last_name: 'Иванов',
  first_name: 'Петя',
  middle_name: '',
  price: 13000,
  active: true,
  parent_ids: [101],
  clubs: [
    { weekday: 1, start: '15:00', name: 'Шахматы' },
    { weekday: 3, start: '15:00', name: 'Шахматы' },
  ],
}

export const anya: Child = {
  id: 2,
  last_name: 'Козлова',
  first_name: 'Аня',
  middle_name: 'Олеговна',
  price: 15000,
  active: true,
  parent_ids: [],
  clubs: [],
}

export const kolya: Child = {
  id: 3,
  last_name: 'Сидоров',
  first_name: 'Коля',
  middle_name: '',
  price: 13000,
  active: false,
  parent_ids: [],
  clubs: [],
}

export const childrenResponse: ChildrenResponse = {
  children: [anya, kolya, petya],
  members: [
    { user_id: 101, name: 'Мама Пети' },
    { user_id: 102, name: 'Папа Ани' },
  ],
}
