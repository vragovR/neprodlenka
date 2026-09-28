import { ref } from 'vue'
import type { MarkChange, Status } from '@/api/types'

export function useMarks(saved: (date: string, childId: number) => Status | null) {
  const changes = ref(new Map<string, MarkChange>())

  const status = (date: string, childId: number): Status | null => {
    const change = changes.value.get(`${date}:${childId}`)
    return change ? change.status : saved(date, childId)
  }

  function toggle(date: string, childId: number) {
    const was = saved(date, childId)
    const current = status(date, childId)
    const value = current === 'present' ? 'absent' : current === 'absent' ? (was === 'pending' ? was : null) : 'present'
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
