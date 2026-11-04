import { ref, watch } from 'vue'

export type ViewMode = 'calendar' | 'table'

const storageKey = 'view'

function readViewMode(): ViewMode {
  try {
    return localStorage.getItem(storageKey) === 'table' ? 'table' : 'calendar'
  } catch {
    return 'calendar'
  }
}

export const viewMode = ref<ViewMode>(readViewMode())

watch(viewMode, (v) => {
  try {
    localStorage.setItem(storageKey, v)
  } catch {
    return
  }
})
