import type { Attendance, AttendanceRow, Status } from '@/api/types'
import { isSchool, month, septemberDates, today } from './dates'

const published = septemberDates.filter((d) => isSchool(d) && d <= today)

function row(
  childId: number,
  lastName: string,
  firstName: string,
  price: number,
  special: Record<string, Status>,
  paid = false,
  weekdays = [1, 2, 3, 4, 5],
): AttendanceRow {
  const marks = Object.fromEntries(published.map((d) => [d, special[d] ?? 'present']))
  const present = Object.values(marks).filter((s) => s === 'present').length
  return {
    child_id: childId,
    last_name: lastName,
    first_name: firstName,
    middle_name: '',
    active: true,
    weekdays,
    marks,
    total: present * price,
    paid,
  }
}

export const attendance: Attendance = {
  month,
  days: septemberDates.map((date) => ({
    date,
    school: isSchool(date),
    published: isSchool(date) && date <= today,
    closed: isSchool(date) && date < today,
  })),
  children: [
    row(1, 'Иванов', 'Петя', 13000, { '2026-09-10': 'absent', [today]: 'pending' }),
    row(2, 'Козлова', 'Аня', 15000, { [today]: 'absent' }, true, [2, 4]),
  ],
  message: { text: 'Питание за сентябрь 2026.', buttons: [] },
  sent: false,
}
