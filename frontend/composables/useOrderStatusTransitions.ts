/** 订单状态转换规则（镜像后端 ValidOrderStatusTransitions + 付款策略） */
const VALID_TRANSITIONS: Record<string, string[]> = {
  pending: ['confirmed', 'cancelled'],
  pending_confirmation: ['pending', 'confirmed', 'cancelled', 'expired'],
  pending_approval: ['pending_confirmation', 'cancelled'],
  confirmed: ['production', 'cancelled'],
  production: ['shipped', 'partially_shipped', 'cancelled'],
  partially_shipped: ['shipped', 'partially_delivered', 'delivered', 'partially_returned', 'returned'],
  shipped: ['delivered', 'partially_delivered', 'partially_returned', 'returned'],
  partially_delivered: ['delivered', 'partially_returned', 'returned'],
  delivered: ['returned', 'partially_returned'],
  partially_returned: [],
  returned: [],
  cancelled: [],
  expired: []
}

const EXECUTION_STATUSES = new Set(['confirmed', 'production', 'shipped', 'delivered'])

export function requiresPaidBeforeExecution(status: string): boolean {
  return EXECUTION_STATUSES.has(String(status || '').toLowerCase().trim())
}

export function isPaidOrPartial(paymentStatus: string): boolean {
  const ps = String(paymentStatus || '').toLowerCase().trim()
  return ps === 'paid' || ps === 'partial'
}

export function paymentInsufficientForExecution(targetStatus: string, paymentStatus: string): boolean {
  if (!requiresPaidBeforeExecution(targetStatus)) return false
  return !isPaidOrPartial(paymentStatus)
}

/** 当前状态下可转换的目标状态列表 */
export function allowedStatusTransitions(currentStatus: string): string[] {
  const cur = String(currentStatus || '').toLowerCase().trim()
  return VALID_TRANSITIONS[cur] ? [...VALID_TRANSITIONS[cur]] : []
}

/** 状态下拉选项：合法转换 + 当前状态；需付款的状态置灰不可选 */
export function statusSelectOptions(currentStatus: string, paymentStatus: string): { value: string; disabled: boolean; paymentBlocked: boolean }[] {
  const cur = String(currentStatus || '').toLowerCase().trim()
  const allowed = new Set(allowedStatusTransitions(cur))
  allowed.add(cur)
  return Object.keys(VALID_TRANSITIONS)
    .filter(s => allowed.has(s))
    .map(value => {
      const paymentBlocked = paymentInsufficientForExecution(value, paymentStatus)
      return {
        value,
        disabled: paymentBlocked && value !== cur,
        paymentBlocked,
      }
    })
}

export function useOrderStatusTransitions() {
  return {
    allowedStatusTransitions,
    statusSelectOptions,
    paymentInsufficientForExecution,
    requiresPaidBeforeExecution,
    isPaidOrPartial
  }
}
