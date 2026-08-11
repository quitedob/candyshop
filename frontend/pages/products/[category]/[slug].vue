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
              :alt="tField(product, 'name')"
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
              <span v-if="product.isOrganic" class="badge badge-accent">
                {{ $t('filters.organic') }}
              </span>
              <span v-if="product.isVegan" class="badge badge-success">
                {{ $t('filters.vegan') }}
              </span>
              <span v-if="product.isGlutenFree" class="badge badge-info">
                {{ $t('filters.gluten_free') }}
              </span>
              <span v-if="product.isSugarFree" class="badge badge-info">
                {{ $t('filters.sugar_free') }}
              </span>
              <span v-if="product.isKosher" class="badge badge-accent">
                {{ $t('filters.kosher') }}
              </span>
            </div>

            <h1 class="product-header__title">{{ tField(product, 'name') }}</h1>
            <p class="product-header__summary">{{ tField(product, 'summary') }}</p>

            <!-- Price -->
            <p v-if="product.basePrice" class="product-header__price">
              {{ $t('product.price_from') }} {{ priceDisplay }}{{ $t('product.price_per_unit') }}
              <span v-if="currency.isConverted.value && currency.ratesLoaded.value" class="product-header__price-note">{{ $t('product.reference_price') }}: {{ referencePriceDisplay }}</span>
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
              <div v-if="product.gtin" class="product-header__spec">
                <Icon name="lucide:barcode" size="18" />
                <span>GTIN: <strong>{{ product.gtin }}</strong></span>
              </div>
            </div>

            <!-- Quick Actions -->
            <div class="product-header__actions">
              <button class="btn btn-highlight btn-lg" @click="openInquiryModal">
                <Icon name="lucide:message-circle" size="20" />
                {{ $t('product.inquire_now') }}
              </button>
              <template v-if="isMounted">
                <button
                  v-if="!isAdmin"
                  class="btn btn-outline btn-lg"
                  :disabled="addingToCart"
                  @click="handleAddToCart"
                >
                  <Icon name="lucide:shopping-cart" size="20" />
                  {{ addingToCart ? $t('common.loading') : $t('product.add_to_cart') }}
                </button>
                <NuxtLink v-if="isAuthenticated && !isAdmin" :to="localePath(`/customer/products/${product.slug}`)" class="btn btn-outline btn-lg">
                  <Icon name="lucide:shopping-bag" size="20" />
                  {{ $t('product.place_order') }}
                </NuxtLink>
                <button v-else-if="!isAuthenticated" class="btn btn-outline btn-lg" @click="openInquiryModal">
                  <Icon name="lucide:package" size="20" />
                  {{ $t('product.request_sample') }}
                </button>
              </template>
              <template v-else>
                <button class="btn btn-outline btn-lg" disabled>
                  <Icon name="lucide:package" size="20" />
                  {{ $t('product.request_sample') }}
                </button>
              </template>
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
            <div class="tab-content__body">{{ tField(product, 'description') }}</div>
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
                  <img :src="pkg.image" :alt="t(`oem.packaging_options.${pkg.id}.name`)" />
                </div>
                <div class="packaging-card__info">
                  <h4>{{ t(`oem.packaging_options.${pkg.id}.name`) }}</h4>
                  <p>{{ t(`oem.packaging_options.${pkg.id}.description`) }}</p>
                  <span class="packaging-card__moq">{{ $t('product.moq_prefix') }}{{ t(`oem.packaging_options.${pkg.id}.moq`) }}</span>
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
            <h4>{{ t(`product.scenarios.${scenario.id}.title`) }}</h4>
            <p>{{ t(`product.scenarios.${scenario.id}.description`) }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Compliance Notices -->
    <section class="compliance section-sm bg-alt">
      <div class="container">
        <div class="compliance__inner">
          <div class="compliance__notice">
            <Icon name="lucide:globe" size="18" class="compliance__icon" />
            <p class="compliance__text">{{ t('legal.jurisdiction_notice_desc') }}</p>
          </div>
          <div class="compliance__notice">
            <Icon name="lucide:shield-check" size="18" class="compliance__icon" />
            <p class="compliance__text">{{ t('legal.food_compliance_desc') }}</p>
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

    <!-- Inquiry Modal（仅客户端挂载，避免 SSR/hydration 与 Teleport 冲突） -->
    <ClientOnly>
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
                :product-id="product.id"
                :product-name="tField(product, 'name')"
                :category="categoryName"
                @success="closeInquiryModal"
              />
            </div>
          </div>
        </div>
      </Teleport>
    </ClientOnly>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n, useLocalePath } from '#i18n'
