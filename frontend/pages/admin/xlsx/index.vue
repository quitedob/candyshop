<template>
  <div class="max-w-5xl mx-auto space-y-8">
    <PageHeader :title="t('admin.xlsx.title')" :description="t('admin.xlsx.description')">
      <template #actions>
        <AiHelpHint topic="xlsx_translate" />
      </template>
    </PageHeader>

    <div v-if="toast" class="rounded-lg px-4 py-3 text-sm" :class="toast.ok ? 'bg-green-50 text-green-800' : 'bg-red-50 text-red-800'">
      {{ toast.message }}
    </div>

    <!-- 数据导出 -->
    <section class="rounded-2xl border border-gray-200 bg-white p-6 shadow-sm">
      <h2 class="text-lg font-semibold text-gray-900">{{ t('admin.xlsx.export_section') }}</h2>
      <p class="mt-1 text-sm text-gray-500">{{ t('admin.xlsx.export_hint') }}</p>

      <div class="mt-4 flex flex-wrap gap-4">
        <div>
          <label for="export-currency" class="block text-xs font-medium text-gray-500 mb-1">{{ t('admin.xlsx.currency') }}</label>
          <select id="export-currency" v-model="exportCurrency" class="rounded-lg border border-gray-300 px-3 py-2 text-sm">
            <option v-for="c in currencies" :key="c" :value="c">{{ c }}</option>
          </select>
        </div>
        <div>
          <label for="export-months" class="block text-xs font-medium text-gray-500 mb-1">{{ t('admin.xlsx.months') }}</label>
          <input id="export-months" v-model.number="exportMonths" type="number" min="1" max="36" class="w-24 rounded-lg border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label for="order-status" class="block text-xs font-medium text-gray-500 mb-1">{{ t('admin.xlsx.order_status') }}</label>
          <select id="order-status" v-model="orderStatus" class="rounded-lg border border-gray-300 px-3 py-2 text-sm min-w-[140px]">
            <option value="">{{ t('admin.xlsx.all_statuses') }}</option>
            <option v-for="s in orderStatuses" :key="s" :value="s">{{ s }}</option>
          </select>
        </div>
        <div>
          <label for="trade-status" class="block text-xs font-medium text-gray-500 mb-1">{{ t('admin.xlsx.trade_status') }}</label>
          <select id="trade-status" v-model="tradeStatus" class="rounded-lg border border-gray-300 px-3 py-2 text-sm min-w-[140px]">
            <option value="">{{ t('admin.xlsx.all_statuses') }}</option>
            <option v-for="s in tradeStatuses" :key="s" :value="s">{{ s }}</option>
          </select>
        </div>
      </div>

      <div class="mt-6 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        <button
          v-for="item in exportItems"
          :key="item.type"
          type="button"
          class="flex items-center gap-3 rounded-xl border border-gray-200 px-4 py-3 text-left text-sm hover:border-orange-300 hover:bg-orange-50 transition-colors disabled:opacity-50"
          :disabled="exporting === item.type"
          @click="runExport(item.type)"
        >
          <Icon :name="item.icon" class="h-5 w-5 text-orange-500 shrink-0" />
          <span class="font-medium text-gray-800">{{ t(item.labelKey) }}</span>
          <Icon v-if="exporting === item.type" name="heroicons:arrow-path" class="h-4 w-4 ml-auto animate-spin text-gray-400" />
        </button>
      </div>
    </section>

    <!-- AI 翻译 -->
    <section class="rounded-2xl border border-gray-200 bg-white p-6 shadow-sm">
      <h2 class="text-lg font-semibold text-gray-900 inline-flex items-center">
        {{ t('admin.xlsx.translate_section') }}
        <AiHelpHint topic="xlsx_translate" size="sm" />
      </h2>
      <p class="mt-1 text-sm text-gray-500">{{ t('admin.xlsx.translate_hint') }}</p>

      <div class="mt-4 flex flex-wrap gap-2">
        <button
          type="button"
          class="px-3 py-1.5 rounded-full text-xs font-medium border transition-colors"
          :class="translateMode === 'single' ? 'bg-orange-600 text-white border-orange-600' : 'border-gray-300 text-gray-600'"
          @click="translateMode = 'single'"
        >
          {{ t('admin.xlsx.translate_mode_single') }}
        </button>
        <button
          type="button"
          class="px-3 py-1.5 rounded-full text-xs font-medium border transition-colors"
          :class="translateMode === 'batch' ? 'bg-orange-600 text-white border-orange-600' : 'border-gray-300 text-gray-600'"
          @click="translateMode = 'batch'"
        >
          {{ t('admin.xlsx.translate_mode_batch') }}
        </button>
      </div>

      <div class="mt-4 grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div>
          <label for="source-lang" class="block text-xs font-medium text-gray-500 mb-1">{{ t('admin.xlsx.source_lang') }}</label>
          <select id="source-lang" v-model="sourceLang" class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm">
            <option value="">{{ t('admin.xlsx.source_auto') }}</option>
            <option v-for="loc in ALL_LOCALES" :key="'src-' + loc" :value="loc">{{ localeLabel(loc) }}</option>
          </select>
        </div>
        <div v-if="translateMode === 'single'">
          <label for="target-lang" class="block text-xs font-medium text-gray-500 mb-1">{{ t('admin.xlsx.target_lang') }}</label>
          <select id="target-lang" v-model="targetLang" class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm">
            <option v-for="loc in ALL_LOCALES" :key="'tgt-' + loc" :value="loc">{{ localeLabel(loc) }}</option>
          </select>
        </div>
        <div v-else class="sm:col-span-2">
          <span class="block text-xs font-medium text-gray-500 mb-2">{{ t('admin.xlsx.target_langs') }}</span>
          <div class="flex flex-wrap gap-2">
            <label v-for="loc in ALL_LOCALES" :key="'batch-' + loc" class="inline-flex items-center gap-1.5 text-sm">
              <input v-model="batchLocales" type="checkbox" :value="loc" class="rounded border-gray-300 text-orange-600" />
              {{ localeLabel(loc) }}
            </label>
          </div>
        </div>
      </div>

      <div class="mt-4">
        <input ref="fileInput" type="file" accept=".xlsx" class="hidden" @change="onFileChange" />
        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-lg border border-dashed border-gray-300 px-4 py-3 text-sm text-gray-600 hover:border-orange-400 hover:text-orange-700 w-full sm:w-auto"
          @click="fileInput?.click()"
        >
          <Icon name="heroicons:document-arrow-up" class="h-5 w-5" />
          {{ t('admin.xlsx.select_file') }}
        </button>
        <p v-if="selectedFile" class="mt-2 text-sm text-gray-600">{{ t('admin.xlsx.file_selected', { name: selectedFile.name }) }}</p>
      </div>

      <button
        type="button"
        class="mt-4 inline-flex items-center gap-2 rounded-lg bg-orange-600 px-5 py-2.5 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-50"
        :disabled="translating || !selectedFile"
        @click="runTranslate"
      >
        <Icon v-if="translating" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" />
        {{ translating ? t('admin.xlsx.translating') : t('admin.xlsx.translate_btn') }}
      </button>
    </section>

    <p class="text-sm text-gray-500">
      {{ t('admin.xlsx.inventory_link') }}
      <NuxtLink :to="localePath('/admin/inventory')" class="text-orange-600 hover:underline">{{ t('admin.nav.inventory') }}</NuxtLink>
    </p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: ['auth'] })

