/**
 * WhatsApp Composable
 * Generates WhatsApp links and handles WhatsApp-related functionality
 */

interface WhatsAppOptions {
  phone?: string
  message?: string
  product?: string
  category?: string
}

interface WhatsAppContactInfo {
  phone: string
  message: string
  fullName?: string
  company?: string
  email?: string
  country?: string
  product?: string
  quantity?: string
}

export const useWhatsApp = () => {
  const config = useRuntimeConfig()
  const { t } = useI18n()

  /**
   * Get the configured WhatsApp number
   */
  const getWhatsAppNumber = (): string => {
    return config.public.whatsappNumber ? String(config.public.whatsappNumber) : '1234567890'
  }

  /**
   * Generate a WhatsApp chat link
   */
  const createWhatsAppLink = (options: WhatsAppOptions = {}): string => {
    const phone = options.phone || getWhatsAppNumber()

    // Build message
    let message = options.message || t('whatsapp.message')

    // Add product info if provided
    if (options.product) {
      const categoryPart = options.category ? t('whatsapp.category_part', { name: options.category }) : ''
      message = t('whatsapp.product_inquiry', { product: options.product, categoryPart })
    }

    const encodedMessage = encodeURIComponent(message)

    return `https://wa.me/${phone}?text=${encodedMessage}`
  }

  /**
   * Generate a WhatsApp link with detailed inquiry info
   */
  const createInquiryLink = (info: WhatsAppContactInfo): string => {
    const phone = info.phone || getWhatsAppNumber()

    let message = t('whatsapp.inquiry_intro')

    if (info.fullName) message += `${t('whatsapp.field_name')}: ${info.fullName}\n`
    if (info.company) message += `${t('whatsapp.field_company')}: ${info.company}\n`
    if (info.email) message += `${t('whatsapp.field_email')}: ${info.email}\n`
    if (info.country) message += `${t('whatsapp.field_country')}: ${info.country}\n`
    if (info.product) message += `${t('whatsapp.field_product')}: ${info.product}\n`
    if (info.quantity) message += `${t('whatsapp.field_quantity')}: ${info.quantity}\n`

    message += `\n${info.message || t('whatsapp.inquiry_followup')}`

    return `https://wa.me/${phone}?text=${encodeURIComponent(message)}`
  }

  /**
   * Open WhatsApp in a new tab
   */
  const openWhatsApp = (options: WhatsAppOptions = {}) => {
    const link = createWhatsAppLink(options)
    window.open(link, '_blank', 'noopener,noreferrer')
  }

  /**
   * Open WhatsApp with product inquiry
   */
  const inquireProduct = (productName: string, category?: string) => {
    openWhatsApp({
      product: productName,
      category
    })
  }

  /**
   * Check if WhatsApp is available (for showing/hiding the button)
   * Can be extended with geo-detection logic
   */
  const isWhatsAppAvailable = ref(true)

  onMounted(() => {
    // You could add logic here to detect user's country
    // and hide WhatsApp if it's not commonly used there
    // For now, we'll keep it always visible
  })

  /**
   * Share on WhatsApp
   */
  const shareOnWhatsApp = (url: string, title: string) => {
    const text = `${title}\n\n${url}`
    const link = `https://wa.me/?text=${encodeURIComponent(text)}`
    window.open(link, '_blank', 'noopener,noreferrer')
  }

  /**
   * Format phone number for display
   */
  const formatPhoneNumber = (phone: string): string => {
    // Remove all non-numeric characters
    const cleaned = phone.replace(/\D/g, '')

    // Format based on length
    if (cleaned.length === 10) {
      return `(${cleaned.slice(0, 3)}) ${cleaned.slice(3, 6)}-${cleaned.slice(6)}`
    } else if (cleaned.length === 11 && cleaned[0] === '1') {
      return `+${cleaned[0]} (${cleaned.slice(1, 4)}) ${cleaned.slice(4, 7)}-${cleaned.slice(7)}`
    }

    // Return as-is if format doesn't match
    return phone
  }

  /**
   * Generate a QR code for WhatsApp contact (useful for print materials)
   */
  const generateWhatsAppQR = (options: WhatsAppOptions = {}): string => {
    const link = createWhatsAppLink(options)

    // Using a public QR code API
    return `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(link)}`
  }

  /**
   * WhatsApp business API helpers (for future implementation)
   */
  const sendTemplateMessage = async (to: string, templateName: string, params: Record<string, string>) => {
    // This would integrate with WhatsApp Business API
    // Not implementing now as it requires backend setup
    console.warn('WhatsApp Business API not implemented')
  }

  /**
   * Track WhatsApp clicks (for analytics)
   */
  const trackWhatsAppClick = (context?: string) => {
    // Integration with analytics can go here
    // Google Analytics, etc.
    if (typeof window !== 'undefined' && (window as any).gtag) {
      ;(window as any).gtag('event', 'whatsapp_click', {
        event_category: 'engagement',
        event_label: context || 'general',
        transport_type: 'beacon'
      })
    }
  }

  /**
   * Open WhatsApp with tracking
   */
  const openWhatsAppTracked = (options: WhatsAppOptions & { context?: string } = {}) => {
    const { context, ...whatsappOptions } = options
    trackWhatsAppClick(context)
    openWhatsApp(whatsappOptions)
  }

  return {
    // Properties
    isWhatsAppAvailable,

    // Methods
    getWhatsAppNumber,
    createWhatsAppLink,
    createInquiryLink,
    openWhatsApp,
    openWhatsAppTracked,
    inquireProduct,
    shareOnWhatsApp,
    formatPhoneNumber,
    generateWhatsAppQR,
    sendTemplateMessage,
    trackWhatsAppClick
  }
}

/**
 * Composable for WhatsApp floating button behavior
 */
export const useWhatsAppFloating = () => {
  const { createWhatsAppLink, getWhatsAppNumber, trackWhatsAppClick } = useWhatsApp()

  const whatsappNumber = ref(getWhatsAppNumber())
  const whatsappUrl = computed(() => createWhatsAppLink())

  // Pulse animation on scroll to draw attention
  const shouldPulse = ref(false)

  let pulseTimeout: ReturnType<typeof setTimeout> | null = null

  const triggerPulse = () => {
    shouldPulse.value = true

    if (pulseTimeout) clearTimeout(pulseTimeout)

    pulseTimeout = setTimeout(() => {
      shouldPulse.value = false
    }, 2000)
  }

  // Auto-pulse after user has been on page for 30 seconds
  onMounted(() => {
    setTimeout(() => {
      triggerPulse()
    }, 30000)
  })

  // Click handler
  const handleClick = (context = 'floating_button') => {
    trackWhatsAppClick(context)
    window.open(whatsappUrl.value, '_blank', 'noopener,noreferrer')
  }

  return {
    whatsappNumber,
    whatsappUrl,
    shouldPulse,
    triggerPulse,
    handleClick
  }
}
