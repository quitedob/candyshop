/**
 * useOrderMessageStream
 *
 * Subscribes to an order conversation's real-time feed over WebSocket, with
 * automatic reconnect and a polling fallback. The fallback guarantees the
 * thread still updates if the WebSocket upgrade is unavailable (e.g. a proxy
 * that does not forward the Upgrade header), so this never regresses behaviour
 * versus the previous 30s polling implementation.
 *
 * Auth is the HttpOnly `auth_token` cookie, which the browser sends
 * automatically on same-origin WebSocket handshakes — no token in the URL.
 *
 * @param path     WS endpoint relative to apiBase, e.g. `/user/orders/123/messages/ws`.
 * @param onEvent  Receives each parsed event envelope `{ type, payload }`.
 * @param refetch  Full REST refetch, used by the polling fallback and on reconnect.
 * @param opts     Optional tuning (poll interval).
 */
export function useOrderMessageStream(
  path: string,
  onEvent: (event: { type: string; payload?: any }) => void,
  refetch: () => Promise<void> | void,
  opts: { pollIntervalMs?: number } = {},
) {
  const config = useRuntimeConfig()
  const apiBase = (config.public.apiBase as string) || '/api/v1'

  const connected = ref(false)
  let ws: WebSocket | null = null
  let pollTimer: ReturnType<typeof setInterval> | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectAttempts = 0
  let closedByUs = false

  const buildWsUrl = (): string | null => {
    if (typeof window === 'undefined') return null
    const full = `${apiBase}${path}`
    // apiBase may be absolute (http(s)://...) or relative (/api/v1).
    if (/^https?:\/\//i.test(full)) {
      return full.replace(/^http/i, 'ws')
    }
    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${proto}//${window.location.host}${full}`
  }

  const startPolling = () => {
    if (pollTimer) return
    const interval = opts.pollIntervalMs ?? 30000
    pollTimer = setInterval(() => { void refetch() }, interval)
  }

  const stopPolling = () => {
    if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
  }

  const scheduleReconnect = () => {
    if (closedByUs || reconnectTimer) return
    // Exponential backoff capped at 30s.
    const delay = Math.min(30000, 1000 * 2 ** reconnectAttempts)
    reconnectAttempts += 1
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      connect()
    }, delay)
  }

  const connect = () => {
    if (typeof window === 'undefined' || ws) return
    closedByUs = false

    const url = buildWsUrl()
    if (!url) { startPolling(); return }

    try {
      ws = new WebSocket(url)
    } catch {
      // Construction failed outright — rely on polling and retry.
      startPolling()
      scheduleReconnect()
      return
    }

    ws.onopen = () => {
      connected.value = true
      reconnectAttempts = 0
      stopPolling()
      // Resync anything missed in the gap before the socket opened.
      void refetch()
    }

    ws.onmessage = (ev) => {
      try {
        const evt = JSON.parse(ev.data)
        if (evt && typeof evt.type === 'string') onEvent(evt)
      } catch { /* ignore malformed frames */ }
    }

    ws.onclose = () => {
      connected.value = false
      ws = null
      if (!closedByUs) {
        startPolling()
        scheduleReconnect()
      }
    }

    ws.onerror = () => {
      // onclose fires next and handles fallback/reconnect.
      try { ws?.close() } catch { /* noop */ }
    }
  }

  const stop = () => {
    closedByUs = true
    stopPolling()
    if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null }
    if (ws) {
      try { ws.close() } catch { /* noop */ }
      ws = null
    }
    connected.value = false
  }

  onUnmounted(stop)

  return { connected, connect, stop }
}
