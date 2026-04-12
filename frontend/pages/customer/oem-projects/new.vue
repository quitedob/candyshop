<template>
  <div>
    <div class="mb-6">
      <NuxtLink to="/customer/oem-projects" class="flex items-center text-sm font-medium text-blue-600 hover:text-blue-500">
        <Icon name="heroicons:arrow-left" class="mr-1 h-4 w-4" />
        {{ t('customer.oemProjects.back') }}
      </NuxtLink>
    </div>

    <div class="max-w-2xl">
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.oemProjects.new_project') }}</h3>
          <p class="mt-1 max-w-2xl text-sm text-gray-500">{{ t('customer.oemProjects.new_project_subtitle') }}</p>
        </div>

        <form @submit.prevent="submitProject" class="px-4 py-5 sm:p-6 space-y-6">
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.oemProjects.select_product') || 'Select Product' }} *</label>
            <select v-model="form.productId" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
              <option value="">{{ t('customer.oemProjects.select_product_placeholder') || 'Choose a product...' }}</option>
              <option v-for="p in availableProducts" :key="p.id" :value="p.id">{{ p.name }} ({{ p.category }})</option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.oemProjects.product_name') }}</label>
            <input v-model="form.productName" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :placeholder="t('customer.oemProjects.product_name_placeholder') || 'Or specify a custom product name'" />
          </div>

          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('customer.oemProjects.flavor') }}</label>
              <input v-model="form.flavor" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :placeholder="t('customer.oemProjects.flavor_placeholder')" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('customer.oemProjects.shape') }}</label>
              <input v-model="form.shape" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :placeholder="t('customer.oemProjects.shape_placeholder')" />
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.oemProjects.packaging') }}</label>
            <input v-model="form.packaging" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :placeholder="t('customer.oemProjects.packaging_placeholder')" />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.oemProjects.target_market') }}</label>
            <input v-model="form.targetMarket" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :placeholder="t('customer.oemProjects.target_market_placeholder')" />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.oemProjects.certifications') }}</label>
            <input v-model="form.certificationsInput" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :placeholder="t('customer.oemProjects.certifications_placeholder')" />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.oemProjects.moq') }}</label>
            <input v-model.number="form.moq" type="number" min="1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :placeholder="t('customer.oemProjects.moq_placeholder')" />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.oemProjects.requirements') }}</label>
            <textarea v-model="form.requirements" rows="4" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :placeholder="t('customer.oemProjects.requirements_placeholder')"></textarea>
          </div>

          <div v-if="errorMessage" class="text-sm text-red-600">{{ errorMessage }}</div>
          <div v-if="successMessage" class="text-sm text-green-600">{{ successMessage }}</div>

          <div class="flex justify-end gap-3">
            <NuxtLink to="/customer/oem-projects" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">
              {{ t('customer.oemProjects.cancel') }}
            </NuxtLink>
            <button type="submit" :disabled="submitting" class="rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 disabled:opacity-50">
              {{ submitting ? t('customer.oemProjects.submitting') : t('customer.oemProjects.submit') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const { t } = useI18n()
const api = useApi()

const submitting = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const availableProducts = ref<any[]>([])

onMounted(async () => {
  try {
    const result = await api.getProducts({ limit: 100 })
    availableProducts.value = result.data || []
  } catch (e) { console.error('Failed to load products', e) }
})

const form = reactive({
  productId: '', productName: '', flavor: '', shape: '', packaging: '',
  targetMarket: '', certificationsInput: '', moq: 0, requirements: ''
})

const parseCSV = (value: string) => value.split(',').map(v => v.trim()).filter(Boolean)

const submitProject = async () => {
  submitting.value = true; errorMessage.value = ''; successMessage.value = ''
  try {
    await api.post('/user/oem-projects', {
      productId: form.productId || null, productName: form.productName,
      flavor: form.flavor, shape: form.shape, packaging: form.packaging,
      targetMarket: form.targetMarket, certifications: parseCSV(form.certificationsInput),
      moq: form.moq, requirements: form.requirements
    })
    successMessage.value = t('customer.oemProjects.project_created')
    setTimeout(() => navigateTo('/customer/oem-projects'), 1500)
  } catch (err: any) {
    errorMessage.value = err?.message || 'Failed to create project'
  } finally { submitting.value = false }
}
</script>
