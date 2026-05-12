<template>
  <div>
    <div class="mb-6">
      <NuxtLink :to="localePath('/admin/users')" class="flex items-center text-sm font-medium text-orange-600 hover:text-orange-500">
        <Icon name="heroicons:arrow-left" class="mr-1 h-4 w-4" />
        {{ t('admin.user_detail.back') }}
      </NuxtLink>
    </div>

    <div v-if="pending" class="text-center py-10">{{ t('admin.user_detail.loading') }}</div>
    <div v-else-if="error" class="text-center text-red-600 py-10">{{ error }}</div>
    <div v-else-if="user" class="bg-white shadow sm:rounded-lg overflow-hidden">
      <div class="px-4 py-5 sm:px-6 flex justify-between items-center bg-gray-50">
        <div>
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.user_detail.title') }}</h3>
          <p class="mt-1 max-w-2xl text-sm text-gray-500">{{ t('admin.user_detail.subtitle') }}</p>
        </div>
        <div>
          <span :class="[
            user.status === 'active' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800',
            'inline-flex rounded-full px-3 py-1 text-sm font-semibold leading-5'
          ]">
            {{ user.status }}
          </span>
        </div>
      </div>
      <div class="border-t border-gray-200 px-4 py-5 sm:px-6">
        <dl class="grid grid-cols-1 gap-x-4 gap-y-8 sm:grid-cols-2">
          <div class="sm:col-span-1">
            <dt class="text-sm font-medium text-gray-500">{{ t('admin.user_detail.full_name') }}</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ user.firstName }} {{ user.lastName }}</dd>
          </div>
          <div class="sm:col-span-1">
            <dt class="text-sm font-medium text-gray-500">{{ t('admin.user_detail.email') }}</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ user.email }}</dd>
          </div>
          <div class="sm:col-span-1">
            <dt class="text-sm font-medium text-gray-500">{{ t('admin.user_detail.company') }}</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ user.company ? user.company : t('display.na') }}</dd>
          </div>
          <div class="sm:col-span-1">
            <dt class="text-sm font-medium text-gray-500">{{ t('admin.user_detail.role') }}</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ displayRole }}</dd>
          </div>
          
          <div class="sm:col-span-2 mt-6">
            <h4 class="text-md font-medium text-gray-900 mb-4 border-b pb-2">{{ t('admin.user_detail.admin_actions') }}</h4>
            
            <form @submit.prevent="updateStatus" class="flex items-end gap-4">
              <div>
                <label for="status" class="block text-sm font-medium text-gray-700">{{ t('admin.user_detail.account_status') }}</label>
                <select id="status" v-model="statusInput" class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-orange-500 focus:border-orange-500 sm:text-sm rounded-md">
                  <option value="active">{{ t('admin.user_detail.status_active') }}</option>
                  <option value="suspended">{{ t('admin.user_detail.status_suspended') }}</option>
                  <option value="inactive">{{ t('admin.user_detail.status_inactive') }}</option>
                </select>
              </div>
              <button type="submit" :disabled="updating" class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-orange-600 hover:bg-orange-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-orange-500 disabled:opacity-50">
                <span v-if="updating">{{ t('admin.user_detail.updating') }}</span>
                <span v-else>{{ t('admin.user_detail.update_status') }}</span>
              </button>
            </form>
            
            <div v-if="updateMessage" class="mt-2 text-sm" :class="updateError ? 'text-red-600' : 'text-green-600'">
              {{ updateMessage }}
            </div>
          </div>

          <div class="sm:col-span-2 mt-6">
            <h4 class="text-md font-medium text-gray-900 mb-4 border-b pb-2">{{ t('admin.user_detail.edit_profile') }}</h4>
            <form @submit.prevent="updateProfile" class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label for="user-firstName" class="block text-sm font-medium text-gray-700">{{ t('admin.user_detail.first_name') }}</label>
                <input id="user-firstName" v-model="profileInput.firstName" name="firstName" type="text" autocomplete="given-name" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-orange-500" />
              </div>
              <div>
                <label for="user-lastName" class="block text-sm font-medium text-gray-700">{{ t('admin.user_detail.last_name') }}</label>
                <input id="user-lastName" v-model="profileInput.lastName" name="lastName" type="text" autocomplete="family-name" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-orange-500" />
              </div>
              <div>
                <label for="user-company" class="block text-sm font-medium text-gray-700">{{ t('admin.user_detail.company') }}</label>
                <input id="user-company" v-model="profileInput.company" name="company" type="text" autocomplete="organization" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-orange-500" />
              </div>
              <div>
                <label for="user-phone" class="block text-sm font-medium text-gray-700">{{ t('admin.user_detail.phone') }}</label>
                <input id="user-phone" v-model="profileInput.phone" name="phone" type="text" autocomplete="tel" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-orange-500" />
              </div>
              <div class="sm:col-span-2">
                <button type="submit" :disabled="savingProfile" class="inline-flex justify-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700 disabled:opacity-50">
                  {{ savingProfile ? t('admin.user_detail.saving') : t('admin.user_detail.save_profile') }}
                </button>
              </div>
            </form>
          </div>

          <div class="sm:col-span-2 mt-6">
            <h4 class="text-md font-medium text-gray-900 mb-4 border-b pb-2">{{ t('admin.user_detail.role_management') }}</h4>
            <form @submit.prevent="updateRole" class="flex flex-wrap items-end gap-4">
              <div>
                <label for="user-role" class="block text-sm font-medium text-gray-700">{{ t('admin.user_detail.role') }}</label>
                <select id="user-role" name="role" v-model="roleInput" class="mt-1 block rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-orange-500">
                  <option value="customer">{{ t('admin.user_detail.role_option_customer') }}</option>
                  <option value="admin">{{ t('admin.user_detail.role_option_admin') }}</option>
                  <option value="superadmin">{{ t('admin.user_detail.role_option_superadmin') }}</option>
                </select>
              </div>
              <button type="submit" :disabled="updatingRole" class="inline-flex justify-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700 disabled:opacity-50">
                {{ updatingRole ? t('admin.user_detail.updating_role') : t('admin.user_detail.update_role') }}
              </button>
            </form>
          </div>

          <div class="sm:col-span-2 mt-6">
            <h4 class="text-md font-medium text-gray-900 mb-4 border-b pb-2">{{ t('admin.user_detail.danger_zone') }}</h4>
            <button type="button" @click="deleteUser" :disabled="deletingUser" class="inline-flex justify-center rounded-md border border-transparent bg-red-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-red-700 disabled:opacity-50">
              {{ deletingUser ? t('admin.user_detail.deleting') : t('admin.user_detail.delete_user') }}
            </button>
          </div>
        </dl>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted, computed } from 'vue'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const route = useRoute()
