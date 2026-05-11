<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="bg-white shadow px-4 py-5 sm:rounded-lg sm:p-6 mb-6">
      <div class="md:flex md:items-center md:justify-between">
        <div class="flex-1 min-w-0">
          <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
            {{ t('customer.trades.dashboard_title', { id: tradeId }) }}
          </h2>
          <div class="mt-1 flex flex-col sm:flex-row sm:flex-wrap sm:mt-0 sm:space-x-6">
            <div class="mt-2 flex items-center text-sm text-gray-500">
              <Icon name="heroicons:tag" class="flex-shrink-0 mr-1.5 h-5 w-5 text-gray-400" />
              {{ t('customer.trades.status_label') }}: {{ trade?.status || t('customer.trades.loading') }}
            </div>
            <div class="mt-2 flex items-center text-sm text-gray-500">
              <Icon name="heroicons:currency-dollar" class="flex-shrink-0 mr-1.5 h-5 w-5 text-gray-400" />
              {{ t('customer.trades.currency_label') }}: {{ trade?.currency }}
            </div>
            <div class="mt-2 flex items-center text-sm text-gray-500">
              <Icon name="heroicons:truck" class="flex-shrink-0 mr-1.5 h-5 w-5 text-gray-400" />
              {{ t('customer.trades.incoterms_info') }}: {{ trade?.incoterms || trade?.terms || t('common.display.tbd') }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div v-if="loadError" class="lg:col-span-3 rounded-md bg-red-50 p-3 text-sm text-red-700">{{ loadError }}</div>

      <!-- Left: AI Chat -->
      <div class="lg:col-span-1 bg-white shadow sm:rounded-lg h-[600px] flex flex-col">
        <div class="px-4 py-5 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.trades.ai_assistant') }}</h3>
          <p class="mt-1 text-sm text-gray-500">{{ t('customer.trades.ai_assistant_desc') }}</p>
        </div>
        <div class="flex-1 p-4 overflow-y-auto bg-gray-50" ref="chatContainer">
          <div v-for="(msg, index) in messages" :key="index" class="mb-4">
            <div :class="['rounded-lg p-3 max-w-sm', msg.role === 'user' ? 'bg-orange-600 text-white ml-auto' : 'bg-white border text-gray-800 mr-auto']">
              <div v-if="msg.sender" class="text-xs font-semibold mb-1 opacity-70">{{ msg.sender }}</div>
              <div class="whitespace-pre-wrap text-sm">{{ msg.content }}</div>
              <div v-if="msg.tool_calls?.length" class="mt-2 border-t pt-2 border-gray-200">
                <p class="text-xs font-bold text-gray-500 mb-1">{{ t('customer.trades.tools_called') }}:</p>
                <div v-for="tc in msg.tool_calls" :key="tc.id" class="text-xs text-orange-800 bg-orange-50 rounded p-1 mb-1">{{ tc.function.name }}</div>
              </div>
            </div>
          </div>
          <div v-if="streamingText" class="mb-4">
            <div class="rounded-lg p-3 max-w-sm bg-white border text-gray-800 mr-auto">
              <div class="text-xs font-semibold mb-1 opacity-70">{{ t('customer.trades.ai_typing') }}</div>
              <div class="whitespace-pre-wrap text-sm">{{ streamingText }}</div>
            </div>
          </div>
        </div>
        <div class="p-4 border-t border-gray-200 flex space-x-2 bg-white rounded-b-lg">
          <input v-model="inputQuery" @keyup.enter="sendMessage" type="text"
            :placeholder="t('customer.trades.ai_input_placeholder')"
            class="flex-1 shadow-sm focus:ring-orange-500 focus:border-orange-500 block w-full sm:text-sm border-gray-300 rounded-md"
            :disabled="isStreaming" />
          <button @click="sendMessage" :disabled="isStreaming || !inputQuery.trim()"
            class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-orange-600 hover:bg-orange-700 disabled:opacity-50">
            {{ t('customer.trades.send') }}
          </button>
        </div>
      </div>

      <!-- Right: Documents + Compliance -->
      <div class="lg:col-span-2 space-y-6">
        <!-- Rich Documents -->
        <div class="bg-white shadow sm:rounded-lg">
          <div class="px-4 py-5 border-b border-gray-200 flex items-center justify-between">
            <div>
              <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.trades.documents_title') }}</h3>
              <p class="mt-1 text-sm text-gray-500">{{ t('customer.trades.documents_subtitle') }}</p>
            </div>
            <button @click="refreshDocs" class="text-xs text-orange-600 hover:text-orange-800">
              <Icon name="heroicons:arrow-path" class="h-4 w-4" />
            </button>
          </div>
          <div class="p-4">
            <div v-if="loadingDocs" class="text-gray-400 text-sm py-4 text-center">{{ t('customer.trades.loading') }}</div>
            <div v-else-if="!richDocs.length && !documents.length" class="text-gray-500 text-sm py-4">
              {{ t('customer.trades.no_documents') }}
            </div>
            <ul v-else class="divide-y divide-gray-200">
              <!-- Rich structured documents -->
              <li v-for="doc in richDocs" :key="doc.key" class="py-3 flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <Icon name="heroicons:document-text" class="h-5 w-5 text-orange-400" />
                  <div>
                    <p class="text-sm font-medium text-gray-900">{{ doc.label }}</p>
                    <p class="text-xs text-gray-500">{{ doc.number }} · {{ doc.status }}</p>
                  </div>
                </div>
                <button @click="viewDoc(doc)" class="text-orange-600 hover:text-orange-900 text-sm font-medium">
                  {{ t('customer.trades.view') }}
                </button>
              </li>
              <!-- Generic trade documents -->
              <li v-for="doc in documents" :key="doc.id" class="py-3 flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <Icon name="heroicons:document" class="h-5 w-5 text-gray-400" />
                  <div>
                    <p class="text-sm font-medium text-gray-900">{{ doc.docType || doc.type }}</p>
                    <p class="text-xs text-gray-500">{{ doc.docNumber }} · {{ doc.status }}</p>
                  </div>
                </div>
              </li>
            </ul>
          </div>
        </div>

        <!-- Document Detail Modal -->
        <div v-if="selectedDoc" class="bg-white shadow sm:rounded-lg border border-orange-100">
          <div class="px-4 py-4 border-b border-gray-200 flex items-center justify-between bg-orange-50">
            <h4 class="font-semibold text-gray-900">{{ selectedDoc.label }}</h4>
            <button @click="selectedDoc = null" class="text-gray-400 hover:text-gray-600">
              <Icon name="heroicons:x-mark" class="h-5 w-5" />
            </button>
          </div>
          <div class="p-4">
            <dl class="grid grid-cols-2 gap-3 text-sm">
              <template v-for="(val, key) in selectedDoc.data" :key="key">
                <div v-if="val !== null && val !== undefined && val !== '' && key !== 'id' && key !== 'transactionId'">
                  <dt class="text-gray-500 capitalize">{{ formatKey(String(key)) }}</dt>
                  <dd class="font-medium text-gray-900 break-words">{{ formatVal(val) }}</dd>
                </div>
              </template>
            </dl>
          </div>
        </div>

        <!-- Compliance -->
        <div class="bg-white shadow sm:rounded-lg">
          <div class="px-4 py-5 border-b border-gray-200">
            <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.trades.compliance_title') }}</h3>
          </div>
          <div class="p-4">
            <div v-if="!compliance.length" class="text-gray-500 text-sm py-4">{{ t('customer.trades.compliance_pending') }}</div>
            <ul v-else class="divide-y divide-gray-200">
              <li v-for="c in compliance" :key="c.id" class="py-3 flex items-center justify-between">
                <div>
                  <p class="text-sm font-medium text-gray-900">{{ c.countryCode }}</p>
                  <p class="text-xs text-gray-500">{{ c.language }}</p>
                </div>
                <span :class="complianceStatusClass(c.status)" class="px-2 py-0.5 text-xs font-medium rounded-full">{{ c.status }}</span>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, nextTick } from 'vue'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const route = useRoute()
