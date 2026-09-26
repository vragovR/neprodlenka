import { reactive } from 'vue'

const storageKey = 'login'

function readLogin(): string {
  try {
    return localStorage.getItem(storageKey) ?? ''
  } catch {
    return ''
  }
}

function writeLogin(login: string) {
  try {
    if (login) {
      localStorage.setItem(storageKey, login)
    } else {
      localStorage.removeItem(storageKey)
    }
  } catch {
    return
  }
}

export const session = reactive({
  login: readLogin(),
  needLogin: false,
  signIn(login: string) {
    this.login = login
    this.needLogin = false
    writeLogin(login)
  },
  signOut() {
    this.login = ''
    this.needLogin = true
    writeLogin('')
  },
  expire() {
    this.needLogin = true
  },
})
