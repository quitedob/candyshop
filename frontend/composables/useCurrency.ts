/**
 * Currency conversion composable
 * Fetches exchange rates from API, caches in localStorage (24h TTL),
 * provides convert() and formatPrice() helpers.
 */

import { ref, computed, watch } from 'vue'
import { useI18n } from '#i18n'

interface ExchangeRates {
  base: string
  rates: Record<string, number>
  updated: string
  disclaimer: string
}

const CACHE_KEY = 'candypro_exchange_rates'
const CACHE_TTL = 24 * 60 * 60 * 1000 // 24 hours
const SUPPORTED_CURRENCIES = ['USD', 'EUR', 'CNY', 'GBP', 'JPY', 'SAR', 'AED', 'KRW', 'SGD', 'MYR', 'THB', 'VND', 'IDR', 'PHP']

const selectedCurrency = ref('USD')
const rates = ref<Record<string, number>>({})
const loading = ref(false)
const lastUpdated = ref('')

const CURRENCY_SYMBOLS: Record<string, string> = {
  USD: '$', EUR: '€', CNY: '¥', GBP: '£', JPY: '¥',
  SAR: '﷼', AED: 'د.إ', KRW: '₩', SGD: 'S$',
  MYR: 'RM', THB: '฿', VND: '₫', IDR: 'Rp', PHP: '₱'
}

const CURRENCY_NAMES: Record<string, string> = {
  USD: 'US Dollar', EUR: 'Euro', CNY: 'Chinese Yuan', GBP: 'British Pound',
  JPY: 'Japanese Yen', SAR: 'Saudi Riyal', AED: 'UAE Dirham',
  KRW: 'Korean Won', SGD: 'Singapore Dollar', MYR: 'Malaysian Ringgit',
  THB: 'Thai Baht', VND: 'Vietnamese Dong', IDR: 'Indonesian Rupiah', PHP: 'Philippine Peso'
}

function loadFromCache(): boolean {
  try {
    const cached = localStorage.getItem(CACHE_KEY)
    if (!cached) return false
    const data = JSON.parse(cached)
    if (Date.now() - data.timestamp > CACHE_TTL) return false
    rates.value = data.rates || {}
    lastUpdated.value = data.updated || ''
    return true
  } catch { return false }
}

function saveToCache(r: Record<string, number>, updated: string) {
  try {
    localStorage.setItem(CACHE_KEY, JSON.stringify({
      rates: r, updated, timestamp: Date.now()
    }))
  } catch { /* localStorage full - ignore */ }
}

async function fetchRates() {
  if (loadFromCache()) return
  loading.value = true
  try {
    const baseURL = useRuntimeConfig().public.apiBase || '/api/v1'
    const resp = await $fetch<ExchangeRates>(`${baseURL}/public/exchange-rates`)
    rates.value = resp.rates || {}
    lastUpdated.value = resp.updated || ''
    saveToCache(rates.value, lastUpdated.value)
  } catch {
    // If API unavailable, use empty rates (no conversion)
    rates.value = {}
  } finally {
    loading.value = false
  }
}

export const useCurrency = () => {
  const { locale } = useI18n()

  // Detect best currency from locale
  const detectCurrency = (): string => {
    const map: Record<string, string> = {
      zh: 'CNY', en: 'USD', ko: 'KRW', ar: 'AED',
      ja: 'JPY', th: 'THB', vi: 'VND', id: 'IDR', ms: 'MYR'
    }
    return map[locale.value] || 'USD'
  }

  // Initialize on first call
  if (!lastUpdated.value) {
    const saved = loadFromCache()
    if (!saved) {
      selectedCurrency.value = detectCurrency()
      fetchRates()
    } else {
      selectedCurrency.value = detectCurrency()
    }
  }

  const isConverted = computed(() => selectedCurrency.value !== 'USD')
  const currencySymbol = computed(() => CURRENCY_SYMBOLS[selectedCurrency.value] || selectedCurrency.value)
  const currencyName = computed(() => CURRENCY_NAMES[selectedCurrency.value] || selectedCurrency.value)

  const supportedCurrencies = computed(() =>
    SUPPORTED_CURRENCIES.filter(c => c === 'USD' || rates.value[c])
  )

  function convert(usdPrice: number): number {
    if (!usdPrice || usdPrice <= 0) return 0
    const target = selectedCurrency.value
    if (target === 'USD') return usdPrice
    const rate = rates.value[target]
    if (!rate || rate <= 0) return usdPrice
    return usdPrice * rate
  }

  function setCurrency(code: string) {
    selectedCurrency.value = code
    if (code !== 'USD' && !rates.value[code]) {
      fetchRates()
    }
  }

  function formatPrice(usdPrice: number, showOriginal = true): string {
    const converted = convert(usdPrice)
    const { locale: loc } = useI18n()
    const formatter = new Intl.NumberFormat(loc.value, { minimumFractionDigits: 2, maximumFractionDigits: 2 })
    const sym = CURRENCY_SYMBOLS[selectedCurrency.value] || selectedCurrency.value
    return `${sym}${formatter.format(converted)}`
  }

  function cur(input?: string | null): string {
    if (input && input.trim()) return input.trim()
    return selectedCurrency.value
  }

  const disclaimer = computed(() => {
    if (selectedCurrency.value === 'USD') return ''
    return 'This is a reference converted price. Actual pricing may vary based on exchange rates. Please confirm with our sales team.'
  })

  return {
    selectedCurrency,
    rates,
    loading,
    lastUpdated,
    isConverted,
    currencySymbol,
    currencyName,
    supportedCurrencies,
    convert,
    formatPrice,
    setCurrency,
    fetchRates,
    disclaimer,
    cur
  }
}
