<template>
  <div class="space-y-6">
    <div class="bg-white shadow overflow-hidden sm:rounded-lg">
      <div class="px-4 py-5 sm:px-6 flex justify-between items-center">
        <div>
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.profile.title') }}</h3>
          <p class="mt-1 max-w-2xl text-sm text-gray-500">{{ t('customer.profile.subtitle') }}</p>
        </div>
        <button v-if="!isEditing" @click="startEditing" class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-orange-600 hover:bg-orange-700">
          {{ t('customer.profile.edit_profile') }}
        </button>
      </div>
      <div v-if="!isEditing" class="border-t border-gray-200 px-4 py-5 sm:p-0">
        <dl class="sm:divide-y sm:divide-gray-200">
          <div class="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.profile.full_name') }}</dt>
            <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{{ user?.firstName }} {{ user?.lastName }}</dd>
          </div>
          <div class="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.profile.role') }}</dt>
            <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2 capitalize">{{ roleLabel }}</dd>
          </div>
          <div class="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.profile.email') }}</dt>
            <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{{ user?.email }}</dd>
          </div>
          <div class="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.profile.company') }}</dt>
            <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{{ user?.company || t('customer.profile.not_provided') }}</dd>
          </div>
          <div class="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.profile.phone') }}</dt>
            <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{{ user?.phone || t('customer.profile.not_provided') }}</dd>
          </div>
        </dl>
      </div>
      <div v-else class="border-t border-gray-200 px-4 py-5">
        <form @submit.prevent="saveProfile" class="space-y-4 max-w-2xl">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label for="profile-firstName" class="block text-sm font-medium text-gray-700">{{ t('customer.profile.first_name') }}</label>
              <input id="profile-firstName" v-model="editForm.firstName" name="firstName" type="text" autocomplete="given-name" class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-orange-500 focus:border-orange-500 sm:text-sm" required>
            </div>
            <div>
              <label for="profile-lastName" class="block text-sm font-medium text-gray-700">{{ t('customer.profile.last_name') }}</label>
              <input id="profile-lastName" v-model="editForm.lastName" name="lastName" type="text" autocomplete="family-name" class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-orange-500 focus:border-orange-500 sm:text-sm" required>
            </div>
          </div>
          <div>
            <label for="profile-company" class="block text-sm font-medium text-gray-700">{{ t('customer.profile.company') }}</label>
            <input id="profile-company" v-model="editForm.company" name="company" type="text" autocomplete="organization" class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-orange-500 focus:border-orange-500 sm:text-sm">
          </div>
          <div>
            <label for="profile-phone" class="block text-sm font-medium text-gray-700">{{ t('customer.profile.phone') }}</label>
            <input id="profile-phone" v-model="editForm.phone" name="phone" type="text" autocomplete="tel" class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-orange-500 focus:border-orange-500 sm:text-sm">
          </div>
          <div class="flex justify-end space-x-3 mt-4">
            <button type="button" @click="cancelEditing" class="bg-white py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none">
              {{ t('customer.profile.cancel') }}
            </button>
            <button type="submit" :disabled="isSaving" class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-orange-600 hover:bg-orange-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-orange-500 disabled:opacity-50">
              {{ isSaving ? t('customer.profile.saving') : t('customer.profile.save_changes') }}
            </button>
          </div>
          <p v-if="profileMessage" :class="profileError ? 'text-red-600' : 'text-green-600'" class="text-sm mt-2">{{ profileMessage }}</p>
        </form>
      </div>
    </div>
    <div class="bg-white shadow overflow-hidden sm:rounded-lg">
      <div class="px-4 py-5 sm:px-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.profile.change_password') }}</h3>
      </div>
      <div class="border-t border-gray-200 px-4 py-5">
        <form @submit.prevent="changePassword" class="space-y-4 max-w-md">
          <div>
            <label for="profile-currentPassword" class="block text-sm font-medium text-gray-700">{{ t('customer.profile.current_password') }}</label>
            <input id="profile-currentPassword" v-model="pwdForm.currentPassword" name="currentPassword" type="password" autocomplete="current-password" class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-orange-500 focus:border-orange-500 sm:text-sm" required>
          </div>
          <div>
            <label for="profile-newPassword" class="block text-sm font-medium text-gray-700">{{ t('customer.profile.new_password') }}</label>
            <input id="profile-newPassword" v-model="pwdForm.newPassword" name="newPassword" type="password" autocomplete="new-password" class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-orange-500 focus:border-orange-500 sm:text-sm" required minlength="8">
          </div>
          <div>
            <label for="profile-confirmPassword" class="block text-sm font-medium text-gray-700">{{ t('customer.profile.confirm_password') }}</label>
            <input id="profile-confirmPassword" v-model="pwdForm.confirmPassword" name="confirmPassword" type="password" autocomplete="new-password" class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-orange-500 focus:border-orange-500 sm:text-sm" required minlength="8">
            <p v-if="pwdForm.confirmPassword && pwdForm.newPassword !== pwdForm.confirmPassword" class="mt-1 text-xs text-red-600">
              {{ t('customer.profile.password_mismatch') }}
            </p>
          </div>
          <div>
            <button type="submit" :disabled="isChangingPwd" class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-orange-600 hover:bg-orange-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-orange-500 disabled:opacity-50">
              {{ isChangingPwd ? t('customer.profile.updating') : t('customer.profile.update_password') }}
            </button>
          </div>
          <p v-if="pwdMessage" :class="pwdError ? 'text-red-600' : 'text-green-600'" class="text-sm mt-2">{{ pwdMessage }}</p>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const { user } = useAuth()
