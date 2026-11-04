import type { Club } from '@/api/types'
import { weekdaysShort } from '@/utils/format'

export interface ClubGroup {
  name: string
  start: string
  weekdays: number[]
}

export const groupKey = (g: Pick<Club, 'name' | 'start'>) => `${g.name}\n${g.start}`

export function groupClubs(clubs: Club[]): ClubGroup[] {
  const groups = new Map<string, ClubGroup>()
  for (const c of clubs) {
    const group = groups.get(groupKey(c)) ?? { name: c.name, start: c.start, weekdays: [] }
    group.weekdays.push(c.weekday)
    groups.set(groupKey(c), group)
  }
  return [...groups.values()]
    .map((g) => ({ ...g, weekdays: [...g.weekdays].sort((a, b) => a - b) }))
    .sort(
      (a, b) =>
        (a.weekdays[0] ?? 0) - (b.weekdays[0] ?? 0) ||
        a.start.localeCompare(b.start) ||
        a.name.localeCompare(b.name, 'ru'),
    )
}

export const expandGroup = (g: ClubGroup): Club[] =>
  g.weekdays.map((weekday) => ({ name: g.name, start: g.start, weekday }))

export const groupLabel = (g: ClubGroup) =>
  `${g.name} · ${g.weekdays.map((d) => weekdaysShort[d - 1]).join(', ')} ${g.start}`
