<template>
  <div class="product-card">
    <NuxtLink
      :to="localePath(`/products/${product.categorySlug || product.category}/${product.slug}`)"
      class="product-card__link"
    >
      <!-- Image -->
      <div class="product-card__image">
        <img
          :src="imgSrc"
          :alt="tField(product, 'name')"
          class="product-card__img"
          loading="lazy"
          @error="handleImageError"
        />

        <!-- Badges -->
        <div v-if="hasBadges" class="product-card__badges">
          <span v-if="product.oemAvailable" class="product-card__badge product-card__badge--oem">
            {{ $t('product.oem_available') }}
          </span>
          <span v-if="product.halalCertified" class="product-card__badge product-card__badge--halal">
            {{ $t('factory.halal') }}
          </span>
          <span v-if="product.featured" class="product-card__badge product-card__badge--featured">
            {{ $t('home.featured.title') }}
          </span>
        </div>

        <!-- Quick Actions -->
        <div class="product-card__actions">
          <button
            class="product-card__action"
            :aria-label="$t('product.inquire_now')"
            @click.prevent="openInquiry"
          >
            <Icon name="lucide:message-circle" size="18" aria-hidden="true" />
          </button>
          <button
            class="product-card__action"
            :aria-label="$t('product.request_sample')"
            @click.prevent="requestSample"
          >
            <Icon name="lucide:package" size="18" aria-hidden="true" />
          </button>
        </div>

        <!-- Overlay -->
        <div class="product-card__overlay">
          <span class="product-card__view">{{ $t('product.view_details') }}</span>
        </div>
      </div>

      <!-- Content -->
      <div class="product-card__content">
        <h3 class="product-card__title">{{ tField(product, 'name') }}</h3>
        <p class="product-card__summary">{{ tField(product, 'summary') }}</p>

        <!-- Price -->
        <p v-if="product.basePrice" class="product-card__price">
          {{ $t('product.price_from') }} {{ priceDisplay }}{{ $t('product.price_per_unit') }}
          <span v-if="currency.isConverted.value && currency.ratesLoaded.value" class="product-card__price-note">{{ $t('product.reference_price') }}: {{ referencePriceDisplay }}</span>
        </p>

        <!-- Meta -->
        <div class="product-card__meta">
          <span v-if="product.moq" class="product-card__meta-item">
            <Icon name="lucide:box" size="14" aria-hidden="true" />
            {{ $t('product.moq') }}: {{ product.moq }}
          </span>
          <span v-if="product.leadTime" class="product-card__meta-item">
            <Icon name="lucide:clock" size="14" aria-hidden="true" />
            {{ product.leadTime }}
          </span>
        </div>

        <button
          class="product-card__inquiry-btn"
          @click.prevent.stop="openInquiry"
        >
          <Icon name="lucide:message-circle" size="16" aria-hidden="true" />
          {{ $t('product.inquire_now') }}
        </button>
      </div>
    </NuxtLink>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n, useLocalePath } from '#i18n'
import { useDisplay } from '~/composables/useDisplay'
import { useTranslation } from '~/composables/useTranslation'

interface Product {
  id: string
  slug: string
  name: string
  summary: string
  category: string
  categorySlug?: string
  thumbnail?: string
  oemAvailable?: boolean
  halalCertified?: boolean
  featured?: boolean
  moq?: number
  leadTime?: string
  basePrice?: number
  translations?: Record<string, Record<string, string>>
}

interface Props {
  product: Product
}

const props = defineProps<Props>()

const { t } = useI18n()
const localePath = useLocalePath()
const { formatNumber, currencyOrDefault: cur } = useDisplay()
const currency = useCurrency()
const { tField } = useTranslation()

const transactionCurrency = computed(() => cur((props.product as { currency?: string }).currency || 'USD'))
const priceDisplay = computed(() => `${transactionCurrency.value} ${formatNumber(props.product.basePrice || 0)}`)
const referencePriceDisplay = computed(() => currency.formatPrice(props.product.basePrice || 0))

const FALLBACK = computed(() => {
  const catSlug = props.product.categorySlug || props.product.category
  return catSlug ? `/images/categories/${catSlug}.jpg` : '/images/categories/gummy-candy.jpg'
})
const imgSrc = ref(props.product.thumbnail || FALLBACK.value)

watch(() => props.product.thumbnail, (val) => {
  imgSrc.value = val || FALLBACK.value
})

const handleImageError = () => {
  imgSrc.value = FALLBACK.value
}

const hasBadges = computed(() => {
  return props.product.oemAvailable || props.product.halalCertified || props.product.featured
})

const { openInquiry: triggerInquiry } = useQuickInquiry()

