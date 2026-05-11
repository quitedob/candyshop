<template>
  <div class="product-page">
    <!-- Breadcrumb -->
    <div class="container">
      <Breadcrumb :items="breadcrumbItems" />
    </div>

    <!-- Product Header -->
    <section class="product-header section-sm">
      <div class="container">
        <div v-if="productStatus === 'pending'" class="product-header__inner">
          <SkeletonLoader width="100%" height="400px" borderRadius="16px" />
        </div>
        <div v-else-if="productStatus === 'error' || !product.id" class="product-header__inner">
          <ErrorState
            :title="t('errors.default')"
            :message="t('offline_message')"
            :actionText="t('errors.tryAgain')"
            @action="refreshProduct"
          />
        </div>
        <div v-else class="product-header__inner">
          <!-- Gallery -->
          <div class="product-header__gallery">
            <ProductGallery
              :images="product.images"
              :alt="product.name"
            />
          </div>

          <!-- Info -->
          <div class="product-header__info">
            <div class="product-header__badges">
              <span v-if="product.oemAvailable" class="badge badge-primary">
                {{ $t('product.oem_available') }}
              </span>
              <span v-if="product.halalCertified" class="badge badge-success">
                {{ $t('factory.halal') }}
              </span>
              <span v-if="product.organic" class="badge badge-accent">
                {{ $t('filters.organic') }}
              </span>
            </div>

            <h1 class="product-header__title">{{ product.name }}</h1>
            <p class="product-header__summary">{{ product.summary }}</p>

            <!-- Price -->
            <p v-if="product.basePrice" class="product-header__price">
              {{ $t('product.price_from') }} {{ priceDisplay }}{{ $t('product.price_per_unit') }}
              <span v-if="currency.isConverted.value" class="product-header__price-note">{{ $t('product.reference_price') }}</span>
            </p>

            <!-- Quick Specs -->
            <div class="product-header__specs">
              <div v-if="product.moq" class="product-header__spec">
                <Icon name="lucide:box" size="18" />
                <span>{{ $t('oem.moq') }}: <strong>{{ product.moq }} pcs</strong></span>
              </div>
              <div v-if="product.leadTime" class="product-header__spec">
                <Icon name="lucide:clock" size="18" />
                <span>{{ $t('oem.lead_time') }}: <strong>{{ product.leadTime }}</strong></span>
              </div>
              <div v-if="product.port" class="product-header__spec">
                <Icon name="lucide:anchor" size="18" />
                <span>{{ $t('factory.port') }}: <strong>{{ product.port }}</strong></span>
              </div>
            </div>

            <!-- Quick Actions -->
            <div class="product-header__actions">
              <button class="btn btn-highlight btn-lg" @click="openInquiryModal">
                <Icon name="lucide:message-circle" size="20" />
                {{ $t('product.inquire_now') }}
              </button>
              <NuxtLink v-if="isAuthenticated && !isAdmin" :to="localePath(`/customer/products/${product.id || product.slug}`)" class="btn btn-outline btn-lg">
                <Icon name="lucide:shopping-bag" size="20" />
                {{ $t('product.place_order') }}
              </NuxtLink>
              <button v-else class="btn btn-outline btn-lg" @click="openInquiryModal">
                <Icon name="lucide:package" size="20" />
                {{ $t('product.request_sample') }}
              </button>
            </div>

            <!-- Certifications -->
            <div v-if="product.certifications?.length" class="product-header__certs">
              <span class="product-header__certs-label">{{ $t('factory.certifications') }}:</span>
              <div class="product-header__certs-list">
                <span v-for="cert in product.certifications" :key="cert" class="cert-badge">
                  {{ cert }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Tabs -->
    <section v-if="product.id" class="product-tabs section-sm">
      <div class="container">
        <div class="tabs">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            class="tabs__tab"
            :class="{ 'tabs__tab--active': activeTab === tab.id }"
            @click="activeTab = tab.id"
          >
            <Icon :name="tab.icon" size="18" />
            {{ tab.label }}
          </button>
        </div>

        <div class="tabs__content">
          <!-- Description -->
          <Transition name="tab-fade" mode="out-in">
          <div v-if="activeTab === 'description'" key="description" class="tab-content">
            <div class="tab-content__body">{{ product.description }}</div>
          </div>

          <!-- Specifications -->
          <div v-else-if="activeTab === 'specs'" key="specs" class="tab-content">
            <SpecTable :rows="specRows" />
          </div>

          <!-- Packaging -->
          <div v-else-if="activeTab === 'packaging'" key="packaging" class="tab-content">
            <div class="packaging-options">
              <div
                v-for="pkg in packagingOptions"
                :key="pkg.id"
                class="packaging-card"
              >
                <div class="packaging-card__image">
                  <img :src="pkg.image" :alt="pkg.name" />
                </div>
                <div class="packaging-card__info">
                  <h4>{{ pkg.name }}</h4>
                  <p>{{ pkg.description }}</p>
                  <span class="packaging-card__moq">MOQ: {{ pkg.moq }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- OEM Options -->
          <div v-else-if="activeTab === 'oem'" key="oem" class="tab-content">
            <OEMOptions
              :options="oemOptions"
              :subtitle="$t('oem.customize_message')"
            />
          </div>
          </Transition>
        </div>
      </div>
    </section>

    <!-- Related Products -->
    <section v-if="relatedProducts.length" class="related section bg-alt">
      <div class="container">
        <div class="section-header">
          <h2>{{ $t('product.related_products') }}</h2>
        </div>

        <div class="related__grid">
          <ProductCard
            v-for="related in relatedProducts"
            :key="related.id"
            :product="related"
          />
        </div>
      </div>
    </section>

    <!-- Application Scenarios -->
    <section class="scenarios section">
      <div class="container">
        <div class="section-header">
          <h2>{{ $t('product.applications') }}</h2>
        </div>

        <div class="scenarios__grid">
          <div v-for="scenario in scenarios" :key="scenario.id" class="scenario-card">
            <div class="scenario-card__icon">
              <Icon :name="scenario.icon" size="32" />
            </div>
            <h4>{{ scenario.title }}</h4>
            <p>{{ scenario.description }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Sticky Inquiry CTA (Mobile) -->
    <div class="sticky-cta hide-desktop">
      <button class="sticky-cta__inquire" @click="openInquiryModal">
        {{ $t('product.inquire_now') }}
      </button>
      <button class="sticky-cta__sample" @click="openInquiryModal">
        {{ $t('product.request_sample') }}
      </button>
    </div>

    <!-- Inquiry Modal -->
    <Teleport to="body">
      <div
        v-if="isInquiryOpen"
        class="modal"
        @click.self="closeInquiryModal"
      >
        <div class="modal__content">
          <button class="modal__close" @click="closeInquiryModal">
            <Icon name="lucide:x" size="24" />
          </button>
          <div class="modal__body">
            <h3>{{ $t('product.inquire_now') }}</h3>
            <InquiryForm
              :product-slug="product.slug"
              :product-name="product.name"
              :category="categoryName"
              @success="closeInquiryModal"
            />
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n, useLocalePath } from '#i18n'
import { useDisplay } from '~/composables/useDisplay'

const { t, locale } = useI18n()
const localePath = useLocalePath()
const route = useRoute()
const { isAuthenticated, isAdmin, isPending } = useAuth()
const { formatNumber } = useDisplay()
const currency = useCurrency()
const api = useApi()

const priceDisplay = computed(() => currency.formatPrice(product.value.basePrice || 0))
const { getProduct, getRelatedProducts } = api

// State
const activeTab = ref('description')
const isInquiryOpen = ref(false)

// Route params
const categorySlug = computed(() => route.params.category as string)
const productSlug = computed(() => route.params.slug as string)

// Fetch product
const { data: productData, status: productStatus, refresh: refreshProduct } = await useAsyncData(
  `product-${productSlug.value}-${locale.value}`,
  async () => {
    return await getProduct(productSlug.value)
  }
)

const product = computed(() => productData.value || {})

const categoryName = computed(() => {
  const slugKey = (product.value.categorySlug || product.value.category || '').replace(/-/g, '_')
  const key = `product.categories.${slugKey}`
  return t(key, product.value.category)
})

// Breadcrumb
const breadcrumbItems = computed(() => [
  { label: t('nav.products'), to: '/products' },
  { label: categoryName.value, to: `/products/${categorySlug.value}` },
  { label: product.value.name }
])

// Tabs
const tabs = [
  { id: 'description', label: t('nav.about'), icon: 'lucide:info' },
  { id: 'specs', label: t('product.specifications'), icon: 'lucide:list' },
  { id: 'packaging', label: t('product.packaging'), icon: 'lucide:package' },
  { id: 'oem', label: t('nav.oem'), icon: 'lucide:settings' }
]

// Spec rows
const specRows = computed(() => [
  { label: t('product.flavors'), value: product.value.flavors?.join(', ') || t('common.display.em_dash'), type: 'text' },
  { label: t('product.shapes'), value: product.value.shapes?.join(', ') || t('common.display.em_dash'), type: 'text' },
  { label: t('product.ingredients'), value: product.value.ingredients || t('common.display.em_dash'), type: 'text' },
  { label: t('product.allergens'), value: product.value.allergens || t('common.display.em_dash'), type: 'text' },
  { label: t('product.shelf_life'), value: product.value.shelfLife || t('common.display.em_dash'), type: 'text' },
  { label: t('product.storage'), value: product.value.storage || t('common.display.em_dash'), type: 'text' }
])

// Packaging options
const packagingOptions = [
  {
    id: 'bag',
    name: 'Stand-up Pouch',
    description: 'Flexible packaging with zip lock',
    moq: '1,000 pcs',
    image: '/images/packaging/bag.jpg'
  },
  {
    id: 'jar',
    name: 'Plastic Jar',
    description: 'Resealable container',
    moq: '500 pcs',
    image: '/images/packaging/jar.jpg'
  },
  {
    id: 'box',
    name: 'Gift Box',
    description: 'Premium presentation box',
    moq: '300 pcs',
    image: '/images/packaging/box.jpg'
  },
  {
    id: 'bulk',
    name: 'Bulk Bag',
    description: 'Large quantity packaging',
    moq: '100 pcs',
    image: '/images/packaging/bulk.jpg'
  }
]

// OEM Options
const oemOptions = [
  {
    icon: 'lucide:flame',
    title: t('product.flavors'),
    summary: 'Choose from our range or custom develop',
    description: 'We offer over 50 fruit flavors or can create custom blends to match your requirements.',
    choices: [
      { id: 'fruit', label: t('product.flavor_options.fruit_mix'), image: '/images/flavors/fruit.jpg' },
      { id: 'citrus', label: t('product.flavor_options.citrus'), image: '/images/flavors/citrus.jpg' },
      { id: 'berry', label: t('product.flavor_options.berry'), image: '/images/flavors/berry.jpg' },
      { id: 'custom', label: t('product.flavor_options.custom'), image: '/images/flavors/custom.jpg' }
    ],
    pricing: { moq: '3,000 pcs', leadTime: '4-6 weeks' }
  },
  {
    icon: 'lucide:shape',
    title: t('product.shapes'),
    summary: 'Standard shapes or custom molds',
    description: 'From classic bears and fruits to your unique brand shapes.',
    choices: [
      { id: 'bear', label: t('product.shape_options.bear'), image: '/images/shapes/bear.jpg' },
      { id: 'fruit', label: t('product.shape_options.fruit'), image: '/images/shapes/fruit.jpg' },
      { id: 'letter', label: t('product.shape_options.letter'), image: '/images/shapes/letter.jpg' },
      { id: 'custom', label: t('product.shape_options.custom'), image: '/images/shapes/custom.jpg' }
    ],
    pricing: { moq: '5,000 pcs', leadTime: '6-8 weeks' }
  },
  {
    icon: 'lucide:palette',
    title: t('product.colors'),
    summary: 'Natural or artificial colors',
    description: 'Available in various color options using natural fruit extracts or food-grade colors.',
    choices: [
      { id: 'natural', label: t('product.color_options.natural'), image: '/images/colors/natural.jpg' },
      { id: 'vibrant', label: t('product.color_options.vibrant'), image: '/images/colors/vibrant.jpg' },
      { id: 'custom', label: t('product.color_options.custom'), image: '/images/colors/custom.jpg' }
    ],
    pricing: { moq: '2,000 pcs', leadTime: '3-4 weeks' }
  }
]

// Related products
const { data: relatedProductsData } = await useAsyncData(
  `related-${productSlug.value}-${locale.value}`,
  () => getRelatedProducts(productSlug.value, 3)
)
const relatedProducts = computed(() => relatedProductsData.value || [])

// Application scenarios
const scenarios = [
  {
    id: 'retail',
    icon: 'lucide:shopping-cart',
    title: 'Retail & Supermarket',
    description: 'Perfect for retail chains and supermarkets'
  },
  {
    id: 'wholesale',
    icon: 'lucide:package',
    title: 'Wholesale & Distribution',
    description: 'Bulk packaging for wholesale and distribution'
  },
  {
    id: 'gift',
    icon: 'lucide:gift',
    title: 'Gift & Premium',
    description: 'Beautiful gift sets for holidays and promotions'
  },
  {
    id: 'vending',
    icon: 'lucide:coffee',
    title: 'Vending Machine',
    description: 'Sized for automatic vending machines'
  }
]

// Modal functions
const openInquiryModal = () => {
  if (isAuthenticated.value) {
    const productName = product.value.name || ''
    const inquiryPath = localePath(`/customer/inquiries/new?name=${encodeURIComponent(productName)}`)
    navigateTo(inquiryPath)
  } else {
    isInquiryOpen.value = true
  }
}

const closeInquiryModal = () => {
  isInquiryOpen.value = false
}

// SEO
useSeo({
  title: product.value.name,
  description: product.value.summary,
  ogImage: product.value.images?.[0],
  ogType: 'product',
  schema: {
    '@context': 'https://schema.org',
    '@type': 'Product',
    '@id': `${siteUrl}/products/${categorySlug.value}/${productSlug.value}#product`,
    name: product.value.name,
    description: product.value.summary,
    image: product.value.images,
    sku: product.value.sku || product.value.slug,
    category: categoryName.value,
    brand: product.value.brand ? { '@type': 'Brand', name: product.value.brand } : undefined,
    offers: {
      '@type': 'Offer',
      price: product.value.basePrice,
      priceCurrency: product.value.currency || 'USD',
      availability: product.value.inStock
        ? 'https://schema.org/InStock'
        : 'https://schema.org/OutOfStock',
    },
    manufacturer: { '@id': `${siteUrl}/#organization` },
  }
})
</script>

<style scoped>
.product-header__inner {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--spacing-3xl);
}

@media (min-width: 1024px) {
  .product-header__inner {
    grid-template-columns: 1fr 1fr;
  }
}

.product-header__badges {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-md);
}

