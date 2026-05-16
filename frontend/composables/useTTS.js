/**
 * TTS (Text-to-Speech) Composable
 * Plays pre-generated audio notifications for B2B candy OEM platform.
 * Locale-aware: plays language-specific audio when available, falls back to English.
 */

// Map toast types to TTS audio files
const toastToTTSMap = {
  success: 'payment_success',
  error: 'payment_failed',
  warning: 'inquiry_received',
  info: 'support_available',
}

const LOCALES_WITH_AUDIO = ['zh']

let currentAudio = null
const audioEnabled = ref(true)
const volume = ref(0.7)

export const useTTS = () => {
  const play = (type) => {
    if (!audioEnabled.value) return

    // Stop any currently playing audio
    if (currentAudio) {
      currentAudio.pause()
      currentAudio.currentTime = 0
    }

    const audioPath = `/audio/${type}.mp3`
    currentAudio = new Audio(audioPath)
    currentAudio.volume = volume.value

    currentAudio.play().catch((err) => {
      console.warn(`TTS playback failed for ${type}:`, err)
    })

    currentAudio.onended = () => {
      currentAudio = null
    }
  }

  const stop = () => {
    if (currentAudio) {
      currentAudio.pause()
      currentAudio.currentTime = 0
      currentAudio = null
    }
  }

  const setVolume = (v) => {
    volume.value = Math.max(0, Math.min(1, v))
    if (currentAudio) {
      currentAudio.volume = volume.value
    }
  }

  const toggle = () => {
    audioEnabled.value = !audioEnabled.value
  }

  /** Resolve a locale-specific audio name, falling back to the default (English) audio. */
  const localizedAudio = (baseName) => {
    try {
      const { locale } = useI18n()
      if (LOCALES_WITH_AUDIO.includes(locale.value)) {
        return `${baseName}_${locale.value}`
      }
    } catch { /* useI18n not available outside setup — use default */ }
    return baseName
  }

  // Play TTS based on toast type
  const playForToast = (toastType) => {
    const ttsType = toastToTTSMap[toastType]
    if (ttsType) {
      play(ttsType)
    }
  }

  // Play welcome message based on time of day
  const playWelcome = () => {
    const hour = new Date().getHours()
    if (hour < 12) {
      play(localizedAudio('greeting_morning'))
    } else if (hour < 18) {
      play(localizedAudio('greeting_afternoon'))
    } else {
      play(localizedAudio('greeting_evening'))
    }
  }

  // Play cart-specific TTS
  const playCartWelcome = () => { play(localizedAudio('cart_welcome')) }
  const playCartEmpty = () => { play(localizedAudio('cart_empty')) }

  // Play order-specific TTS
  const playOrderPlaced = () => { play(localizedAudio('order_placed')) }
  const playOrderConfirmed = () => { play('order_confirmed') }

  // Play payment TTS
  const playPaymentSuccess = () => { play(localizedAudio('payment_success')) }
  const playPaymentFailed = () => { play(localizedAudio('payment_failed')) }

  // Play shipping TTS
  const playShippingDispatched = () => { play(localizedAudio('shipping_dispatched')) }

  // Play OEM project TTS
  const playOEMProjectUpdate = () => { play(localizedAudio('oem_project_update')) }

  // Play inquiry TTS
  const playInquiryReceived = () => { play(localizedAudio('inquiry_received')) }
  const playInquiryThanks = () => { play('inquiry_thanks') }

  // Play account TTS
  const playAccountApproved = () => { play(localizedAudio('account_approved')) }

  // Play factory tour TTS
  const playFactoryTour = () => { play('factory_tour') }
  const playQualityAssurance = () => { play('quality_assurance') }
  const playProductionCapacity = () => { play('production_capacity') }
  const playCertifications = () => { play('certifications') }

  return {
    // State
    audioEnabled: readonly(audioEnabled),
    volume: readonly(volume),
    // Methods
    play,
    stop,
    setVolume,
    toggle,
    playForToast,
    playWelcome,
    playCartWelcome,
    playCartEmpty,
    playOrderPlaced,
    playOrderConfirmed,
    playPaymentSuccess,
    playPaymentFailed,
    playShippingDispatched,
    playOEMProjectUpdate,
    playInquiryReceived,
    playInquiryThanks,
    playAccountApproved,
    playFactoryTour,
    playQualityAssurance,
    playProductionCapacity,
    playCertifications,
  }
}
