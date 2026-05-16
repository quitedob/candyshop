<template>
  <div>
    <div class="mb-6">
      <NuxtLink :to="localePath('/admin/inquiries')" class="flex items-center text-sm font-medium text-orange-600 hover:text-orange-500">
        <Icon name="heroicons:arrow-left" class="mr-1 h-4 w-4" aria-hidden="true" />
        {{ t('admin.inquiryDetail.back') }}
      </NuxtLink>
    </div>

    <div v-if="pending" class="text-center py-10">
      <p class="text-gray-500">{{ t('admin.inquiryDetail.loading_details') }}</p>
    </div>

    <div v-else-if="error" class="bg-red-50 p-4 rounded-md">
      <p class="text-red-700">{{ error }}</p>
    </div>

    <div v-else-if="inquiry" class="space-y-6">
      <!-- Inquiry Info Card -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 flex justify-between items-center bg-gray-50 border-b border-gray-200">
          <div>
            <h3 class="text-lg leading-6 font-medium text-gray-900">Inquiry #{{ String(inquiry.id).substring(0, 8) }}</h3>
            <p class="mt-1 max-w-2xl text-sm text-gray-500">
              {{ t('admin.inquiryDetail.submitted_on') }} {{ formatDate(inquiry.createdAt, { dateStyle: 'medium', timeStyle: 'short' }) }}
            </p>
          </div>
          <div class="flex items-center gap-3">
            <span :class="[statusClass(inquiry.status), 'inline-flex rounded-full px-3 py-1 text-sm font-semibold leading-5']">
              {{ enumLabel('inquiry_status', inquiry.status) }}
            </span>
            <span v-if="inquiry.priority" :class="[priorityClass(inquiry.priority), 'inline-flex rounded-full px-3 py-1 text-sm font-semibold leading-5']">
              {{ enumLabel('inquiry_priority', inquiry.priority) }}
            </span>
          </div>
        </div>

        <div class="px-4 py-5 sm:p-6">
          <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
            <div v-if="inquiry.userId">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.user') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.userId }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.company_name') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ cell(inquiry.companyName) }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.contact_person') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ cell(inquiry.contactPerson) }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.email') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ cell(inquiry.email) }}</dd>
            </div>
            <div v-if="inquiry.whatsapp">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.whatsapp') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.whatsapp }}</dd>
            </div>
            <div v-if="inquiry.targetCountry">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.target_country') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.targetCountry }}</dd>
            </div>
            <div v-if="inquiry.estimatedQuantity">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.estimated_quantity') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.estimatedQuantity }}</dd>
            </div>
            <div v-if="inquiry.expectedDelivery">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.expected_delivery') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.expectedDelivery }}</dd>
            </div>
            <div v-if="inquiry.assignedTo">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.assigned_to') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.assignedTo }}</dd>
            </div>
            <div v-if="inquiry.quotedAmount" class="sm:col-span-2">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.quoted_amount') }}</dt>
              <dd class="mt-1 text-sm font-medium text-gray-900">{{ cur(inquiry.currency) }} {{ formatNumber(inquiry.quotedAmount || 0) }}</dd>
            </div>
            <div class="sm:col-span-2" v-if="inquiry.interestedProducts && inquiry.interestedProducts.length">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.interested_products') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ Array.isArray(inquiry.interestedProducts) ? inquiry.interestedProducts.join(', ') : inquiry.interestedProducts }}</dd>
            </div>
            <div v-if="inquiry.packagingRequirements">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.packaging_requirements') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.packagingRequirements }}</dd>
            </div>
            <div v-if="inquiry.flavorRequirements">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.flavor_requirements') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.flavorRequirements }}</dd>
            </div>
            <div v-if="inquiry.oemNeeded">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.oem_needed') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ t('admin.inquiryDetail.yes') }}</dd>
            </div>
            <div class="sm:col-span-2" v-if="inquiry.message">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.message') }}</dt>
              <dd class="mt-1 text-sm text-gray-900 whitespace-pre-line">{{ inquiry.message }}</dd>
            </div>
            <div class="sm:col-span-2" v-if="inquiry.customerNotes">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.business_notes') }}</dt>
              <dd class="mt-1 text-sm text-gray-900 whitespace-pre-line">{{ inquiry.customerNotes }}</dd>
            </div>
          </dl>
        </div>
      </div>

      <!-- AI Compliance Check Results -->
      <div v-if="inquiry.aiComplianceCheck" class="bg-amber-50 shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-amber-100 border-b border-amber-200">
          <h3 class="text-lg leading-6 font-medium text-amber-900">{{ t('admin.inquiryDetail.ai_compliance_check') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <dl class="grid grid-cols-1 gap-x-4 gap-y-4 sm:grid-cols-2">
            <div>
              <dt class="text-sm font-medium text-amber-700">{{ t('admin.inquiryDetail.compliance_status') }}</dt>
              <dd class="mt-1">
                <span :class="[inquiry.aiComplianceCheck.passed ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800', 'inline-flex rounded-full px-2 text-xs font-semibold leading-5']">
                  {{ inquiry.aiComplianceCheck.passed ? t('admin.inquiryDetail.passed') : t('admin.inquiryDetail.failed') }}
                </span>
              </dd>
            </div>
            <div v-if="inquiry.aiComplianceCheck.warnings && inquiry.aiComplianceCheck.warnings.length">
              <dt class="text-sm font-medium text-amber-700">{{ t('admin.inquiryDetail.warnings') }}</dt>
              <dd class="mt-1 text-sm text-amber-900">
                <ul class="list-disc list-inside">
                  <li v-for="(warning, idx) in inquiry.aiComplianceCheck.warnings" :key="idx">{{ warning }}</li>
                </ul>
              </dd>
            </div>
            <div v-if="inquiry.aiComplianceCheck.summary" class="sm:col-span-2">
              <dt class="text-sm font-medium text-amber-700">{{ t('admin.inquiryDetail.summary') }}</dt>
              <dd class="mt-1 text-sm text-amber-900">{{ inquiry.aiComplianceCheck.summary }}</dd>
            </div>
          </dl>
        </div>
      </div>

      <!-- Quotation Section -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.inquiryDetail.quotation') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <form @submit.prevent="submitQuote" class="space-y-4">
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label for="inquiry-quotedAmount" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.quoted_amount') }}</label>
                <input id="inquiry-quotedAmount" v-model.number="quoteForm.quotedAmount" name="quotedAmount" type="number" min="0" step="0.01" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="inquiry-validUntil" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.valid_until') }}</label>
                <input id="inquiry-validUntil" v-model="quoteForm.validUntil" name="validUntil" type="datetime-local" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
            </div>
            <div>
              <label for="inquiry-notes" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.notes') }}</label>
              <textarea id="inquiry-notes" v-model="quoteForm.customerNotes" name="customerNotes" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div v-if="quoteMessage" class="text-sm" :class="quoteError ? 'text-red-600' : 'text-green-600'">
              {{ quoteMessage }}
            </div>
            <div class="flex justify-end">
              <button type="submit" :disabled="submittingQuote" class="inline-flex justify-center rounded-md border border-transparent bg-emerald-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-emerald-700 disabled:opacity-50">
                {{ submittingQuote ? t('admin.inquiryDetail.submitting') : t('admin.inquiryDetail.submit_quote') }}
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- Negotiation Section -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.inquiryDetail.negotiation') }}</h3>
          <p class="mt-1 text-sm text-gray-500">{{ t('admin.inquiryDetail.negotiation_description') }}</p>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <div v-if="negotiationOffers.length === 0 && !showNegotiationForm" class="text-center text-gray-500 py-6">
            <p class="mb-4">{{ t('admin.inquiryDetail.no_offers') }}</p>
            <button type="button" @click="showNegotiationForm = true" class="inline-flex justify-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700">
              {{ t('admin.inquiryDetail.create_offer') }}
            </button>
          </div>

          <div v-if="negotiationOffers.length > 0" class="space-y-4 mb-6">
            <div v-for="(offer, oidx) in negotiationOffers" :key="offer.id" class="border rounded-lg p-4" :class="offer.status === 'accepted' ? 'border-green-300 bg-green-50' : offer.status === 'rejected' ? 'border-red-300 bg-red-50' : 'border-gray-200'">
              <div class="flex items-center justify-between mb-2">
                <span class="text-sm font-medium text-gray-900">{{ t('admin.inquiryDetail.offer_round') }} #{{ oidx + 1 }}</span>
                <span class="text-xs px-2 py-0.5 rounded-full" :class="offer.status === 'accepted' ? 'bg-green-100 text-green-800' : offer.status === 'rejected' ? 'bg-red-100 text-red-800' : 'bg-yellow-100 text-yellow-800'">
                  {{ offer.status === 'accepted' ? t('admin.inquiryDetail.offer_accepted_status') : offer.status === 'rejected' ? t('admin.inquiryDetail.offer_rejected_status') : t('admin.inquiryDetail.offer_pending') }}
                </span>
              </div>
              <dl class="grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
                <div><dt class="text-gray-500">{{ t('admin.inquiryDetail.offer_by') }}</dt><dd class="text-gray-900">{{ offer.senderType }}</dd></div>
                <div><dt class="text-gray-500">{{ t('admin.inquiryDetail.offer_unit_price') }}</dt><dd class="text-gray-900">{{ cur(null) }} {{ offer.unitPrice }}</dd></div>
                <div><dt class="text-gray-500">{{ t('admin.inquiryDetail.offer_quantity') }}</dt><dd class="text-gray-900">{{ offer.quantity }}</dd></div>
                <div><dt class="text-gray-500">{{ t('admin.inquiryDetail.offer_total_amount') }}</dt><dd class="text-gray-900 font-semibold">{{ cur(null) }} {{ formatNumber(offer.totalAmount) }}</dd></div>
                <div v-if="offer.incoterms"><dt class="text-gray-500">{{ t('admin.inquiryDetail.offer_incoterms') }}</dt><dd class="text-gray-900">{{ offer.incoterms }}</dd></div>
                <div v-if="offer.paymentTerms"><dt class="text-gray-500">{{ t('admin.inquiryDetail.offer_payment_terms') }}</dt><dd class="text-gray-900">{{ offer.paymentTerms }}</dd></div>
                <div v-if="offer.deliveryDate"><dt class="text-gray-500">{{ t('admin.inquiryDetail.offer_delivery_date') }}</dt><dd class="text-gray-900">{{ offer.deliveryDate }}</dd></div>
              </dl>
              <p v-if="offer.message" class="mt-2 text-sm text-gray-600 bg-white rounded p-2">{{ offer.message }}</p>
              <div v-if="offer.status === 'pending' && offer.senderType === 'customer'" class="mt-3 flex gap-2">
                <button type="button" @click="acceptOffer(offer.id)" class="text-sm rounded border border-transparent bg-emerald-600 px-3 py-1 text-white hover:bg-emerald-700">{{ t('admin.inquiryDetail.offer_accept') }}</button>
                <button type="button" @click="rejectOffer(offer.id)" class="text-sm rounded border border-gray-300 bg-white px-3 py-1 text-gray-700 hover:bg-gray-50">{{ t('admin.inquiryDetail.offer_reject') }}</button>
              </div>
            </div>
            <button type="button" v-if="!showNegotiationForm" @click="showNegotiationForm = true" class="text-sm text-orange-600 hover:text-orange-500">{{ t('admin.inquiryDetail.create_offer') }}</button>
          </div>

          <form v-if="showNegotiationForm" @submit.prevent="submitOffer" class="border-t pt-4 space-y-4">
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label for="nego-unitPrice" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.offer_unit_price') }}</label>
                <input id="nego-unitPrice" v-model.number="negotiationForm.unitPrice" name="nego-unitPrice" type="number" min="0" step="0.01" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="nego-quantity" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.offer_quantity') }}</label>
                <input id="nego-quantity" v-model.number="negotiationForm.quantity" name="nego-quantity" type="number" min="1" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="nego-totalAmount" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.offer_total_amount') }} *</label>
                <input id="nego-totalAmount" v-model.number="negotiationForm.totalAmount" name="nego-totalAmount" type="number" min="0" step="0.01" autocomplete="off" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="nego-currency" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.offer_currency') }}</label>
                <select id="nego-currency" v-model="negotiationForm.currency" name="nego-currency" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                  <option value="USD">USD</option>
                  <option value="EUR">EUR</option>
                  <option value="GBP">GBP</option>
                  <option value="CNY">CNY</option>
                </select>
              </div>
              <div>
                <label for="nego-incoterms" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.offer_incoterms') }}</label>
                <input id="nego-incoterms" v-model="negotiationForm.incoterms" name="nego-incoterms" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="nego-paymentTerms" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.offer_payment_terms') }}</label>
                <input id="nego-paymentTerms" v-model="negotiationForm.paymentTerms" name="nego-paymentTerms" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="nego-deliveryDate" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.offer_delivery_date') }}</label>
                <input id="nego-deliveryDate" v-model="negotiationForm.deliveryDate" name="nego-deliveryDate" type="date" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="nego-validUntil" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.offer_valid_until') }}</label>
                <input id="nego-validUntil" v-model="negotiationForm.validUntil" name="nego-validUntil" type="datetime-local" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
            </div>
            <div>
              <label for="nego-message" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.offer_message') }}</label>
              <textarea id="nego-message" v-model="negotiationForm.message" name="nego-message" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div v-if="negotiationMessage" class="text-sm" :class="negotiationError ? 'text-red-600' : 'text-green-600'">
              {{ negotiationMessage }}
            </div>
            <div class="flex justify-end gap-3">
              <button type="button" @click="showNegotiationForm = false" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700">{{ t('admin.pricing.cancel') }}</button>
              <button type="submit" :disabled="submittingOffer" class="inline-flex justify-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700 disabled:opacity-50">
                {{ submittingOffer ? t('admin.inquiryDetail.offer_submitting') : t('admin.inquiryDetail.create_offer') }}
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- Packaging & Standards Confirmation (Admin) -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.inquiryDetail.confirm_specs_title') }}</h3>
          <p class="mt-1 text-sm text-gray-500">{{ t('admin.inquiryDetail.confirm_specs_desc') }}</p>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <!-- Current customer specs -->
          <div v-if="inquiry.packagingType || inquiry.packagingWeight || inquiry.packagingSize || inquiry.qualityStandard" class="mb-4 p-3 bg-gray-50 rounded-md">
            <p class="text-sm font-medium text-gray-700 mb-2">{{ t('admin.inquiryDetail.current_packaging') }}</p>
            <dl class="grid grid-cols-2 gap-2 text-sm">
              <div v-if="inquiry.packagingType"><dt class="text-gray-500 inline">{{ t('admin.inquiryDetail.packaging_type') }}:</dt><dd class="text-gray-900 inline ml-1">{{ inquiry.packagingType }}</dd></div>
              <div v-if="inquiry.packagingWeight"><dt class="text-gray-500 inline">{{ t('admin.inquiryDetail.packaging_weight') }}:</dt><dd class="text-gray-900 inline ml-1">{{ inquiry.packagingWeight }}{{ t('admin.inquiryDetail.packaging_weight_unit') }}</dd></div>
              <div v-if="inquiry.packagingSize"><dt class="text-gray-500 inline">{{ t('admin.inquiryDetail.packaging_size') }}:</dt><dd class="text-gray-900 inline ml-1">{{ inquiry.packagingSize }}</dd></div>
              <div v-if="inquiry.qualityStandard"><dt class="text-gray-500 inline">{{ t('admin.inquiryDetail.quality_standard') }}:</dt><dd class="text-gray-900 inline ml-1">{{ inquiry.qualityStandard }}</dd></div>
            </dl>
          </div>
          <div v-else class="mb-4 text-sm text-gray-400">{{ t('admin.inquiryDetail.packaging_not_set') }}</div>

          <!-- Confirmation status -->
          <div class="mb-4 flex items-center gap-4 text-sm">
            <span :class="inquiry.customerConfirmed ? 'text-green-700' : 'text-gray-400'">
              <Icon :name="inquiry.customerConfirmed ? 'heroicons:check-circle' : 'heroicons:clock'" class="inline h-5 w-5 mr-1" aria-hidden="true" />
              {{ inquiry.customerConfirmed ? t('admin.inquiryDetail.customer_already_confirmed') : t('admin.inquiryDetail.waiting_for_customer') }}
            </span>
            <span v-if="inquiry.customerConfirmed && inquiry.adminConfirmed" class="text-green-700 font-medium">{{ t('admin.inquiryDetail.both_confirmed') }}</span>
          </div>

          <!-- View confirmation notes -->
          <div v-if="inquiry.confirmationNotes" class="mb-4 p-3 bg-yellow-50 rounded-md text-sm whitespace-pre-line">
            <p class="font-medium text-yellow-800 mb-1">{{ t('admin.inquiryDetail.confirmation_notes') }}:</p>
            <p class="text-yellow-700">{{ inquiry.confirmationNotes }}</p>
          </div>

          <!-- Customer files -->
          <div v-if="inquiry.files && inquiry.files.length > 0" class="mb-4">
            <p class="text-sm font-medium text-gray-700 mb-1">{{ t('admin.inquiryDetail.customer_files') }} ({{ inquiry.files.length }})</p>
            <ul class="space-y-1">
              <li v-for="(f, idx) in inquiry.files" :key="idx" class="text-sm">
                <a :href="f" target="_blank" class="text-orange-600 hover:text-orange-500">
                  <Icon name="heroicons:paper-clip" class="inline h-4 w-4 mr-1" aria-hidden="true" />{{ f.split('/').pop() || f }}
                </a>
              </li>
            </ul>
          </div>

          <!-- Admin confirm form -->
          <form @submit.prevent="adminConfirmSpecs" class="space-y-4 border-t pt-4">
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label for="apkg-type" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.packaging_type') }}</label>
                <input id="apkg-type" v-model="adminConfirmForm.packagingType" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="apkg-weight" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.packaging_weight') }}</label>
                <input id="apkg-weight" v-model.number="adminConfirmForm.packagingWeight" type="number" min="0" step="0.01" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="apkg-size" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.packaging_size') }}</label>
                <input id="apkg-size" v-model="adminConfirmForm.packagingSize" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="aquality-std" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.quality_standard') }}</label>
                <input id="aquality-std" v-model="adminConfirmForm.qualityStandard" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
            </div>
            <div>
              <label for="aconfirm-notes" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.admin_notes') }}</label>
              <textarea id="aconfirm-notes" v-model="adminConfirmForm.notes" rows="2" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div v-if="adminConfirmMessage" class="text-sm" :class="adminConfirmError ? 'text-red-600' : 'text-green-600'">
              {{ adminConfirmMessage }}
            </div>
            <div class="flex justify-end">
              <button type="submit" :disabled="adminConfirming" class="inline-flex justify-center rounded-md border border-transparent bg-emerald-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-emerald-700 disabled:opacity-50">
                {{ adminConfirming ? t('admin.inquiryDetail.confirming_specs') : t('admin.inquiryDetail.confirm_specs_button') }}
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- Convert to Order -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.inquiryDetail.convert_to_order') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <p class="text-sm text-gray-600 mb-4">{{ t('admin.inquiryDetail.convert_description') }}</p>
          <form @submit.prevent="convertToOrder" class="space-y-4">
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label for="inquiry-street" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.shipping_street') }}</label>
                <input id="inquiry-street" v-model="convertForm.street" name="street" type="text" autocomplete="street-address" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="inquiry-city" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.shipping_city') }}</label>
                <input id="inquiry-city" v-model="convertForm.city" name="city" type="text" autocomplete="address-level2" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="inquiry-state" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.shipping_state') }}</label>
                <input id="inquiry-state" v-model="convertForm.state" name="state" type="text" autocomplete="address-level1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="inquiry-zipCode" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.shipping_zip') }}</label>
                <input id="inquiry-zipCode" v-model="convertForm.zipCode" name="zipCode" type="text" autocomplete="postal-code" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="inquiry-country" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.shipping_country') }}</label>
                <input id="inquiry-country" v-model="convertForm.country" name="country" type="text" autocomplete="country-name" :placeholder="inquiry.targetCountry || ''" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
            </div>
            <div v-if="convertMessage" class="text-sm" :class="convertError ? 'text-red-600' : 'text-green-600'">
              {{ convertMessage }}
            </div>
            <div class="flex justify-end">
              <button type="submit" :disabled="converting" class="inline-flex justify-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700 disabled:opacity-50">
                {{ converting ? t('admin.inquiryDetail.converting') : t('admin.inquiryDetail.convert_button') }}
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- Activity Log -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.inquiryDetail.activity_log') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <div v-if="!inquiry.statusHistory || inquiry.statusHistory.length === 0" class="text-center text-gray-500">
            {{ t('admin.inquiryDetail.no_activity') }}
          </div>
          <ul v-else class="border-l-2 border-gray-200 ml-3 space-y-4">
            <li v-for="(entry, idx) in inquiry.statusHistory" :key="idx" class="ml-4">
              <div class="flex items-start">
                <div class="flex-shrink-0 w-2 h-2 mt-2 rounded-full bg-gray-400"></div>
                <div class="ml-3">
                  <p class="text-sm text-gray-900">{{ enumLabel('inquiry_status', entry.status) }} {{ entry.note ? `- ${entry.note}` : '' }}</p>
                  <p class="text-xs text-gray-500">{{ formatDate(entry.timestamp, { dateStyle: 'medium', timeStyle: 'short' }) }}</p>
                </div>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive } from 'vue'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const route = useRoute()
const { token } = useAuth()
const { t } = useI18n()
const localePath = useLocalePath()
const { currencyOrDefault: cur, cell, enumLabel, formatNumber, formatDate } = useDisplay()
const config = useRuntimeConfig()
const baseURL = config.public.apiBase || '/api/v1'

const inquiry = ref<any>(null)
const pending = ref(true)
const error = ref('')

const submittingQuote = ref(false)
const quoteMessage = ref('')
const quoteError = ref(false)

const converting = ref(false)
const convertMessage = ref('')
const convertError = ref(false)

const adminConfirming = ref(false)
const adminConfirmMessage = ref('')
const adminConfirmError = ref(false)
const adminConfirmForm = reactive({
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

const quoteForm = reactive({
  quotedAmount: 0,
  validUntil: '',
  customerNotes: ''
})

const convertForm = reactive({
  street: '',
  city: '',
  state: '',
  zipCode: '',
  country: ''
})

const fetchInquiry = async () => {
  pending.value = true
  error.value = ''
  try {
    inquiry.value = await $fetch<any>(`${baseURL}/admin/inquiries/${route.params.id}`, {
      headers: { Authorization: `Bearer ${token.value}` }
    })
    quoteForm.quotedAmount = inquiry.value.quotedAmount || 0
    quoteForm.validUntil = inquiry.value.validUntil ? new Date(inquiry.value.validUntil).toISOString().slice(0, 16) : ''
    quoteForm.customerNotes = inquiry.value.customerNotes || ''
  } catch (err: any) {
    error.value = err?.data?.message || err.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const submitQuote = async () => {
  submittingQuote.value = true
  quoteMessage.value = ''
  quoteError.value = false

  try {
    await $fetch(`${baseURL}/admin/inquiries/${route.params.id}/quote`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token.value}` },
      body: {
        quotedAmount: quoteForm.quotedAmount,
        validUntil: quoteForm.validUntil ? new Date(quoteForm.validUntil).toISOString() : '',
        customerNotes: quoteForm.customerNotes
      }
    })
    quoteMessage.value = t('admin.inquiryDetail.quote_submitted')
    await fetchInquiry()
  } catch (err: any) {
    quoteError.value = true
    quoteMessage.value = err?.data?.message || err.message || t('errors.api.quote_failed')
  } finally {
    submittingQuote.value = false
  }
}

const convertToOrder = async () => {
  converting.value = true
  convertMessage.value = ''
  convertError.value = false

  try {
    const result = await $fetch<any>(`${baseURL}/admin/inquiries/${route.params.id}/convert-to-order`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token.value}` },
      body: {
        shippingAddress: {
          street: convertForm.street,
          city: convertForm.city,
          state: convertForm.state,
          zipCode: convertForm.zipCode,
          country: convertForm.country
        }
      }
    })
    convertMessage.value = t('admin.inquiryDetail.converted_success') + (result.orderId ? ` Order ID: ${result.orderId}` : '')
    await fetchInquiry()
  } catch (err: any) {
    convertError.value = true
    convertMessage.value = err?.data?.message || err.message || t('errors.api.inquiry_convert_failed')
  } finally {
    converting.value = false
  }
}

const fetchNegotiationOffers = async () => {
  try {
    const data = await $fetch<any>(`${baseURL}/admin/inquiries/${route.params.id}/negotiations`, {
      headers: { Authorization: `Bearer ${token.value}` }
    })
    negotiationOffers.value = data?.data || data || []
  } catch (err: any) {
    console.error('Failed to fetch negotiation offers:', err)
  }
}

const submitOffer = async () => {
  if (!negotiationForm.totalAmount || negotiationForm.totalAmount <= 0) {
    negotiationMessage.value = t('admin.inquiryDetail.offer_amount_required')
    negotiationError.value = true
    return
  }
  submittingOffer.value = true
  negotiationMessage.value = ''
  negotiationError.value = false
  try {
    const body: any = { ...negotiationForm }
    if (body.validUntil) body.validUntil = new Date(body.validUntil).toISOString()
    await $fetch(`${baseURL}/admin/inquiries/${route.params.id}/negotiations`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token.value}` },
      body
    })
    negotiationMessage.value = t('admin.inquiryDetail.offer_created')
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
    await $fetch(`${baseURL}/admin/inquiries/${route.params.id}/negotiations/${offerId}/accept`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token.value}` }
    })
    await fetchNegotiationOffers()
  } catch (err: any) {
    alert(err?.data?.message || err.message || t('errors.api.save_failed'))
  }
}

