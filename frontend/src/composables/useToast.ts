import { reactive } from 'vue'

export type ToastTone = 'success' | 'danger'

export interface Toast {
  id: number
  text: string
  tone: ToastTone
}

let nextId = 1

export const toasts = reactive<Toast[]>([])

export function dismissToast(id: number) {
  const i = toasts.findIndex((t) => t.id === id)
  if (i >= 0) {
    toasts.splice(i, 1)
  }
}

export function toast(text: string, tone: ToastTone = 'success') {
  const id = nextId++
  toasts.push({ id, text, tone })
  setTimeout(() => dismissToast(id), 4000)
}
