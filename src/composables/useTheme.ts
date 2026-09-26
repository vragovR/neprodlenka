import { ref, watch } from 'vue'

export type Theme = 'light' | 'dark' | 'system'

const storageKey = 'theme'

function readTheme(): Theme {
  try {
    const t = localStorage.getItem(storageKey)
    return t === 'light' || t === 'dark' ? t : 'system'
  } catch {
    return 'system'
  }
}

function writeTheme(t: Theme) {
  try {
    if (t === 'system') {
      localStorage.removeItem(storageKey)
    } else {
      localStorage.setItem(storageKey, t)
    }
  } catch {
    return
  }
}

export const theme = ref<Theme>(readTheme())

watch(
  theme,
  (t) => {
    const root = document.documentElement
    if (t === 'system') {
      delete root.dataset.theme
    } else {
      root.dataset.theme = t
    }
    writeTheme(t)
  },
  { immediate: true },
)
