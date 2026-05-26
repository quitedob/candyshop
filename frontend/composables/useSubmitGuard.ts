/**
 * useSubmitGuard — 全局表单双提交保护组合式（L-4）
 *
 * 解决问题：
 *   - 用户网络慢时连续点击提交按钮，导致后端创建多份订单/询价/付款；
 *   - 部分页面已有 `isSubmitting` 局部状态，但实现各异，体验不一致。
 *
 * 设计：
 *   - guard(asyncFn) 包装任何返回 Promise 的提交动作，确保未完成前再次调用直接 short-circuit；
 *   - submitting 是响应式 boolean，可直接绑定到 :disabled；
 *   - error 暴露最近一次的错误对象，便于 UI 展示；
 *   - 与后端 Idempotency-Key 头互补：后端是最终防线，前端只防 UI 抖动。
 *
 * 用法：
 *   const submit = useSubmitGuard()
 *   const onClick = () => submit.guard(async () => { await api.post('/orders', body) })
 *   <button :disabled="submit.submitting.value">Save</button>
 */
import { ref, type Ref } from 'vue'

export interface SubmitGuard {
  submitting: Ref<boolean>
  error: Ref<unknown>
  /**
   * 包装异步操作；并发调用时第二次开始的调用立即解析为 undefined 而不会触发 fn。
   */
  guard: <T>(fn: () => Promise<T>) => Promise<T | undefined>
  /**
   * 重置错误状态。
   */
  reset: () => void
}

export function useSubmitGuard(): SubmitGuard {
  const submitting = ref(false)
  const error = ref<unknown>(null)

  const guard = async <T>(fn: () => Promise<T>): Promise<T | undefined> => {
    if (submitting.value) {
      return undefined
    }
    submitting.value = true
    error.value = null
    try {
      return await fn()
    } catch (e) {
      error.value = e
      throw e
    } finally {
      submitting.value = false
    }
  }

  const reset = () => {
    error.value = null
  }

  return { submitting, error, guard, reset }
}

/**
 * generateIdempotencyKey 给单次提交生成可复用的 Idempotency-Key 值，
 * 用于配合后端 M-17 中间件去重。每次组件挂载/打开表单调用一次即可，
 * 同一表单提交在重试链路上保持同一 key，但跨提交各自独立。
 */
export function generateIdempotencyKey(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return `idem-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}