import { useDisplay } from '~/composables/useDisplay'
import { useTranslation } from '~/composables/useTranslation'

const { t, te, locale } = useI18n()
const { tField } = useTranslation()
const localePath = useLocalePath()
const route = useRoute()
const config = useRuntimeConfig()
const siteUrl = config.public.siteUrl || 'https://candypro-oem.com'
const { isAuthenticated, isAdmin, isPending } = useAuth()
const { formatNumber, currencyOrDefault: cur } = useDisplay()
const currency = useCurrency()
const api = useApi()
const toast = useToast()

const transactionCurrency = computed(() => cur((product.value as { currency?: string })?.currency || 'USD'))
const priceDisplay = computed(() => `${transactionCurrency.value} ${formatNumber(product.value.basePrice || 0)}`)
const referencePriceDisplay = computed(() => currency.formatPrice(product.value.basePrice || 0))
const { getProduct, getRelatedProducts } = api

// State
const activeTab = ref('description')
const isInquiryOpen = ref(false)
const addingToCart = ref(false)
const isMounted = ref(false)

onMounted(async () => {
  isMounted.value = true
  if (route.query.addToCart === '1' && isAuthenticated.value && !isAdmin.value && product.value?.id) {
    await handleAddToCart()
    const query = { ...route.query }
    delete query.addToCart
    await navigateTo({ path: route.path, query, replace: true })
  }
})

// Route params
const categorySlug = computed(() => route.params.category as string)
const productSlug = computed(() => route.params.slug as string)

// 并行拉取产品与关联产品，避免多次 await 导致 hydration 不一致
const [
  { data: productData, status: productStatus, refresh: refreshProduct },
  { data: relatedProductsData }
] = await Promise.all([
  useAsyncData(
    `product-${productSlug.value}-${locale.value}`,
    () => getProduct(productSlug.value)
  ),
  useAsyncData(
    `related-${productSlug.value}-${locale.value}`,
    () => getRelatedProducts(productSlug.value, 3)
  )
])

const product = computed(() => productData.value || {})

const categoryName = computed(() => {
  const slug = product.value.categorySlug || categorySlug.value
  const slugKey = slug.replace(/-/g, '_')
  if (!slugKey) return categorySlug.value.replace(/-/g, ' ')
  const key = `product.categories.${slugKey}`
  if (te(key)) return t(key)
  return product.value.category || categorySlug.value.replace(/-/g, ' ')
})

// Breadcrumb — SSR 阶段使用 slug 兜底，避免 hydration 标签不一致
const breadcrumbItems = computed(() => [
  { label: t('nav.products'), to: '/products' },
  { label: categoryName.value || categorySlug.value, to: `/products/${categorySlug.value}` },
  { label: tField(product.value, 'name') || productSlug.value.replace(/-/g, ' ') }
])

// Tabs
const tabs = [
  { id: 'description', label: t('nav.about'), icon: 'lucide:info' },
  { id: 'specs', label: t('product.specifications'), icon: 'lucide:list' },
  { id: 'packaging', label: t('product.packaging'), icon: 'lucide:package' },
  { id: 'oem', label: t('nav.oem'), icon: 'lucide:settings' }
]

