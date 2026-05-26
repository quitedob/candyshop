<template>
  <div>
    <PageHeader :title="t('admin.returns.title')" :description="t('admin.returns.description')">
      <template #actions>
        <select v-model="filterStatus" class="rounded-md border border-gray-300 px-3 py-2 text-sm">
          <option value="">{{ t('admin.returns.filter_all') }}</option>
          <option value="pending">{{ t('admin.returns.status_pending') }}</option>
          <option value="approved">{{ t('admin.returns.status_approved') }}</option>
          <option value="received">{{ t('admin.returns.status_received') }}</option>
          <option value="refunded">{{ t('admin.returns.status_refunded') }}</option>
          <option value="rejected">{{ t('admin.returns.status_rejected') }}</option>
        </select>
      </template>
    </PageHeader>

    <AdminTable
      class="mt-4"
      :columns="columns"
      :rows="returns"
      :loading="pending"
      :error="!!error"
      :error-message="error"
      :empty-text="t('admin.returns.no_data')"
      @retry="fetchReturns"
    >
      <template #cell-id="{ row }">
        <span class="font-mono text-sm">{{ row.id }}</span>
      </template>
      <template #cell-orderId="{ row }">
        <NuxtLink :to="localePath(`/admin/orders/${row.orderId}`)" class="text-orange-600 hover:text-orange-900">{{ row.orderId }}</NuxtLink>
      </template>
      <template #cell-status="{ row }">
        <StatusBadge :status="row.status" :label="t(`admin.returns.status_${row.status}`, row.status)" />
      </template>
      <template #cell-createdAt="{ row }">
        {{ formatDate(row.createdAt) }}
      </template>
      <template #cell-actions="{ row }">
        <button v-if="row.status === 'pending'" type="button" class="text-green-600 hover:text-green-900 mr-2" @click="doAction(row.id, 'approve')">{{ t('admin.returns.approve') }}</button>
        <button v-if="row.status === 'approved'" type="button" class="text-blue-600 hover:text-blue-900 mr-2" @click="doAction(row.id, 'receive')">{{ t('admin.returns.receive') }}</button>
        <button v-if="row.status === 'received'" type="button" class="text-emerald-600 hover:text-emerald-900 mr-2" @click="doAction(row.id, 'refund')">{{ t('admin.returns.refund') }}</button>
        <button v-if="['pending', 'approved'].includes(row.status)" type="button" class="text-red-600 hover:text-red-900" @click="openReject(row)">{{ t('admin.returns.reject') }}</button>
      </template>
      <template #bottom>
        <div v-if="pagination" class="flex items-center justify-between">
          <div class="text-sm text-gray-700">
            {{ t('admin.returns.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, pagination.total), total: pagination.total }) }}
          </div>
          <div class="flex gap-2">
            <button type="button" class="rounded-md border px-3 py-2 text-sm disabled:opacity-50" :disabled="page <= 1" @click="prevPage">{{ t('admin.returns.previous') }}</button>
            <button type="button" class="rounded-md border px-3 py-2 text-sm disabled:opacity-50" :disabled="page >= pagination.totalPages" @click="nextPage">{{ t('admin.returns.next') }}</button>
          </div>
        </div>
      </template>
    </AdminTable>

    <AdminModal :open="showRejectModal" :title="t('admin.returns.reject_title')" @close="showRejectModal = false">
      <form @submit.prevent="confirmReject">
        <label class="block text-sm font-medium text-gray-700">{{ t('admin.returns.reject_reason') }}</label>
        <textarea v-model="rejectReason" rows="3" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        <div class="mt-4 flex justify-end gap-2">
          <button type="button" class="rounded-md border px-4 py-2 text-sm" @click="showRejectModal = false">{{ t('admin.returns.cancel') }}</button>
          <button type="submit" :disabled="acting" class="rounded-md bg-red-600 px-4 py-2 text-sm text-white disabled:opacity-50">{{ t('admin.returns.reject') }}</button>
        </div>
      </form>
    </AdminModal>

    <p v-if="actionMessage" class="mt-4 text-sm" :class="actionError ? 'text-red-600' : 'text-green-600'">{{ actionMessage }}</p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: ['auth'] })

const { t } = useI18n()
const { formatDate } = useDisplay()
const localePath = useLocalePath()
const api = useApi()

const columns = [
  { key: 'id', label: t('admin.returns.col_id') },
  { key: 'orderId', label: t('admin.returns.col_order') },
  { key: 'reason', label: t('admin.returns.col_reason') },
  { key: 'status', label: t('admin.returns.col_status') },
  { key: 'createdAt', label: t('admin.returns.col_date') },
  { key: 'actions', label: '' },
]

const returns = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20
const filterStatus = ref('')
const acting = ref(false)
const actionMessage = ref('')
const actionError = ref(false)
const showRejectModal = ref(false)
const rejectId = ref('')
const rejectReason = ref('')

const fetchReturns = async () => {
  pending.value = true
  error.value = ''
  try {
    const params: Record<string, any> = { page: page.value, limit: pageSize }
    if (filterStatus.value) params.status = filterStatus.value
    const res = await api.get<any>('/admin/returns', params)
    returns.value = res.data || []
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const doAction = async (id: string, action: 'approve' | 'receive' | 'refund') => {
  acting.value = true
  actionMessage.value = ''
  actionError.value = false
  try {
    await api.put(`/admin/returns/${id}/${action}`, {})
    actionMessage.value = t(`admin.returns.${action}_success`)
    await fetchReturns()
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.message || t('errors.api.action_failed')
  } finally {
    acting.value = false
  }
}

const openReject = (row: any) => {
  rejectId.value = row.id
  rejectReason.value = ''
  showRejectModal.value = true
}

const confirmReject = async () => {
  acting.value = true
  try {
    await api.put(`/admin/returns/${rejectId.value}/reject`, { reason: rejectReason.value })
    showRejectModal.value = false
    actionMessage.value = t('admin.returns.reject_success')
    await fetchReturns()
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.message || t('errors.api.action_failed')
  } finally {
    acting.value = false
  }
}

const prevPage = () => { if (page.value > 1) page.value -= 1 }
const nextPage = () => { if (pagination.value && page.value < pagination.value.totalPages) page.value += 1 }

watch(page, fetchReturns)
watch(filterStatus, () => { page.value = 1; fetchReturns() })
onMounted(fetchReturns)
</script>
