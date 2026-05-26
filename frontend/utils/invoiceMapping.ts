/**
 * 发票字段映射工具
 *
 * 后端 Invoice 模型字段（实际）：
 *   - invoiceNo      (发票编号)
 *   - createdAt      (创建/开票日期)
 *   - paidAt         (已付款时间戳，仅在状态=paid 时填充)
 *   - totalAmount    (总额)
 *
 * 前端 UI 历史使用的字段名（不存在于后端）：
 *   - invoiceNumber  → 实际为 invoiceNo
 *   - invoiceDate    → 实际为 createdAt
 *   - paidAmount     → 实际无独立字段；以 totalAmount 推导（仅 status=paid 时）
 *
 * 该模块用于在前端层做字段对齐，避免在后端模型上引入冗余列。
 *
 * 单一进入点：normalizeInvoice() / normalizeInvoiceList()。所有 UI 仅消费规范化后的对象。
 */

export interface RawInvoice {
  id?: string
  invoiceNo?: string
  invoiceNumber?: string
  type?: string
  status?: string
  amount?: number
  taxAmount?: number
  totalAmount?: number
  currency?: string
  notes?: string
  items?: string
  orderId?: string
  orderNumber?: string
  customerName?: string
  customerEmail?: string
  paidAt?: string | null
  sentAt?: string | null
  dueDate?: string | null
  createdAt?: string | null
  updatedAt?: string | null
  invoiceDate?: string | null
  paidAmount?: number
  [key: string]: unknown
}

export interface NormalizedInvoice extends RawInvoice {
  invoiceNumber: string
  invoiceDate: string
  paidAmount: number
}

/**
 * 把后端裸 Invoice 对象规范化为 UI 期望的形状。
 *
 * - invoiceNumber：优先取 invoiceNo（后端真实字段），否则回退到老字段名或 id。
 * - invoiceDate：优先 invoiceDate（兼容历史），否则取 createdAt。
 * - paidAmount：若已是数字直接保留；否则当 status=paid 时显示 totalAmount，其它状态显示 0。
 *   这里的 "paid" 推导是一种近似（无部分付款追踪），等后端引入独立字段后可下线该映射。
 */
export function normalizeInvoice<T extends RawInvoice>(invoice: T | null | undefined): NormalizedInvoice {
  const inv = (invoice ?? {}) as RawInvoice
  const invoiceNumber = String(inv.invoiceNo ?? inv.invoiceNumber ?? inv.id ?? '')
  const invoiceDate = String(inv.invoiceDate ?? inv.createdAt ?? '')
  let paidAmount: number
  if (typeof inv.paidAmount === 'number') {
    paidAmount = inv.paidAmount
  } else if ((inv.status ?? '').toLowerCase() === 'paid') {
    paidAmount = Number(inv.totalAmount ?? inv.amount ?? 0) || 0
  } else {
    paidAmount = 0
  }
  return {
    ...inv,
    invoiceNumber,
    invoiceDate,
    paidAmount,
  }
}

export function normalizeInvoiceList<T extends RawInvoice>(invoices: T[] | null | undefined): NormalizedInvoice[] {
  if (!Array.isArray(invoices)) return []
  return invoices.map((inv) => normalizeInvoice(inv))
}

/**
 * 安全提取 ISO 日期前缀（YYYY-MM-DD），用于 <input type="date"> 双向绑定。
 * 输入为空/无效时返回空字符串而非崩溃（修复 split('T')[0] 在 undefined 上抛异常的问题）。
 */
export function isoDatePrefix(value: string | null | undefined): string {
  if (!value || typeof value !== 'string') return ''
  const idx = value.indexOf('T')
  return idx >= 0 ? value.slice(0, idx) : value
}
