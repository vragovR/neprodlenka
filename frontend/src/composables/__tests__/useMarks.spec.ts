import { describe, expect, it } from 'vite-plus/test'
import type { Status } from '@/api/types'
import { nextStatus, useMarks } from '../useMarks'

describe('nextStatus', () => {
  it('cycles present → absent → none → present', () => {
    expect(nextStatus('present', null)).toBe('absent')
    expect(nextStatus('absent', null)).toBeNull()
    expect(nextStatus(null, null)).toBe('present')
  })

  it('returns to pending when it was saved', () => {
    expect(nextStatus('pending', 'pending')).toBe('present')
    expect(nextStatus('absent', 'pending')).toBe('pending')
  })

  it('toggles only absent on future days', () => {
    expect(nextStatus(null, null, true)).toBe('absent')
    expect(nextStatus('absent', null, true)).toBeNull()
  })
})

describe('useMarks', () => {
  const saved: Record<string, Status> = { '2026-09-01:1': 'present', '2026-09-01:2': 'pending' }
  const marks = () => useMarks((date, childId) => saved[`${date}:${childId}`] ?? null)

  it('keeps a change until the saved status comes back', () => {
    const { changes, status, toggle } = marks()

    toggle('2026-09-01', 1)
    expect(status('2026-09-01', 1)).toBe('absent')
    expect([...changes.value.values()]).toEqual([
      { date: '2026-09-01', child_id: 1, status: 'absent' },
    ])

    toggle('2026-09-01', 1)
    expect(status('2026-09-01', 1)).toBeNull()
    expect(changes.value.get('2026-09-01:1')?.status).toBeNull()

    toggle('2026-09-01', 1)
    expect(status('2026-09-01', 1)).toBe('present')
    expect(changes.value.size).toBe(0)
  })

  it('cycles pending through present and absent back to pending', () => {
    const { changes, status, toggle } = marks()

    toggle('2026-09-01', 2)
    toggle('2026-09-01', 2)
    expect(status('2026-09-01', 2)).toBe('absent')
    toggle('2026-09-01', 2)
    expect(status('2026-09-01', 2)).toBe('pending')
    expect(changes.value.size).toBe(0)
  })

  it('resets changes', () => {
    const { changes, status, toggle, reset } = marks()

    toggle('2026-09-02', 1)
    reset()
    expect(changes.value.size).toBe(0)
    expect(status('2026-09-02', 1)).toBeNull()
  })
})