const rejectOffer = async (offerId: string) => {
  try {
    await $fetch(`${baseURL}/admin/inquiries/${route.params.id}/negotiations/${offerId}/reject`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token.value}` }
    })
    await fetchNegotiationOffers()
  } catch (err: any) {
    alert(err?.data?.message || err.message || t('errors.api.save_failed'))
  }
}
const adminConfirmSpecs = async () => {
  adminConfirming.value = true
  adminConfirmMessage.value = ''
  adminConfirmError.value = false
  try {
    const result = await $fetch<any>(`${baseURL}/admin/inquiries/${route.params.id}/confirm`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${token.value}` },
      body: {
        packagingType: adminConfirmForm.packagingType,
        packagingWeight: adminConfirmForm.packagingWeight,
        packagingSize: adminConfirmForm.packagingSize,
        qualityStandard: adminConfirmForm.qualityStandard,
        notes: adminConfirmForm.notes
      }
    })
    adminConfirmMessage.value = result.message || t('admin.inquiryDetail.confirm_specs_success')
    await fetchInquiry()
  } catch (err: any) {
    adminConfirmError.value = true
    adminConfirmMessage.value = err?.data?.message || err.message || t('admin.inquiryDetail.confirm_specs_error')
  } finally {
    adminConfirming.value = false
  }
}

const statusClass = (status: string) => {
  if (status === 'pending') return 'bg-yellow-100 text-yellow-800'
  if (status === 'contacted' || status === 'quoted') return 'bg-orange-100 text-orange-800'
  if (status === 'negotiating') return 'bg-amber-100 text-amber-800'
  if (status === 'won') return 'bg-green-100 text-green-800'
  if (status === 'lost' || status === 'closed') return 'bg-gray-100 text-gray-800'
  return 'bg-gray-100 text-gray-800'
}

const priorityClass = (priority: string) => {
  if (priority === 'high' || priority === 'urgent') return 'bg-red-100 text-red-800'
  if (priority === 'normal') return 'bg-orange-100 text-orange-800'
  return 'bg-gray-100 text-gray-800'
}

onMounted(() => { fetchInquiry(); fetchNegotiationOffers() })
</script>
