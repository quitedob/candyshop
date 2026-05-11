import { useI18n } from '#i18n'
import { useCurrency } from './useCurrency'

/**
 * 统一展示：默认货币、空单元格、枚举状态（common.enum.*），无硬编码英文兜底
 */
export function useDisplay() {
  const { t, te, locale } = useI18n()

  const currencyOrDefault = (code?: string | null) => {
    if (code != null && String(code).trim() !== '') return String(code).trim()
    try { return useCurrency().cur() } catch { return 'USD' }
  }

  /** 贸易术语空值默认 FOB */
  const incotermsOrDefault = (v?: string | null) =>
    v != null && String(v).trim() !== '' ? String(v).trim() : 'FOB'

  const cell = (value?: string | number | null | undefined) => {
    if (value === null || value === undefined) return t('common.display.em_dash')
    const s = String(value).trim()
    return s === '' ? t('common.display.em_dash') : s
  }

  /**
   * @param group common.enum 下的分组名，如 order_status
   * @param raw 后端原始值；空值时用 defaultKey（默认 pending / unpaid 等）
   */
  const enumLabel = (group: string, raw: string | null | undefined, defaultKey = 'unknown') => {
    const v0 = raw != null && String(raw).trim() !== '' ? String(raw) : defaultKey
    const v = v0.toLowerCase().replace(/[^a-z0-9_]/g, '_')
    const key = `common.enum.${group}.${v}`
    if (te(key)) return t(key)
    const unk = `common.enum.${group}.unknown`
    return te(unk) ? t(unk) : String(raw ?? '')
  }

  const formatNumber = (num: number | null | undefined, opts?: Intl.NumberFormatOptions) => {
    if (num == null || Number.isNaN(num)) return cell(null)
    return new Intl.NumberFormat(locale.value, { minimumFractionDigits: 2, maximumFractionDigits: 2, ...opts }).format(num)
  }

  const formatDate = (d: string | null | undefined, opts?: Intl.DateTimeFormatOptions) => {
    if (!d) return cell(null)
    const date = new Date(d)
    if (Number.isNaN(date.getTime())) return cell(null)
    return new Intl.DateTimeFormat(locale.value, opts ?? { dateStyle: 'medium' }).format(date)
  }

  return { currencyOrDefault, incotermsOrDefault, cell, enumLabel, formatNumber, formatDate }
}
