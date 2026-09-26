import { session } from '@/composables/useSession'

export class ApiError extends Error {
  readonly status: number
  readonly code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.status = status
    this.code = code
  }
}

export async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  let res: Response
  try {
    res = await fetch(`/api${path}`, {
      method,
      credentials: 'same-origin',
      headers: body === undefined ? undefined : { 'Content-Type': 'application/json' },
      body: body === undefined ? undefined : JSON.stringify(body),
    })
  } catch {
    throw new ApiError(0, 'network', 'Нет связи с сервером')
  }

  if (res.status === 401 && path !== '/login') {
    session.expire()
  }
  if (!res.ok) {
    const err = await res.json().catch(() => null)
    throw new ApiError(res.status, err?.error ?? 'unknown', err?.message ?? 'Ошибка сервера')
  }
  if (res.status === 204) {
    return undefined as T
  }
  return res.json()
}

export function errorMessage(e: unknown): string {
  return e instanceof ApiError ? e.message : 'Что-то пошло не так'
}
