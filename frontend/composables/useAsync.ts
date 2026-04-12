import type { Ref } from 'vue'

interface AsyncState<T> {
  data: Ref<T | null>
  error: Ref<Error | null>
  loading: Ref<boolean>
  execute: () => Promise<T | null>
}

export const useAsync = <T>(
  asyncFn: () => Promise<T>,
  immediate = true
): AsyncState<T> => {
  const data = ref<T | null>(null) as Ref<T | null>
  const error = ref<Error | null>(null) as Ref<Error | null>
  const loading = ref(false)

  const execute = async (): Promise<T | null> => {
    loading.value = true
    error.value = null

    try {
      const result = await asyncFn()
      data.value = result
      return result
    } catch (e) {
      error.value = e instanceof Error ? e : new Error(String(e))
      return null
    } finally {
      loading.value = false
    }
  }

  if (immediate) {
    execute()
  }

  return {
    data,
    error,
    loading,
    execute
  }
}
