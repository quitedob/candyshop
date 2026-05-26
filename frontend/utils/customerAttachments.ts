/** 客户侧通用附件（OEM / 订单消息 / 询价 / 发货）允许的 MIME */
export const CUSTOMER_ATTACHMENT_MIME = [
  'application/pdf',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'application/vnd.ms-excel',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  'text/csv',
  'application/csv',
  'text/comma-separated-values',
  'image/png',
  'image/jpeg',
  'video/mp4',
] as const

/** 允许的扩展名 */
export const CUSTOMER_ATTACHMENT_EXTENSIONS = [
  '.pdf',
  '.doc',
  '.docx',
  '.xlsx',
  '.csv',
  '.png',
  '.jpeg',
  '.jpg',
  '.mp4',
] as const

export const CUSTOMER_ATTACHMENT_MAX_FILES = 5
export const CUSTOMER_ATTACHMENT_MAX_MB = 64
export const ORDER_MESSAGE_MAX_FILES = 3
export const SHIPMENT_ATTACHMENT_MAX_FILES = 5

export const CUSTOMER_ATTACHMENT_ACCEPT = [
  ...CUSTOMER_ATTACHMENT_MIME,
  ...CUSTOMER_ATTACHMENT_EXTENSIONS,
].join(',')

/** 判断是否为允许的客户附件 */
export function isCustomerAttachmentAllowed(file: File): boolean {
  if (file.type && (CUSTOMER_ATTACHMENT_MIME as readonly string[]).includes(file.type)) {
    return true
  }
  const match = file.name.toLowerCase().match(/\.[^.]+$/)
  if (!match) return false
  return (CUSTOMER_ATTACHMENT_EXTENSIONS as readonly string[]).includes(match[0])
}

/** 付款凭证：仅 PDF / JPG / PNG */
export const PAYMENT_PROOF_MIME = [
  'application/pdf',
  'image/jpeg',
  'image/png',
] as const

export const PAYMENT_PROOF_EXTENSIONS = ['.pdf', '.jpg', '.jpeg', '.png'] as const
export const PAYMENT_PROOF_MAX_MB = 10

export const PAYMENT_PROOF_ACCEPT = [
  ...PAYMENT_PROOF_MIME,
  ...PAYMENT_PROOF_EXTENSIONS,
].join(',')

export function isPaymentProofAllowed(file: File): boolean {
  if (file.type && (PAYMENT_PROOF_MIME as readonly string[]).includes(file.type)) {
    return true
  }
  const match = file.name.toLowerCase().match(/\.[^.]+$/)
  if (!match) return false
  return (PAYMENT_PROOF_EXTENSIONS as readonly string[]).includes(match[0])
}

// 兼容 OEM 页面旧引用
export const OEM_ALLOWED_MIME_TYPES = CUSTOMER_ATTACHMENT_MIME
export const OEM_ALLOWED_EXTENSIONS = CUSTOMER_ATTACHMENT_EXTENSIONS
export const OEM_MAX_FILES = CUSTOMER_ATTACHMENT_MAX_FILES
export const OEM_MAX_FILE_MB = CUSTOMER_ATTACHMENT_MAX_MB
export const OEM_ACCEPT_ATTR = CUSTOMER_ATTACHMENT_ACCEPT
export const isOemAttachmentAllowed = isCustomerAttachmentAllowed