.product-header__title {
  font-size: clamp(1.75rem, 3vw, 2.5rem);
  margin-bottom: var(--spacing-md);
}

.product-header__summary {
  font-size: var(--text-lg);
  color: var(--color-text-light);
  margin-bottom: var(--spacing-xl);
}

.product-header__price {
  font-size: var(--text-2xl);
  font-weight: 700;
  color: var(--color-highlight);
  margin-bottom: var(--spacing-lg);
}

.product-header__price-note {
  display: block;
  font-size: var(--text-xs);
  font-weight: 400;
  color: var(--color-text-light);
  margin-top: var(--spacing-xs);
}

.product-header__specs {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-xl);
  padding: var(--spacing-lg);
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-lg);
}

.product-header__spec {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  font-size: var(--text-sm);
}

.product-header__actions {
  display: flex;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-xl);
  flex-wrap: wrap;
}

.product-header__certs {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--spacing-sm);
  padding-top: var(--spacing-lg);
  border-top: 1px solid var(--color-border-light);
}

.product-header__certs-label {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text-light);
}

.product-header__certs-list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-xs);
}

.cert-badge {
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: var(--text-xs);
  font-weight: 600;
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-sm);
  color: var(--color-accent);
}

/* Tabs */
.tabs {
  display: flex;
  gap: var(--spacing-sm);
  border-bottom: 1px solid var(--color-border);
  margin-bottom: var(--spacing-xl);
  overflow-x: auto;
}

