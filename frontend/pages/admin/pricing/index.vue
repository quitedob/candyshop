<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.pricing.title') }}</h1>
        <p class="mt-2 text-sm text-gray-700">{{ t('admin.pricing.description') }}</p>
      </div>
    </div>

    <!-- Tabs -->
    <div class="mt-6 border-b border-gray-200">
      <nav class="-mb-px flex space-x-8">
        <button type="button" @click="activeTab = 'price-lists'" :class="[activeTab === 'price-lists' ? 'border-orange-500 text-orange-600' : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'], 'whitespace-nowrap border-b-2 py-4 px-1 text-sm font-medium'">
          {{ t('admin.pricing.price_lists') }}
        </button>
        <button type="button" @click="activeTab = 'product-pricing'" :class="[activeTab === 'product-pricing' ? 'border-orange-500 text-orange-600' : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'], 'whitespace-nowrap border-b-2 py-4 px-1 text-sm font-medium'">
          {{ t('admin.pricing.product_pricing') }}
        </button>
      </nav>
    </div>

    <!-- Price Lists Tab -->
    <div v-if="activeTab === 'price-lists'" class="mt-6 space-y-6">
      <div class="flex justify-end">
        <button type="button" @click="openCreatePriceList" class="inline-flex items-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700">
          {{ t('admin.pricing.create_price_list') }}
        </button>
      </div>

      <div class="overflow-hidden rounded-lg bg-white shadow ring-1 ring-black ring-opacity-5">
        <table class="min-w-full divide-y divide-gray-300">
          <thead class="bg-gray-50">
            <tr>
              <th class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">{{ t('admin.pricing.col_name') }}</th>
              <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.pricing.col_currency') }}</th>
              <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.pricing.col_status') }}</th>
              <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.pricing.col_products') }}</th>
              <th class="relative py-3.5 pl-3 pr-4 sm:pr-6"><span class="sr-only">{{ t('admin.pricing.actions') }}</span></th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 bg-white">
            <tr v-if="loadingPriceLists">
              <td colspan="5" class="py-5 text-center text-sm text-gray-500">{{ t('admin.pricing.loading') }}</td>
            </tr>
            <tr v-else-if="priceLists.length === 0">
              <td colspan="5" class="py-5 text-center text-sm text-gray-500">{{ t('admin.pricing.no_price_lists') }}</td>
            </tr>
            <tr v-else v-for="pl in priceLists" :key="pl.id">
              <td class="py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">{{ pl.name }}</td>
              <td class="px-3 py-4 text-sm text-gray-500">{{ cur(pl.currency) }}</td>
              <td class="px-3 py-4 text-sm">
                <span :class="[pl.status === 'active' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800', 'inline-flex rounded-full px-2 text-xs font-semibold leading-5']">
                  {{ enumLabel('product_status', pl.status) }}
                </span>
              </td>
              <td class="px-3 py-4 text-sm text-gray-500">{{ pl.productCount || 0 }}</td>
              <td class="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                <button type="button" class="text-orange-600 hover:text-orange-900 mr-3" @click="openEditPriceList(pl)">{{ t('admin.pricing.edit') }}</button>
                <button type="button" class="text-red-600 hover:text-red-900" @click="deletePriceList(pl.id)">{{ t('admin.pricing.delete') }}</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Product Pricing Tab -->
    <div v-if="activeTab === 'product-pricing'" class="mt-6 space-y-6">
      <div class="flex flex-wrap items-center gap-4">
        <div class="flex-1 min-w-[300px]">
          <label for="pricing-product" class="block text-sm font-medium text-gray-700">{{ t('admin.pricing.select_product') }}</label>
          <select id="pricing-product" v-model="selectedProductId" name="productId" @change="fetchProductPrices" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
            <option value="">{{ t('admin.pricing.select_product_placeholder') }}</option>
            <option v-for="product in products" :key="product.id" :value="product.id">{{ product.name }}</option>
          </select>
        </div>
      </div>

      <div v-if="selectedProductId" class="overflow-hidden rounded-lg bg-white shadow ring-1 ring-black ring-opacity-5">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200 flex justify-between items-center">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.pricing.pricing_rules') }}</h3>
          <button type="button" @click="openAddPriceRule" class="inline-flex items-center rounded-md border border-transparent bg-orange-600 px-3 py-1 text-sm font-medium text-white shadow-sm hover:bg-orange-700">
            {{ t('admin.pricing.add_rule') }}
          </button>
        </div>
        <table class="min-w-full divide-y divide-gray-300">
          <thead class="bg-gray-50">
            <tr>
              <th class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">{{ t('admin.pricing.price_list') }}</th>
              <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.pricing.min_qty') }}</th>
              <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.pricing.unit_price') }}</th>
              <th class="relative py-3.5 pl-3 pr-4 sm:pr-6"><span class="sr-only">{{ t('admin.pricing.actions') }}</span></th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 bg-white">
            <tr v-if="loadingPrices">
              <td colspan="4" class="py-5 text-center text-sm text-gray-500">{{ t('admin.pricing.loading') }}</td>
            </tr>
            <tr v-else-if="productPrices.length === 0">
              <td colspan="4" class="py-5 text-center text-sm text-gray-500">{{ t('admin.pricing.no_prices') }}</td>
            </tr>
            <tr v-else v-for="price in productPrices" :key="price.id">
              <td class="py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">{{ price.priceListName || price.priceListId }}</td>
              <td class="px-3 py-4 text-sm text-gray-500">{{ price.minQuantity || 0 }}</td>
              <td class="px-3 py-4 text-sm text-gray-500">{{ cur(price.currency) }} {{ formatNumber(price.unitPrice || 0) }}</td>
              <td class="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                <button type="button" class="text-red-600 hover:text-red-900" @click="deletePrice(price.id)">{{ t('admin.pricing.delete') }}</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Price List Modal -->
    <div v-if="showPriceListModal" class="fixed inset-0 z-10 overflow-y-auto" role="dialog" aria-modal="true">
      <div class="flex min-h-screen items-end justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <button type="button" class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity w-full border-0 cursor-pointer" @click="closePriceListModal" :aria-label="t('common.close')"></button>
        <span class="hidden sm:inline-block sm:h-screen sm:align-middle" aria-hidden="true">&#8203;</span>
        <div class="inline-block w-full transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left align-bottom shadow-xl sm:my-8 sm:max-w-lg sm:p-6 sm:align-middle">
          <h3 class="text-lg font-medium leading-6 text-gray-900">{{ editingPriceListId ? t('admin.pricing.edit_price_list') : t('admin.pricing.create_price_list') }}</h3>
          <form @submit.prevent="savePriceList" class="mt-4 space-y-4">
            <div>
              <label for="pricelist-name" class="block text-sm font-medium text-gray-700">{{ t('admin.pricing.name') }}</label>
              <input id="pricelist-name" v-model="priceListForm.name" name="name" autocomplete="off" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="pricelist-currency" class="block text-sm font-medium text-gray-700">{{ t('admin.pricing.currency') }}</label>
              <select id="pricelist-currency" v-model="priceListForm.currency" name="currency" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                <option value="USD">USD</option>
                <option value="EUR">EUR</option>
                <option value="GBP">GBP</option>
                <option value="CNY">CNY</option>
              </select>
            </div>
            <div>
              <label for="pricelist-status" class="block text-sm font-medium text-gray-700">{{ t('admin.pricing.status') }}</label>
              <select id="pricelist-status" v-model="priceListForm.status" name="status" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                <option value="active">{{ enumLabel('product_status', 'active') }}</option>
                <option value="inactive">{{ enumLabel('product_status', 'inactive') }}</option>
              </select>
            </div>
            <div v-if="priceListError" class="text-sm text-red-600">{{ priceListError }}</div>
            <div class="flex justify-end gap-3">
              <button type="button" @click="closePriceListModal" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700">{{ t('admin.pricing.cancel') }}</button>
              <button type="submit" :disabled="savingPriceList" class="rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm text-white disabled:opacity-50">{{ t('admin.pricing.save') }}</button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Price Rule Modal -->
    <div v-if="showPriceRuleModal" class="fixed inset-0 z-10 overflow-y-auto" role="dialog" aria-modal="true">
      <div class="flex min-h-screen items-end justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <button type="button" class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity w-full border-0 cursor-pointer" @click="closePriceRuleModal" :aria-label="t('common.close')"></button>
        <span class="hidden sm:inline-block sm:h-screen sm:align-middle" aria-hidden="true">&#8203;</span>
        <div class="inline-block w-full transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left align-bottom shadow-xl sm:my-8 sm:max-w-lg sm:p-6 sm:align-middle">
          <h3 class="text-lg font-medium leading-6 text-gray-900">{{ t('admin.pricing.add_pricing_rule') }}</h3>
          <form @submit.prevent="savePriceRule" class="mt-4 space-y-4">
            <div>
              <label for="pricerule-priceListId" class="block text-sm font-medium text-gray-700">{{ t('admin.pricing.price_list') }}</label>
              <select id="pricerule-priceListId" v-model="priceRuleForm.priceListId" name="priceListId" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                <option value="">{{ t('admin.pricing.select_price_list') }}</option>
                <option v-for="pl in priceLists" :key="pl.id" :value="pl.id">{{ pl.name }}</option>
              </select>
            </div>
            <div>
              <label for="pricerule-minQuantity" class="block text-sm font-medium text-gray-700">{{ t('admin.pricing.min_qty') }}</label>
              <input id="pricerule-minQuantity" v-model.number="priceRuleForm.minQuantity" name="minQuantity" type="number" min="1" autocomplete="off" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="pricerule-unitPrice" class="block text-sm font-medium text-gray-700">{{ t('admin.pricing.unit_price') }}</label>
              <input id="pricerule-unitPrice" v-model.number="priceRuleForm.unitPrice" name="unitPrice" type="number" min="0" step="0.01" autocomplete="off" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div v-if="priceRuleError" class="text-sm text-red-600">{{ priceRuleError }}</div>
            <div class="flex justify-end gap-3">
              <button type="button" @click="closePriceRuleModal" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700">{{ t('admin.pricing.cancel') }}</button>
              <button type="submit" :disabled="savingPriceRule" class="rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm text-white disabled:opacity-50">{{ t('admin.pricing.save') }}</button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const api = useApi()
