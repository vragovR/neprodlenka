export const today = '2026-09-25'
export const month = '2026-09'

export const septemberDates = Array.from({ length: 30 }, (_, i) => `2026-09-${String(i + 1).padStart(2, '0')}`)

export const isSchool = (date: string) => {
  const day = new Date(`${date}T00:00:00Z`).getUTCDay()
  return day !== 0 && day !== 6
}
