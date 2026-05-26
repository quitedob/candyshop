<template>
  <div>
    <PageHeader :title="t('admin.organizations.title')" :description="t('admin.organizations.description')">
      <template #actions>
        <button type="button" class="inline-flex items-center rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700" @click="openCreateOrg">
          {{ t('admin.organizations.create') }}
        </button>
      </template>
    </PageHeader>

    <div class="mb-4 border-b border-gray-200">
      <nav class="-mb-px flex gap-6">
        <button type="button" :class="tabClass('orgs')" @click="activeTab = 'orgs'">{{ t('admin.organizations.tab_orgs') }}</button>
        <button type="button" :class="tabClass('members')" @click="activeTab = 'members'">{{ t('admin.organizations.tab_members') }}</button>
      </nav>
    </div>

    <AdminTable
      v-if="activeTab === 'orgs'"
      :columns="orgColumns"
      :rows="organizations"
      :loading="pending"
      :error="!!error"
      :error-message="error"
      :empty-text="t('admin.organizations.no_data')"
      @retry="fetchOrganizations"
    >
      <template #cell-isActive="{ row }">
        <span :class="row.isActive !== false ? 'text-emerald-600' : 'text-gray-400'">{{ row.isActive !== false ? t('admin.status_options.active') : t('admin.status_options.inactive') }}</span>
      </template>
      <template #cell-actions="{ row }">
        <button type="button" class="text-orange-600 hover:text-orange-900" @click="openEditOrg(row)">{{ t('admin.organizations.edit') }}</button>
        <button type="button" class="ml-3 text-gray-600 hover:text-gray-900" @click="selectOrgForMembers(row)">{{ t('admin.organizations.members') }}</button>
      </template>
    </AdminTable>

    <div v-else class="space-y-4">
      <div class="flex flex-wrap items-end gap-3 rounded-lg border border-gray-200 bg-white p-4">
        <div>
          <label class="block text-xs font-medium text-gray-600">{{ t('admin.organizations.select_org') }}</label>
          <select v-model="selectedOrgId" class="mt-1 min-w-[200px] rounded-md border border-gray-300 px-3 py-2 text-sm" @change="fetchMembers">
            <option value="">{{ t('admin.organizations.select_org_placeholder') }}</option>
            <option v-for="org in organizations" :key="org.id" :value="org.id">{{ org.name }}</option>
          </select>
        </div>
        <button type="button" class="rounded-md bg-orange-600 px-4 py-2 text-sm text-white hover:bg-orange-700" :disabled="!selectedOrgId" @click="openAddMember">
          {{ t('admin.organizations.add_member') }}
        </button>
      </div>
      <AdminTable
        :columns="memberColumns"
        :rows="members"
        :loading="membersPending"
        :error="!!membersError"
        :error-message="membersError"
        :empty-text="t('admin.organizations.no_members')"
        @retry="fetchMembers"
      >
        <template #cell-actions="{ row }">
          <button type="button" class="text-red-600 hover:text-red-900" @click="removeMember(row.id)">{{ t('admin.organizations.remove') }}</button>
        </template>
      </AdminTable>
    </div>

    <AdminModal :open="showOrgModal" :title="editingOrgId ? t('admin.organizations.edit') : t('admin.organizations.create')" width="md" @close="closeOrgModal">
      <form class="space-y-4" @submit.prevent="saveOrg">
        <div>
          <label class="block text-sm font-medium text-gray-700">{{ t('admin.organizations.name') }}</label>
          <input v-model="orgForm.name" required class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('admin.organizations.credit_limit') }}</label>
            <input v-model.number="orgForm.creditLimit" type="number" step="0.01" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('admin.organizations.approval_threshold') }}</label>
            <input v-model.number="orgForm.approvalThreshold" type="number" step="0.01" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700">{{ t('admin.organizations.payment_terms') }}</label>
          <input v-model="orgForm.paymentTerms" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <label v-if="editingOrgId" class="inline-flex items-center gap-2 text-sm">
          <input v-model="orgForm.isActive" type="checkbox" class="rounded border-gray-300" />
          {{ t('admin.organizations.active') }}
        </label>
        <p v-if="formError" class="text-sm text-red-600">{{ formError }}</p>
        <div class="flex justify-end gap-3 pt-2">
          <button type="button" class="rounded-md border border-gray-300 px-4 py-2 text-sm" @click="closeOrgModal">{{ t('admin.organizations.cancel') }}</button>
          <button type="submit" :disabled="saving" class="rounded-md bg-orange-600 px-4 py-2 text-sm text-white disabled:opacity-50">{{ t('admin.organizations.save') }}</button>
        </div>
      </form>
    </AdminModal>

    <AdminModal :open="showMemberModal" :title="t('admin.organizations.add_member')" width="md" @close="showMemberModal = false">
      <form class="space-y-4" @submit.prevent="saveMember">
        <div>
          <label class="block text-sm font-medium text-gray-700">{{ t('admin.organizations.user_id') }}</label>
          <input v-model="memberForm.userId" required class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm font-mono" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700">{{ t('admin.organizations.role') }}</label>
          <select v-model="memberForm.role" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
            <option value="buyer">{{ t('admin.organizations.role_buyer') }}</option>
            <option value="approver">{{ t('admin.organizations.role_approver') }}</option>
            <option value="admin">{{ t('admin.organizations.role_admin') }}</option>
          </select>
        </div>
        <p v-if="memberError" class="text-sm text-red-600">{{ memberError }}</p>
        <div class="flex justify-end gap-3 pt-2">
          <button type="button" class="rounded-md border border-gray-300 px-4 py-2 text-sm" @click="showMemberModal = false">{{ t('admin.organizations.cancel') }}</button>
          <button type="submit" :disabled="memberSaving" class="rounded-md bg-orange-600 px-4 py-2 text-sm text-white disabled:opacity-50">{{ t('admin.organizations.save') }}</button>
        </div>
      </form>
    </AdminModal>
  </div>
