<template>
  <div>
    <div class="sm:flex sm:items-center">
      <div class="sm:flex-auto">
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.users.title') }}</h1>
        <p class="mt-2 text-sm text-gray-700">{{ t('admin.users.description') }}</p>
      </div>
      <div class="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
        <button @click="openCreateModal" type="button" class="inline-flex items-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700">
          {{ t('admin.users.new_user') }}
        </button>
      </div>
    </div>
    
    <div class="mt-8 flex flex-col">
      <div class="-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8">
        <div class="inline-block min-w-full py-2 align-middle md:px-6 lg:px-8">
          <div class="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
            <table class="min-w-full divide-y divide-gray-300">
              <thead class="bg-gray-50">
                <tr>
                  <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">{{ t('admin.users.col_name') }}</th>
                  <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.users.col_company') }}</th>
                  <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.users.col_role') }}</th>
                  <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.users.col_status') }}</th>
                  <th scope="col" class="relative py-3.5 pl-3 pr-4 sm:pr-6">
                    <span class="sr-only">{{ t('admin.users.edit') }}</span>
                  </th>
                  <th scope="col" class="relative py-3.5 pl-3 pr-4 sm:pr-6">
                    <span class="sr-only">{{ t('admin.users.delete') }}</span>
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200 bg-white">
                <tr v-if="pending">
                  <td colspan="6" class="py-4 text-center text-sm text-gray-500">{{ t('admin.users.loading') }}</td>
                </tr>
                <tr v-else-if="error">
                  <td colspan="6" class="py-4 text-center text-sm text-red-500">{{ error }}</td>
                </tr>
                <tr v-else v-for="user in users" :key="user.id">
                  <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm sm:pl-6">
                    <div class="flex items-center">
                      <div class="h-10 w-10 flex-shrink-0">
                        <div class="h-10 w-10 rounded-full bg-gray-200 flex items-center justify-center font-bold text-gray-600">
                          {{ user.firstName?.charAt(0) }}{{ user.lastName?.charAt(0) }}
                        </div>
                      </div>
                      <div class="ml-4">
                        <div class="font-medium text-gray-900">{{ user.firstName }} {{ user.lastName }}</div>
                        <div class="text-gray-500">{{ user.email }}</div>
                      </div>
                    </div>
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                    <div class="text-gray-900">{{ user.company }}</div>
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                    <span class="inline-flex rounded-full bg-orange-100 px-2 text-xs font-semibold leading-5 text-orange-800">{{ user.role?.name || user.role || t('admin.users.role_customer') }}</span>
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                    <span :class="[
                      user.status === 'active' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800',
                      'inline-flex rounded-full px-2 text-xs font-semibold leading-5'
                    ]">
                      {{ user.status }}
                    </span>
                  </td>
                  <td class="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                    <NuxtLink :to="localePath(`/admin/users/${user.id}`)" class="text-orange-600 hover:text-orange-900">{{ t('admin.users.edit') }}</NuxtLink>
                  </td>
                  <td class="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                    <button @click="removeUser(user.id)" class="text-red-600 hover:text-red-800">{{ t('admin.users.delete') }}</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          
          <!-- Pagination simple -->
          <div v-if="pagination" class="mt-4 flex items-center justify-between border-t border-gray-200 bg-white px-4 py-3 sm:px-6 shadow sm:rounded-lg">
            <div class="flex flex-1 justify-between sm:hidden">
              <button @click="prevPage" :disabled="page <= 1" class="relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50">{{ t('common.previous') }}</button>
              <button @click="nextPage" :disabled="page >= pagination.totalPages" class="relative ml-3 inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50">{{ t('common.next') }}</button>
            </div>
            <div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
              <div>
                <p class="text-sm text-gray-700">
                  {{ t('admin.users.showing', { from: ((page - 1) * 20) + 1, to: Math.min(page * 20, pagination.total), total: pagination.total }) }}
                </p>
              </div>
              <div>
                <nav class="isolate inline-flex -space-x-px rounded-md shadow-sm" aria-label="Pagination">
                  <button @click="prevPage" :disabled="page <= 1" class="relative inline-flex items-center rounded-l-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50">
                    <span class="sr-only">{{ t('common.previous') }}</span>
                    <Icon name="heroicons:chevron-left" class="h-5 w-5" />
                  </button>
                  <button @click="nextPage" :disabled="page >= pagination.totalPages" class="relative inline-flex items-center rounded-r-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50">
                    <span class="sr-only">{{ t('common.next') }}</span>
                    <Icon name="heroicons:chevron-right" class="h-5 w-5" />
                  </button>
                </nav>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showCreateModal" class="fixed inset-0 z-10 overflow-y-auto" role="dialog" aria-modal="true">
      <div class="flex min-h-screen items-end justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" @click="closeCreateModal"></div>
        <span class="hidden sm:inline-block sm:h-screen sm:align-middle" aria-hidden="true">&#8203;</span>
        <div class="inline-block w-full transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left align-bottom shadow-xl transition-all sm:my-8 sm:max-w-2xl sm:p-6 sm:align-middle">
          <h3 class="text-lg font-medium leading-6 text-gray-900">{{ t('admin.users.create_user') }}</h3>
          <form class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2" @submit.prevent="createUser">
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.users.first_name') }}</label>
              <input v-model="createForm.firstName" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.users.last_name') }}</label>
              <input v-model="createForm.lastName" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div class="sm:col-span-2">
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.users.email') }}</label>
              <input v-model="createForm.email" type="email" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div class="sm:col-span-2">
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.users.password') }}</label>
              <input v-model="createForm.password" type="password" required minlength="8" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.users.company') }}</label>
              <input v-model="createForm.company" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.users.phone') }}</label>
              <input v-model="createForm.phone" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.users.role') }}</label>
              <select v-model="createForm.roleName" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                <option value="customer">{{ t('admin.users.role_customer') }}</option>
                <option value="admin">{{ t('admin.users.role_admin') }}</option>
                <option value="superadmin">{{ t('admin.users.role_superadmin') }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.users.status') }}</label>
              <select v-model="createForm.status" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                <option value="active">{{ t('admin.users.status_active') }}</option>
                <option value="inactive">{{ t('admin.users.status_inactive') }}</option>
                <option value="suspended">{{ t('admin.users.status_suspended') }}</option>
              </select>
            </div>
            <div v-if="createError" class="sm:col-span-2 text-sm text-red-600">{{ createError }}</div>
            <div class="sm:col-span-2 mt-2 flex justify-end gap-3">
              <button type="button" @click="closeCreateModal" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700">{{ t('admin.users.cancel') }}</button>
              <button type="submit" :disabled="creating" class="rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white disabled:opacity-50">
                {{ creating ? t('admin.users.creating') : t('admin.users.create_user') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch, onMounted } from 'vue'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const { token, user: authUser } = useAuth()
const { t } = useI18n()
const localePath = useLocalePath()

// Guard: only superadmin can access user management
if (authUser.value?.role !== 'superadmin') {
  navigateTo(localePath('/admin'))
}
const api = useApi()

const users = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const showCreateModal = ref(false)
const creating = ref(false)
const createError = ref('')
const createForm = reactive({
  email: '', password: '', firstName: '', lastName: '',
  company: '', phone: '', roleName: 'customer', status: 'active'
})

const fetchUsers = async () => {
  pending.value = true; error.value = ''
  try {
    const res = await api.get<any>(`/admin/users?page=${page.value}&limit=20`)
    users.value = res.data; pagination.value = res.pagination
  } catch (err: any) { error.value = err?.message || t('errors.api.load_failed') }
  finally { pending.value = false }
}

const nextPage = () => { if (page.value < pagination.value.totalPages) page.value++ }
const prevPage = () => { if (page.value > 1) page.value-- }

const resetCreateForm = () => { Object.assign(createForm, { email: '', password: '', firstName: '', lastName: '', company: '', phone: '', roleName: 'customer', status: 'active' }) }
const openCreateModal = () => { createError.value = ''; resetCreateForm(); showCreateModal.value = true }
const closeCreateModal = () => { showCreateModal.value = false }

const createUser = async () => {
  creating.value = true; createError.value = ''
  try { await api.post('/admin/users', createForm); closeCreateModal(); await fetchUsers() }
  catch (err: any) { createError.value = err?.message || t('errors.api.user_create_failed') }
  finally { creating.value = false }
}

const removeUser = async (id: string) => {
  if (!confirm(t('admin.users.confirm_delete'))) return
  try { await api.del(`/admin/users/${id}`); await fetchUsers() }
  catch (err: any) { error.value = err?.message || t('errors.api.delete_failed') }
}

watch(page, fetchUsers)
onMounted(fetchUsers)
</script>
