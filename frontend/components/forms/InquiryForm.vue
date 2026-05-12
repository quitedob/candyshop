<template>
  <form ref="formRef" class="inquiry-form" @submit.prevent="handleSubmit">
    <!-- Progress Steps (for multi-step form) -->
    <div v-if="isMultiStep" class="inquiry-form__steps">
      <div
        v-for="(step, index) in steps"
        :key="index"
        class="inquiry-form__step"
        :class="{ 'inquiry-form__step--active': currentStep >= index, 'inquiry-form__step--completed': currentStep > index }"
      >
        <span class="inquiry-form__step-number">{{ index + 1 }}</span>
        <span class="inquiry-form__step-label">{{ step.label }}</span>
      </div>
    </div>

    <!-- Error Alert -->
    <div v-if="form.errors._form && form.errors._form.length > 0" class="inquiry-form__alert inquiry-form__alert--error">
      <Icon name="lucide:alert-circle" size="20" />
      <span>{{ form.errors._form[0] }}</span>
    </div>

    <!-- Success Message -->
    <div v-if="form.isSubmitted" class="inquiry-form__success">
      <div class="inquiry-form__success-icon">
        <Icon name="lucide:check-circle" size="48" />
      </div>
      <h3>{{ $t('form.success').split('.')[0] }}</h3>
      <p>{{ $t('form.success') }}</p>
    </div>
    <div v-else class="inquiry-form__fields">
      <!-- Company Information -->
      <div class="inquiry-form__section">
        <h4 class="inquiry-form__section-title">{{ $t('contact.title') }}</h4>

        <InputText
          id="companyName"
          v-model="form.state.companyName"
          :label="$t('form.company_name')"
          :error="form.errors.companyName?.[0]"
          autocomplete="organization"
          required
          @blur="form.validateAndUpdate('companyName')"
        />

        <InputText
          id="contactPerson"
          v-model="form.state.contactPerson"
          :label="$t('form.contact_person')"
          :error="form.errors.contactPerson?.[0]"
          autocomplete="name"
          required
          @blur="form.validateAndUpdate('contactPerson')"
        />

        <InputText
          id="email"
          v-model="form.state.email"
          type="email"
          :label="$t('form.email')"
          :error="form.errors.email?.[0]"
          autocomplete="email"
          required
          @blur="form.validateAndUpdate('email')"
        />

        <InputText
          id="whatsapp"
          v-model="form.state.whatsapp"
          type="tel"
          :label="$t('form.whatsapp')"
          :error="form.errors.whatsapp?.[0]"
          autocomplete="tel"
          placeholder="+86 123 4567 890"
        >
          <template #hint>
            {{ $t('whatsapp.us') }}
          </template>
        </InputText>
      </div>

      <!-- Product Information -->
      <div class="inquiry-form__section">
        <h4 class="inquiry-form__section-title">{{ $t('nav.products') }}</h4>

        <InputSelect
          id="targetCountry"
          v-model="form.state.targetCountry"
          :label="$t('form.target_country')"
          :options="countryOptions"
          :error="form.errors.targetCountry?.[0]"
          autocomplete="country-name"
        />

        <InputSelect
          id="estimatedQuantity"
          v-model="form.state.estimatedQuantity"
          :label="$t('form.estimated_quantity')"
          :options="quantityOptions"
          :error="form.errors.estimatedQuantity?.[0]"
        />

        <InputText
          id="interestedProducts"
          v-model="productsInput"
          :label="$t('form.interested_products')"
          :placeholder="$t('product.inquire_now')"
        >
          <template #hint>
            {{ $t("form.interested_products_hint") }}
          </template>
        </InputText>

        <InputTextarea
          id="packagingRequirements"
          v-model="form.state.packagingRequirements"
          :label="$t('form.packaging_requirements')"
          :rows="2"
        />

        <InputTextarea
          id="flavorRequirements"
          v-model="form.state.flavorRequirements"
          :label="$t('form.flavor_requirements')"
          :rows="2"
        />
      </div>

      <!-- OEM & Delivery -->
      <div class="inquiry-form__section">
        <h4 class="inquiry-form__section-title">{{ $t('oem.title') }}</h4>

        <InputRadioGroup
          id="oemNeeded"
          name="oem_needed"
          v-model="form.state.oemNeeded"
          :label="$t('form.oem_needed')"
          :options="[
            { value: true, label: $t('form.yes') },
            { value: false, label: $t('form.no') }
          ]"
        />

        <InputText
          id="expectedDelivery"
          v-model="form.state.expectedDelivery"
          type="date"
          :label="$t('form.expected_delivery')"
        />
      </div>

      <!-- Additional Message -->
      <div class="inquiry-form__section">
        <h4 class="inquiry-form__section-title">{{ $t('form.message') }}</h4>

        <InputTextarea
          id="message"
          v-model="form.state.message"
          :label="$t('form.message')"
          :rows="4"
          :placeholder="t('form.message_placeholder')"
        />
      </div>

      <!-- File Upload -->
      <div class="inquiry-form__section">
        <InputFile
          id="files"
          :label="$t('form.upload_files')"
          :error="form.errors.files?.[0]"
          :max-files="5"
          :max-size="5"
          accept="image/jpeg,image/png,image/webp,application/pdf"
          @files-selected="form.handleFileUpload"
        />
        <div v-if="form.state.files.length > 0" class="inquiry-form__files">
          <div
            v-for="(file, index) in form.state.files"
            :key="index"
            class="inquiry-form__file"
          >
            <Icon name="lucide:file" size="16" />
            <span class="inquiry-form__file-name">{{ file.name }}</span>
            <button
              type="button"
              class="inquiry-form__file-remove"
              @click="form.removeFile(index)"
            >
              <Icon name="lucide:x" size="14" />
            </button>
          </div>
        </div>
      </div>

      <!-- Submit Button -->
      <div class="inquiry-form__actions">
        <Button
          type="submit"
          variant="highlight"
          size="lg"
          :loading="form.isSubmitting"
          block
        >
          {{ form.isSubmitting ? $t('form.submitting') : $t('form.submit') }}
        </Button>

        <p class="inquiry-form__privacy">
          <Icon name="lucide:shield-check" size="14" />
          {{ t('form.privacy_note') }}
        </p>
      </div>
    </div>
  </form>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useInquiry } from '~/composables/useInquiry'