// Spec rows
const noVal = () => t('display.em_dash')
const specRows = computed(() => {
  const p = product.value
  return [
    // Basic
    { label: t('product.flavors'), value: p.flavors?.join(', ') || noVal(), type: 'text' },
    { label: t('product.shapes'), value: p.shapes?.join(', ') || noVal(), type: 'text' },
    // Weight & Dimensions
    { label: t('product.net_weight_per_piece'), value: p.netWeightPerPiece ? `${p.netWeightPerPiece}g` : noVal(), type: 'text' },
    { label: t('product.net_weight_per_pack'), value: p.netWeightPerPack ? `${p.netWeightPerPack}g` : noVal(), type: 'text' },
    { label: t('product.gross_weight_per_carton'), value: p.grossWeightPerCarton ? `${p.grossWeightPerCarton}kg` : noVal(), type: 'text' },
    { label: t('product.pieces_per_pack'), value: p.piecesPerPack || noVal(), type: 'text' },
    { label: t('product.packs_per_carton'), value: p.packsPerCarton || noVal(), type: 'text' },
    { label: t('product.product_length'), value: p.productLengthMM ? `${p.productLengthMM}mm` : noVal(), type: 'text' },
    { label: t('product.product_width'), value: p.productWidthMM ? `${p.productWidthMM}mm` : noVal(), type: 'text' },
    { label: t('product.product_height'), value: p.productHeightMM ? `${p.productHeightMM}mm` : noVal(), type: 'text' },
    // Ingredients
    { label: t('product.ingredients'), value: p.ingredients || noVal(), type: 'text' },
    { label: t('product.allergens'), value: p.allergens || noVal(), type: 'text' },
    // Nutrition (per 100g)
    { label: t('product.energy_kj'), value: p.energyKj ? `${p.energyKj}kJ` : noVal(), type: 'text' },
    { label: t('product.energy_kcal'), value: p.energyKcal ? `${p.energyKcal}kcal` : noVal(), type: 'text' },
    { label: t('product.total_fat'), value: p.totalFatG != null ? `${p.totalFatG}g` : noVal(), type: 'text' },
    { label: t('product.saturated_fat'), value: p.saturatedFatG != null ? `${p.saturatedFatG}g` : noVal(), type: 'text' },
    { label: t('product.carbohydrates'), value: p.carbohydratesG != null ? `${p.carbohydratesG}g` : noVal(), type: 'text' },
    { label: t('product.sugars'), value: p.sugarsG != null ? `${p.sugarsG}g` : noVal(), type: 'text' },
    { label: t('product.protein'), value: p.proteinG != null ? `${p.proteinG}g` : noVal(), type: 'text' },
    { label: t('product.salt'), value: p.saltG != null ? `${p.saltG}g` : noVal(), type: 'text' },
    { label: t('product.fiber'), value: p.fiberG != null ? `${p.fiberG}g` : noVal(), type: 'text' },
    // Ingredient Compliance
    { label: t('product.additives'), value: p.additives?.join(', ') || noVal(), type: 'text' },
    { label: t('product.sweetener_type'), value: p.sweetenerType || noVal(), type: 'text' },
    { label: t('product.cocoa_solids'), value: p.cocoaSolidsPct != null ? `${p.cocoaSolidsPct}%` : noVal(), type: 'text' },
    { label: t('product.milk_solids'), value: p.milkSolidsPct != null ? `${p.milkSolidsPct}%` : noVal(), type: 'text' },
    { label: t('product.gmo_status'), value: p.gmoStatus || noVal(), type: 'text' },
    { label: t('product.may_contain'), value: p.mayContain?.join(', ') || noVal(), type: 'text' },
    { label: t('product.water_activity'), value: p.waterActivity != null ? String(p.waterActivity) : noVal(), type: 'text' },
    // Trade
    { label: t('product.gtin'), value: p.gtin || noVal(), type: 'text' },
    { label: t('product.hs_code'), value: p.hsCode || noVal(), type: 'text' },
    { label: t('product.primary_packaging'), value: p.primaryPackaging || noVal(), type: 'text' },
    { label: t('product.inner_pack_config'), value: p.innerPackConfig || noVal(), type: 'text' },
    { label: t('product.pallet_config'), value: p.palletConfig || noVal(), type: 'text' },
    // Storage
    { label: t('product.shelf_life'), value: p.shelfLife || noVal(), type: 'text' },
    { label: t('product.storage'), value: p.storage || noVal(), type: 'text' },
    // Sample Specs
    { label: t('product.sample_moq'), value: p.sampleMOQ || noVal(), type: 'text' },
    { label: t('product.sample_lead_time'), value: p.sampleLeadTime || noVal(), type: 'text' },
    { label: t('product.sample_price'), value: p.samplePrice ? `${cur((p as { currency?: string })?.currency)} ${formatNumber(p.samplePrice)}` : noVal(), type: 'text' },
  ]
})

// Packaging options (names/descriptions/MOQ come from oem.packaging_options locale keys)
const packagingOptions = [
  { id: 'bag', image: '/images/packaging/bag.jpg' },
  { id: 'jar', image: '/images/packaging/jar.jpg' },
  { id: 'box', image: '/images/packaging/box.jpg' },
  { id: 'bulk', image: '/images/packaging/bulk.jpg' }
]

