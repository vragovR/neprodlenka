import { describe, expect, it } from 'vitest'
import { expandGroup, groupClubs, groupLabel } from '../clubs'

describe('clubs', () => {
  const clubs = [
    { weekday: 3, start: '15:00', name: 'Шахматы' },
    { weekday: 1, start: '15:00', name: 'Шахматы' },
    { weekday: 1, start: '14:30', name: 'Робототехника' },
    { weekday: 2, start: '16:00', name: 'Шахматы' },
  ]

  it('groups by name and time, sorted by first day, time and name', () => {
    expect(groupClubs(clubs)).toEqual([
      { name: 'Робототехника', start: '14:30', weekdays: [1] },
      { name: 'Шахматы', start: '15:00', weekdays: [1, 3] },
      { name: 'Шахматы', start: '16:00', weekdays: [2] },
    ])
  })

  it('expands groups back to clubs', () => {
    const back = groupClubs(clubs).flatMap(expandGroup)
    const key = (c: (typeof clubs)[number]) => `${c.weekday} ${c.start} ${c.name}`
    expect(back.map(key).sort()).toEqual(clubs.map(key).sort())
  })

  it('labels a group', () => {
    expect(groupLabel({ name: 'Шахматы', start: '15:00', weekdays: [1, 3] })).toBe('Шахматы · Пн, Ср 15:00')
  })
})
