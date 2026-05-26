<template>
  <div>
    <div class="mb-6">
      <NuxtLink :to="localePath('/admin/trades')" class="flex items-center text-sm font-medium text-orange-600 hover:text-orange-500">
        <Icon name="heroicons:arrow-left" class="mr-1 h-4 w-4" aria-hidden="true" />
        {{ t('admin.trades.back') }}
      </NuxtLink>
    </div>

    <div v-if="pending" class="text-center py-10 text-gray-500">{{ t('admin.trades.loading_details') }}</div>
    <div v-else-if="error" class="bg-red-50 p-4 rounded-md text-red-700">{{ error }}</div>

    <div v-else-if="trade" class="space-y-6">
      <!-- Trade Info -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 flex justify-between items-center bg-gray-50 border-b border-gray-200">
          <div>
            <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.trades.transaction') }} #{{ trade.reference || trade.id }}</h3>
            <p class="mt-1 text-sm text-gray-500">{{ t('admin.trades.created_on') }} {{ formatDate(trade.createdAt) }}</p>
          </div>
          <span :class="statusBadgeClass(trade.status)" class="inline-flex rounded-full px-3 py-1 text-sm font-semibold">{{ enumLabel('trade_status', trade.status) }}</span>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <dl class="grid grid-cols-1 gap-x-4 gap-y-4 sm:grid-cols-3 text-sm">
            <div>
              <dt class="text-gray-500">{{ t('admin.trades.customer') }}</dt>
              <dd class="mt-1 font-medium text-gray-900">
                <template v-if="trade.user">{{ trade.user.firstName }} {{ trade.user.lastName }} ({{ trade.user.email }})</template>
                <template v-else>{{ trade.userId }}</template>
              </dd>
            </div>
            <div>
              <dt class="text-gray-500">{{ t('admin.trades.incoterms') }}</dt>
              <dd class="mt-1 font-medium text-gray-900">{{ cell(trade.terms || trade.incoterms) }}</dd>
            </div>
            <div>
              <dt class="text-gray-500">{{ t('admin.trades.total_amount') }}</dt>
              <dd class="mt-1 font-medium text-gray-900">{{ cur(trade.currency) }} {{ formatNumber(trade.totalAmount || 0) }}</dd>
            </div>
            <div v-if="trade.orderId">
              <dt class="text-gray-500">{{ t('admin.trades.order_id') }}</dt>
              <dd class="mt-1">
                <NuxtLink :to="localePath(`/admin/orders/${trade.orderId}`)" class="text-orange-600 hover:underline font-medium">{{ trade.orderId }}</NuxtLink>
              </dd>
            </div>
            <div v-if="trade.inquiryId">
              <dt class="text-gray-500">{{ t('admin.trades.inquiry_id') }}</dt>
              <dd class="mt-1">
                <NuxtLink :to="localePath(`/admin/inquiries/${trade.inquiryId}`)" class="text-orange-600 hover:underline font-medium">{{ trade.inquiryId }}</NuxtLink>
              </dd>
            </div>
          </dl>
        </div>
      </div>

      <!-- Status Update -->
      <div class="bg-white shadow sm:rounded-lg">
        <div class="px-4 py-4 border-b border-gray-200 bg-gray-50">
          <h3 class="text-base font-medium text-gray-900">{{ t('admin.trades.update_status') }}</h3>
        </div>
        <div class="px-4 py-4 flex items-end gap-4">
          <div>
            <label for="trade-status" class="block text-sm font-medium text-gray-700 mb-1">{{ t('admin.trades.status') }}</label>
            <select id="trade-status" name="status" v-model="statusInput" class="rounded-md border border-gray-300 px-3 py-2 text-sm">
              <option value="draft">{{ enumLabel('trade_status', 'draft') }}</option>
              <option value="pending">{{ enumLabel('trade_status', 'pending') }}</option>
              <option value="confirmed">{{ enumLabel('trade_status', 'confirmed') }}</option>
              <option value="paid">{{ enumLabel('trade_status', 'paid') }}</option>
              <option value="shipped">{{ enumLabel('trade_status', 'shipped') }}</option>
              <option value="completed">{{ enumLabel('trade_status', 'completed') }}</option>
              <option value="cancelled">{{ enumLabel('trade_status', 'cancelled') }}</option>
            </select>
          </div>
          <button @click="updateStatus" :disabled="updatingStatus"
            class="px-4 py-2 bg-orange-600 text-white text-sm font-medium rounded-md hover:bg-orange-700 disabled:opacity-50">
            {{ updatingStatus ? t('admin.trades.updating') : t('admin.trades.update') }}
          </button>
          <p v-if="statusMessage" class="text-sm" :class="statusError ? 'text-red-600' : 'text-green-600'">{{ statusMessage }}</p>
        </div>
      </div>

      <!-- Quick doc generation -->
      <div class="bg-white shadow sm:rounded-lg">
        <div class="px-4 py-4 border-b border-gray-200 bg-gray-50">
          <h3 class="text-base font-medium text-gray-900">{{ t('admin.trades.quick_generate') }}</h3>
          <p class="text-xs text-gray-500 mt-0.5">{{ t('admin.trades.quick_generate_desc') }}</p>
        </div>
        <div class="p-4 flex flex-wrap gap-2">
          <button
            v-for="btn in quickDocButtons"
            :key="btn.key"
            type="button"
            :disabled="generatingDoc === btn.key"
            class="px-3 py-1.5 text-xs font-medium rounded-md border border-orange-200 text-orange-700 bg-orange-50 hover:bg-orange-100 disabled:opacity-50"
            @click="quickGenerateDoc(btn)"
          >
            {{ btn.label }}
          </button>
        </div>
      </div>

      <!-- Settlements -->
      <div class="bg-white shadow sm:rounded-lg">
        <div class="px-4 py-4 border-b border-gray-200 bg-gray-50 flex items-center justify-between">
          <h3 class="text-base font-medium text-gray-900">{{ t('admin.trades.settlements') }}</h3>
          <button type="button" class="text-xs font-medium text-orange-600 hover:text-orange-800" @click="openSettlementModal()">{{ t('admin.trades.add_settlement') }}</button>
        </div>
        <div class="p-4">
          <div v-if="!settlements.length" class="text-sm text-gray-500">{{ t('admin.trades.no_settlements') }}</div>
          <table v-else class="min-w-full divide-y divide-gray-200 text-sm">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-3 py-2 text-left text-xs font-medium text-gray-500">{{ t('admin.trades.settlement_method') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium text-gray-500">{{ t('admin.trades.settlement_amount') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium text-gray-500">{{ t('admin.trades.settlement_due') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium text-gray-500">{{ t('admin.trades.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              <tr v-for="s in settlements" :key="s.id">
                <td class="px-3 py-3">{{ s.paymentMethod }}</td>
                <td class="px-3 py-3">{{ cur(s.currency) }} {{ formatNumber(s.amountDue || 0) }}</td>
                <td class="px-3 py-3">{{ s.dueDate ? formatDate(s.dueDate) : '-' }}</td>
                <td class="px-3 py-3">
                  <button type="button" class="text-red-600 text-xs hover:text-red-800" @click="deleteSettlement(s.id)">{{ t('admin.trades.delete') }}</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Compliance -->
      <div class="bg-white shadow sm:rounded-lg">
        <div class="px-4 py-4 border-b border-gray-200 bg-gray-50">
          <h3 class="text-base font-medium text-gray-900">{{ t('admin.trades.compliance') }}</h3>
        </div>
        <div class="p-4">
          <div v-if="!complianceItems.length" class="text-sm text-gray-500">{{ t('admin.trades.no_compliance') }}</div>
          <table v-else class="min-w-full divide-y divide-gray-200 text-sm">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-3 py-2 text-left text-xs font-medium text-gray-500">{{ t('admin.trades.compliance_item') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium text-gray-500">{{ t('admin.trades.doc_status') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium text-gray-500">{{ t('admin.trades.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              <tr v-for="c in complianceItems" :key="c.id">
                <td class="px-3 py-3">{{ c.requirement || c.name || c.id }}</td>
                <td class="px-3 py-3">
                  <select :value="c.status" class="rounded border border-gray-300 px-2 py-1 text-xs" @change="updateCompliance(c.id, ($event.target as HTMLSelectElement).value)">
                    <option value="pending">pending</option>
                    <option value="in_progress">in_progress</option>
                    <option value="completed">completed</option>
                    <option value="waived">waived</option>
                  </select>
                </td>
                <td class="px-3 py-3 text-xs text-gray-500">{{ c.notes || '' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- AI Trade Assistant -->
      <div class="bg-white shadow sm:rounded-lg">
        <div class="px-4 py-4 border-b border-gray-200 flex items-center justify-between">
          <div>
            <h3 class="text-base font-medium text-gray-900 inline-flex items-center">
              {{ t('admin.trades.ai_assistant') }}
              <AiHelpHint topic="trades_assistant" size="sm" />
            </h3>
            <p class="text-xs text-gray-500 mt-0.5">{{ t('admin.trades.ai_assistant_desc') }}</p>
          </div>
          <button @click="aiChatExpanded = !aiChatExpanded"
            class="text-gray-400 hover:text-gray-600 text-xs font-medium">
            {{ aiChatExpanded ? t('admin.trades.collapse') : t('admin.trades.expand') }}
          </button>
        </div>
        <div v-if="aiChatExpanded" class="flex flex-col h-[420px]">
          <div class="flex-1 p-3 overflow-y-auto bg-gray-50 text-sm" ref="adminChatContainer">
            <div v-for="(msg, i) in aiMessages" :key="i" class="mb-3">
              <div :class="['rounded-lg p-2.5 max-w-[85%] text-sm', msg.role === 'user' ? 'bg-orange-600 text-white ml-auto' : 'bg-white border text-gray-800 mr-auto']">
                <div v-if="msg.sender" class="text-xs font-semibold mb-0.5 opacity-70">{{ msg.sender }}</div>
                <div class="whitespace-pre-wrap">{{ msg.content }}</div>
                <div v-if="msg.tool_calls?.length" class="mt-1.5 border-t pt-1.5 border-gray-200">
                  <p class="text-xs font-bold text-gray-500 mb-0.5">{{ t('admin.trades.tools_called') }}:</p>
                  <div v-for="tc in msg.tool_calls" :key="tc.id" class="text-xs text-orange-800 bg-orange-50 rounded px-1 py-0.5 mb-0.5">{{ tc.function.name }}</div>
                </div>
                <div v-if="msg.documentType" class="text-xs text-green-700 mt-1 font-medium">📄 {{ msg.documentType }}</div>
              </div>
            </div>
            <div v-if="aiStreamingText" class="mb-3">
              <div class="rounded-lg p-2.5 max-w-[85%] bg-white border text-gray-800 mr-auto text-sm">
                <div class="text-xs font-semibold mb-0.5 opacity-70">{{ t('admin.trades.ai_typing') }}</div>
                <div class="whitespace-pre-wrap">{{ aiStreamingText }}</div>
              </div>
            </div>
          </div>
          <div class="p-3 border-t border-gray-200 flex gap-2 bg-white rounded-b-lg">
            <input v-model="aiInputQuery" @keyup.enter="aiSendMessage" type="text"
              :placeholder="t('admin.trades.ai_input_placeholder')"
              class="flex-1 shadow-sm focus:ring-orange-500 focus:border-orange-500 block w-full sm:text-sm border-gray-300 rounded-md"
              :disabled="aiStreaming" />
            <button @click="aiSendMessage" :disabled="aiStreaming || !aiInputQuery.trim()"
              class="inline-flex items-center px-3 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-orange-600 hover:bg-orange-700 disabled:opacity-50">
              {{ t('admin.trades.send') }}
            </button>
          </div>
        </div>
      </div>

      <!-- Trade Documents (generic) -->
      <div class="bg-white shadow sm:rounded-lg">
        <div class="px-4 py-4 border-b border-gray-200 bg-gray-50 flex items-center justify-between">
          <h3 class="text-base font-medium text-gray-900">{{ t('admin.trades.documents') }}</h3>
          <div class="inline-flex items-center gap-1">
            <button @click="aiGenerateDocument" :disabled="aiGenerating"
              class="px-3 py-1.5 bg-orange-600 text-white text-xs font-medium rounded-md hover:bg-orange-700 disabled:opacity-50 flex items-center gap-1">
              <Icon name="heroicons:sparkles" class="h-3.5 w-3.5" aria-hidden="true" />
              {{ aiGenerating ? t('admin.trades.ai_generating') : t('admin.trades.ai_generate_doc') }}
            </button>
            <AiHelpHint topic="trades_generate_doc" size="sm" />
          </div>
        </div>
        <div class="p-4">
          <div v-if="!documents.length" class="text-gray-500 text-sm">{{ t('admin.trades.no_documents') }}</div>
          <table v-else class="min-w-full divide-y divide-gray-200 text-sm">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.trades.doc_type') }}</th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.trades.doc_name') }}</th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.trades.doc_status') }}</th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.trades.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              <tr v-for="doc in documents" :key="doc.id">
                <td class="px-4 py-3 font-medium text-gray-900">{{ enumLabel('document_type', doc.type) }}</td>
                <td class="px-4 py-3 text-gray-500">{{ doc.docNumber }}</td>
                <td class="px-4 py-3">
                  <span :class="docStatusBadge(doc.status)" class="px-2 py-0.5 text-xs font-semibold rounded-full">{{ enumLabel('document_status', doc.status) }}</span>
                </td>
                <td class="px-4 py-3">
                  <button v-if="doc.status === 'DRAFT'" @click="confirmDoc(doc.id)"
                    class="text-green-600 hover:text-green-800 text-xs font-medium">{{ t('admin.trades.confirm') }}</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Rich Documents: PI / CI / B/L -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div v-for="rdoc in richDocCards" :key="rdoc.key" class="bg-white shadow sm:rounded-lg">
          <div class="px-4 py-3 border-b border-gray-200 bg-gray-50 flex items-center justify-between">
            <h4 class="text-sm font-semibold text-gray-900">{{ rdoc.label }}</h4>
            <span v-if="rdoc.data" :class="docStatusBadge(rdoc.data.status)" class="px-2 py-0.5 text-xs font-semibold rounded-full">{{ enumLabel('document_status', rdoc.data.status) }}</span>
          </div>
          <div class="p-4">
            <div v-if="!rdoc.data" class="text-gray-400 text-xs mb-3">{{ t('admin.trades.not_created') }}</div>
            <div v-else class="text-xs text-gray-600 space-y-1 mb-3">
              <div><span class="font-medium">{{ rdoc.numberLabel }}:</span> {{ rdoc.data[rdoc.numberField] }}</div>
              <div v-if="rdoc.data.totalAmount"><span class="font-medium">{{ dl('amount') }}:</span> {{ cur(rdoc.data.currency) }} {{ formatNumber(rdoc.data.totalAmount) }}</div>
              <div v-if="rdoc.data.incoterms"><span class="font-medium">{{ dl('incoterms') }}:</span> {{ rdoc.data.incoterms }}</div>
            </div>
            <div class="flex gap-3">
              <button @click="openRichDocModal(rdoc)" class="text-xs text-orange-600 hover:text-orange-800 font-medium">
                {{ rdoc.data ? t('admin.trades.edit') : t('admin.trades.create') }}
              </button>
              <button @click="aiGenerateRichDoc(rdoc)" :disabled="aiGenerating"
                class="text-xs text-purple-600 hover:text-purple-800 font-medium disabled:opacity-40">
                <Icon name="heroicons:sparkles" class="h-3 w-3 inline" aria-hidden="true" /> AI
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Settlement Modal -->
      <div v-if="showSettlementModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-40">
        <div class="bg-white rounded-xl shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ t('admin.trades.add_settlement') }}</h3>
          <form class="space-y-3" @submit.prevent="saveSettlement">
            <div>
              <label class="block text-xs font-medium text-gray-700">{{ t('admin.trades.settlement_method') }}</label>
              <input v-model="settlementForm.paymentMethod" required class="mt-1 w-full border border-gray-300 rounded-md px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700">{{ t('admin.trades.settlement_amount') }}</label>
              <input v-model.number="settlementForm.amountDue" type="number" step="0.01" required class="mt-1 w-full border border-gray-300 rounded-md px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700">{{ t('admin.trades.settlement_due') }}</label>
              <input v-model="settlementForm.dueDate" type="datetime-local" class="mt-1 w-full border border-gray-300 rounded-md px-3 py-2 text-sm" />
            </div>
            <div class="flex gap-3 pt-2">
              <button type="submit" :disabled="savingSettlement" class="px-4 py-2 bg-orange-600 text-white text-sm rounded-md disabled:opacity-50">{{ t('admin.trades.save') }}</button>
              <button type="button" class="px-4 py-2 bg-gray-100 text-sm rounded-md" @click="showSettlementModal = false">{{ t('admin.trades.cancel') }}</button>
            </div>
          </form>
        </div>
      </div>

      <!-- AI Generate Document Modal -->
      <div v-if="aiGenModal.docType" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-40">
        <div class="bg-white rounded-xl shadow-xl w-full max-w-md p-6">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-semibold text-gray-900">{{ t('admin.trades.ai_generate_title', { type: aiGenModal.docType }) }}</h3>
            <button @click="closeAiGenModal" class="text-gray-400 hover:text-gray-600" :aria-label="t('close')">
              <Icon name="heroicons:x-mark" class="h-5 w-5" aria-hidden="true" />
            </button>
          </div>
          <div class="space-y-3">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">{{ t('admin.trades.ai_context_label') }}</label>
              <textarea v-model="aiGenModal.context" rows="3" :placeholder="t('admin.trades.ai_context_placeholder')"
                class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-1 focus:ring-orange-500"></textarea>
            </div>
            <div class="flex gap-3 pt-2">
              <button @click="aiSubmitGenerate" :disabled="aiGenerating"
                class="px-4 py-2 bg-purple-600 text-white text-sm font-medium rounded-md hover:bg-purple-700 disabled:opacity-50 flex items-center gap-1">
                <Icon name="heroicons:sparkles" class="h-4 w-4" aria-hidden="true" />
                {{ aiGenerating ? t('admin.trades.ai_generating') : t('admin.trades.ai_generate') }}
              </button>
              <button @click="closeAiGenModal"
                class="px-4 py-2 bg-gray-100 text-gray-700 text-sm font-medium rounded-md hover:bg-gray-200">
                {{ t('admin.trades.cancel') }}
              </button>
            </div>
            <p v-if="aiGenError" class="text-sm text-red-600">{{ aiGenError }}</p>
            <p v-if="aiGenResult" class="text-sm text-green-700">{{ aiGenResult }}</p>
          </div>
        </div>
      </div>

      <!-- Rich Doc Edit Modal -->
      <div v-if="editingRichDoc" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-40">
        <div class="bg-white rounded-xl shadow-xl w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-semibold text-gray-900">{{ editingRichDoc.label }}</h3>
            <button @click="editingRichDoc = null" class="text-gray-400 hover:text-gray-600" :aria-label="t('close')">
              <Icon name="heroicons:x-mark" class="h-5 w-5" aria-hidden="true" />
            </button>
          </div>
          <form @submit.prevent="saveRichDoc" class="space-y-3">
            <template v-for="field in editingRichDoc.fields" :key="field.key">
              <div>
                <label class="block text-xs font-medium text-gray-700 mb-0.5">{{ field.label }}</label>
                <input v-model="richDocForm[field.key]" :type="field.type || 'text'"
                  class="w-full border border-gray-300 rounded-md px-3 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-orange-500" />
              </div>
            </template>
            <div class="flex gap-3 pt-2">
              <button type="submit" :disabled="savingRichDoc"
                class="px-4 py-2 bg-orange-600 text-white text-sm font-medium rounded-md hover:bg-orange-700 disabled:opacity-50">
                {{ savingRichDoc ? t('admin.trades.saving') : t('admin.trades.save') }}
              </button>
              <button type="button" @click="editingRichDoc = null"
                class="px-4 py-2 bg-gray-100 text-gray-700 text-sm font-medium rounded-md hover:bg-gray-200">
                {{ t('admin.trades.cancel') }}
              </button>
            </div>
            <p v-if="richDocError" class="text-sm text-red-600">{{ richDocError }}</p>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive, computed, nextTick } from 'vue'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const route = useRoute()
const api = useApi()
const { t } = useI18n()
const localePath = useLocalePath()
const { currencyOrDefault: cur, cell, enumLabel, formatNumber, formatDate } = useDisplay()

const trade = ref<any>(null)
const documents = ref<any[]>([])
const pending = ref(true)
const error = ref('')
const statusInput = ref('')
const updatingStatus = ref(false)
const statusMessage = ref('')
const statusError = ref(false)

// Rich documents
const piData = ref<any>(null)
const ciData = ref<any>(null)
const blData = ref<any>(null)
const scData = ref<any>(null)
const plData = ref<any>(null)
const cooData = ref<any>(null)
const hcData = ref<any>(null)
const settlements = ref<any[]>([])
const complianceItems = ref<any[]>([])
const generatingDoc = ref('')
const showSettlementModal = ref(false)
const savingSettlement = ref(false)
const settlementForm = reactive({ paymentMethod: 'T/T', amountDue: 0, dueDate: '', currency: 'USD' })
const editingRichDoc = ref<any>(null)
const richDocForm = reactive<Record<string, any>>({})
const savingRichDoc = ref(false)
const richDocError = ref('')

const dl = (k: string) => t(`admin.docLabels.${k}`)

// ── AI Chat State ──
const aiChatExpanded = ref(false)
const aiMessages = ref<{role: string, content: string, sender?: string, tool_calls?: any[], documentType?: string}[]>([])
const aiInputQuery = ref('')
const aiStreaming = ref(false)
const aiStreamingText = ref('')
const adminChatContainer = ref<HTMLElement | null>(null)
const { token } = useAuth()
const config = useRuntimeConfig()
const baseURL = config.public.apiBase || '/api/v1'

// ── AI Doc Generation State ──
const aiGenerating = ref(false)
const aiGenModal = reactive({ docType: '', context: '' })
const aiGenError = ref('')
const aiGenResult = ref('')

const docTypeToKey: Record<string, string> = {
  'PROFORMA_INVOICE': 'proforma-invoice',
  'COMMERCIAL_INVOICE': 'commercial-invoice',
  'SALES_CONTRACT': 'sales-contract',
  'PACKING_LIST': 'packing-list',
  'ORIGIN_CERTIFICATE': 'certificate-of-origin',
  'HEALTH_CERTIFICATE': 'health-certificate',
  'BILL_OF_LADING': 'bill-of-lading',
}

const richDocKeyToType = (key: string) => {
  const map: Record<string, string> = {
    pi: 'PROFORMA_INVOICE', ci: 'COMMERCIAL_INVOICE', bl: 'BILL_OF_LADING',
    sc: 'SALES_CONTRACT', pl: 'PACKING_LIST', coo: 'ORIGIN_CERTIFICATE', hc: 'HEALTH_CERTIFICATE',
  }
  return map[key] || ''
}

/** 快捷单证生成按钮配置 */
const quickDocButtons = computed(() => [
  { key: 'pi', label: 'PI', endpoint: 'proforma-invoice', aiType: 'PROFORMA_INVOICE' },
  { key: 'ci', label: 'CI', endpoint: 'commercial-invoice', aiType: 'COMMERCIAL_INVOICE' },
  { key: 'sc', label: 'SC', endpoint: 'sales-contract', aiType: 'SALES_CONTRACT' },
  { key: 'pl', label: 'PL', endpoint: 'packing-list', aiType: 'PACKING_LIST' },
  { key: 'coo', label: 'COO', endpoint: 'certificate-of-origin', aiType: 'ORIGIN_CERTIFICATE' },
  { key: 'hc', label: 'HC', endpoint: 'health-certificate', aiType: 'HEALTH_CERTIFICATE' },
  { key: 'bl', label: 'BL', endpoint: 'bill-of-lading', aiType: 'BILL_OF_LADING' },
])

const quickGenerateDoc = async (btn: { key: string; endpoint: string; aiType: string }) => {
  generatingDoc.value = btn.key
  try {
    await api.post(`/admin/trades/${route.params.id}/${btn.endpoint}`, { status: 'DRAFT' })
    await fetchRichDocs()
  } catch {
    try {
      await api.post(`/admin/trades/${route.params.id}/documents/ai-generate`, { docType: btn.aiType })
      await fetchRichDocs()
      await fetchDocuments()
    } catch (err: any) {
      alert(err?.message || t('errors.api.document_save_failed'))
    }
  } finally {
    generatingDoc.value = ''
  }
}

const fetchSettlements = async () => {
  try {
    settlements.value = await api.get<any[]>(`/admin/trades/${route.params.id}/settlements`) || []
  } catch { settlements.value = [] }
}

const fetchCompliance = async () => {
  try {
    complianceItems.value = await api.get<any[]>(`/admin/trades/${route.params.id}/compliance`) || []
  } catch { complianceItems.value = [] }
}

const openSettlementModal = () => {
  settlementForm.paymentMethod = 'T/T'
  settlementForm.amountDue = trade.value?.totalAmount || 0
  settlementForm.currency = trade.value?.currency || 'USD'
  settlementForm.dueDate = ''
  showSettlementModal.value = true
}

const saveSettlement = async () => {
  savingSettlement.value = true
  try {
    const payload: Record<string, unknown> = {
      paymentMethod: settlementForm.paymentMethod,
      amountDue: settlementForm.amountDue,
      currency: settlementForm.currency,
    }
    if (settlementForm.dueDate) payload.dueDate = new Date(settlementForm.dueDate).toISOString()
    await api.post(`/admin/trades/${route.params.id}/settlements`, payload)
    showSettlementModal.value = false
    await fetchSettlements()
  } catch (err: any) {
    alert(err?.message || t('errors.api.save_failed'))
  } finally {
    savingSettlement.value = false
  }
}

const deleteSettlement = async (id: number) => {
  if (!confirm(t('admin.confirm_delete'))) return
  try {
    await api.delete(`/admin/trades/${route.params.id}/settlements/${id}`)
    await fetchSettlements()
  } catch (err: any) {
    alert(err?.message || t('errors.api.delete_failed'))
  }
}

const updateCompliance = async (compId: number, status: string) => {
  try {
    await api.put(`/admin/trades/${route.params.id}/compliance/${compId}`, { status })
    await fetchCompliance()
  } catch (err: any) {
    alert(err?.message || t('errors.api.status_failed'))
  }
}

const aiScrollToBottom = () => {
  nextTick(() => { if (adminChatContainer.value) adminChatContainer.value.scrollTop = adminChatContainer.value.scrollHeight })
}

const aiSendMessage = () => {
  if (!aiInputQuery.value.trim() || aiStreaming.value) return
  const query = aiInputQuery.value
  aiMessages.value.push({ role: 'user', content: query })
  aiInputQuery.value = ''
  aiStreaming.value = true
  aiStreamingText.value = ''
  aiScrollToBottom()
  const url = `${baseURL}/admin/trades/${route.params.id}/ai-chat?query=${encodeURIComponent(query)}`
  aiStreamFetch(url)
}

const aiStreamFetch = async (url: string) => {
  try {
    const resp = await fetch(url, { credentials: 'include' })
    if (!resp.ok) {
      aiMessages.value.push({ role: 'ai', content: `*Connection error*`, sender: 'System' })
      aiStreaming.value = false
      return
    }
    const reader = resp.body?.getReader()
    if (!reader) {
      // Non-streaming fallback
      const data = await resp.json()
      aiMessages.value.push({ role: 'ai', content: data.reply || '', sender: 'AI' })
      aiStreaming.value = false
      return
    }
    const decoder = new TextDecoder()
    let partial = ''
    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      partial += decoder.decode(value, { stream: true })
      const parts = partial.split('\n\n')
      partial = parts.pop() || ''
      for (const part of parts) {
        if (part.startsWith('data: ')) {
          try { aiHandleSSE(JSON.parse(part.replace('data: ', ''))) } catch(e) {}
        }
      }
    }
  } catch(e) {
    aiMessages.value.push({ role: 'ai', content: `*${t('admin.trades.ai_stream_interrupted')}*`, sender: 'System' })
  } finally {
    aiStreaming.value = false
    if (aiStreamingText.value) {
      aiMessages.value.push({ role: 'ai', content: aiStreamingText.value, sender: 'AI' })
      aiStreamingText.value = ''
    }
    aiScrollToBottom()
    fetchRichDocs()
    fetchDocuments()
  }
}

const aiHandleSSE = (data: any) => {
  if (data.type === 'message' || data.type === 'tool_result') {
    if (aiStreamingText.value) {
      aiMessages.value.push({ role: 'ai', content: aiStreamingText.value, sender: data.agent_name || 'AI' })
      aiStreamingText.value = ''
    }
    aiMessages.value.push({
      role: 'ai', content: data.content || '',
      sender: data.agent_name || 'AI',
      tool_calls: data.tool_calls,
      documentType: data.document_type,
    })
  } else if (data.type === 'stream_chunk') {
    aiStreamingText.value += data.content
  } else if (data.type === 'error') {
    aiMessages.value.push({ role: 'ai', content: `**${t('admin.trades.ai_error_prefix')}:** ${data.error}`, sender: 'System' })
  }
  aiScrollToBottom()
}

// ── AI Document Generation ──
const closeAiGenModal = () => {
  aiGenModal.docType = ''
  aiGenModal.context = ''
  aiGenError.value = ''
  aiGenResult.value = ''
}

const aiGenerateDocument = () => {
  closeAiGenModal()
  aiGenModal.docType = 'PROFORMA_INVOICE'
}

const aiGenerateRichDoc = (rdoc: any) => {
  closeAiGenModal()
  aiGenModal.docType = richDocKeyToType(rdoc.key)
}

const aiSubmitGenerate = async () => {
  if (!aiGenModal.docType || aiGenerating.value) return
  aiGenerating.value = true
  aiGenError.value = ''
  aiGenResult.value = ''
  try {
    const result = await api.post<any>(`/admin/trades/${route.params.id}/documents/ai-generate`, {
      docType: aiGenModal.docType,
      context: aiGenModal.context,
      prompt: t('admin.trades.ai_generate_prompt', { type: aiGenModal.docType }),
    })
    aiGenResult.value = result?.message || t('admin.trades.ai_doc_generated')
    closeAiGenModal()
    await fetchRichDocs()
    await fetchDocuments()
  } catch (err: any) {
    aiGenError.value = err?.message || t('admin.trades.ai_generation_failed')
  } finally {
    aiGenerating.value = false
  }
}
const richDocCards = computed(() => [
  {
    key: 'pi', label: dl('proforma_invoice'), data: piData.value,
    numberLabel: dl('pi_no'), numberField: 'piNumber',
    endpoint: 'proforma-invoice',
    fields: [
      { key: 'buyerName', label: dl('buyer_name') },
      { key: 'sellerName', label: dl('seller_name') },
      { key: 'incoterms', label: dl('incoterms') },
      { key: 'termsOfPayment', label: dl('payment_terms') },
      { key: 'totalAmount', label: dl('total_amount'), type: 'number' },
      { key: 'currency', label: dl('currency') },
      { key: 'bankDetails', label: dl('bank_details') },
      { key: 'notes', label: dl('notes') },
      { key: 'status', label: dl('status') },
    ]
  },
  {
    key: 'ci', label: dl('commercial_invoice'), data: ciData.value,
    numberLabel: dl('ci_no'), numberField: 'ciNumber',
    endpoint: 'commercial-invoice',
    fields: [
      { key: 'buyerName', label: dl('buyer_name') },
      { key: 'sellerName', label: dl('seller_name') },
      { key: 'incoterms', label: dl('incoterms') },
      { key: 'termsOfPayment', label: dl('payment_terms') },
      { key: 'totalAmount', label: dl('total_amount'), type: 'number' },
      { key: 'currency', label: dl('currency') },
      { key: 'piNumber', label: dl('pi_reference') },
      { key: 'bankDetails', label: dl('bank_details') },
      { key: 'status', label: dl('status') },
    ]
  },
  {
    key: 'bl', label: dl('bill_of_lading'), data: blData.value,
    numberLabel: dl('bl_no'), numberField: 'blNumber',
    endpoint: 'bill-of-lading',
    fields: [
      { key: 'shipper', label: dl('shipper') },
      { key: 'consignee', label: dl('consignee') },
      { key: 'notifyParty', label: dl('notify_party') },
      { key: 'carrierName', label: dl('carrier') },
      { key: 'vesselVoyage', label: dl('vessel_voyage') },
      { key: 'portOfLoading', label: dl('port_of_loading') },
      { key: 'portOfDischarge', label: dl('port_of_discharge') },
      { key: 'freightTerms', label: dl('freight_terms') },
      { key: 'goodsDescription', label: dl('goods_description') },
      { key: 'grossWeight', label: dl('gross_weight'), type: 'number' },
      { key: 'measurement', label: dl('measurement'), type: 'number' },
      { key: 'numberOfPackages', label: dl('no_of_packages'), type: 'number' },
      { key: 'status', label: dl('status') },
    ]
  },
  {
    key: 'sc', label: dl('sales_contract'), data: scData.value,
    numberLabel: dl('contract_no'), numberField: 'contractNo',
    endpoint: 'sales-contract',
    fields: [
      { key: 'contractNo', label: dl('contract_no') },
      { key: 'buyerName', label: dl('buyer_name') },
      { key: 'sellerName', label: dl('seller_name') },
      { key: 'incoterms', label: dl('incoterms') },
      { key: 'paymentTerms', label: dl('payment_terms') },
      { key: 'totalAmount', label: dl('total_amount'), type: 'number' },
      { key: 'currency', label: dl('currency') },
      { key: 'deliveryDate', label: dl('delivery_date'), type: 'date' },
      { key: 'specialTerms', label: dl('special_terms') },
      { key: 'status', label: dl('status') },
    ]
  },
  {
    key: 'pl', label: dl('packing_list'), data: plData.value,
    numberLabel: dl('pl_no'), numberField: 'plNumber',
    endpoint: 'packing-list',
    fields: [
      { key: 'plNumber', label: dl('pl_no') },
      { key: 'exporterName', label: dl('exporter_name') },
      { key: 'importerName', label: dl('importer_name') },
      { key: 'totalPackages', label: dl('total_packages'), type: 'number' },
      { key: 'totalGrossWeight', label: dl('total_gross_weight'), type: 'number' },
      { key: 'totalNetWeight', label: dl('total_net_weight'), type: 'number' },
      { key: 'totalVolume', label: dl('total_volume'), type: 'number' },
      { key: 'shippingMark', label: dl('shipping_mark') },
      { key: 'notes', label: dl('notes') },
    ]
  },
  {
    key: 'coo', label: dl('cert_of_origin'), data: cooData.value,
    numberLabel: dl('cert_no'), numberField: 'certificateNo',
    endpoint: 'certificate-of-origin',
    fields: [
      { key: 'certificateNo', label: dl('cert_no') },
      { key: 'exporterName', label: dl('exporter_name') },
      { key: 'importerName', label: dl('importer_name') },
      { key: 'countryOfOrigin', label: dl('country_of_origin') },
      { key: 'destinationCountry', label: dl('destination_country') },
      { key: 'hsCode', label: dl('hs_code') },
      { key: 'ftaType', label: dl('fta_type') },
      { key: 'goodsDescription', label: dl('goods_description') },
      { key: 'grossWeight', label: dl('gross_weight'), type: 'number' },
      { key: 'issuingAuthority', label: dl('issuing_authority') },
    ]
  },
  {
    key: 'hc', label: dl('health_cert'), data: hcData.value,
    numberLabel: dl('cert_no'), numberField: 'certificateNo',
    endpoint: 'health-certificate',
    fields: [
      { key: 'certificateNo', label: dl('cert_no') },
      { key: 'exporterName', label: dl('exporter_name') },
      { key: 'importerName', label: dl('importer_name') },
      { key: 'productName', label: dl('product_name') },
      { key: 'batchNumber', label: dl('batch_number') },
      { key: 'productionDate', label: dl('production_date'), type: 'date' },
      { key: 'expiryDate', label: dl('expiry_date'), type: 'date' },
      { key: 'issuingAuthority', label: dl('issuing_authority') },
      { key: 'inspectionResult', label: dl('inspection_result') },
      { key: 'notes', label: dl('notes') },
    ]
  },
])

const fetchTrade = async () => {
  pending.value = true
  error.value = ''
  try {
    trade.value = await api.get<any>(`/admin/trades/${route.params.id}`)
    statusInput.value = trade.value.status || 'DRAFT'
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const fetchDocuments = async () => {
  try {
    documents.value = await api.get<any[]>(`/admin/trades/${route.params.id}/documents`) || []
  } catch { documents.value = [] }
}

const fetchRichDocs = async () => {
  const tryFetch = async (path: string) => {
    try { return await api.get<any>(`/admin/trades/${route.params.id}/${path}`) } catch { return null }
  }
  const [pi, ci, bl, sc, pl, coo, hc] = await Promise.all([
    tryFetch('proforma-invoice'),
    tryFetch('commercial-invoice'),
    tryFetch('bill-of-lading'),
    tryFetch('sales-contract'),
    tryFetch('packing-list'),
    tryFetch('certificate-of-origin'),
    tryFetch('health-certificate'),
  ])
  piData.value = pi; ciData.value = ci; blData.value = bl
  scData.value = sc; plData.value = pl; cooData.value = coo; hcData.value = hc
}

const updateStatus = async () => {
  updatingStatus.value = true
  statusMessage.value = ''
  statusError.value = false
  try {
    await api.put(`/admin/trades/${route.params.id}/status`, { status: statusInput.value })
    trade.value.status = statusInput.value
    statusMessage.value = t('admin.trades.status_updated')
    setTimeout(() => { statusMessage.value = '' }, 3000)
  } catch (err: any) {
    statusError.value = true
    statusMessage.value = err?.message || t('errors.api.status_failed')
  } finally {
    updatingStatus.value = false
  }
}

const confirmDoc = async (docId: string) => {
  try {
    await api.put(`/admin/trades/${route.params.id}/documents/${docId}`, { status: 'CONFIRMED' })
    await fetchDocuments()
  } catch (err: any) {
    alert(err?.message || t('errors.api.document_confirm_failed'))
  }
}

const openRichDocModal = (rdoc: any) => {
  editingRichDoc.value = rdoc
  richDocError.value = ''
  // Pre-fill form with existing data
  Object.keys(richDocForm).forEach(k => delete richDocForm[k])
  if (rdoc.data) {
    rdoc.fields.forEach((f: any) => { richDocForm[f.key] = rdoc.data[f.key] ?? '' })
  } else {
    rdoc.fields.forEach((f: any) => { richDocForm[f.key] = '' })
  }
}

const saveRichDoc = async () => {
  if (!editingRichDoc.value) return
  savingRichDoc.value = true
  richDocError.value = ''
  const { endpoint, data } = editingRichDoc.value
  try {
    const method = data ? 'put' : 'post'
    const result = await api[method]<any>(`/admin/trades/${route.params.id}/${endpoint}`, { ...richDocForm })
    // Update local data
    if (endpoint === 'proforma-invoice') piData.value = result
    else if (endpoint === 'commercial-invoice') ciData.value = result
    else if (endpoint === 'bill-of-lading') blData.value = result
    else if (endpoint === 'sales-contract') scData.value = result
    else if (endpoint === 'packing-list') plData.value = result
    else if (endpoint === 'certificate-of-origin') cooData.value = result
    else if (endpoint === 'health-certificate') hcData.value = result
    editingRichDoc.value = null
  } catch (err: any) {
    richDocError.value = err?.message || t('errors.api.document_save_failed')
  } finally {
    savingRichDoc.value = false
  }
}


const statusBadgeClass = (status: string) => {
  const map: Record<string, string> = {
    DRAFT: 'bg-gray-100 text-gray-800', PENDING: 'bg-yellow-100 text-yellow-800',
    CONFIRMED: 'bg-orange-100 text-orange-800', PAID: 'bg-amber-100 text-amber-800',
    SHIPPED: 'bg-amber-100 text-amber-800', COMPLETED: 'bg-green-100 text-green-800',
    CANCELLED: 'bg-red-100 text-red-800',
  }
  return map[status] || 'bg-gray-100 text-gray-800'
}

const docStatusBadge = (status: string) => {
  if (status === 'CONFIRMED' || status === 'ISSUED' || status === 'PAID') return 'bg-green-100 text-green-800'
  if (status === 'DRAFT') return 'bg-gray-100 text-gray-800'
  return 'bg-yellow-100 text-yellow-800'
}

onMounted(() => {
  fetchTrade()
  fetchDocuments()
  fetchRichDocs()
  fetchSettlements()
  fetchCompliance()
})
</script>