const openInquiry = () => {
  triggerInquiry(props.product, false)
}

const requestSample = () => {
  triggerInquiry(props.product, true)
}
</script>

<style scoped>
.product-card {
  background-color: var(--color-bg);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-md);
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
}

.product-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-xl);
}

.product-card__link {
  display: block;
  text-decoration: none;
  color: inherit;
}

.product-card__image {
  position: relative;
  aspect-ratio: 4/3;
  overflow: hidden;
  background-color: var(--color-bg-alt);
}

.product-card__img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition-slow);
}

.product-card:hover .product-card__img {
  transform: scale(1.05);
}

.product-card__badges {
  position: absolute;
  top: var(--spacing-sm);
  left: var(--spacing-sm);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  z-index: 2;
}

.product-card__badge {
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: var(--text-xs);
  font-weight: 600;
  border-radius: var(--radius-sm);
  backdrop-filter: blur(8px);
}

.product-card__badge--oem {
  background-color: rgba(26, 31, 58, 0.9);
  color: white;
}

.product-card__badge--halal {
  background-color: rgba(34, 197, 94, 0.9);
  color: white;
}

.product-card__badge--featured {
  background-color: rgba(var(--color-highlight-rgb), 0.9);
  color: white;
}

.product-card__actions {
  position: absolute;
  top: var(--spacing-sm);
  right: var(--spacing-sm);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  opacity: 0;
  transform: translateX(10px);
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
  z-index: 2;
}

.product-card:hover .product-card__actions {
  opacity: 1;
  transform: translateX(0);
}

.product-card__action {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background-color: var(--color-bg);
  border-radius: var(--radius-full);
  color: var(--color-primary);
  box-shadow: var(--shadow-md);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.product-card__action:hover {
  background-color: var(--color-highlight);
  color: var(--color-text-on-primary);
  transform: scale(1.1);
}

.product-card__overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(to top, rgba(26, 31, 58, 0.8) 0%, transparent 100%);
  display: flex;
  align-items: flex-end;
  justify-content: center;
  padding: var(--spacing-lg);
  opacity: 0;
  transition: opacity var(--transition-base);
}

.product-card:hover .product-card__overlay {
  opacity: 1;
}

.product-card__view {
  color: white;
  font-weight: 600;
  font-size: var(--text-sm);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.product-card__content {
  padding: var(--spacing-md);
}

.product-card__inquiry-btn {
  width: 100%;
  margin-top: var(--spacing-md);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-xs);
  padding: 0.875rem 1rem;
  border: none;
  border-radius: var(--radius-md);
  background: var(--color-highlight);
  color: white;
  font-weight: 600;
  font-size: var(--text-sm);
  cursor: pointer;
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.product-card__inquiry-btn:hover {
  background: var(--color-accent);
  transform: translateY(-1px);
}

.product-card__title {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-primary);
  margin-bottom: var(--spacing-xs);
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  line-height: 1.4;
}

.product-card__summary {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  line-height: 1.5;
  margin-bottom: var(--spacing-md);
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.product-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
}

.product-card__meta-item {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: var(--text-xs);
  color: var(--color-text-light);
  padding: var(--spacing-xs) var(--spacing-sm);
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-sm);
}

.product-card__price {
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--color-highlight);
  margin-bottom: var(--spacing-sm);
}

.product-card__price-note {
  display: block;
  font-size: var(--text-xs);
  font-weight: 400;
  color: var(--color-text-light);
  margin-top: 2px;
}

/* Compact variant */
.product-card--compact .product-card__image {
  aspect-ratio: 1/1;
}

.product-card--compact .product-card__title {
  font-size: var(--text-base);
  -webkit-line-clamp: 1;
}

.product-card--compact .product-card__summary {
  display: none;
}

/* Grid layout variant */
.product-card--grid {
  display: grid;
  grid-template-columns: 120px 1fr;
  gap: var(--spacing-md);
  align-items: center;
}

.product-card--grid .product-card__image {
  aspect-ratio: 1/1;
  border-radius: var(--radius-md);
}

.product-card--grid .product-card__content {
  padding: var(--spacing-sm);
}

.product-card--grid .product-card__actions {
  opacity: 1;
  transform: translateX(0);
  flex-direction: row;
  top: auto;
  bottom: var(--spacing-sm);
  right: var(--spacing-sm);
}

@media (hover: none), (max-width: 640px) {
  .product-card__actions {
    opacity: 1;
    transform: translateX(0);
  }
}

@media (max-width: 640px) {
  .product-card--grid {
    grid-template-columns: 80px 1fr;
  }
}
</style>