interface Props {
  productSlug?: string
  productName?: string
  category?: string
  userId?: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  success: []
}>()

const { t } = useI18n()
const form = useInquiry({
  onSuccess: (response) => {
    emit('success')
  }
})

// Multi-step form state
const isMultiStep = ref(false)
const currentStep = ref(0)
const steps = [
  { label: t('contact.title') },
  { label: t('nav.products') },
  { label: t('form.oem_needed') },
  { label: t('form.message') }
]

// Products input (comma-separated)
const productsInput = ref('')

// Watch products input to update form state
watch(productsInput, (value) => {
  form.state.interestedProducts = value
    .split(',')
    .map(p => p.trim())
    .filter(p => p.length > 0)
})

// Options
const countryOptions = computed(() => [
  { value: 'US', label: t('form.options.countries.US') },
  { value: 'UK', label: t('form.options.countries.UK') },
  { value: 'DE', label: t('form.options.countries.DE') },
  { value: 'FR', label: t('form.options.countries.FR') },
  { value: 'CA', label: t('form.options.countries.CA') },
  { value: 'AU', label: t('form.options.countries.AU') },
  { value: 'JP', label: t('form.options.countries.JP') },
  { value: 'KR', label: t('form.options.countries.KR') },
  { value: 'SG', label: t('form.options.countries.SG') },
  { value: 'MY', label: t('form.options.countries.MY') },
  { value: 'TH', label: t('form.options.countries.TH') },
  { value: 'VN', label: t('form.options.countries.VN') },
  { value: 'ID', label: t('form.options.countries.ID') },
  { value: 'PH', label: t('form.options.countries.PH') },
  { value: 'OTHER', label: t('form.options.countries.OTHER') }
])

const quantityOptions = computed(() => [
  { value: '100-500', label: t('form.options.quantities.100-500') },
  { value: '500-1000', label: t('form.options.quantities.500-1000') },
  { value: '1000-5000', label: t('form.options.quantities.1000-5000') },
  { value: '5000-10000', label: t('form.options.quantities.5000-10000') },
  { value: '10000+', label: t('form.options.quantities.10000+') }
])

