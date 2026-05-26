<template>
  <div>
    <PageHeader :title="t('admin.users.title')" :description="t('admin.users.description')">
      <template #actions>
        <button
          type="button"
          class="inline-flex items-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700"
          @click="openCreateModal"
        >
          {{ t('admin.users.new_user') }}
        </button>
      </template>
    </PageHeader>

    <AdminTable
      :columns="columns"
      :rows="users"
      :loading="pending"
      :error="!!error"
      :error-message="error"
      :empty-text="t('admin.users.no_data')"
      @retry="fetchUsers"
    >
      <template #cell-name="{ row }">
        <div class="flex items-center">
          <div class="h-10 w-10 flex-shrink-0 rounded-full bg-gray-200 flex items-center justify-center font-bold text-gray-600">
            {{ row.firstName?.charAt(0) }}{{ row.lastName?.charAt(0) }}
          </div>
          <div class="ml-4">
            <div class="font-medium text-gray-900">{{ row.firstName }} {{ row.lastName }}</div>
            <div class="text-gray-500">{{ row.email }}</div>
          </div>
        </div>
      </template>

      <template #cell-role="{ row }">
        <span class="inline-flex rounded-full bg-orange-100 px-2 text-xs font-semibold leading-5 text-orange-800">
          {{ row.role?.name || row.role || t('admin.users.role_customer') }}
        </span>
      </template>

      <template #cell-status="{ row }">
        <StatusBadge :status="row.status" type="user" />
      </template>

      <template #cell-actions="{ row }">
        <NuxtLink :to="localePath(`/admin/users/${row.id}`)" class="text-orange-600 hover:text-orange-900 mr-3">
          {{ t('admin.users.edit') }}
        </NuxtLink>
        <button type="button" class="text-red-600 hover:text-red-800" @click="removeUser(row.id)">
          {{ t('admin.users.delete') }}
        </button>
      </template>

      <template #bottom>
        <div v-if="pagination" class="flex items-center justify-between">
          <p class="text-sm text-gray-700">
            {{ t('admin.users.showing', { from: ((page - 1) * 20) + 1, to: Math.min(page * 20, pagination.total), total: pagination.total }) }}
          </p>
          <nav class="isolate inline-flex -space-x-px rounded-md shadow-sm" :aria-label="t('pagination.nav_label')">
            <button
              type="button"
              :disabled="page <= 1"
              class="relative inline-flex items-center rounded-l-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50"
              @click="prevPage"
            >
              <span class="sr-only">{{ t('common.previous') }}</span>
              <Icon name="heroicons:chevron-left" class="h-5 w-5" aria-hidden="true" />
            </button>
            <button
              type="button"
              :disabled="page >= pagination.totalPages"
              class="relative inline-flex items-center rounded-r-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20 focus:outline-offset-0 disabled:opacity-50"
              @click="nextPage"
            >
              <span class="sr-only">{{ t('common.next') }}</span>
              <Icon name="heroicons:chevron-right" class="h-5 w-5" aria-hidden="true" />
            </button>
          </nav>
        </div>
      </template>
    </AdminTable>

    <!-- Create Modal -->
    <AdminModal :open="showCreateModal" :title="t('admin.users.create_user')" @close="closeCreateModal">
      <form class="grid grid-cols-1 gap-4 sm:grid-cols-2" @submit.prevent="createUser">
        <div>
          <label for="user-firstName" class="block text-sm font-medium text-gray-700">{{ t('admin.users.first_name') }}</label>
          <input id="user-firstName" v-model="createForm.firstName" name="firstName" autocomplete="given-name" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label for="user-lastName" class="block text-sm font-medium text-gray-700">{{ t('admin.users.last_name') }}</label>
          <input id="user-lastName" v-model="createForm.lastName" name="lastName" autocomplete="family-name" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div class="sm:col-span-2">
          <label for="user-email" class="block text-sm font-medium text-gray-700">{{ t('admin.users.email') }}</label>
          <input id="user-email" v-model="createForm.email" name="email" type="email" autocomplete="email" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div class="sm:col-span-2">
          <label for="user-password" class="block text-sm font-medium text-gray-700">{{ t('admin.users.password') }}</label>
          <input id="user-password" v-model="createForm.password" name="password" type="password" autocomplete="new-password" required minlength="8" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label for="user-company" class="block text-sm font-medium text-gray-700">{{ t('admin.users.company') }}</label>
          <input id="user-company" v-model="createForm.company" name="company" autocomplete="organization" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label for="user-phone" class="block text-sm font-medium text-gray-700">{{ t('admin.users.phone') }}</label>
          <input id="user-phone" v-model="createForm.phone" name="phone" autocomplete="tel" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label for="user-role" class="block text-sm font-medium text-gray-700">{{ t('admin.users.role') }}</label>
          <select id="user-role" v-model="createForm.roleName" name="roleName" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
            <option value="customer">{{ t('admin.users.role_customer') }}</option>
            <option value="admin">{{ t('admin.users.role_admin') }}</option>
            <option value="superadmin">{{ t('admin.users.role_superadmin') }}</option>
          </select>
        </div>
        <div>
          <label for="user-status" class="block text-sm font-medium text-gray-700">{{ t('admin.users.status') }}</label>
          <select id="user-status" v-model="createForm.status" name="status" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
            <option value="active">{{ t('admin.users.status_active') }}</option>
            <option value="inactive">{{ t('admin.users.status_inactive') }}</option>
            <option value="suspended">{{ t('admin.users.status_suspended') }}</option>
          </select>
        </div>
        <div v-if="createError" class="sm:col-span-2 text-sm text-red-600">{{ createError }}</div>
        <div class="sm:col-span-2 mt-2 flex justify-end gap-3">
          <button type="button" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700" @click="closeCreateModal">
            {{ t('admin.users.cancel') }}
          </button>
          <button type="submit" :disabled="creating" class="rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white disabled:opacity-50">
            {{ creating ? t('admin.users.creating') : t('admin.users.create_user') }}
          </button>
        </div>
      </form>
    </AdminModal>
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
const { enumLabel } = useDisplay()
const localePath = useLocalePath()

if (authUser.value?.role !== 'superadmin') {
  navigateTo(localePath('/admin'))
}
const api = useApi()

const columns = [
  { key: 'name', label: t('admin.users.col_name') },
  { key: 'company', label: t('admin.users.col_company') },
  { key: 'role', label: t('admin.users.col_role') },
  { key: 'status', label: t('admin.users.col_status') },
  { key: 'actions', label: '' },
]

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
