<template>
  <div>
    <div class="mb-6">
      <NuxtLink :to="localePath('/customer/inquiries')" class="flex items-center text-sm font-medium text-orange-600 hover:text-orange-500">
        <Icon name="heroicons:arrow-left" class="mr-1 h-4 w-4" />
        {{ t('customer.inquiries.back') }}
      </NuxtLink>
    </div>

    <div v-if="pending" class="text-center py-10">
      <p class="text-gray-500">{{ t('customer.inquiries.loading_details') }}</p>
    </div>

    <div v-else-if="error" class="bg-red-50 p-4 rounded-md">
      <p class="text-red-700">{{ error }}</p>
    </div>

    <div v-else-if="inquiry" class="space-y-6">
      <!-- Confirmation Status Banner -->
      <div v-if="inquiry.customerConfirmed || inquiry.adminConfirmed || inquiry.status === 'pending_confirmation' || inquiry.status === 'confirmed'" :class="[confirmationBannerClass, 'shadow overflow-hidden sm:rounded-lg']">
        <div class="px-4 py-4 sm:px-6">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-medium">{{ t('customer.inquiries.confirmation_status') }}</h3>
            <span :class="[confirmationStatusBadge, 'inline-flex rounded-full px-3 py-1 text-sm font-semibold']">
              {{ confirmationLabel }}
            </span>
          </div>
          <div class="mt-3 grid grid-cols-2 gap-4 text-sm">
            <div>
              <span class="font-medium">{{ t('customer.inquiries.customer_confirmed') }}:</span>
              <Icon v-if="inquiry.customerConfirmed" name="heroicons:check-circle" class="inline ml-1 h-5 w-5 text-green-600" />
              <Icon v-else name="heroicons:x-circle" class="inline ml-1 h-5 w-5 text-gray-400" />
            </div>
            <div>
              <span class="font-medium">{{ t('customer.inquiries.admin_confirmed') }}:</span>
              <Icon v-if="inquiry.adminConfirmed" name="heroicons:check-circle" class="inline ml-1 h-5 w-5 text-green-600" />
              <Icon v-else name="heroicons:x-circle" class="inline ml-1 h-5 w-5 text-gray-400" />
            </div>
          </div>
        </div>
      </div>

      <!-- Inquiry Info Card -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 flex justify-between items-center bg-gray-50 border-b border-gray-200">
          <div>
            <h3 class="text-lg leading-6 font-medium text-gray-900">Inquiry #{{ inquiry.id.substring(0, 8) }}</h3>
            <p class="mt-1 max-w-2xl text-sm text-gray-500">
              {{ t('customer.inquiries.submitted_on') }} {{ formatDate(inquiry.createdAt) }}
            </p>
          </div>
          <span :class="[statusClass(inquiry.status), 'inline-flex rounded-full px-3 py-1 text-sm font-semibold leading-5']">
            {{ enumLabel('inquiry_status', inquiry.status) }}
          </span>
        </div>

        <div class="px-4 py-5 sm:p-6">
          <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.company') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.companyName || t('customer.common.na') }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.contact_person') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.contactPerson || t('customer.common.na') }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.email') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.email || t('customer.common.na') }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.whatsapp') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.whatsapp || t('customer.common.na') }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.target_country') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.targetCountry || t('customer.common.na') }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.estimated_quantity') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.estimatedQuantity || t('customer.common.na') }}</dd>
            </div>
            <div class="sm:col-span-2">
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.interested_products') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ formatProducts(inquiry.interestedProducts) }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.packaging_requirements') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.packagingRequirements || t('customer.common.na') }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.flavor_requirements') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.flavorRequirements || t('customer.common.na') }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.oem_needed') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.oemNeeded ? t('customer.common.yes') : t('customer.common.no') }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.expected_delivery') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.expectedDelivery || t('customer.common.na') }}</dd>
            </div>
            <div class="sm:col-span-2">
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.message') }}</dt>
              <dd class="mt-1 text-sm text-gray-900 whitespace-pre-line">{{ inquiry.message || t('customer.inquiries.no_message') }}</dd>
            </div>
            <div class="sm:col-span-2" v-if="inquiry.customerNotes">
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.business_notes') }}</dt>
              <dd class="mt-1 text-sm text-gray-900 whitespace-pre-line">{{ inquiry.customerNotes }}</dd>
            </div>
            <!-- Current packaging specs -->
            <div class="sm:col-span-2" v-if="inquiry.packagingType || inquiry.packagingWeight || inquiry.packagingSize || inquiry.qualityStandard">
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.confirmation_title') }}</dt>
              <dd class="mt-2 text-sm text-gray-900 grid grid-cols-2 gap-2">
                <span v-if="inquiry.packagingType">{{ t('customer.inquiries.packaging_type') }}: {{ inquiry.packagingType }}</span>
                <span v-if="inquiry.packagingWeight">{{ t('customer.inquiries.packaging_weight') }}: {{ inquiry.packagingWeight }}g</span>
                <span v-if="inquiry.packagingSize">{{ t('customer.inquiries.packaging_size') }}: {{ inquiry.packagingSize }}</span>
                <span v-if="inquiry.qualityStandard">{{ t('customer.inquiries.quality_standard') }}: {{ inquiry.qualityStandard }}</span>
              </dd>
            </div>
            <!-- Confirmation notes -->
            <div class="sm:col-span-2" v-if="inquiry.confirmationNotes">
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.confirmation_notes') }}</dt>
              <dd class="mt-1 text-sm text-gray-900 whitespace-pre-line">{{ inquiry.confirmationNotes }}</dd>
            </div>
          </dl>
        </div>
      </div>

      <!-- Files Section -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.inquiries.files_title') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <!-- Existing files -->
          <div v-if="inquiry.files && inquiry.files.length > 0" class="mb-4">
            <p class="text-sm font-medium text-gray-500 mb-2">{{ t('customer.inquiries.current_files') }} ({{ inquiry.files.length }})</p>
            <ul class="space-y-1">
              <li v-for="(fileUrl, idx) in inquiry.files" :key="idx" class="text-sm">
                <a :href="fileUrl" target="_blank" class="text-orange-600 hover:text-orange-500 truncate block">
                  <Icon name="heroicons:paper-clip" class="inline h-4 w-4 mr-1" />
                  {{ fileUrl.split('/').pop() || fileUrl }}
                </a>
              </li>
            </ul>
          </div>
          <div v-else class="text-sm text-gray-400 mb-4">{{ t('customer.inquiries.no_files') }}</div>

          <!-- Upload form -->
          <form @submit.prevent="uploadAttachment" class="space-y-3 border-t pt-4">
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries.attachment_note') }}</label>
              <input v-model="attachmentNote" type="text" :placeholder="t('customer.inquiries.attachment_note_placeholder')" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries.upload_attachment') }}</label>
              <input ref="fileInput" type="file" multiple accept=".docx,.doc,.pdf,.png,.jpg,.jpeg,.mkv,.mp4,.mp3,.txt,.svg" class="mt-1 block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-orange-50 file:text-orange-700 hover:file:bg-orange-100" />
              <p class="mt-1 text-xs text-gray-400">docx, doc, pdf, png, jpeg, mkv, mp4, mp3, txt, svg (max 32MB)</p>
            </div>
            <div v-if="uploadMessage" class="text-sm" :class="uploadError ? 'text-red-600' : 'text-green-600'">
              {{ uploadMessage }}
            </div>
            <div class="flex justify-end">
              <button type="submit" :disabled="uploading" class="inline-flex justify-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700 disabled:opacity-50">
                <Icon v-if="uploading" name="heroicons:arrow-path" class="animate-spin h-4 w-4 mr-1" />
                {{ uploading ? t('customer.inquiries.uploading') : t('customer.inquiries.upload_attachment_btn') }}
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- Packaging Confirmation Form (show if not yet confirmed by customer) -->
      <div v-if="!inquiry.customerConfirmed" class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.inquiries.confirmation_title') }}</h3>
          <p class="mt-1 text-sm text-gray-500">{{ t('customer.inquiries.waiting_admin') }}</p>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <form @submit.prevent="confirmInquiry" class="space-y-4">
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label for="pkg-type" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries.packaging_type') }}</label>
                <input id="pkg-type" v-model="confirmForm.packagingType" type="text" :placeholder="t('customer.inquiries.packaging_type_placeholder')" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="pkg-weight" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries.packaging_weight') }}</label>
                <input id="pkg-weight" v-model.number="confirmForm.packagingWeight" type="number" min="0" step="0.01" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="pkg-size" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries.packaging_size') }}</label>
                <input id="pkg-size" v-model="confirmForm.packagingSize" type="text" :placeholder="t('customer.inquiries.packaging_size_placeholder')" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="quality-std" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries.quality_standard') }}</label>
                <input id="quality-std" v-model="confirmForm.qualityStandard" type="text" :placeholder="t('customer.inquiries.quality_standard_placeholder')" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
            </div>
            <div>
              <label for="confirm-notes" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries.confirmation_notes') }}</label>
              <textarea id="confirm-notes" v-model="confirmForm.notes" rows="2" :placeholder="t('customer.inquiries.confirmation_notes_placeholder')" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div v-if="confirmMessage" class="text-sm" :class="confirmError ? 'text-red-600' : 'text-green-600'">
              {{ confirmMessage }}
            </div>
            <div class="flex justify-end">
              <button type="submit" :disabled="confirming" class="inline-flex justify-center rounded-md border border-transparent bg-emerald-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-emerald-700 disabled:opacity-50">
                {{ confirming ? t('customer.inquiries.confirming') : t('customer.inquiries.confirm_button') }}
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- Negotiation Section -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.inquiries.negotiation') }}</h3>
          <p class="mt-1 text-sm text-gray-500">{{ t('customer.inquiries.negotiation_description') }}</p>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <div v-if="negotiationOffers.length === 0 && !showNegotiationForm" class="text-center text-gray-500 py-6">
            <p class="mb-4">{{ t('customer.inquiries.no_offers') }}</p>
            <button type="button" @click="showNegotiationForm = true" class="inline-flex justify-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700">
              {{ t('customer.inquiries.create_offer') }}
            </button>
          </div>

          <div v-if="negotiationOffers.length > 0" class="space-y-4 mb-6">
            <div v-for="(offer, oidx) in negotiationOffers" :key="offer.id" class="border rounded-lg p-4" :class="offer.status === 'accepted' ? 'border-green-300 bg-green-50' : offer.status === 'rejected' ? 'border-red-300 bg-red-50' : 'border-gray-200'">
              <div class="flex items-center justify-between mb-2">
                <span class="text-sm font-medium text-gray-900">{{ t('customer.inquiries.offer_round') }} #{{ oidx + 1 }}</span>
                <span class="text-xs px-2 py-0.5 rounded-full" :class="offer.status === 'accepted' ? 'bg-green-100 text-green-800' : offer.status === 'rejected' ? 'bg-red-100 text-red-800' : 'bg-yellow-100 text-yellow-800'">
                  {{ offer.status === 'accepted' ? t('customer.inquiries.offer_accepted_status') : offer.status === 'rejected' ? t('customer.inquiries.offer_rejected_status') : t('customer.inquiries.offer_pending') }}
                </span>
              </div>
              <dl class="grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
                <div><dt class="text-gray-500">{{ t('customer.inquiries.offer_by') }}</dt><dd class="text-gray-900">{{ offer.senderType }}</dd></div>
                <div><dt class="text-gray-500">{{ t('customer.inquiries.offer_unit_price') }}</dt><dd class="text-gray-900">{{ offer.unitPrice }}</dd></div>
                <div><dt class="text-gray-500">{{ t('customer.inquiries.offer_quantity') }}</dt><dd class="text-gray-900">{{ offer.quantity }}</dd></div>
                <div><dt class="text-gray-500">{{ t('customer.inquiries.offer_total_amount') }}</dt><dd class="text-gray-900 font-semibold">{{ offer.totalAmount }}</dd></div>
                <div v-if="offer.incoterms"><dt class="text-gray-500">{{ t('customer.inquiries.offer_incoterms') }}</dt><dd class="text-gray-900">{{ offer.incoterms }}</dd></div>
                <div v-if="offer.paymentTerms"><dt class="text-gray-500">{{ t('customer.inquiries.offer_payment_terms') }}</dt><dd class="text-gray-900">{{ offer.paymentTerms }}</dd></div>
                <div v-if="offer.deliveryDate"><dt class="text-gray-500">{{ t('customer.inquiries.offer_delivery_date') }}</dt><dd class="text-gray-900">{{ offer.deliveryDate }}</dd></div>
              </dl>
              <p v-if="offer.message" class="mt-2 text-sm text-gray-600 bg-white rounded p-2">{{ offer.message }}</p>
              <div v-if="offer.status === 'pending' && offer.senderType === 'admin'" class="mt-3 flex gap-2">
                <button type="button" @click="acceptOffer(offer.id)" class="text-sm rounded border border-transparent bg-emerald-600 px-3 py-1 text-white hover:bg-emerald-700">{{ t('customer.inquiries.offer_accept') }}</button>
                <button type="button" @click="rejectOffer(offer.id)" class="text-sm rounded border border-gray-300 bg-white px-3 py-1 text-gray-700 hover:bg-gray-50">{{ t('customer.inquiries.offer_reject') }}</button>
              </div>
            </div>
            <button type="button" v-if="!showNegotiationForm" @click="showNegotiationForm = true" class="text-sm text-orange-600 hover:text-orange-500">{{ t('customer.inquiries.create_offer') }}</button>
          </div>

          <form v-if="showNegotiationForm" @submit.prevent="submitOffer" class="border-t pt-4 space-y-4">
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label for="nego-unitPrice" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries.offer_unit_price') }}</label>
                <input id="nego-unitPrice" v-model.number="negotiationForm.unitPrice" name="nego-unitPrice" type="number" min="0" step="0.01" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="nego-quantity" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries.offer_quantity') }}</label>
                <input id="nego-quantity" v-model.number="negotiationForm.quantity" name="nego-quantity" type="number" min="1" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="nego-totalAmount" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries.offer_total_amount') }} *</label>
                <input id="nego-totalAmount" v-model.number="negotiationForm.totalAmount" name="nego-totalAmount" type="number" min="0" step="0.01" autocomplete="off" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="nego-currency" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries.offer_currency') }}</label>
                <select id="nego-currency" v-model="negotiationForm.currency" name="nego-currency" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                  <option value="USD">USD</option>
                  <option value="EUR">EUR</option>
                  <option value="GBP">GBP</option>
                  <option value="CNY">CNY</option>
                </select>
              </div>
              <div>
                <label for="nego-incoterms" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries.offer_incoterms') }}</label>
                <input id="nego-incoterms" v-model="negotiationForm.incoterms" name="nego-incoterms" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="nego-paymentTerms" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries.offer_payment_terms') }}</label>
                <input id="nego-paymentTerms" v-model="negotiationForm.paymentTerms" name="nego-paymentTerms" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="nego-deliveryDate" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries.offer_delivery_date') }}</label>
                <input id="nego-deliveryDate" v-model="negotiationForm.deliveryDate" name="nego-deliveryDate" type="date" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="nego-validUntil" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries.offer_valid_until') }}</label>
                <input id="nego-validUntil" v-model="negotiationForm.validUntil" name="nego-validUntil" type="datetime-local" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
            </div>
            <div>
              <label for="nego-message" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries.offer_message') }}</label>
              <textarea id="nego-message" v-model="negotiationForm.message" name="nego-message" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div v-if="negotiationMessage" class="text-sm" :class="negotiationError ? 'text-red-600' : 'text-green-600'">
              {{ negotiationMessage }}
            </div>
            <div class="flex justify-end gap-3">
              <button type="button" @click="showNegotiationForm = false" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700">{{ t('customer.inquiries.offer_reject') }}</button>
              <button type="submit" :disabled="submittingOffer" class="inline-flex justify-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700 disabled:opacity-50">
                {{ submittingOffer ? t('customer.inquiries.offer_submitting') : t('customer.inquiries.create_offer') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive } from 'vue'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const route = useRoute()
const api = useApi()
const { t } = useI18n()
const { enumLabel, formatDate } = useDisplay()
const localePath = useLocalePath()

const inquiry = ref<any>(null)
const pending = ref(true)
const error = ref('')

const fileInput = ref<HTMLInputElement | null>(null)
const attachmentNote = ref('')
const uploading = ref(false)
const uploadMessage = ref('')
const uploadError = ref(false)

const confirming = ref(false)
const confirmMessage = ref('')
const confirmError = ref(false)
const confirmForm = reactive({
  packagingType: '',
  packagingWeight: 0,
  packagingSize: '',
  qualityStandard: '',
  notes: ''
})

const negotiationOffers = ref<any[]>([])
const showNegotiationForm = ref(false)
const submittingOffer = ref(false)
const negotiationMessage = ref('')
const negotiationError = ref(false)
const negotiationForm = reactive({
  unitPrice: 0,
  quantity: 1,
  totalAmount: 0,
  currency: 'USD',
  incoterms: '',
  paymentTerms: '',
  deliveryDate: '',
  validUntil: '',
  message: ''
})

const fetchInquiry = async () => {
  pending.value = true; error.value = ''
  try { inquiry.value = await api.get<any>(`/user/inquiries/${route.params.id}`) }
  catch (err: any) { error.value = err?.message || t('errors.api.load_failed') }
  finally { pending.value = false }
}

const uploadAttachment = async () => {
  const files = fileInput.value?.files
  if (!files || files.length === 0) {
    uploadError.value = true
    uploadMessage.value = t('customer.inquiries.upload_error')
    return
  }
  uploading.value = true
  uploadMessage.value = ''
  uploadError.value = false
  try {
    const formData = new FormData()
    for (let i = 0; i < files.length; i++) formData.append('files', files[i])
    if (attachmentNote.value.trim()) formData.append('note', attachmentNote.value.trim())
    await api.customerUploadInquiryAttachment(route.params.id as string, formData)
    uploadMessage.value = t('customer.inquiries.upload_success')
    attachmentNote.value = ''
    if (fileInput.value) fileInput.value.value = ''
    await fetchInquiry()
  } catch (err: any) {
    uploadError.value = true
    uploadMessage.value = err?.message || t('customer.inquiries.upload_error')
  } finally { uploading.value = false }
}

const confirmInquiry = async () => {
  confirming.value = true
  confirmMessage.value = ''
  confirmError.value = false
  try {
    await api.customerConfirmInquiry(route.params.id as string, {
      packagingType: confirmForm.packagingType,
      packagingWeight: confirmForm.packagingWeight,
      packagingSize: confirmForm.packagingSize,
      qualityStandard: confirmForm.qualityStandard,
      notes: confirmForm.notes
    })
    confirmMessage.value = t('customer.inquiries.confirm_success')
    await fetchInquiry()
  } catch (err: any) {
    confirmError.value = true
    confirmMessage.value = err?.message || t('customer.inquiries.confirm_error')
  } finally { confirming.value = false }
}

const formatProducts = (products: unknown) => {
  if (Array.isArray(products)) return products.length > 0 ? products.join(', ') : t('customer.inquiries.general_inquiry')
  if (typeof products === 'string') return products.trim() !== '' ? products : t('customer.inquiries.general_inquiry')
  return t('customer.inquiries.general_inquiry')
}

const statusClass = (status: string) => {
  if (status === 'pending') return 'bg-yellow-100 text-yellow-800'
  if (status === 'quoted' || status === 'contacted') return 'bg-orange-100 text-orange-800'
  if (status === 'negotiating') return 'bg-amber-100 text-amber-800'
  if (status === 'won' || status === 'confirmed') return 'bg-green-100 text-green-800'
  if (status === 'lost' || status === 'closed') return 'bg-gray-100 text-gray-800'
  if (status === 'pending_confirmation') return 'bg-blue-100 text-blue-800'
  return 'bg-gray-100 text-gray-800'
}

const confirmationBannerClass = computed(() => {
  if (inquiry.value?.status === 'confirmed' || (inquiry.value?.customerConfirmed && inquiry.value?.adminConfirmed)) return 'bg-green-50 border border-green-200'
  if (inquiry.value?.status === 'pending_confirmation') return 'bg-blue-50 border border-blue-200'
  return 'bg-yellow-50 border border-yellow-200'
})

const confirmationStatusBadge = computed(() => {
  if (inquiry.value?.status === 'confirmed' || (inquiry.value?.customerConfirmed && inquiry.value?.adminConfirmed)) return 'bg-green-200 text-green-800'
  if (inquiry.value?.status === 'pending_confirmation') return 'bg-blue-200 text-blue-800'
  return 'bg-yellow-200 text-yellow-800'
})

const confirmationLabel = computed(() => {
  if (inquiry.value?.status === 'confirmed' || (inquiry.value?.customerConfirmed && inquiry.value?.adminConfirmed)) return t('customer.inquiries.fully_confirmed')
  if (inquiry.value?.customerConfirmed && !inquiry.value?.adminConfirmed) return t('customer.inquiries.waiting_admin')
  if (!inquiry.value?.customerConfirmed && inquiry.value?.adminConfirmed) return t('customer.inquiries.waiting_customer')
  return t('customer.inquiries.pending_confirmation')
})

const fetchNegotiationOffers = async () => {
  try {
    const data = await api.get<any>(`/user/inquiries/${route.params.id}/negotiations`)
    negotiationOffers.value = data?.data || data || []
  } catch (err: any) {
    console.error('Failed to fetch negotiation offers:', err)
  }
}

const submitOffer = async () => {
  if (!negotiationForm.totalAmount || negotiationForm.totalAmount <= 0) {
    negotiationMessage.value = t('customer.inquiries.offer_amount_required')
    negotiationError.value = true
    return
  }
  submittingOffer.value = true
  negotiationMessage.value = ''
  negotiationError.value = false
  try {
    const body: any = { ...negotiationForm }
    if (body.validUntil) body.validUntil = new Date(body.validUntil).toISOString()
    await api.post(`/user/inquiries/${route.params.id}/negotiations`, body)
    negotiationMessage.value = t('customer.inquiries.offer_created')
    showNegotiationForm.value = false
    Object.assign(negotiationForm, { unitPrice: 0, quantity: 1, totalAmount: 0, currency: 'USD', incoterms: '', paymentTerms: '', deliveryDate: '', validUntil: '', message: '' })
    await fetchNegotiationOffers()
  } catch (err: any) {
    negotiationError.value = true
    negotiationMessage.value = err?.data?.message || err.message || t('errors.api.save_failed')
  } finally {
    submittingOffer.value = false
  }
}

const acceptOffer = async (offerId: string) => {
  try {
    await api.post(`/user/inquiries/${route.params.id}/negotiations/${offerId}/accept`, {})
    await fetchNegotiationOffers()
  } catch (err: any) {
    alert(err?.data?.message || err.message || t('errors.api.save_failed'))
  }
}

const rejectOffer = async (offerId: string) => {
  try {
    await api.post(`/user/inquiries/${route.params.id}/negotiations/${offerId}/reject`, {})
    await fetchNegotiationOffers()
  } catch (err: any) {
    alert(err?.data?.message || err.message || t('errors.api.save_failed'))
  }
}

onMounted(() => { fetchInquiry(); fetchNegotiationOffers() })
</script>
