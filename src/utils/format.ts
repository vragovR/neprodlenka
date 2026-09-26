import type { PersonName } from '@/api/types'

export const monthNames = [
  'январь', 'февраль', 'март', 'апрель', 'май', 'июнь',
  'июль', 'август', 'сентябрь', 'октябрь', 'ноябрь', 'декабрь',
]

export const weekdaysShort = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']

const pad = (n: number) => String(n).padStart(2, '0')

export function fullName(p: PersonName): string {
  return [p.last_name, p.first_name, p.middle_name].filter(Boolean).join(' ')
}

export function shortName(p: PersonName): string {
  const initials = [p.first_name, p.middle_name].filter(Boolean).map((n) => `${n.charAt(0)}.`).join('')
  return [p.last_name, initials].filter(Boolean).join(' ')
}

export function sortByName<T extends PersonName>(list: T[]): T[] {
  return [...list].sort(
    (a, b) => a.last_name.localeCompare(b.last_name, 'ru') || a.first_name.localeCompare(b.first_name, 'ru'),
  )
}

export function formatMoney(kopecks: number): string {
  return new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    minimumFractionDigits: kopecks % 100 === 0 ? 0 : 2,
  }).format(kopecks / 100)
}

export function toISODate(d: Date): string {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

export function parseISODate(date: string): Date {
  const [y = 0, m = 1, d = 1] = date.split('-').map(Number)
  return new Date(y, m - 1, d)
}

export const today = () => toISODate(new Date())

export const monthOf = (date: string) => date.slice(0, 7)

export const isMonthKey = (s: string) => /^\d{4}-(0[1-9]|1[0-2])$/.test(s)

export const isISODate = (s: string) => /^\d{4}-\d{2}-\d{2}$/.test(s) && toISODate(parseISODate(s)) === s

export function addMonths(month: string, n: number): string {
  const d = parseISODate(`${month}-01`)
  d.setMonth(d.getMonth() + n)
  return monthOf(toISODate(d))
}

export function monthTitle(month: string): string {
  const d = parseISODate(`${month}-01`)
  const name = monthNames[d.getMonth()] ?? ''
  return `${name} ${d.getFullYear()}`
}

export function monthDates(month: string): string[] {
  const d = parseISODate(`${month}-01`)
  const dates: string[] = []
  while (monthOf(toISODate(d)) === month) {
    dates.push(toISODate(d))
    d.setDate(d.getDate() + 1)
  }
  return dates
}

export const weekday = (date: string) => (parseISODate(date).getDay() + 6) % 7

export const isWeekend = (date: string) => weekday(date) >= 5

export const dayOfMonth = (date: string) => parseISODate(date).getDate()

export const capitalize = (s: string) => s.charAt(0).toUpperCase() + s.slice(1)

export const formatDate = (date: string) => date.split('-').reverse().join('.')