const { token } = useAuth()
const { t } = useI18n()
const { formatDate } = useDisplay()
const api = useApi()
const config = useRuntimeConfig()
const baseURL = config.public.apiBase || '/api/v1'
const tradeId = route.params.id

const trade = ref<any>(null)
const loadError = ref('')
const loadingDocs = ref(false)
const compliance = ref<any[]>([])
const selectedDoc = ref<any>(null)

const messages = ref<{role: string, content: string, sender?: string, tool_calls?: any[]}[]>([])
const inputQuery = ref('')
const isStreaming = ref(false)
const streamingText = ref('')
const chatContainer = ref<HTMLElement | null>(null)

// Raw generic documents from trade
const documents = computed(() => {
  if (!trade.value) return []
  return Array.isArray(trade.value.documents) ? trade.value.documents
    : Array.isArray(trade.value.Documents) ? trade.value.Documents : []
})

// Rich structured documents fetched individually
const piData = ref<any>(null)
const ciData = ref<any>(null)
const blData = ref<any>(null)
const scData = ref<any>(null)
const plData = ref<any>(null)
const cooData = ref<any>(null)
const hcData = ref<any>(null)

const dash = () => t('common.display.em_dash')
const richDocs = computed(() => {
  const list: any[] = []
  if (piData.value) list.push({ key: 'pi', label: t('customer.trades.doc_pi'), number: piData.value.piNumber, status: piData.value.status, data: piData.value })
  if (ciData.value) list.push({ key: 'ci', label: t('customer.trades.doc_ci'), number: ciData.value.ciNumber, status: ciData.value.status, data: ciData.value })
  if (blData.value) list.push({ key: 'bl', label: t('customer.trades.doc_bl'), number: blData.value.blNumber, status: blData.value.status, data: blData.value })
  if (scData.value) list.push({ key: 'sc', label: t('customer.trades.doc_sc'), number: scData.value.contractNo, status: scData.value.status || dash(), data: scData.value })
  if (plData.value) list.push({ key: 'pl', label: t('customer.trades.doc_pl'), number: plData.value.plNumber, status: dash(), data: plData.value })
  if (cooData.value) list.push({ key: 'coo', label: t('customer.trades.doc_coo'), number: cooData.value.certificateNo, status: dash(), data: cooData.value })
  if (hcData.value) list.push({ key: 'hc', label: t('customer.trades.doc_hc'), number: hcData.value.certificateNo, status: dash(), data: hcData.value })
  return list
})

