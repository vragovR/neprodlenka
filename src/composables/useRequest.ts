import { ref, shallowRef } from 'vue'
import { ApiError } from '@/api/client'

export function useRequest<T>(fn: () => Promise<T>) {
  const data = shallowRef<T>()
  const error = shallowRef<ApiError | null>(null)
  const loading = ref(false)

  async function run() {
    loading.value = true
    error.value = null
    try {
      data.value = await fn()
    } catch (e) {
      error.value = e instanceof ApiError ? e : new ApiError(0, 'unknown', 'Что-то пошло не так')
    } finally {
      loading.value = false
    }
  }

  return { data, error, loading, run }
}