// OEM Options (summaries/descriptions/MOQ come from oem.oem_options locale keys)
const oemOptions = [
  {
    icon: 'lucide:flame',
    title: t('product.flavors'),
    summary: t('oem.oem_options.flavors.summary'),
    description: t('oem.oem_options.flavors.description'),
    choices: [
      { id: 'fruit', label: t('product.flavor_options.fruit_mix'), image: '/images/flavors/fruit.jpg' },
      { id: 'citrus', label: t('product.flavor_options.citrus'), image: '/images/flavors/citrus.jpg' },
      { id: 'berry', label: t('product.flavor_options.berry'), image: '/images/flavors/berry.jpg' },
      { id: 'custom', label: t('product.flavor_options.custom'), image: '/images/flavors/custom.jpg' }
    ],
    pricing: { moq: t('oem.oem_options.flavors.moq'), leadTime: t('product.category_placeholder_lead') }
  },
  {
    icon: 'lucide:shape',
    title: t('product.shapes'),
    summary: t('oem.oem_options.shapes.summary'),
    description: t('oem.oem_options.shapes.description'),
    choices: [
      { id: 'bear', label: t('product.shape_options.bear'), image: '/images/shapes/bear.jpg' },
      { id: 'fruit', label: t('product.shape_options.fruit'), image: '/images/shapes/fruit.jpg' },
      { id: 'letter', label: t('product.shape_options.letter'), image: '/images/shapes/letter.jpg' },
      { id: 'custom', label: t('product.shape_options.custom'), image: '/images/shapes/custom.jpg' }
    ],
    pricing: { moq: t('oem.oem_options.shapes.moq'), leadTime: t('product.category_placeholder_lead') }
  },
  {
    icon: 'lucide:palette',
    title: t('product.colors'),
    summary: t('oem.oem_options.colors.summary'),
    description: t('oem.oem_options.colors.description'),
    choices: [
      { id: 'natural', label: t('product.color_options.natural'), image: '/images/colors/natural.jpg' },
      { id: 'vibrant', label: t('product.color_options.vibrant'), image: '/images/colors/vibrant.jpg' },
      { id: 'custom', label: t('product.color_options.custom'), image: '/images/colors/custom.jpg' }
    ],
    pricing: { moq: t('oem.oem_options.colors.moq'), leadTime: t('product.category_placeholder_lead') }
  }
]

const relatedProducts = computed(() => relatedProductsData.value || [])

// Application scenarios (titles/descriptions come from product.scenarios locale keys)
const scenarios = [
  { id: 'retail', icon: 'lucide:shopping-cart' },
  { id: 'wholesale', icon: 'lucide:package' },
  { id: 'gift', icon: 'lucide:gift' },
  { id: 'vending', icon: 'lucide:coffee' }
]

// Modal functions
const openInquiryModal = () => {
  if (isAuthenticated.value) {
    const productName = tField(product.value, 'name')
    const params = new URLSearchParams({ name: productName })
    if (product.value.id) params.set('productId', product.value.id)
    const inquiryPath = localePath(`/customer/inquiries/new?${params.toString()}`)
    navigateTo(inquiryPath)
  } else {
    isInquiryOpen.value = true
  }
}

const handleAddToCart = async () => {
  if (!isAuthenticated.value) {
    const sep = route.fullPath.includes('?') ? '&' : '?'
    const redirect = `${route.fullPath}${sep}addToCart=1`
    await navigateTo(localePath(`/auth/login?redirect=${encodeURIComponent(redirect)}`))
    return
  }
  if (!product.value.id) return
  addingToCart.value = true
  try {
    await api.addToCart(product.value.id, {
      quantity: product.value.moq || 1,
      unitPrice: product.value.basePrice || 0
    })
    toast.success(t('product.added_to_cart'))
  } catch (err: any) {
    notifyError(err, t('errors.api.cart_submit_failed'))
  } finally {
    addingToCart.value = false
  }
}

const closeInquiryModal = () => {
  isInquiryOpen.value = false
}

// SEO
usePageOgImage({
  title: tField(product.value, 'name'),
  description: tField(product.value, 'summary'),
  ogImage: product.value.ogImage,
  thumbnail: product.value.thumbnail,
  images: product.value.images,
  ogType: 'product',
  schema: {
    '@context': 'https://schema.org',
    '@type': 'Product',
    '@id': `${siteUrl}/products/${categorySlug.value}/${productSlug.value}#product`,
    name: tField(product.value, 'name'),
    description: tField(product.value, 'summary'),
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

/* Compliance Notices */
.compliance__inner {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
  max-width: 800px;
  margin: 0 auto;
}

.compliance__notice {
  display: flex;
  gap: var(--spacing-md);
  align-items: flex-start;
  padding: var(--spacing-md);
  background-color: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}

.compliance__icon {
  flex-shrink: 0;
  color: var(--color-text-light);
  margin-top: 2px;
}

.compliance__text {
  font-size: var(--text-xs);
  color: var(--color-text-light);
  line-height: 1.6;
  margin: 0;
}

/* Bottom padding to prevent sticky CTA from covering content on mobile */
@media (max-width: 767px) {
  .product-page {
    padding-bottom: 80px;
  }
}
</style>