const { t } = useI18n()
const localePath = useLocalePath()
const api = useApi()
// ALL_LOCALES 由 Nuxt 自动导入

const currencies = ['USD', 'EUR', 'CNY', 'GBP', 'JPY', 'SAR', 'AED', 'KRW', 'SGD', 'MYR', 'THB']
const orderStatuses = ['pending', 'pending_confirmation', 'pending_approval', 'confirmed', 'production', 'shipped', 'delivered', 'cancelled']
const tradeStatuses = ['DRAFT', 'PENDING', 'CONFIRMED', 'PAID', 'SHIPPED', 'COMPLETED', 'CANCELLED']

const exportItems = [
  { type: 'products' as const, labelKey: 'admin.xlsx.export_products', icon: 'material-symbols:inventory-2' },
  { type: 'orders' as const, labelKey: 'admin.xlsx.export_orders', icon: 'material-symbols:shopping-cart' },
  { type: 'revenue' as const, labelKey: 'admin.xlsx.export_revenue', icon: 'material-symbols:payments' },
  { type: 'trades' as const, labelKey: 'admin.xlsx.export_trades', icon: 'material-symbols:attach-money' },
  { type: 'customers' as const, labelKey: 'admin.xlsx.export_customers', icon: 'material-symbols:group' },
]

