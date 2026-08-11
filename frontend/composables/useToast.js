/**
 * Toast Notification Composable
 * Visual + Audio notifications for B2B candy OEM platform
 */

const toastState = reactive({
  visible: false,
  message: '',
  type: 'info'
})

// Single hide timer owned by the currently-visible toast. Cleared whenever a new toast
// is shown (or the current one is hidden) so an older toast's timer can never hide a
// newer, longer-lived toast — previously each show() leaked a timer that would fire
// later and dismiss whatever was on screen.
let hideTimer = null

export const useToast = () => {
  const tts = useTTS()

  const show = (options) => {
    // Clear any pending hide timer from a previous toast before showing the new one,
    // including for duration === 0 (persist) which must outlive earlier timers.
    if (hideTimer) {
      clearTimeout(hideTimer)
      hideTimer = null
    }

    toastState.message = options.message
    toastState.type = options.type || 'info'
    toastState.visible = true

    // Play TTS sound for this toast type
    tts.playForToast(toastState.type)

    if (options.duration !== 0) {
      hideTimer = setTimeout(() => {
        hide()
      }, options.duration || 3000)
    }
  }

  const hide = () => {
    toastState.visible = false
    if (hideTimer) {
      clearTimeout(hideTimer)
      hideTimer = null
    }
  }

  const success = (message, duration) => {
    show({ message, type: 'success', duration })
  }

  const error = (message, duration) => {
    show({ message, type: 'error', duration })
  }

  const warning = (message, duration) => {
    show({ message, type: 'warning', duration })
  }

  const info = (message, duration) => {
    show({ message, type: 'info', duration })
  }

  return {
    state: readonly(toastState),
    show,
    hide,
    success,
    error,
    warning,
    info
  }
}
