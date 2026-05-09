<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <NuxtLink :to="localePath('/customer/orders')" class="inline-flex items-center text-sm font-medium text-orange-600 hover:text-orange-500">
        <Icon name="heroicons:arrow-left" class="mr-1 h-4 w-4" />
        {{ t('customer.common.back_to_orders') }}
      </NuxtLink>
    </div>

    <div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
      <section class="rounded-lg border border-blue-200 bg-white shadow-sm">
        <div class="border-b border-orange-100 bg-orange-50 px-5 py-4">
          <h2 class="text-lg font-semibold text-blue-900">{{ t('customer.orders_new.ai_title') }}</h2>
          <p class="mt-1 text-sm text-orange-800">{{ t('customer.orders_new.ai_desc') }}</p>
        </div>
        <div class="space-y-4 px-5 py-5">
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.order_intention') }}</label>
            <textarea v-model="aiForm.prompt" rows="4" :placeholder="t('customer.orders_new.order_intention_placeholder')" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
          </div>
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.target_country') }}</label>
              <input v-model="aiForm.targetCountry" type="text" :placeholder="t('customer.orders_new.target_country_placeholder')" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.requested_quantity') }}</label>
              <input v-model.number="aiForm.quantity" type="number" min="1" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.budget') }}</label>
              <input v-model.number="aiForm.budget" type="number" min="0" step="0.01" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.currency') }}</label>
              <input v-model="aiForm.currency" type="text" :placeholder="t('common.defaults.currency')" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm uppercase focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.estimated_tax') }}</label>
              <input v-model.number="aiForm.taxAmount" type="number" min="0" step="0.01" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.estimated_shipping') }}</label>
              <input v-model.number="aiForm.shippingAmount" type="number" min="0" step="0.01" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.additional_requirements') }}</label>
            <textarea v-model="aiForm.additionalRequirements" rows="3" :placeholder="t('customer.orders_new.additional_requirements_placeholder')" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
          </div>
          <button :disabled="aiSubmitting" @click="submitAIDraft" class="inline-flex items-center rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-60">
            <Icon name="heroicons:sparkles" class="mr-1.5 h-4 w-4" />
            {{ aiSubmitting ? t('customer.orders_new.generating') : t('customer.orders_new.generate_ai') }}
          </button>
          <p v-if="aiError" class="text-sm text-red-600">{{ aiError }}</p>
          <ul v-if="aiViolationDetails.length > 0" class="list-disc pl-5 text-sm text-red-700 space-y-1">
            <li v-for="(item, idx) in aiViolationDetails" :key="`ai-violation-${idx}`">{{ item }}</li>
          </ul>
          <p v-if="aiSuccess" class="text-sm text-green-700">{{ aiSuccess }}</p>
        </div>
      </section>

      <section class="rounded-lg border border-gray-200 bg-white shadow-sm">
        <div class="border-b border-gray-200 px-5 py-4">
          <h2 class="text-lg font-semibold text-gray-900">{{ t('customer.orders_new.shipping_title') }}</h2>
          <p class="mt-1 text-sm text-gray-600">{{ t('customer.orders_new.shipping_desc') }}</p>
        </div>
        <div class="grid grid-cols-1 gap-4 px-5 py-5 sm:grid-cols-2">
          <div class="sm:col-span-2">
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.street') }}</label>
            <input v-model="shippingAddress.street" type="text" :class="[fieldErrors.street ? 'border-red-500 ring-1 ring-red-500' : 'border-gray-300', 'mt-1 w-full rounded-md px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500']" />
            <p v-if="fieldErrors.street" class="mt-1 text-xs text-red-600">{{ fieldErrors.street }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.city') }}</label>
            <input v-model="shippingAddress.city" type="text" :class="[fieldErrors.city ? 'border-red-500 ring-1 ring-red-500' : 'border-gray-300', 'mt-1 w-full rounded-md px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500']" />
            <p v-if="fieldErrors.city" class="mt-1 text-xs text-red-600">{{ fieldErrors.city }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.state') }}</label>
            <input v-model="shippingAddress.state" type="text" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.zip_code') }}</label>
            <input v-model="shippingAddress.zipCode" type="text" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.country') }}</label>
            <input v-model="shippingAddress.country" type="text" :class="[fieldErrors.country ? 'border-red-500 ring-1 ring-red-500' : 'border-gray-300', 'mt-1 w-full rounded-md px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500']" />
            <p v-if="fieldErrors.country" class="mt-1 text-xs text-red-600">{{ fieldErrors.country }}</p>
          </div>
        </div>
      </section>
    </div>

    <section v-if="draftOrder" class="rounded-lg border border-orange-200 bg-white shadow-sm">
      <div class="border-b border-orange-100 bg-orange-50 px-5 py-4">
        <h2 class="text-lg font-semibold text-orange-900">{{ t('customer.orders_new.draft_ready') }}</h2>
        <p class="mt-1 text-sm text-orange-800">{{ t('customer.orders_new.draft_review') }}</p>
      </div>
      <div class="space-y-5 px-5 py-5">
        <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
          <div class="rounded-md border border-gray-200 bg-gray-50 p-3">
            <p class="text-xs uppercase tracking-wide text-gray-500">{{ t('customer.orders_new.order_number') }}</p>
            <p class="mt-1 text-sm font-semibold text-gray-900">{{ draftOrder.orderNumber }}</p>
          </div>
          <div class="rounded-md border border-gray-200 bg-gray-50 p-3">
            <p class="text-xs uppercase tracking-wide text-gray-500">{{ t('customer.orders_new.status') }}</p>
            <p class="mt-1 text-sm font-semibold text-orange-700">{{ enumLabel('order_status', draftOrder.status) }}</p>
          </div>
          <div class="rounded-md border border-gray-200 bg-gray-50 p-3">
            <p class="text-xs uppercase tracking-wide text-gray-500">{{ t('customer.orders_new.estimated_total') }}</p>
            <p class="mt-1 text-sm font-semibold text-gray-900">{{ cur(draftOrder.currency) }} {{ formatNumber(draftOrder.totalAmount || 0) }}</p>
          </div>
        </div>

        <div v-if="selectedProducts.length > 0" class="overflow-x-auto rounded-md border border-gray-200">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-4 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('customer.orders_new.col_product') }}</th>
                <th class="px-4 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('customer.orders_new.col_quantity') }}</th>
                <th class="px-4 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('customer.orders_new.col_unit_price') }}</th>
                <th class="px-4 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('customer.orders_new.col_reason') }}</th>
                <th class="px-4 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('customer.orders_new.col_actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 bg-white">
              <tr v-for="(item, idx) in selectedProducts" :key="item.id">
                <td class="px-4 py-2 text-sm text-gray-700">
                  <div class="font-medium text-gray-900">{{ item.name }}</div>
                  <div class="text-xs text-gray-500">ID: {{ item.id }}</div>
                </td>
                <td class="px-4 py-2 text-sm text-gray-700">
                  <input v-model.number="selectedProducts[idx].quantity" type="number" min="1"
                    class="w-20 rounded border border-gray-300 px-2 py-1 text-sm focus:border-orange-500 focus:outline-none" />
                </td>
                <td class="px-4 py-2 text-sm text-gray-700">
                  <input v-model.number="selectedProducts[idx].unitPrice" type="number" min="0" step="0.01"
                    class="w-24 rounded border border-gray-300 px-2 py-1 text-sm focus:border-orange-500 focus:outline-none" />
                </td>
                <td class="px-4 py-2 text-sm text-gray-600">{{ cell(item.reason) }}</td>
                <td class="px-4 py-2 text-sm">
                  <button @click="selectedProducts.splice(idx, 1)" class="text-red-500 hover:text-red-700 text-xs font-medium">
                    {{ t('customer.orders_new.remove') }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
          <div v-if="compliance.paymentPolicy.requiresFullPrepayment" class="rounded-md border border-red-200 bg-red-50 p-4">
            <h3 class="text-sm font-semibold text-red-900">{{ t('customer.orders_new.payment_policy') }}</h3>
            <p class="mt-2 text-sm text-red-800">{{ compliance.paymentPolicy.note || '' }}</p>
            <ul v-if="compliance.paymentPolicy.allowedTerms.length > 0" class="mt-2 list-disc space-y-1 pl-5 text-sm text-red-800">
              <li v-for="(item, idx) in compliance.paymentPolicy.allowedTerms" :key="`pay-term-${idx}`">{{ item }}</li>
            </ul>
          </div>
          <div class="rounded-md border border-green-200 bg-green-50 p-4">
            <h3 class="text-sm font-semibold text-green-900">{{ t('customer.orders_new.compliance_checklist') }}</h3>
            <ul class="mt-2 list-disc space-y-1 pl-5 text-sm text-green-800">
              <li v-for="(item, idx) in compliance.checklist" :key="`check-${idx}`">{{ item }}</li>
            </ul>
          </div>
          <div class="rounded-md border border-amber-200 bg-amber-50 p-4">
            <h3 class="text-sm font-semibold text-amber-900">{{ t('customer.orders_new.warnings') }}</h3>
            <ul class="mt-2 list-disc space-y-1 pl-5 text-sm text-amber-800">
              <li v-for="(item, idx) in compliance.warnings" :key="`warn-${idx}`">{{ item }}</li>
            </ul>
            <p v-if="compliance.warnings.length === 0" class="mt-2 text-sm text-amber-800">{{ t('customer.orders_new.no_warnings') }}</p>
          </div>
          <div v-if="!compliance.hasOfficialEvidence" class="rounded-md border border-red-300 bg-red-50 p-4">
            <h3 class="text-sm font-semibold text-red-900">{{ t('customer.orders_new.manual_compliance') }}</h3>
            <p class="mt-2 text-sm text-red-800">{{ t('customer.orders_new.manual_compliance_desc') }}</p>
            <label class="mt-3 inline-flex items-start gap-2 text-sm text-red-900">
              <input v-model="confirmComplianceAck" type="checkbox" class="mt-0.5 rounded border-red-300 text-red-600 focus:ring-red-500" />
              <span>{{ t('customer.common.compliance_ack_label') }}</span>
            </label>
          </div>
        </div>

        <div v-if="compliance.references.length > 0" class="rounded-md border border-gray-200 bg-gray-50 p-4">
          <h3 class="text-sm font-semibold text-gray-900">{{ t('customer.orders_new.compliance_references') }}</h3>
          <p v-if="compliance.retrievalNote" class="mt-2 text-xs text-gray-600">{{ compliance.retrievalNote }}</p>
          <ul class="mt-2 space-y-2 text-sm text-gray-700">
            <li v-for="(ref, idx) in compliance.references" :key="`ref-${idx}`">
              <span class="font-medium">{{ ref.source }} / {{ ref.section }}</span>
              <span v-if="ref.official" class="ml-2 inline-flex rounded-full bg-emerald-100 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-emerald-700">{{ t('customer.orders_new.official') }}</span>
              <span>: {{ ref.snippet }}</span>
              <a v-if="ref.url" :href="ref.url" target="_blank" rel="noopener noreferrer" class="ml-2 text-xs font-medium text-orange-600 hover:text-orange-500">{{ t('customer.orders_new.source') }}</a>
            </li>
          </ul>
        </div>

        <div v-if="aiSummary" class="rounded-md border border-gray-200 bg-gray-50 p-4">
          <h3 class="text-sm font-semibold text-gray-900">{{ t('customer.orders_new.ai_summary') }}</h3>
          <p class="mt-2 text-sm text-gray-700 whitespace-pre-line">{{ aiSummary }}</p>
        </div>

        <div v-if="missingInformation.length > 0" class="rounded-md border border-gray-200 bg-gray-50 p-4">
          <h3 class="text-sm font-semibold text-gray-900">{{ t('customer.orders_new.missing_info') }}</h3>
          <ul class="mt-2 list-disc space-y-1 pl-5 text-sm text-gray-700">
            <li v-for="(item, idx) in missingInformation" :key="`miss-${idx}`">{{ item }}</li>
          </ul>
        </div>

        <div class="flex flex-wrap items-center gap-3">
          <button :disabled="confirmingDraft || (!compliance.hasOfficialEvidence && !confirmComplianceAck)" @click="confirmDraftOrder" class="inline-flex items-center rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-60">
            <Icon name="heroicons:check-circle" class="mr-1.5 h-4 w-4" />
            {{ confirmingDraft ? t('customer.orders_new.confirming_draft') : t('customer.orders_new.confirm_draft') }}
          </button>
          <NuxtLink :to="localePath(`/customer/orders/${draftOrder.id}`)" class="text-sm font-medium text-orange-600 hover:text-orange-500">{{ t('customer.orders_new.view_draft') }}</NuxtLink>
        </div>
      </div>
    </section>

    <section class="rounded-lg border border-gray-200 bg-white shadow-sm">
      <div class="border-b border-gray-200 px-5 py-4">
        <h2 class="text-lg font-semibold text-gray-900">{{ t('customer.orders_new.manual_title') }}</h2>
        <p class="mt-1 text-sm text-gray-600">{{ t('customer.orders_new.manual_desc') }}</p>
      </div>
      <div class="space-y-4 px-5 py-5">
        <div class="grid grid-cols-1 gap-4 md:grid-cols-4">
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.currency') }}</label>
            <input v-model="manualForm.currency" type="text" :placeholder="t('common.defaults.currency')" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm uppercase focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.tax_amount') }}</label>
            <input v-model.number="manualForm.taxAmount" type="number" min="0" step="0.01" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.shipping_amount') }}</label>
            <input v-model.number="manualForm.shippingAmount" type="number" min="0" step="0.01" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.related_inquiry') }}</label>
            <input v-model="manualForm.inquiryId" type="text" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
          </div>
        </div>
        <div class="space-y-3">
          <div v-for="(item, idx) in manualForm.items" :key="`manual-item-${idx}`" class="grid grid-cols-1 gap-3 rounded-md border border-gray-200 p-3 md:grid-cols-12">
            <div class="md:col-span-4">
              <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.product_id') }}</label>
              <input v-model="item.productId" type="text" :placeholder="t('customer.orders_new.product_id_placeholder')" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
            </div>
            <div class="md:col-span-2">
              <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.quantity') }}</label>
              <input v-model.number="item.quantity" type="number" min="1" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
            </div>
            <div class="md:col-span-2">
              <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.unit_price') }}</label>
              <input v-model.number="item.unitPrice" type="number" min="0" step="0.01" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
            </div>
            <div class="md:col-span-3">
              <label class="block text-sm font-medium text-gray-700">{{ t('customer.orders_new.specifications') }}</label>
              <input v-model="item.specifications" type="text" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500" />
            </div>
            <div class="md:col-span-1 flex items-end justify-end">
              <button v-if="manualForm.items.length > 1" @click="removeManualItem(idx)" type="button" class="rounded-md border border-red-200 px-2 py-2 text-xs font-medium text-red-700 hover:bg-red-50">{{ t('customer.orders_new.remove') }}</button>
            </div>
          </div>
          <button type="button" @click="addManualItem" class="inline-flex items-center rounded-md border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50">
            <Icon name="heroicons:plus" class="mr-1 h-4 w-4" />
            {{ t('customer.orders_new.add_item') }}
          </button>
        </div>
        <div class="flex flex-wrap items-center gap-3">
          <button :disabled="manualSubmitting" @click="submitManualOrder" class="inline-flex items-center rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-800 disabled:opacity-60">
            <Icon name="heroicons:shopping-bag" class="mr-1.5 h-4 w-4" />
            {{ manualSubmitting ? t('customer.orders_new.submitting_manual') : t('customer.orders_new.create_manual') }}
          </button>
          <div class="space-y-1">
            <p v-if="manualError" class="text-sm text-red-600">{{ manualError }}</p>
            <ul v-if="manualViolationDetails.length > 0" class="list-disc pl-5 text-sm text-red-700 space-y-1">
              <li v-for="(item, idx) in manualViolationDetails" :key="`manual-violation-${idx}`">{{ item }}</li>
            </ul>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

type ManualItem = { productId: string; quantity: number; unitPrice: number; specifications: string }

const { t } = useI18n()
const localePath = useLocalePath()
const { currencyOrDefault: cur, cell, enumLabel, formatNumber } = useDisplay()
const api = useApi()

const shippingAddress = reactive({ street: '', city: '', state: '', zipCode: '', country: '' })

const aiForm = reactive({
  prompt: '', targetCountry: '', quantity: 1, budget: 0,
  currency: cur(null), taxAmount: 0, shippingAmount: 0,
  additionalRequirements: '', inquiryId: ''
})

const manualForm = reactive({
  currency: cur(null), taxAmount: 0, shippingAmount: 0, inquiryId: '',
  items: [{ productId: '', quantity: 1, unitPrice: 0, specifications: '' }] as ManualItem[]
})

const aiSubmitting = ref(false)
const aiError = ref('')
const aiSuccess = ref('')
const aiViolationDetails = ref<string[]>([])
const manualSubmitting = ref(false)
const manualError = ref('')
const manualViolationDetails = ref<string[]>([])
const confirmingDraft = ref(false)
const confirmComplianceAck = ref(false)
const draftOrder = ref<any>(null)
const selectedProducts = ref<any[]>([])
const compliance = reactive({
  checklist: [] as string[], warnings: [] as string[], hasOfficialEvidence: false,
  paymentPolicy: { requiresFullPrepayment: false, allowedTerms: [] as string[], note: '' },
  references: [] as any[], retrievalNote: ''
})
const aiSummary = ref('')
const missingInformation = ref<string[]>([])

const fieldErrors = reactive<Record<string, string>>({})

// Clear individual field errors as user types
watch(() => shippingAddress.street, () => { if (fieldErrors.street) fieldErrors.street = '' })
watch(() => shippingAddress.city, () => { if (fieldErrors.city) fieldErrors.city = '' })
watch(() => shippingAddress.country, () => { if (fieldErrors.country) fieldErrors.country = '' })
const validateShippingAddress = () => {
  // Clear previous field errors
  fieldErrors.street = ''
  fieldErrors.city = ''
  fieldErrors.country = ''

  const errors: string[] = []
  if (!shippingAddress.street.trim()) {
    fieldErrors.street = t('customer.common.shipping_street_required')
    errors.push(fieldErrors.street)
  }
  if (!shippingAddress.city.trim()) {
    fieldErrors.city = t('customer.common.shipping_city_required')
    errors.push(fieldErrors.city)
  }
  if (!shippingAddress.country.trim()) {
    fieldErrors.country = t('customer.common.shipping_country_required')
    errors.push(fieldErrors.country)
  }
  return errors
}

const buildSharedAddressPayload = () => ({
  street: shippingAddress.street.trim(), city: shippingAddress.city.trim(),
  state: shippingAddress.state.trim(), zipCode: shippingAddress.zipCode.trim(),
  country: shippingAddress.country.trim()
})

const submitAIDraft = async () => {
  aiError.value = ''; aiSuccess.value = ''; aiViolationDetails.value = []
  const addressErrors = validateShippingAddress()
  if (addressErrors.length > 0) { aiError.value = addressErrors.join('; '); return }
  if (!aiForm.prompt.trim()) { aiError.value = t('customer.common.provide_intention'); return }
  const targetCountry = aiForm.targetCountry.trim() || shippingAddress.country.trim()
  if (!targetCountry) { aiError.value = t('customer.common.provide_country'); return }

  aiSubmitting.value = true
  try {
    const res = await api.createAIAssistOrder({
      prompt: aiForm.prompt.trim(), targetCountry,
      quantity: Math.max(1, Number(aiForm.quantity) || 1),
      budget: Number(aiForm.budget) || 0,
      currency: cur(aiForm.currency).toUpperCase(),
      taxAmount: Number(aiForm.taxAmount) || 0,
      shippingAmount: Number(aiForm.shippingAmount) || 0,
      shippingAddress: buildSharedAddressPayload(),
      additionalRequirements: aiForm.additionalRequirements.trim(),
      inquiryId: aiForm.inquiryId.trim() || undefined
    })
    draftOrder.value = res.order
    selectedProducts.value = Array.isArray(res.selectedProducts) ? res.selectedProducts : []
    // Handle "no products matched" response (order is null, suggestions provided)
    if (!res.order && Array.isArray(res.suggestions)) {
      aiError.value = res.message || t('customer.orders_new.ai_no_products_matched')
      aiViolationDetails.value = res.suggestions
      return
    }
    compliance.checklist = Array.isArray(res.compliance?.checklist) ? res.compliance.checklist : []
    compliance.warnings = Array.isArray(res.compliance?.warnings) ? res.compliance.warnings : []
    compliance.paymentPolicy.requiresFullPrepayment = Boolean(res.compliance?.paymentPolicy?.requiresFullPrepayment)
    compliance.paymentPolicy.allowedTerms = Array.isArray(res.compliance?.paymentPolicy?.allowedTerms) ? res.compliance.paymentPolicy.allowedTerms : []
    compliance.paymentPolicy.note = res.compliance?.paymentPolicy?.note || ''
    compliance.references = Array.isArray(res.compliance?.references) ? res.compliance.references : []
    compliance.retrievalNote = res.compliance?.retrievalNote || ''
    compliance.hasOfficialEvidence = Boolean(res.compliance?.hasOfficialEvidence)
    confirmComplianceAck.value = false
    aiSummary.value = res.ai?.summary || ''
    missingInformation.value = Array.isArray(res.ai?.missingInformation) ? res.ai.missingInformation : []
    aiSuccess.value = res.message || t('customer.orders_new.ai_draft_generated')
  } catch (err: any) {
    const violations = err?.details?.violations
    aiViolationDetails.value = Array.isArray(violations) ? violations : []
    aiError.value = err?.message || t('errors.api.ai_draft_failed')
  } finally { aiSubmitting.value = false }
}

const confirmDraftOrder = async () => {
  if (!draftOrder.value?.id) return
  if (!compliance.hasOfficialEvidence && !confirmComplianceAck.value) {
    aiError.value = t('customer.common.compliance_review_required'); return
  }
  confirmingDraft.value = true; aiError.value = ''
  try {
    // M1: Send edited quantities/prices back so the backend can update the draft
    const editedItems = selectedProducts.value.map((p: any) => ({
      productId: p.id,
      quantity: Math.max(1, Number(p.quantity) || 1),
      unitPrice: Math.max(0, Number(p.unitPrice) || 0),
      specifications: p.specifications || ''
    }))
    await api.post(`/user/orders/${draftOrder.value.id}/confirm`, {
      complianceAck: compliance.hasOfficialEvidence ? true : confirmComplianceAck.value,
      items: editedItems
    })
    await navigateTo(localePath(`/customer/orders/${draftOrder.value.id}`))
  } catch (err: any) {
    aiError.value = err?.message || t('errors.api.ai_confirm_failed')
  } finally { confirmingDraft.value = false }
}

const addManualItem = () => { manualForm.items.push({ productId: '', quantity: 1, unitPrice: 0, specifications: '' }) }
const removeManualItem = (idx: number) => { manualForm.items.splice(idx, 1) }

const submitManualOrder = async () => {
  manualError.value = ''; manualViolationDetails.value = []
  const addressErrors = validateShippingAddress()
  if (addressErrors.length > 0) { manualError.value = addressErrors.join('; '); return }
  const cleanItems = manualForm.items
    .map(item => ({ productId: item.productId.trim(), quantity: Math.max(1, Number(item.quantity) || 1), unitPrice: Math.max(0, Number(item.unitPrice) || 0), specifications: item.specifications.trim() }))
    .filter(item => item.productId.length > 0)
  if (cleanItems.length === 0) { manualError.value = t('customer.common.at_least_one_item'); return }

  manualSubmitting.value = true
  try {
    const res = await api.createOrder({
      items: cleanItems, currency: cur(manualForm.currency).toUpperCase(),
      taxAmount: Number(manualForm.taxAmount) || 0, shippingAmount: Number(manualForm.shippingAmount) || 0,
      shippingAddress: buildSharedAddressPayload(), inquiryId: manualForm.inquiryId.trim() || undefined
    })
    const createdId = res.order?.id
    if (createdId) { await navigateTo(localePath(`/customer/orders/${createdId}`)); return }
    await navigateTo(localePath('/customer/orders'))
  } catch (err: any) {
    const violations = err?.details?.violations
    manualViolationDetails.value = Array.isArray(violations) ? violations : []
    manualError.value = err?.message || t('errors.api.manual_order_failed')
  } finally { manualSubmitting.value = false }
}
</script>