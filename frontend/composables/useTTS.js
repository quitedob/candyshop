/**
 * TTS (Text-to-Speech) Composable
 * Plays pre-generated audio notifications for B2B candy OEM platform
 */

// Map toast types to TTS audio files
const toastToTTSMap = {
  success: 'payment_success',
  error: 'payment_failed',
  warning: 'inquiry_received',
  info: 'support_available',
}

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
      play('greeting_morning')
    } else if (hour < 18) {
      play('greeting_afternoon')
    } else {
      play('greeting_evening')
    }
  }

  // Play cart-specific TTS
  const playCartWelcome = (isZh = false) => {
    play(isZh ? 'cart_welcome_zh' : 'cart_welcome')
  }

  const playCartEmpty = (isZh = false) => {
    play(isZh ? 'cart_empty_zh' : 'cart_empty')
  }

  // Play order-specific TTS
  const playOrderPlaced = (isZh = false) => {
    play(isZh ? 'order_placed_zh' : 'order_placed')
  }

  const playOrderConfirmed = () => {
    play('order_confirmed')
  }

  // Play payment TTS
  const playPaymentSuccess = (isZh = false) => {
    play(isZh ? 'payment_success_zh' : 'payment_success')
  }

  const playPaymentFailed = (isZh = false) => {
    play(isZh ? 'payment_failed_zh' : 'payment_failed')
  }

  // Play shipping TTS
  const playShippingDispatched = (isZh = false) => {
    play(isZh ? 'shipping_dispatched_zh' : 'shipping_dispatched')
  }

  // Play OEM project TTS
  const playOEMProjectUpdate = (isZh = false) => {
    play(isZh ? 'oem_project_update_zh' : 'oem_project_update')
  }

  // Play inquiry TTS
  const playInquiryReceived = (isZh = false) => {
    play(isZh ? 'inquiry_received_zh' : 'inquiry_received')
  }

  const playInquiryThanks = () => {
    play('inquiry_thanks')
  }

  // Play account TTS
  const playAccountApproved = (isZh = false) => {
    play(isZh ? 'account_approved_zh' : 'account_approved')
  }

  // Play factory tour TTS
  const playFactoryTour = () => {
    play('factory_tour')
  }

  const playQualityAssurance = () => {
    play('quality_assurance')
  }

  const playProductionCapacity = () => {
    play('production_capacity')
  }

  const playCertifications = () => {
    play('certifications')
  }

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
