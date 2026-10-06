import { ref } from 'vue'
import type { MarkChange, Status } from '@/api/types'
import { today } from '@/utils/format'

export const nextStatus = (current: Status | null, saved: Status | null, future = false): Status | null =>
  future
    ? current === 'absent' ? null : 'absent'
    : current === 'present' ? 'absent' : current === 'absent' ? (saved === 'pending' ? saved : null) : 'present'

export function useMarks(saved: (date: string, childId: number) => Status | null) {
  const changes = ref(new Map<string, MarkChange>())

  const status = (date: string, childId: number): Status | null => {
    const change = changes.value.get(`${date}:${childId}`)
    return change ? change.status : saved(date, childId)
  }

  function toggle(date: string, childId: number) {
    const was = saved(date, childId)
    const value = nextStatus(status(date, childId), was, date > today())
    const key = `${date}:${childId}`
    if (value === was) {
      changes.value.delete(key)
    } else {
      changes.value.set(key, { date, child_id: childId, status: value })
    }
  }

  const reset = () => (changes.value = new Map())

  return { changes, status, toggle, reset }
}