const exportCurrency = ref('USD')
const exportMonths = ref(12)
const orderStatus = ref('')
const tradeStatus = ref('')
const exporting = ref<string | null>(null)

const translateMode = ref<'single' | 'batch'>('single')
const sourceLang = ref('')
const targetLang = ref('en')
const batchLocales = ref<string[]>(['en', 'ja'])
const selectedFile = ref<File | null>(null)
const fileInput = ref<HTMLInputElement>()
const translating = ref(false)
const toast = ref<{ ok: boolean; message: string } | null>(null)

const localeLabels: Record<string, string> = {
  en: 'English', zh: '中文', ko: '한국어', ar: 'العربية', ja: '日本語',
  th: 'ไทย', vi: 'Tiếng Việt', id: 'Indonesia', ms: 'Melayu',
}
const localeLabel = (code: string) => localeLabels[code] || code

function showToast(ok: boolean, message: string) {
  toast.value = { ok, message }
  setTimeout(() => { toast.value = null }, 4000)
}

function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

function buildExportParams(type: string): Record<string, string> {
  const params: Record<string, string> = {}
  if (type !== 'customers') params.currency = exportCurrency.value
  if (type === 'revenue') params.months = String(exportMonths.value)
  if (type === 'orders' && orderStatus.value) params.status = orderStatus.value
  if (type === 'trades' && tradeStatus.value) params.status = tradeStatus.value
  return params
}

async function runExport(type: 'products' | 'orders' | 'revenue' | 'trades' | 'customers') {
  exporting.value = type
  try {
    const blob = await api.exportAdminXlsx(type, buildExportParams(type))
    const date = new Date().toISOString().split('T')[0]
    downloadBlob(blob as Blob, `candypro_${type}_${date}.xlsx`)
    showToast(true, t('admin.xlsx.export_success'))
  } catch (err: any) {
    showToast(false, err?.message || t('admin.xlsx.export_failed'))
  } finally {
    exporting.value = null
  }
}

function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (file && !file.name.toLowerCase().endsWith('.xlsx')) {
    showToast(false, t('admin.xlsx.invalid_file'))
    selectedFile.value = null
    return
  }
  selectedFile.value = file || null
}

async function runTranslate() {
  if (!selectedFile.value) return
  if (translateMode.value === 'batch' && batchLocales.value.length === 0) {
    showToast(false, t('admin.xlsx.select_target'))
    return
  }
  translating.value = true
  try {
    const blob = await api.translateAdminXlsx(selectedFile.value, {
      batch: translateMode.value === 'batch',
      targetLang: translateMode.value === 'single' ? targetLang.value : undefined,
      targetLangs: translateMode.value === 'batch' ? batchLocales.value : undefined,
      sourceLang: sourceLang.value || undefined,
    })
    const suffix = translateMode.value === 'batch' ? 'multi' : targetLang.value
    downloadBlob(blob as Blob, `translated_${suffix}_${Date.now()}.xlsx`)
    showToast(true, t('admin.xlsx.translate_success'))
  } catch (err: any) {
    showToast(false, err?.message || t('admin.xlsx.translate_failed'))
  } finally {
    translating.value = false
  }
}
</script>