const { t } = useI18n()
const { currencyOrDefault: cur, enumLabel, formatNumber, formatDate } = useDisplay()

const activeTab = ref('price-lists')
const priceLists = ref<any[]>([])
const loadingPriceLists = ref(false)
const showPriceListModal = ref(false)
const editingPriceListId = ref('')
const savingPriceList = ref(false)
const priceListError = ref('')
const priceListForm = reactive({ name: '', currency: cur(null), status: 'active' })
const products = ref<any[]>([])
const selectedProductId = ref('')
const productPrices = ref<any[]>([])
const loadingPrices = ref(false)
const showPriceRuleModal = ref(false)
const savingPriceRule = ref(false)
const priceRuleError = ref('')
const priceRuleForm = reactive({ priceListId: '', minQuantity: 1, unitPrice: 0 })

const fetchPriceLists = async () => {
  loadingPriceLists.value = true
  try { const res = await api.get<any>('/admin/price-lists'); priceLists.value = res.data || [] }
  catch (err: any) { console.error('Failed to fetch price lists:', err) }
  finally { loadingPriceLists.value = false }
}

const fetchProducts = async () => {
  try { const res = await api.get<any>('/admin/products?limit=100'); products.value = res.data || [] }
  catch (err: any) { console.error('Failed to fetch products:', err) }
}

