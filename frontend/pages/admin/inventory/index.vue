<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('admin.inventory.title') }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ t('admin.inventory.description') }}</p>
      </div>
      <div class="mt-4 sm:mt-0 flex items-center gap-3">
        <button @click="exportInventory" class="inline-flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors">
          <Icon name="heroicons:arrow-down-tray" class="h-4 w-4" />
          {{ t('admin.inventory.export') }}
        </button>
        <button @click="openAdjustmentModal()" class="inline-flex items-center gap-2 px-4 py-2 bg-orange-600 text-white rounded-lg text-sm font-medium hover:bg-orange-700 transition-colors">
          <Icon name="heroicons:adjustments-horizontal" class="h-4 w-4" />
          {{ t('admin.inventory.adjust_stock') }}
        </button>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="mt-6 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500">{{ t('admin.inventory.total_products') }}</p>
            <p class="mt-1 text-2xl font-bold text-gray-900">{{ stats.totalProducts }}</p>
          </div>
          <div class="h-12 w-12 rounded-lg bg-orange-50 flex items-center justify-center">
            <Icon name="heroicons:cube" class="h-6 w-6 text-orange-600" />
          </div>
        </div>
      </div>
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500">{{ t('admin.inventory.low_stock') }}</p>
            <p class="mt-1 text-2xl font-bold text-red-600">{{ stats.lowStock }}</p>
          </div>
          <div class="h-12 w-12 rounded-lg bg-red-50 flex items-center justify-center">
            <Icon name="heroicons:exclamation-triangle" class="h-6 w-6 text-red-600" />
          </div>
        </div>
      </div>
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500">{{ t('admin.inventory.out_of_stock') }}</p>
            <p class="mt-1 text-2xl font-bold text-orange-600">{{ stats.outOfStock }}</p>
          </div>
          <div class="h-12 w-12 rounded-lg bg-orange-50 flex items-center justify-center">
            <Icon name="heroicons:x-circle" class="h-6 w-6 text-orange-600" />
          </div>
        </div>
      </div>
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500">{{ t('admin.inventory.total_value') }}</p>
            <p class="mt-1 text-2xl font-bold text-emerald-600">${{ stats.totalValue.toLocaleString() }}</p>
          </div>
          <div class="h-12 w-12 rounded-lg bg-emerald-50 flex items-center justify-center">
            <Icon name="heroicons:currency-dollar" class="h-6 w-6 text-emerald-600" />
          </div>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 p-4">
      <div class="flex flex-wrap items-center gap-4">
        <div class="flex-1 min-w-[200px]">
          <div class="relative">
            <Icon name="heroicons:magnifying-glass" class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
            <input v-model="searchQuery" type="text" :placeholder="t('admin.inventory.search')" class="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500" />
          </div>
        </div>
        <select v-model="stockFilter" class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500">
          <option value="all">{{ t('admin.inventory.filter_all') }}</option>
          <option value="low">{{ t('admin.inventory.filter_low') }}</option>
          <option value="out">{{ t('admin.inventory.filter_out') }}</option>
          <option value="in">{{ t('admin.inventory.filter_in') }}</option>
        </select>
        <select v-model="categoryFilter" class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500">
          <option value="">{{ t('admin.inventory.all_categories') }}</option>
          <option v-for="cat in categories" :key="cat" :value="cat">{{ cat }}</option>
        </select>
      </div>
    </div>

    <!-- Inventory Table -->
    <div class="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.inventory.col_product') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.inventory.col_category') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.inventory.col_sku') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider text-right">{{ t('admin.inventory.col_stock') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider text-right">{{ t('admin.inventory.col_moq') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider text-right">{{ t('admin.inventory.col_unit_value') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider text-right">{{ t('admin.inventory.col_total_value') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.inventory.col_status') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.inventory.col_actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 bg-white">
            <tr v-if="pending">
              <td colspan="9" class="px-6 py-10 text-center text-sm text-gray-500">{{ t('admin.inventory.loading') }}</td>
            </tr>
            <tr v-else-if="error">
              <td colspan="9" class="px-6 py-10 text-center text-sm text-red-600">{{ error }}</td>
            </tr>
            <tr v-else-if="filteredItems.length === 0">
              <td colspan="9" class="px-6 py-10 text-center text-sm text-gray-500">{{ t('admin.inventory.no_data') }}</td>
            </tr>
            <tr v-else v-for="item in filteredItems" :key="item.id" class="hover:bg-gray-50 transition-colors">
              <td class="px-6 py-4">
                <div class="flex items-center gap-3">
                  <div class="h-10 w-10 rounded-lg bg-gradient-to-br from-orange-50 to-amber-50 flex items-center justify-center overflow-hidden">
                    <img v-if="item.thumbnail" :src="item.thumbnail" class="h-full w-full object-cover" />
                    <Icon v-else name="heroicons:cube" class="h-5 w-5 text-blue-300" />
                  </div>
                  <div>
                    <div class="font-medium text-gray-900">{{ item.name }}</div>
                    <div class="text-xs text-gray-500">{{ item.slug }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 text-sm text-gray-600">{{ item.category || '-' }}</td>
              <td class="px-6 py-4 text-sm text-gray-600 font-mono">{{ item.sku || item.id.substring(0, 8).toUpperCase() }}</td>
              <td class="px-6 py-4 text-sm text-right">
                <div class="flex items-center justify-end gap-2">
                  <span :class="stockStatusClass(item.stockQuantity, item.moq)" class="font-semibold">{{ item.stockQuantity || 0 }}</span>
                  <button @click="openAdjustmentModal(item)" class="p-1 text-orange-600 hover:bg-orange-50 rounded transition-colors">
                    <Icon name="heroicons:pencil" class="h-4 w-4" />
                  </button>
                </div>
              </td>
              <td class="px-6 py-4 text-sm text-right text-gray-600">{{ item.moq || 0 }}</td>
              <td class="px-6 py-4 text-sm text-right text-gray-600">${{ (item.unitValue || 0).toLocaleString() }}</td>
              <td class="px-6 py-4 text-sm text-right font-medium text-gray-900">${{ ((item.stockQuantity || 0) * (item.unitValue || 0)).toLocaleString() }}</td>
              <td class="px-6 py-4">
                <span :class="stockBadgeClass(item.stockQuantity, item.moq)" class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold">
                  {{ stockStatusLabel(item.stockQuantity, item.moq) }}
                </span>
              </td>
              <td class="px-6 py-4 text-sm">
                <button @click="viewHistory(item)" class="text-orange-600 hover:text-orange-900 mr-3">{{ t('admin.inventory.history') }}</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="pagination" class="px-6 py-4 border-t border-gray-200 flex items-center justify-between">
        <div class="text-sm text-gray-600">
          {{ t('admin.inventory.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, pagination.total), total: pagination.total }) }}
        </div>
        <div class="flex items-center gap-2">
          <button @click="prevPage" :disabled="page <= 1" class="px-3 py-1.5 border border-gray-200 rounded-lg text-sm hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
            {{ t('admin.inventory.previous') }}
          </button>
          <button @click="nextPage" :disabled="page >= pagination.totalPages" class="px-3 py-1.5 border border-gray-200 rounded-lg text-sm hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
            {{ t('admin.inventory.next') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Stock Adjustment Modal -->
    <div v-if="showAdjustmentModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-screen items-end justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" @click="closeAdjustmentModal"></div>
        <div class="inline-block w-full transform overflow-hidden rounded-xl bg-white text-left align-bottom shadow-xl transition-all sm:my-8 sm:max-w-lg sm:align-middle">
          <div class="bg-gradient-to-r from-orange-500 to-amber-600 px-6 py-4">
            <h3 class="text-lg font-semibold text-white">{{ t('admin.inventory.adjust_title') }}</h3>
            <p class="text-sm text-blue-100 mt-0.5">{{ adjustmentProduct?.name }}</p>
          </div>
          <div class="p-6 space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.inventory.current_stock') }}</label>
                <p class="mt-1 text-2xl font-bold text-gray-900">{{ adjustmentProduct?.stockQuantity || 0 }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.inventory.new_stock') }}</label>
                <input v-model.number="newStock" type="number" min="0" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.inventory.adjustment_type') }}</label>
              <div class="mt-2 flex gap-4">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input v-model="adjustmentType" type="radio" value="set" class="text-orange-600 focus:ring-orange-500" />
                  <span class="text-sm text-gray-700">{{ t('admin.inventory.type_set') }}</span>
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                  <input v-model="adjustmentType" type="radio" value="add" class="text-orange-600 focus:ring-orange-500" />
                  <span class="text-sm text-gray-700">{{ t('admin.inventory.type_add') }}</span>
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                  <input v-model="adjustmentType" type="radio" value="subtract" class="text-orange-600 focus:ring-orange-500" />
                  <span class="text-sm text-gray-700">{{ t('admin.inventory.type_subtract') }}</span>
                </label>
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.inventory.reason') }}</label>
              <select v-model="adjustmentReason" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500">
                <option value="restock">{{ t('admin.inventory.reason_restock') }}</option>
                <option value="correction">{{ t('admin.inventory.reason_correction') }}</option>
                <option value="damage">{{ t('admin.inventory.reason_damage') }}</option>
                <option value="return">{{ t('admin.inventory.reason_return') }}</option>
                <option value="audit">{{ t('admin.inventory.reason_audit') }}</option>
                <option value="other">{{ t('admin.inventory.reason_other') }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.inventory.notes') }}</label>
              <textarea v-model="adjustmentNotes" rows="2" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"></textarea>
            </div>
            <div v-if="adjustmentError" class="text-sm text-red-600">{{ adjustmentError }}</div>
          </div>
          <div class="px-6 py-4 bg-gray-50 flex justify-end gap-3">
            <button @click="closeAdjustmentModal" class="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition-colors">
              {{ t('admin.inventory.cancel') }}
            </button>
            <button @click="submitAdjustment" :disabled="savingAdjustment" class="px-4 py-2 bg-orange-600 text-white rounded-lg text-sm font-medium hover:bg-orange-700 disabled:opacity-50 transition-colors flex items-center gap-2">
              <Icon v-if="savingAdjustment" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" />
              {{ t('admin.inventory.save') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- History Modal -->
    <div v-if="showHistoryModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-screen items-end justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" @click="showHistoryModal = false"></div>
        <div class="inline-block w-full transform overflow-hidden rounded-xl bg-white text-left align-bottom shadow-xl transition-all sm:my-8 sm:max-w-2xl sm:align-middle">
          <div class="bg-gradient-to-r from-gray-700 to-gray-900 px-6 py-4 flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold text-white">{{ t('admin.inventory.history_title') }}</h3>
              <p class="text-sm text-gray-300 mt-0.5">{{ historyProduct?.name }}</p>
            </div>
            <button @click="showHistoryModal = false" class="text-gray-300 hover:text-white">
              <Icon name="heroicons:x-mark" class="h-5 w-5" />
            </button>
          </div>
          <div class="p-6 max-h-96 overflow-y-auto">
            <div v-if="stockHistory.length === 0" class="text-center py-8 text-gray-500">
              {{ t('admin.inventory.no_history') }}
            </div>
            <div v-else class="space-y-3">
              <div v-for="entry in stockHistory" :key="entry.id" class="flex items-start gap-4 p-3 rounded-lg border border-gray-100 hover:bg-gray-50">
                <div :class="entry.change > 0 ? 'bg-emerald-100 text-emerald-700' : 'bg-red-100 text-red-700'" class="flex-shrink-0 w-16 h-16 rounded-lg flex items-center justify-center font-bold text-lg">
                  {{ entry.change > 0 ? '+' : '' }}{{ entry.change }}
                </div>
                <div class="flex-1">
                  <div class="flex items-center justify-between">
                    <span class="font-medium text-gray-900">{{ entry.reason }}</span>
                    <span class="text-xs text-gray-500">{{ new Date(entry.createdAt).toLocaleString() }}</span>
                  </div>
                  <p class="text-sm text-gray-600 mt-0.5">{{ entry.notes }}</p>
                  <div class="text-xs text-gray-400 mt-1">
                    {{ t('admin.inventory.from_to', { from: entry.previousQty, to: entry.newQty }) }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()

const items = ref<any[]>([])
const categories = ref<string[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20
const searchQuery = ref('')
const stockFilter = ref('all')
const categoryFilter = ref('')
const showAdjustmentModal = ref(false)
const adjustmentProduct = ref<any>(null)
const newStock = ref(0)
const adjustmentType = ref('set')
const adjustmentReason = ref('restock')
const adjustmentNotes = ref('')
const savingAdjustment = ref(false)
const adjustmentError = ref('')
const showHistoryModal = ref(false)
const historyProduct = ref<any>(null)
const stockHistory = ref<any[]>([])

const stats = computed(() => {
  const total = items.value.length
  const low = items.value.filter(i => i.stockQuantity > 0 && i.stockQuantity <= (i.moq || 10)).length
  const out = items.value.filter(i => !i.stockQuantity || i.stockQuantity <= 0).length
  const value = items.value.reduce((sum, i) => sum + ((i.stockQuantity || 0) * (i.unitValue || 0)), 0)
  return { totalProducts: total, lowStock: low, outOfStock: out, totalValue: value }
})

const filteredItems = computed(() => items.value.filter(item => {
  const matchesSearch = !searchQuery.value || item.name?.toLowerCase().includes(searchQuery.value.toLowerCase()) || item.slug?.toLowerCase().includes(searchQuery.value.toLowerCase())
  const matchesCategory = !categoryFilter.value || item.category === categoryFilter.value
  const matchesStock = (stockFilter.value === 'all') || (stockFilter.value === 'low' && item.stockQuantity > 0 && item.stockQuantity <= (item.moq || 10)) || (stockFilter.value === 'out' && (!item.stockQuantity || item.stockQuantity <= 0)) || (stockFilter.value === 'in' && item.stockQuantity > (item.moq || 10))
  return matchesSearch && matchesCategory && matchesStock
}))

const fetchInventory = async () => {
  pending.value = true; error.value = ''
  try {
    const res = await api.get<any>('/admin/inventory', { page: page.value, limit: pageSize })
    items.value = res.data || []
    pagination.value = res.pagination
    categories.value = Array.from(new Set(items.value.map((i: any) => i.category).filter(Boolean)))
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally { pending.value = false }
}

const openAdjustmentModal = (product?: any) => {
  adjustmentProduct.value = product || null; newStock.value = product?.stockQuantity || 0
  adjustmentType.value = 'set'; adjustmentReason.value = 'restock'; adjustmentNotes.value = ''; adjustmentError.value = ''; showAdjustmentModal.value = true
}
const closeAdjustmentModal = () => { showAdjustmentModal.value = false; savingAdjustment.value = false }

const submitAdjustment = async () => {
  if (!adjustmentProduct.value) return
  savingAdjustment.value = true; adjustmentError.value = ''
  let finalQty = newStock.value
  if (adjustmentType.value === 'add') finalQty = (adjustmentProduct.value.stockQuantity || 0) + newStock.value
  else if (adjustmentType.value === 'subtract') finalQty = Math.max(0, (adjustmentProduct.value.stockQuantity || 0) - newStock.value)
  try {
    await api.put(`/admin/inventory/${adjustmentProduct.value.id}`, { stockQuantity: finalQty, reason: adjustmentReason.value, notes: adjustmentNotes.value })
    closeAdjustmentModal(); await fetchInventory()
  } catch (err: any) { adjustmentError.value = err?.message || t('errors.api.stock_adjust_failed') }
  finally { savingAdjustment.value = false }
}

const viewHistory = async (product: any) => {
  historyProduct.value = product; showHistoryModal.value = true
  try { stockHistory.value = await api.get<any[]>(`/admin/inventory/${product.id}/history`) || [] }
  catch { stockHistory.value = [] }
}

const exportInventory = () => {
  const csv = [['Name', 'Category', 'SKU', 'Stock', 'MOQ', 'Unit Value', 'Total Value'].join(','), ...filteredItems.value.map(i => [`"${i.name}"`, `"${i.category || ''}"`, `"${i.sku || i.id}"`, i.stockQuantity || 0, i.moq || 0, i.unitValue || 0, ((i.stockQuantity || 0) * (i.unitValue || 0)).toFixed(2)].join(','))].join('\n')
  const blob = new Blob([csv], { type: 'text/csv' }); const url = URL.createObjectURL(blob)
  const a = document.createElement('a'); a.href = url; a.download = `inventory-${new Date().toISOString().split('T')[0]}.csv`; a.click()
}

const stockStatusClass = (qty: number, moq: number) => { if (!qty || qty <= 0) return 'text-red-600'; if (qty <= (moq || 10)) return 'text-orange-600'; return 'text-emerald-600' }
const stockBadgeClass = (qty: number, moq: number) => { if (!qty || qty <= 0) return 'bg-red-100 text-red-800'; if (qty <= (moq || 10)) return 'bg-orange-100 text-orange-800'; return 'bg-emerald-100 text-emerald-800' }
const stockStatusLabel = (qty: number, moq: number) => { if (!qty || qty <= 0) return 'Out of Stock'; if (qty <= (moq || 10)) return 'Low Stock'; return 'In Stock' }
const prevPage = () => { if (page.value > 1) { page.value -= 1; fetchInventory() } }
const nextPage = () => { if (pagination.value && page.value < pagination.value.totalPages) { page.value += 1; fetchInventory() } }

onMounted(fetchInventory)
</script>