const { token, user: authUser } = useAuth()
const { t, te } = useI18n()
const localePath = useLocalePath()

// Guard: only superadmin can access user management
if (authUser.value?.role !== 'superadmin') {
  navigateTo(localePath('/admin'))
}

const config = useRuntimeConfig()
const baseURL = config.public.apiBase || '/api/v1'
const id = route.params.id as string

const user = ref<any>(null)
const pending = ref(true)
const error = ref('')

const displayRole = computed(() => {
  if (!user.value) return t('roles.customer')
  const raw = String(user.value.role?.name || user.value.role || 'customer').toLowerCase()
  const key = `roles.${raw}`
  return te(key) ? t(key) : t('roles.customer')
})

const statusInput = ref('active')
const updating = ref(false)
const updateMessage = ref('')
const updateError = ref(false)
const roleInput = ref('customer')
const updatingRole = ref(false)
const deletingUser = ref(false)
const savingProfile = ref(false)
const profileInput = reactive({
  firstName: '',
  lastName: '',
  company: '',
  phone: ''
})

const fetchUser = async () => {
  pending.value = true
  
  try {
    const res = await $fetch<any>(`${baseURL}/admin/users/${id}`, {
      headers: { Authorization: `Bearer ${token.value}` }
    })
    user.value = res
    statusInput.value = res.status
    roleInput.value = res.role?.name || res.role || 'customer'
    profileInput.firstName = res.firstName || ''
    profileInput.lastName = res.lastName || ''
    profileInput.company = res.company || ''
    profileInput.phone = res.phone || ''
  } catch (err: any) {
    error.value = err.message ? err.message : t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const updateStatus = async () => {
  updating.value = true
  updateMessage.value = ''
  updateError.value = false
  
  try {
    await $fetch(`${baseURL}/admin/users/${id}/status`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${token.value}` },
      body: { status: statusInput.value }
    })
    user.value.status = statusInput.value
    updateMessage.value = t('admin.user_detail.status_updated')
    
    setTimeout(() => { updateMessage.value = '' }, 3000)
  } catch (err: any) {
    updateError.value = true
    updateMessage.value = err?.data?.message || err.message || t('errors.api.status_failed')
  } finally {
    updating.value = false
  }
}

const updateProfile = async () => {
  savingProfile.value = true
  updateMessage.value = ''
  updateError.value = false

  try {
    const res = await $fetch<any>(`${baseURL}/admin/users/${id}`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${token.value}` },
      body: {
        firstName: profileInput.firstName,
        lastName: profileInput.lastName,
        company: profileInput.company,
        phone: profileInput.phone
      }
    })

    user.value = res
    updateMessage.value = t('admin.user_detail.profile_updated')
  } catch (err: any) {
    updateError.value = true
    updateMessage.value = err?.data?.message || err.message || t('errors.api.save_failed')
  } finally {
    savingProfile.value = false
  }
}

const updateRole = async () => {
  updatingRole.value = true
  updateMessage.value = ''
  updateError.value = false

  try {
    await $fetch<any>(`${baseURL}/admin/users/${id}/role`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${token.value}` },
      body: {
        roleName: roleInput.value
      }
    })

    if (user.value) {
      user.value.role = { ...(user.value.role || {}), name: roleInput.value }
    }
    updateMessage.value = t('admin.user_detail.role_updated')
  } catch (err: any) {
    updateError.value = true
    updateMessage.value = err?.data?.message || err.message || t('errors.api.save_failed')
  } finally {
    updatingRole.value = false
  }
}

const deleteUser = async () => {
  if (!confirm(t('admin.user_detail.confirm_delete'))) return

  deletingUser.value = true
  updateMessage.value = ''
  updateError.value = false

  try {
    await $fetch(`${baseURL}/admin/users/${id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${token.value}` }
    })
    await navigateTo(localePath('/admin/users'))
  } catch (err: any) {
    updateError.value = true
    updateMessage.value = err?.data?.message || err.message || t('errors.api.delete_failed')
  } finally {
    deletingUser.value = false
  }
}

onMounted(fetchUser)
</script>
