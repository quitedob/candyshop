<template>
  <div class="pricing-page">
    <header class="pricing-page__header">
      <div>
        <h1 class="pricing-page__title">{{ t('customer.pricing.title') }}</h1>
        <p class="pricing-page__subtitle">{{ t('customer.pricing.subtitle') }}</p>
      </div>
    </header>

    <div v-if="pending" class="pricing-page__loading">
      <Icon name="heroicons:arrow-path" class="h-8 w-8 animate-spin text-orange-500" aria-hidden="true" />
    </div>

    <div v-else-if="error" class="pricing-alert pricing-alert--error" role="alert">
      <Icon name="heroicons:exclamation-circle" class="h-5 w-5 shrink-0" aria-hidden="true" />
      {{ error }}
    </div>

    <div v-else-if="!priceList" class="pricing-empty">
      <div class="pricing-empty__icon">
        <Icon name="heroicons:tag" class="h-8 w-8" aria-hidden="true" />
      </div>
      <h2 class="pricing-empty__title">
        {{ emptyTitle }}
      </h2>
      <p class="pricing-empty__text">{{ emptyHint }}</p>
      <div class="pricing-empty__actions">
        <NuxtLink
          v-if="emptyStatus === 'no_company_profile'"
          :to="localePath('/customer/company')"
          class="pricing-btn pricing-btn--primary"
        >
          {{ t('customer.pricing.setup_company') }}
        </NuxtLink>
        <NuxtLink :to="localePath('/customer/inquiries/new')" class="pricing-btn pricing-btn--ghost">
          {{ t('customer.pricing.contact_sales') }}
        </NuxtLink>
      </div>
    </div>

    <div v-else class="pricing-card">
      <div class="pricing-card__head">
        <div>
          <h2 class="pricing-card__name">{{ priceList.name }}</h2>
          <p v-if="priceList.description" class="pricing-card__desc">{{ priceList.description }}</p>
        </div>
        <span class="pricing-card__badge">{{ t('customer.pricing.active') }}</span>
      </div>

      <div v-if="priceList.prices?.length" class="pricing-table-wrap">
        <table class="pricing-table">
          <thead>
            <tr>
              <th>{{ t('customer.pricing.col_product') }}</th>
              <th>{{ t('customer.pricing.col_min_qty') }}</th>
              <th>{{ t('customer.pricing.col_unit_price') }}</th>
              <th>{{ t('customer.pricing.col_currency') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in priceList.prices" :key="p.id">
              <td class="pricing-table__product">{{ productNames[p.productId] || p.productId }}</td>
              <td>{{ formatNumber(p.minQuantity || 1) }}</td>
              <td class="pricing-table__price">{{ p.unitPrice }}</td>
              <td>{{ cur(p.currency || priceList.currency) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <p v-else class="pricing-card__empty">{{ t('customer.pricing.no_items') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useTranslation } from '~/composables/useTranslation'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const { t } = useI18n()
const { tField } = useTranslation()
const { currencyOrDefault: cur } = useDisplay()
const localePath = useLocalePath()
const api = useApi()

const priceList = ref<any>(null)
const emptyStatus = ref('')
const pending = ref(true)
const error = ref('')
const productNames = ref<Record<string, string>>({})

/** 空状态标题 */
const emptyTitle = computed(() => {
  if (emptyStatus.value === 'no_company_profile') return t('customer.pricing.no_company_profile')
  return t('customer.pricing.no_price_list')
})

/** 空状态说明 */
const emptyHint = computed(() => {
  if (emptyStatus.value === 'no_company_profile') return t('customer.pricing.no_company_profile_hint')
  return t('customer.pricing.contact_sales')
})

const formatNumber = (n: number) => new Intl.NumberFormat().format(n)

/** 加载专属价格表 */
const load = async () => {
  pending.value = true
  error.value = ''
  emptyStatus.value = ''
  try {
    const res = await api.get<{ priceList?: any; status?: string }>('/user/price-list')
    priceList.value = res.priceList || null
    emptyStatus.value = res.status || ''
    if (priceList.value?.prices?.length) {
      const ids: string[] = [...new Set<string>(priceList.value.prices.map((p: any) => p.productId as string))]
      await Promise.all(ids.map(async (id: string) => {
        try {
          const product = await api.getProduct(id)
          if (product) productNames.value[id] = tField(product, 'name')
        } catch {
          // 保留 UUID 作为后备显示
        }
      }))
    }
  } catch (e: any) {
    error.value = e?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.pricing-page {
  max-width: 56rem;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.pricing-page__header {
  padding: 1.25rem 1.5rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg, 0.75rem);
}

.pricing-page__title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-text);
}

.pricing-page__subtitle {
  margin: 0.375rem 0 0;
  font-size: 0.875rem;
  color: var(--color-text-lighter);
}

.pricing-page__loading {
  display: flex;
  justify-content: center;
  padding: 4rem 0;
}

.pricing-alert {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  padding: 0.875rem 1rem;
  border-radius: var(--radius-md, 0.5rem);
  font-size: 0.875rem;
}

.pricing-alert--error {
  background: rgba(var(--color-error-rgb), 0.08);
  color: var(--color-error);
  border: 1px solid rgba(var(--color-error-rgb), 0.2);
}

.pricing-empty {
  text-align: center;
  padding: 3rem 1.5rem;
  background: var(--color-bg);
  border: 1px dashed var(--color-border);
  border-radius: var(--radius-lg, 0.75rem);
}

.pricing-empty__icon {
  width: 4rem;
  height: 4rem;
  margin: 0 auto 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: var(--color-bg-alt);
  color: var(--color-text-lighter);
}

.pricing-empty__title {
  margin: 0;
  font-size: 1.0625rem;
  font-weight: 700;
  color: var(--color-text);
}

.pricing-empty__text {
  margin: 0.5rem auto 0;
  max-width: 28rem;
  font-size: 0.875rem;
  color: var(--color-text-lighter);
  line-height: 1.5;
}

.pricing-empty__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.75rem;
  margin-top: 1.25rem;
}

.pricing-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.5625rem 1.125rem;
  border-radius: var(--radius-md, 0.5rem);
  font-size: 0.875rem;
  font-weight: 600;
  text-decoration: none;
  transition: background 0.15s, filter 0.15s;
}

.pricing-btn--primary {
  background: var(--color-highlight);
  color: white;
}

.pricing-btn--primary:hover { filter: brightness(1.05); }

.pricing-btn--ghost {
  background: var(--color-bg-alt);
  color: var(--color-text);
  border: 1px solid var(--color-border);
}

.pricing-card {
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg, 0.75rem);
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.pricing-card__head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--color-border-light);
  background: var(--color-bg-alt);
}

.pricing-card__name {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--color-text);
}

.pricing-card__desc {
  margin: 0.25rem 0 0;
  font-size: 0.8125rem;
  color: var(--color-text-lighter);
}

.pricing-card__badge {
  padding: 0.25rem 0.625rem;
  border-radius: var(--radius-full, 9999px);
  font-size: 0.6875rem;
  font-weight: 700;
  background: rgba(var(--color-success-rgb), 0.12);
  color: #065f46;
}

.pricing-table-wrap { overflow-x: auto; }

.pricing-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.pricing-table th {
  padding: 0.75rem 1rem;
  text-align: left;
  font-size: 0.6875rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-lighter);
  background: var(--color-bg-alt);
  border-bottom: 1px solid var(--color-border-light);
}

.pricing-table td {
  padding: 0.875rem 1rem;
  border-bottom: 1px solid var(--color-border-light);
  color: var(--color-text);
}

.pricing-table__product { font-weight: 600; }
.pricing-table__price { font-weight: 700; color: var(--color-highlight); }

.pricing-card__empty {
  padding: 1.5rem;
  margin: 0;
  font-size: 0.875rem;
  color: var(--color-text-lighter);
}
</style>
