export const today = '2026-09-25'
export const month = '2026-09'

export const septemberDates = Array.from({ length: 30 }, (_, i) => `2026-09-${String(i + 1).padStart(2, '0')}`)

export const isSchool = (date: string) => {
  const day = new Date(`${date}T00:00:00Z`).getUTCDay()
  return day !== 0 && day !== 6
}

export const weekends = (month: string) => {
  const [year = 0, mon = 0] = month.split('-').map(Number)
  const length = new Date(Date.UTC(year, mon, 0)).getUTCDate()
  return Array.from({ length }, (_, i) => `${month}-${String(i + 1).padStart(2, '0')}`).filter((d) => !isSchool(d))
}
