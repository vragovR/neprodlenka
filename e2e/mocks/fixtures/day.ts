import type { Day } from '@/api/types'
import { today } from './dates'

export const publishedDay: Day = {
  date: today,
  school: true,
  published: true,
  closed: false,
  children: [
    {
      child_id: 1,
      last_name: 'Иванов',
      first_name: 'Петя',
      middle_name: '',
      active: true,
      status: 'present',
      clubs: [{ start: '15:00', name: 'Шахматы' }],
    },
    {
      child_id: 2,
      last_name: 'Козлова',
      first_name: 'Аня',
      middle_name: 'Олеговна',
      active: true,
      status: 'absent',
      clubs: [],
    },
  ],
  message: {
    text: 'Обед 25 сентября.\n\n🎨 Кружки и секции:\n<b>15:00 Шахматы</b>: Иванов Петя',
    buttons: ['Не обедает'],
  },
}

export const unpublishedDay: Day = {
  ...publishedDay,
  published: false,
  children: publishedDay.children.map((c) => ({ ...c, status: c.child_id === 1 ? 'absent' : null })),
}

export const pastUnmarkedDay: Day = {
  date: '2026-09-24',
  school: true,
  published: false,
  closed: false,
  children: [],
  message: null,
}

export const noSchoolDay: Day = {
  date: '2026-09-26',
  school: false,
  published: false,
  closed: false,
  children: [],
  message: null,
}