.tabs__tab {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-md) var(--spacing-lg);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text-light);
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
  white-space: nowrap;
}

.tabs__tab:hover {
  color: var(--color-primary);
}

.tabs__tab--active {
  color: var(--color-highlight);
  border-bottom-color: var(--color-highlight);
}

.tab-content__body {
  font-size: var(--text-base);
  line-height: 1.8;
  color: var(--color-text);
}

.tab-content__body :deep(p) {
  margin-bottom: var(--spacing-md);
}

/* Packaging */
.packaging-options {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-lg);
}

.packaging-card {
  background-color: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
}

.packaging-card:hover {
  border-color: var(--color-accent);
  box-shadow: var(--shadow-md);
}

.packaging-card__image {
  aspect-ratio: 4/3;
  background-color: var(--color-bg-alt);
}

.packaging-card__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.packaging-card__info {
  padding: var(--spacing-md);
}

.packaging-card__info h4 {
  font-size: var(--text-base);
  margin-bottom: var(--spacing-xs);
}

.packaging-card__info p {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin-bottom: var(--spacing-sm);
}

.packaging-card__moq {
  display: inline-block;
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-highlight);
}

/* Related */
.related__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: var(--spacing-xl);
}

/* Scenarios */
.scenarios__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-xl);
}

.scenario-card {
  text-align: center;
  padding: var(--spacing-xl);
}

