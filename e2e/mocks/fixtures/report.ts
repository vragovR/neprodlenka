import type { ChildReport, Report, ReportRow } from '@/api/types'
import { isSchool, month, septemberDates } from './dates'

export const reportUID = 'r-sep'

const days = septemberDates.map((date) => ({ date, school: isSchool(date) }))

const petyaRow: ReportRow = {
  id: 1,
  last_name: 'Иванов',
  first_name: 'Петя',
  middle_name: '',
  marks: septemberDates.map((d) => (isSchool(d) && d < '2026-09-05' ? 'present' : null)),
  present: 3,
  absent: 0,
  total: 39000,
}

export const report: Report = {
  month,
  days,
  children: [petyaRow],
  total: 39000,
}

export const childReport: ChildReport = {
  month,
  days,
  child: petyaRow,
  price: 13000,
  prev_uid: 'r-aug',
  next_uid: '',
}
