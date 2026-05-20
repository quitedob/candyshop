/**
 * Composable for locale-aware product field access.
 * Falls back to the original field value when no translation exists.
 */
export function useTranslation() {
  const { locale } = useI18n()

  /** Returns the translated value of a product field, or the original field value. */
  const tField = (item: any, field: string): string => {
    const trans = item?.translations?.[locale.value]?.[field]
    if (trans != null && String(trans).trim() !== '') return String(trans)
    return item?.[field] ?? ''
  }

  /** Returns the translated array of a product field, or the original array value. */
  const tArray = (item: any, field: string): string[] => {
    const trans = item?.translations?.[locale.value]?.[field]
    if (Array.isArray(trans) && trans.length > 0) return trans
    return item?.[field] ?? []
  }

  return { tField, tArray }
}