const fetchDoc = async (path: string) => {
  try {
    return await api.get<any>(`/user/trades/${tradeId}/${path}`)
  } catch { return null }
}

const fetchTradeDetails = async () => {
  try {
    trade.value = await api.get<any>(`/user/trades/${tradeId}`)
  } catch (err: any) {
    loadError.value = err?.message || t('errors.api.trade_load_failed')
  }
}

const fetchAllDocs = async () => {
  loadingDocs.value = true
  const [pi, ci, bl, sc, pl, coo, hc] = await Promise.all([
    fetchDoc('proforma-invoice'),
    fetchDoc('commercial-invoice'),
    fetchDoc('bill-of-lading'),
    fetchDoc('sales-contract'),
    fetchDoc('packing-list'),
    fetchDoc('certificate-of-origin'),
    fetchDoc('health-certificate'),
  ])
  piData.value = pi; ciData.value = ci; blData.value = bl
  scData.value = sc; plData.value = pl; cooData.value = coo; hcData.value = hc
  loadingDocs.value = false
}

const refreshDocs = () => fetchAllDocs()

const viewDoc = (doc: any) => { selectedDoc.value = doc }

const formatKey = (key: string) => key.replace(/([A-Z])/g, ' $1').replace(/^./, s => s.toUpperCase())
const formatVal = (val: any) => {
  if (val === null || val === undefined) return '-'
  if (typeof val === 'boolean') return val ? 'Yes' : 'No'
  if (typeof val === 'string' && val.match(/^\d{4}-\d{2}-\d{2}T/)) return formatDate(val)
  return String(val)
}

const complianceStatusClass = (status: string) => {
  if (status === 'PASSED') return 'bg-green-100 text-green-800'
  if (status === 'FAILED') return 'bg-red-100 text-red-800'
  return 'bg-yellow-100 text-yellow-800'
}

const scrollToBottom = () => {
  nextTick(() => { if (chatContainer.value) chatContainer.value.scrollTop = chatContainer.value.scrollHeight })
}

const sendMessage = () => {
  if (!inputQuery.value.trim() || isStreaming.value) return
  const query = inputQuery.value
  messages.value.push({ role: 'user', content: query })
  inputQuery.value = ''
  isStreaming.value = true
  streamingText.value = ''
  scrollToBottom()
  const url = `${baseURL}/user/ai/stream?query=${encodeURIComponent(query)}&tradeId=${encodeURIComponent(String(tradeId))}`
  streamFetch(url)
}

const streamFetch = async (url: string) => {
  try {
    const response = await fetch(url, { headers: { 'Authorization': `Bearer ${token.value}` } })
    if (!response.ok) {
      messages.value.push({ role: 'ai', content: `*${t('customer.trades.error_connecting')}*`, sender: 'System' })
      isStreaming.value = false
      return
    }
    const reader = response.body?.getReader()
    const decoder = new TextDecoder()
    let partialEvent = ''
    while (true) {
      const { done, value } = await reader!.read()
      if (done) break
      const chunk = decoder.decode(value, { stream: true })
      partialEvent += chunk
      const parts = partialEvent.split('\n\n')
      partialEvent = parts.pop() || ''
      for (const part of parts) {
        if (part.startsWith('data: ')) {
          try { handleSSEEvent(JSON.parse(part.replace('data: ', ''))) } catch(e) {}
        }
      }
    }
  } catch(e) {
    messages.value.push({ role: 'ai', content: `*${t('customer.trades.connection_interrupted')}*`, sender: 'System' })
  } finally {
    isStreaming.value = false
    if (streamingText.value) {
      messages.value.push({ role: 'ai', content: streamingText.value, sender: 'AI' })
      streamingText.value = ''
    }
    scrollToBottom()
    // Refresh docs after AI interaction (AI may have generated new documents)
    fetchAllDocs()
  }
}

const handleSSEEvent = (data: any) => {
  if (data.type === 'message' || data.type === 'tool_result') {
    if (streamingText.value) {
      messages.value.push({ role: 'ai', content: streamingText.value, sender: data.agent_name || t('customer.trades.agent_ai') })
      streamingText.value = ''
    }
    messages.value.push({
      role: 'ai',
      content: data.content || '',
      sender: data.agent_name || t('customer.trades.agent_ai'),
      tool_calls: data.tool_calls,
    })
    // When AI generates a document, refresh the document list
    if (data.document_type) {
      fetchAllDocs()
    }
  } else if (data.type === 'stream_chunk') {
    streamingText.value += data.content
  } else if (data.type === 'error') {
    messages.value.push({ role: 'ai', content: `**Error:** ${data.error}`, sender: 'System' })
  }
  scrollToBottom()
}

onMounted(async () => {
  await fetchTradeDetails()
  fetchAllDocs()
})
</script>