const fetchProductPrices = async () => {
  if (!selectedProductId.value) return
  loadingPrices.value = true
  try { productPrices.value = await api.get<any[]>(`/admin/products/${selectedProductId.value}/prices`) || [] }
  catch (err: any) { console.error('Failed to fetch product prices:', err) }
  finally { loadingPrices.value = false }
}

const openCreatePriceList = () => { editingPriceListId.value = ''; Object.assign(priceListForm, { name: '', currency: cur(null), status: 'active' }); priceListError.value = ''; showPriceListModal.value = true }
const openEditPriceList = (pl: any) => { editingPriceListId.value = pl.id; Object.assign(priceListForm, { name: pl.name, currency: cur(pl.currency), status: pl.status || 'active' }); priceListError.value = ''; showPriceListModal.value = true }
const closePriceListModal = () => { showPriceListModal.value = false; savingPriceList.value = false; priceListError.value = '' }

const savePriceList = async () => {
  savingPriceList.value = true; priceListError.value = ''
  try {
    if (editingPriceListId.value) { await api.put(`/admin/price-lists/${editingPriceListId.value}`, priceListForm) }
    else { await api.post('/admin/price-lists', priceListForm) }
    closePriceListModal(); await fetchPriceLists()
  } catch (err: any) { priceListError.value = err?.message || t('errors.api.save_failed') }
  finally { savingPriceList.value = false }
}

const deletePriceList = async (id: string) => {
  if (!confirm(t('admin.pricing.confirm_delete_price_list'))) return
  try { await api.del(`/admin/price-lists/${id}`); await fetchPriceLists() }
  catch (err: any) { alert(err?.message || t('errors.api.delete_failed')) }
}

const openAddPriceRule = () => { Object.assign(priceRuleForm, { priceListId: '', minQuantity: 1, unitPrice: 0 }); priceRuleError.value = ''; showPriceRuleModal.value = true }
const closePriceRuleModal = () => { showPriceRuleModal.value = false; savingPriceRule.value = false; priceRuleError.value = '' }

const savePriceRule = async () => {
  savingPriceRule.value = true; priceRuleError.value = ''
  try { await api.post(`/admin/products/${selectedProductId.value}/prices`, priceRuleForm); closePriceRuleModal(); await fetchProductPrices() }
  catch (err: any) { priceRuleError.value = err?.message || t('errors.api.save_failed') }
  finally { savingPriceRule.value = false }
}

const deletePrice = async (id: string) => {
  if (!confirm(t('admin.pricing.confirm_delete_price'))) return
  try { await api.del(`/admin/products/${selectedProductId.value}/prices/${id}`); await fetchProductPrices() }
  catch (err: any) { alert(err?.message || t('errors.api.delete_failed')) }
}

onMounted(() => { fetchPriceLists(); fetchProducts() })
</script>