.scenario-card__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  margin-bottom: var(--spacing-md);
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-xl);
  color: var(--color-accent);
}

.scenario-card h4 {
  font-size: var(--text-lg);
  margin-bottom: var(--spacing-xs);
}

.scenario-card p {
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

/* Sticky CTA */
.sticky-cta {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  gap: var(--spacing-sm);
  padding: var(--spacing-md);
  background-color: white;
  border-top: 1px solid var(--color-border);
  box-shadow: var(--shadow-lg);
  z-index: var(--z-sticky);
}

.sticky-cta__sample,
.sticky-cta__inquire {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-md);
  border-radius: var(--radius-md);
  font-weight: 600;
}

.sticky-cta__sample {
  background-color: var(--color-accent);
  color: white;
}

.sticky-cta__inquire {
  background-color: var(--color-highlight);
  color: white;
}

/* Modal */
.modal {
  position: fixed;
  inset: 0;
  z-index: var(--z-modal);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-md);
  background-color: rgba(0, 0, 0, 0.5);
}

.modal__content {
  width: 100%;
  max-width: 600px;
  max-height: 90vh;
  background-color: white;
  border-radius: var(--radius-xl);
  overflow: hidden;
}

.modal__close {
  position: absolute;
  top: var(--spacing-md);
  right: var(--spacing-md);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-full);
  color: var(--color-text-light);
  z-index: 1;
}

.modal__body {
  padding: var(--spacing-xl);
  overflow-y: auto;
  max-height: 90vh;
}

.modal__body h3 {
  margin-bottom: var(--spacing-lg);
}

/* Tab transition */
.tab-fade-enter-active,
.tab-fade-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.tab-fade-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

.tab-fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* Bottom padding to prevent sticky CTA from covering content on mobile */
@media (max-width: 767px) {
  .product-page {
    padding-bottom: 80px;
  }
}
</style>
