/** 询价附件允许的 MIME 类型 */
export const INQUIRY_ALLOWED_MIME_TYPES = [
  'application/pdf',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'text/csv',
  'application/csv',
  'text/comma-separated-values',
  'application/vnd.ms-excel',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  'image/png',
  'image/jpeg',
  'image/jpg',
  'video/x-matroska',
  'video/mp4'
] as const

/** 询价附件允许的扩展名（浏览器 MIME 识别不准时的兜底） */
export const INQUIRY_ALLOWED_EXTENSIONS = [
  '.pdf',
  '.doc',
  '.docx',
  '.csv',
  '.xlsx',
  '.png',
  '.jpeg',
  '.jpg',
  '.mkv',
  '.mp4'
] as const

export const INQUIRY_MAX_FILES = 5
export const INQUIRY_MAX_FILE_MB = 64

/** input accept 属性：MIME + 扩展名 */
export const INQUIRY_ACCEPT_ATTR = [
  ...INQUIRY_ALLOWED_MIME_TYPES,
  ...INQUIRY_ALLOWED_EXTENSIONS
].join(',')

/** 判断文件是否为允许的询价附件 */
export function isInquiryAttachmentAllowed(file: File): boolean {
  if (file.type && (INQUIRY_ALLOWED_MIME_TYPES as readonly string[]).includes(file.type)) {
    return true
  }
  const match = file.name.toLowerCase().match(/\.[^.]+$/)
  if (!match) return false
  return (INQUIRY_ALLOWED_EXTENSIONS as readonly string[]).includes(match[0])
}
