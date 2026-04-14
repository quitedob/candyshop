/**
 * Inquiry Form Composable
 * Handles inquiry form state, validation, and submission
 */

import { reactive, ref } from 'vue'
import { useI18n, useLocalePath } from '#i18n'
import { useRoute, useRouter, useNuxtApp } from '#app'

interface InquiryFormState {
  companyName: string
  contactPerson: string
  email: string
  whatsapp: string
  targetCountry: string
  estimatedQuantity: string
  interestedProducts: string[]
  packagingRequirements: string
  flavorRequirements: string
  oemNeeded: boolean | null
  expectedDelivery: string
  message: string
  files: File[]
}

interface ValidationRule {
  required?: boolean
  email?: boolean
  phone?: boolean
  minLength?: number
  pattern?: RegExp
  custom?: (value: unknown) => boolean | string
}

interface ValidationErrors {
  [key: string]: string[]
}

interface InquiryFormOptions {
  onSuccess?: (response: { success: boolean; message: string; inquiryId?: string }) => void
  onError?: (error: { message: string }) => void
}

export const useInquiry = (options: InquiryFormOptions = {}) => {
  const { t } = useI18n()
  const { submitInquiry } = useApi()

  // Form state
  const state = reactive<InquiryFormState>({
    companyName: '',
    contactPerson: '',
    email: '',
    whatsapp: '',
    targetCountry: '',
    estimatedQuantity: '',
    interestedProducts: [],
    packagingRequirements: '',
    flavorRequirements: '',
    oemNeeded: null,
    expectedDelivery: '',
    message: '',
    files: []
  })

  // UI state
  const isSubmitting = ref(false)
  const isSubmitted = ref(false)
  const errors = ref<ValidationErrors>({})

  // Validation rules
  const rules: Record<keyof InquiryFormState, ValidationRule> = {
    companyName: { required: true, minLength: 2 },
    contactPerson: { required: true, minLength: 2 },
    email: { required: true, email: true },
    whatsapp: { phone: true },
    targetCountry: {},
    estimatedQuantity: {},
    interestedProducts: {},
    packagingRequirements: {},
    flavorRequirements: {},
    oemNeeded: {},
    expectedDelivery: {},
    message: {},
    files: {}
  }

  /**
   * Validate a single field
   */
  const validateField = (field: keyof InquiryFormState, value: unknown): string[] => {
    const fieldErrors: string[] = []
    const rule = rules[field]

    if (!rule) return fieldErrors

    // Required validation
    if (rule.required) {
      if (Array.isArray(value) && value.length === 0) {
        fieldErrors.push(t(`validation.required.${field}`) || `${field} is required`)
      } else if (!value) {
        fieldErrors.push(t(`validation.required.${field}`) || `${field} is required`)
      }
    }

    // Email validation
    if (rule.email && typeof value === 'string' && value) {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
      if (!emailRegex.test(value)) {
        fieldErrors.push(t('validation.email') || 'Invalid email format')
      }
    }

    // Phone validation
    if (rule.phone && typeof value === 'string' && value) {
      const phoneRegex = /^[\d\s+\-()]+$/
      if (!phoneRegex.test(value)) {
        fieldErrors.push(t('validation.phone') || 'Invalid phone format')
      }
    }

    // Min length validation
    if (rule.minLength && typeof value === 'string' && value.length < rule.minLength) {
      fieldErrors.push(
        t(`validation.minLength.${field}`, { min: rule.minLength }) ||
        `${field} must be at least ${rule.minLength} characters`
      )
    }

    // Pattern validation
    if (rule.pattern && typeof value === 'string' && value) {
      if (!rule.pattern.test(value)) {
        fieldErrors.push(t(`validation.pattern.${field}`) || `${field} format is invalid`)
      }
    }

    // Custom validation
    if (rule.custom) {
      const result = rule.custom(value)
      if (result === true) {
        // Valid
      } else if (typeof result === 'string') {
        fieldErrors.push(result)
      } else if (result === false) {
        fieldErrors.push(t(`validation.invalid.${field}`) || `${field} is invalid`)
      }
    }

    return fieldErrors
  }

  /**
   * Validate all fields
   */
  const validate = (): boolean => {
    const newErrors: ValidationErrors = {}

    // Only validate required fields for basic submission
    const requiredFields: (keyof InquiryFormState)[] = [
      'companyName',
      'contactPerson',
      'email'
    ]

    requiredFields.forEach(field => {
      const fieldErrors = validateField(field, state[field])
      if (fieldErrors.length > 0) {
        newErrors[field] = fieldErrors
      }
    })

    errors.value = newErrors
    return Object.keys(newErrors).length === 0
  }

  /**
   * Validate a specific field and update errors
   */
  const validateAndUpdate = (field: keyof InquiryFormState) => {
    const fieldErrors = validateField(field, state[field])
    if (fieldErrors.length > 0) {
      errors.value[field] = fieldErrors
    } else {
      delete errors.value[field]
    }
  }

  /**
   * Clear all errors
   */
  const clearErrors = () => {
    errors.value = {}
  }

  /**
   * Clear a specific field error
   */
  const clearFieldError = (field: keyof InquiryFormState) => {
    delete errors.value[field]
  }

  /**
   * Reset form to initial state
   */
  const reset = () => {
    state.companyName = ''
    state.contactPerson = ''
    state.email = ''
    state.whatsapp = ''
    state.targetCountry = ''
    state.estimatedQuantity = ''
    state.interestedProducts = []
    state.packagingRequirements = ''
    state.flavorRequirements = ''
    state.oemNeeded = null
    state.expectedDelivery = ''
    state.message = ''
    state.files = []
    errors.value = {}
    isSubmitted.value = false
  }

  /**
   * Handle file upload
   */
  const handleFileUpload = (files: FileList | null) => {
    if (!files) return

    const maxFiles = 5
    const maxSize = 5 * 1024 * 1024 // 5MB
    const allowedTypes = ['image/jpeg', 'image/png', 'image/webp', 'application/pdf']

    Array.from(files).forEach(file => {
      // Check file count
      if (state.files.length >= maxFiles) {
        errors.value.files = [`Maximum ${maxFiles} files allowed`]
        return
      }

      // Check file size
      if (file.size > maxSize) {
        errors.value.files = [`File ${file.name} is too large. Max size is 5MB`]
        return
      }

      // Check file type
      if (!allowedTypes.includes(file.type)) {
        errors.value.files = [`File ${file.name} is not supported. Use JPG, PNG, WebP, or PDF`]
        return
      }

      state.files.push(file)
    })

    // Clear file errors if successful
    if (state.files.length > 0) {
      delete errors.value.files
    }
  }

  /**
   * Remove a file
   */
  const removeFile = (index: number) => {
    state.files.splice(index, 1)
  }

  /**
   * Submit the form
   */
  const submit = async () => {
    // Validate before submission
    if (!validate()) {
      return false
    }

    isSubmitting.value = true

    try {
      const response = await submitInquiry({
        companyName: state.companyName,
        contactPerson: state.contactPerson,
        email: state.email,
        whatsapp: state.whatsapp || undefined,
        targetCountry: state.targetCountry || undefined,
        estimatedQuantity: state.estimatedQuantity || undefined,
        interestedProducts: state.interestedProducts,
        packagingRequirements: state.packagingRequirements || undefined,
        flavorRequirements: state.flavorRequirements || undefined,
        oemNeeded: state.oemNeeded === true,
        expectedDelivery: state.expectedDelivery || undefined,
        message: state.message || undefined,
        files: state.files
      })

      isSubmitted.value = true
      options.onSuccess?.(response)

      return response
    } catch (err: unknown) {
      const error = err as { message?: string }
      const errorMessage = error?.message || t('form.error') || 'Submission failed'
      errors.value._form = [errorMessage]
      options.onError?.({ message: errorMessage })
      return false
    } finally {
      isSubmitting.value = false
    }
  }

  /**
   * Populate form from URL query params (for pre-filled links)
   */
  const populateFromQuery = () => {
    const route = useRoute()

    if (route.query.company) state.companyName = String(route.query.company)
    if (route.query.contact) state.contactPerson = String(route.query.contact)
    if (route.query.email) state.email = String(route.query.email)
    if (route.query.product) state.interestedProducts = [String(route.query.product)]
    if (route.query.quantity) state.estimatedQuantity = String(route.query.quantity)
    if (route.query.country) state.targetCountry = String(route.query.country)
    if (route.query.message) state.message = String(route.query.message)
  }

  return reactive({
    // State
    state,
    isSubmitting,
    isSubmitted,
    errors,

    // Actions
    validate,
    validateAndUpdate,
    clearErrors,
    clearFieldError,
    reset,
    handleFileUpload,
    removeFile,
    submit,
    populateFromQuery
  })
}

/**
 * Composable for quick inquiry button
 */
export const useQuickInquiry = () => {
  const { $localePath } = useNuxtApp()
  const localePath = $localePath || useLocalePath()

  const openInquiry = (product?: { name: string; category: string }, isSample = false) => {
    const router = useRouter()
    const route = useRoute()
    const { isAuthenticated } = useAuth()

    // Redirect to login if not authenticated
    if (!isAuthenticated.value) {
      router.push({ path: localePath('/auth/login'), query: { redirect: route.fullPath } })
      return
    }

    // Build query params for pre-filled form
    const query: Record<string, string> = {}

    if (product) {
      query.product = product.name
      query.category = product.category
      if (isSample) {
        query.message = 'I would like to request a sample for this product.'
      }
    }

    router.push({
      path: localePath('/contact'),
      query
    })
  }

  return {
    openInquiry
  }
}
