/**
 * Toast Notification Composable
 * Visual + Audio notifications for B2B candy OEM platform
 */

const toastState = reactive({
  visible: false,
  message: '',
  type: 'info'
})

export const useToast = () => {
  const tts = useTTS()

  const show = (options) => {
    toastState.message = options.message
    toastState.type = options.type || 'info'
    toastState.visible = true

    // Play TTS sound for this toast type
    tts.playForToast(toastState.type)

    if (options.duration !== 0) {
      setTimeout(() => {
        hide()
      }, options.duration || 3000)
    }
  }

  const hide = () => {
    toastState.visible = false
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
