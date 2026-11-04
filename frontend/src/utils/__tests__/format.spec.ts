import { describe, expect, it } from 'vite-plus/test'
import {
  addDays,
  addMonths,
  formatDate,
  formatMoney,
  fullName,
  isISODate,
  isMonthKey,
  monthDates,
  monthTitle,
  plural,
  shortName,
  sortByName,
  weekday,
} from '../format'

const person = (last_name: string, first_name: string, middle_name = '') => ({
  last_name,
  first_name,
  middle_name,
})

describe('names', () => {
  it('builds full and short names', () => {
    expect(fullName(person('Иванов', 'Пётр', 'Олегович'))).toBe('Иванов Пётр Олегович')
    expect(fullName(person('Иванов', 'Пётр'))).toBe('Иванов Пётр')
    expect(shortName(person('Иванов', 'Пётр', 'Олегович'))).toBe('Иванов П.О.')
    expect(shortName(person('Иванов', 'Пётр'))).toBe('Иванов П.')
  })

  it('sorts by last and first name without mutating', () => {
    const list = [
      person('Козлова', 'Аня'),
      person('Ёлкин', 'Олег'),
      person('Иванов', 'Петя'),
      person('Иванов', 'Андрей'),
    ]
    expect(sortByName(list).map((p) => fullName(p))).toEqual([
      'Ёлкин Олег',
      'Иванов Андрей',
      'Иванов Петя',
      'Козлова Аня',
    ])
    expect(list[0]?.last_name).toBe('Козлова')
  })
})

describe('money', () => {
  it('shows kopecks only when there are any', () => {
    expect(formatMoney(13000)).toBe('130 ₽')
    expect(formatMoney(13050)).toBe('130,50 ₽')
    expect(formatMoney(1234500)).toBe('12 345 ₽')
  })
})

describe('dates', () => {
  it('validates dates and months', () => {
    expect(isISODate('2026-02-28')).toBe(true)
    expect(isISODate('2026-02-30')).toBe(false)
    expect(isISODate('28.02.2026')).toBe(false)
    expect(isMonthKey('2026-09')).toBe(true)
    expect(isMonthKey('2026-13')).toBe(false)
  })

  it('adds days and months across boundaries', () => {
    expect(addDays('2026-12-31', 1)).toBe('2027-01-01')
    expect(addDays('2026-03-01', -1)).toBe('2026-02-28')
    expect(addMonths('2026-12', 1)).toBe('2027-01')
    expect(addMonths('2026-01', -1)).toBe('2025-12')
  })

  it('lists days of a month', () => {
    expect(monthDates('2026-02')).toHaveLength(28)
    expect(monthDates('2028-02').at(-1)).toBe('2028-02-29')
    expect(monthDates('2026-09')[0]).toBe('2026-09-01')
  })

  it('counts weekdays from Monday', () => {
    expect(weekday('2026-09-28')).toBe(0)
    expect(weekday('2026-09-27')).toBe(6)
  })

  it('formats titles and dates', () => {
    expect(monthTitle('2026-09')).toBe('сентябрь 2026')
    expect(formatDate('2026-09-05')).toBe('05.09.2026')
  })
})

describe('plural', () => {
  it.each([
    [1, 'обед'],
    [2, 'обеда'],
    [5, 'обедов'],
    [11, 'обедов'],
    [12, 'обедов'],
    [21, 'обед'],
    [22, 'обеда'],
    [112, 'обедов'],
  ])('%i → %s', (n, want) => {
    expect(plural(n, ['обед', 'обеда', 'обедов'])).toBe(want)
  })
})