const { t, te } = useI18n()
const api = useApi()

const isEditing = ref(false)
const isSaving = ref(false)
const profileMessage = ref('')
const profileError = ref(false)
const editForm = reactive({ firstName: '', lastName: '', company: '', phone: '' })

const roleLabel = computed(() => {
  const raw = String(user.value?.role || 'customer').toLowerCase()
  const key = `roles.${raw}`
  return te(key) ? t(key) : t('roles.customer')
})

const startEditing = () => {
  editForm.firstName = user.value?.firstName || ''
  editForm.lastName = user.value?.lastName || ''
  editForm.company = user.value?.company || ''
  editForm.phone = user.value?.phone || ''
  isEditing.value = true; profileMessage.value = ''
}
const cancelEditing = () => { isEditing.value = false }

const saveProfile = async () => {
  isSaving.value = true; profileMessage.value = ''; profileError.value = false
  try {
    await api.put('/user/profile', editForm)
    if (user.value) { user.value.firstName = editForm.firstName; user.value.lastName = editForm.lastName; user.value.company = editForm.company; user.value.phone = editForm.phone }
    profileMessage.value = t('customer.profile.profile_updated')
    setTimeout(() => { isEditing.value = false }, 1500)
  } catch (e: any) { profileError.value = true; profileMessage.value = e?.message || t('customer.profile.profile_update_failed') }
  finally { isSaving.value = false }
}

const isChangingPwd = ref(false)
const pwdMessage = ref('')
const pwdError = ref(false)
const pwdForm = reactive({ currentPassword: '', newPassword: '', confirmPassword: '' })

const changePassword = async () => {
  if (pwdForm.newPassword !== pwdForm.confirmPassword) {
    pwdError.value = true
    pwdMessage.value = t('customer.profile.password_mismatch')
    return
  }
  isChangingPwd.value = true; pwdMessage.value = ''; pwdError.value = false
  try {
    await api.post('/user/change-password', { currentPassword: pwdForm.currentPassword, newPassword: pwdForm.newPassword })
    pwdMessage.value = t('customer.profile.password_updated')
    pwdForm.currentPassword = ''; pwdForm.newPassword = ''; pwdForm.confirmPassword = ''
  } catch (e: any) { pwdError.value = true; pwdMessage.value = e?.message || t('customer.profile.password_update_failed') }
  finally { isChangingPwd.value = false }
}
</script>