// Pre-fill product info if provided
if (props.productName) {
  productsInput.value = props.productName
  form.state.interestedProducts = [props.productName]
}
if (props.productSlug) {
  form.state.message = form.state.message
    ? form.state.message
    : `Product: ${props.productName || props.productSlug}${props.category ? ` (${props.category})` : ''}`
}

// Ensure the form never shows a stale submitted state from a previous navigation or SSR
onMounted(() => {
  form.isSubmitted = false
  // Associate user ID if logged in
  if (props.userId) {
    ;(form.state as any).userId = props.userId
  }
  // Re-apply pre-fill after mount (reset clears it)
  if (props.productName) {
    productsInput.value = props.productName
    form.state.interestedProducts = [props.productName]
  }
  if (props.productSlug) {
    form.state.message = `Product: ${props.productName || props.productSlug}${props.category ? ` (${props.category})` : ''}`
  }
})

// Handle form submission
const handleSubmit = async () => {
  const result = await form.submit()
  return result
}

// Expose form ref for parent access
const formRef = ref<HTMLFormElement>()
defineExpose({
  formRef,
  reset: form.reset,
  validate: form.validate
})
</script>

<style scoped>
.inquiry-form {
  width: 100%;
}

/* Steps */
.inquiry-form__steps {
  display: flex;
  justify-content: space-between;
  margin-bottom: var(--spacing-xl);
  padding-bottom: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-light);
}

.inquiry-form__step {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.inquiry-form__step-number {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-text-light);
  background-color: var(--color-bg-alt);
  border: 2px solid var(--color-border);
  border-radius: var(--radius-full);
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
}

.inquiry-form__step--active .inquiry-form__step-number {
  color: var(--color-primary);
  background-color: var(--color-accent);
  border-color: var(--color-accent);
}

.inquiry-form__step--completed .inquiry-form__step-number {
  color: white;
  background-color: var(--color-primary);
  border-color: var(--color-primary);
}

.inquiry-form__step-label {
  display: none;
  font-size: var(--text-sm);
  font-weight: 500;
}

@media (min-width: 640px) {
  .inquiry-form__step-label {
    display: inline;
  }
}

/* Sections */
.inquiry-form__section {
  margin-bottom: var(--spacing-xl);
}

.inquiry-form__section-title {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-primary);
  margin-bottom: var(--spacing-md);
}

/* Alert */
.inquiry-form__alert {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-md);
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-lg);
}

.inquiry-form__alert--error {
  background-color: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
}

/* Success */
.inquiry-form__success {
  text-align: center;
  padding: var(--spacing-3xl) var(--spacing-lg);
}

/* Success pop animation */
.success-pop-enter-active {
  animation: successPop 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275) forwards;
}

@keyframes successPop {
  0% {
    opacity: 0;
    transform: scale(0.8);
  }
  60% {
    opacity: 1;
    transform: scale(1.05);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
}

.inquiry-form__success-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  height: 80px;
  margin-bottom: var(--spacing-lg);
  color: var(--color-success);
  background-color: #f0fdf4;
  border-radius: var(--radius-full);
}

.inquiry-form__success h3 {
  margin-bottom: var(--spacing-sm);
}

.inquiry-form__success p {
  color: var(--color-text-light);
}

/* Files */
.inquiry-form__files {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  margin-top: var(--spacing-md);
}

.inquiry-form__file {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-md);
}

.inquiry-form__file-name {
  flex: 1;
  font-size: var(--text-sm);
  color: var(--color-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.inquiry-form__file-remove {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  color: var(--color-text-light);
  background-color: transparent;
  border-radius: var(--radius-sm);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.inquiry-form__file-remove:hover {
  color: var(--color-error);
  background-color: #fef2f2;
}

/* Actions */
.inquiry-form__actions {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
  margin-top: var(--spacing-2xl);
}

.inquiry-form__privacy {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-xs);
  font-size: var(--text-xs);
  color: var(--color-text-light);
  margin: 0;
}

@media (min-width: 640px) {
  .inquiry-form {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: var(--spacing-xl);
  }

  .inquiry-form__section:first-child {
    grid-column: 1;
  }

  .inquiry-form__section:nth-child(2) {
    grid-column: 2;
  }

  .inquiry-form__section:nth-child(3) {
    grid-column: 1 / -1;
  }

  .inquiry-form__actions {
    grid-column: 1 / -1;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
  }

  .inquiry-form__privacy {
    margin-bottom: 0;
  }
}
</style>
