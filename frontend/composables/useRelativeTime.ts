/**
 * Locale-aware relative time formatting for admin dashboards.
 */
export function useRelativeTime(prefix = 'admin.dashboard', options?: { plainKeys?: boolean }) {
  const { t } = useI18n()
  const { formatDate } = useDisplay()

  const relKey = (name: string) => {
    if (options?.plainKeys) return `${prefix}.${name}`
    return `${prefix}.time_${name}`
  }

  const formatRelativeTime = (input?: string | Date | null) => {
    if (!input) return options?.plainKeys ? t('display.em_dash') : t(relKey('unknown'))
    const date = input instanceof Date ? input : new Date(input)
    if (Number.isNaN(date.getTime())) return options?.plainKeys ? t('display.em_dash') : t(relKey('unknown'))

    const diffMs = Date.now() - date.getTime()
    const diffSec = Math.floor(diffMs / 1000)
    const diffMin = Math.floor(diffSec / 60)
    const diffHour = Math.floor(diffMin / 60)
    const diffDay = Math.floor(diffHour / 24)

    if (diffSec < 60) return t(relKey('just_now'))
    if (diffMin < 60) return t(relKey('minutes_ago'), { count: diffMin })
    if (diffHour < 24) return t(relKey('hours_ago'), { count: diffHour })
    if (diffDay < 30) return t(relKey('days_ago'), { count: diffDay })
    return formatDate(date.toISOString())
  }

  return { formatRelativeTime }
}