</template>

<script setup lang="ts">
/** 采购组织 CRUD + 成员管理 — /admin/organizations */
definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()

const activeTab = ref<'orgs' | 'members'>('orgs')
const organizations = ref<any[]>([])
const members = ref<any[]>([])
const pending = ref(true)
const membersPending = ref(false)
const error = ref('')
const membersError = ref('')
const selectedOrgId = ref('')
const showOrgModal = ref(false)
const showMemberModal = ref(false)
const editingOrgId = ref<number | null>(null)
const saving = ref(false)
const memberSaving = ref(false)
const formError = ref('')
const memberError = ref('')

const orgForm = reactive({ name: '', creditLimit: 0, paymentTerms: '', approvalThreshold: 0, isActive: true })
const memberForm = reactive({ userId: '', role: 'buyer' })

const orgColumns = computed(() => [
  { key: 'name', label: t('admin.organizations.col_name') },
  { key: 'creditLimit', label: t('admin.organizations.col_credit') },
  { key: 'paymentTerms', label: t('admin.organizations.col_terms') },
  { key: 'isActive', label: t('admin.organizations.col_status') },
  { key: 'actions', label: t('admin.organizations.actions') },
])

const memberColumns = computed(() => [
  { key: 'userId', label: t('admin.organizations.col_user') },
  { key: 'role', label: t('admin.organizations.col_role') },
  { key: 'actions', label: t('admin.organizations.actions') },
])

const tabClass = (tab: string) => [
  'whitespace-nowrap border-b-2 py-3 text-sm font-medium',
  activeTab.value === tab ? 'border-orange-500 text-orange-600' : 'border-transparent text-gray-500 hover:text-gray-700',
]

const fetchOrganizations = async () => {
  pending.value = true
  error.value = ''
  try {
    const res = await api.get<any>('/admin/organizations')
    organizations.value = res.data || []
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const fetchMembers = async () => {
  if (!selectedOrgId.value) { members.value = []; return }
  membersPending.value = true
  membersError.value = ''
  try {
    const res = await api.get<any>('/admin/organizations/members', { orgId: selectedOrgId.value })
    members.value = res.data || []
  } catch (err: any) {
    membersError.value = err?.message || t('errors.api.load_failed')
  } finally {
    membersPending.value = false
  }
}

const selectOrgForMembers = (org: any) => {
  selectedOrgId.value = String(org.id)
  activeTab.value = 'members'
  fetchMembers()
}

const openCreateOrg = () => {
  editingOrgId.value = null
  Object.assign(orgForm, { name: '', creditLimit: 0, paymentTerms: 'Net 30', approvalThreshold: 0, isActive: true })
  formError.value = ''
  showOrgModal.value = true
}

const openEditOrg = (row: any) => {
  editingOrgId.value = row.id
  Object.assign(orgForm, {
    name: row.name, creditLimit: row.creditLimit || 0,
    paymentTerms: row.paymentTerms || '', approvalThreshold: row.approvalThreshold || 0,
    isActive: row.isActive !== false,
  })
  formError.value = ''
  showOrgModal.value = true
}

const closeOrgModal = () => { showOrgModal.value = false; saving.value = false }

const saveOrg = async () => {
  saving.value = true
  formError.value = ''
  try {
    if (editingOrgId.value) {
      await api.put(`/admin/organizations/${editingOrgId.value}`, { ...orgForm })
    } else {
      await api.post('/admin/organizations', { ...orgForm })
    }
    closeOrgModal()
    await fetchOrganizations()
  } catch (err: any) {
    formError.value = err?.message || t('errors.api.save_failed')
  } finally {
    saving.value = false
  }
}

const openAddMember = () => {
  memberForm.userId = ''
  memberForm.role = 'buyer'
  memberError.value = ''
  showMemberModal.value = true
}

const saveMember = async () => {
  if (!selectedOrgId.value) return
  memberSaving.value = true
  memberError.value = ''
  try {
    await api.post('/admin/organizations/members', {
      organizationId: Number(selectedOrgId.value),
      userId: memberForm.userId,
      role: memberForm.role,
    })
    showMemberModal.value = false
    await fetchMembers()
  } catch (err: any) {
    memberError.value = err?.message || t('errors.api.save_failed')
  } finally {
    memberSaving.value = false
  }
}

const removeMember = async (id: number) => {
  if (!confirm(t('admin.confirm_delete'))) return
  try {
    await api.delete(`/admin/organizations/members/${id}`)
    await fetchMembers()
  } catch (err: any) {
    alert(err?.message || t('errors.api.delete_failed'))
  }
}

onMounted(fetchOrganizations)
</script>
