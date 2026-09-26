function readDebug(): boolean {
  try {
    return localStorage.getItem('debug') !== null
  } catch {
    return false
  }
}

export const debugEnabled = readDebug()
