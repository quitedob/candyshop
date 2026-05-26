/** 从 CSS 变量读取 Chart.js 主题色 */
export function useChartTheme() {
  const cssVar = (name: string, fallback: string) => {
    if (import.meta.client) {
      const v = getComputedStyle(document.documentElement).getPropertyValue(name).trim()
      if (v) return v
    }
    return fallback
  }

  const colors = computed(() => ({
    highlight: cssVar('--color-highlight', '#fd933d'),
    accent: cssVar('--color-accent', '#944a00'),
    success: cssVar('--color-success', '#22c55e'),
    warning: cssVar('--color-warning', '#f59e0b'),
    error: cssVar('--color-error', '#ba1a1a'),
    info: cssVar('--color-info', '#3b82f6'),
    muted: cssVar('--color-text-lighter', '#7e7570'),
    grid: cssVar('--color-outline-variant', '#d0c4be'),
  }))

  const rgba = (hex: string, alpha: number) => {
    const h = hex.replace('#', '')
    if (h.length !== 6) return hex
    const r = parseInt(h.slice(0, 2), 16)
    const g = parseInt(h.slice(2, 4), 16)
    const b = parseInt(h.slice(4, 6), 16)
    return `rgba(${r}, ${g}, ${b}, ${alpha})`
  }

  return { colors, rgba }
}
