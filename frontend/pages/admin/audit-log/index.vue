<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('admin.auditLog.title') }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ t('admin.auditLog.description') }}</p>
      </div>
    </div>

    <!-- Filters -->
    <div class="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 p-4">
      <div class="flex flex-wrap items-center gap-4">
        <div class="flex-1 min-w-[200px]">
          <div class="relative">
            <Icon name="heroicons:magnifying-glass" class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
            <input id="auditlog-search" v-model="searchQuery" name="search" type="text" autocomplete="off" :placeholder="t('admin.auditLog.search')" class="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500" />
          </div>
        </div>
        <select id="auditlog-entityType" name="entityTypeFilter" v-model="entityTypeFilter" @change="resetAndFetch" class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500">
          <option value="">{{ t('admin.auditLog.filter_all_types') }}</option>
          <option value="order">{{ t('admin.auditLog.entity_order') }}</option>
          <option value="inquiry">{{ t('admin.auditLog.entity_inquiry') }}</option>
          <option value="product">{{ t('admin.auditLog.entity_product') }}</option>
          <option value="user">{{ t('admin.auditLog.entity_user') }}</option>
          <option value="trade">{{ t('admin.auditLog.entity_trade') }}</option>
        </select>
        <select id="auditlog-action" name="actionFilter" v-model="actionFilter" @change="resetAndFetch" class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500">
          <option value="">{{ t('admin.auditLog.filter_all_actions') }}</option>
          <option value="create">{{ t('admin.auditLog.action_create') }}</option>
          <option value="update">{{ t('admin.auditLog.action_update') }}</option>
          <option value="delete">{{ t('admin.auditLog.action_delete') }}</option>
          <option value="status_change">{{ t('admin.auditLog.action_status_change') }}</option>
        </select>
      </div>
    </div>

    <!-- Log Table -->
    <div class="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.auditLog.col_time') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.auditLog.col_user') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.auditLog.col_action') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.auditLog.col_entity_type') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.auditLog.col_entity_id') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.auditLog.col_ip') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 bg-white">
            <tr v-if="pending">
              <td colspan="6" class="px-6 py-10 text-center text-sm text-gray-500">{{ t('admin.auditLog.loading') }}</td>
            </tr>
            <tr v-else-if="error">
              <td colspan="6" class="px-6 py-10 text-center text-sm text-red-600">{{ error }}</td>
            </tr>
            <tr v-else-if="filteredLogs.length === 0">
              <td colspan="6" class="px-6 py-10 text-center text-sm text-gray-500">{{ t('admin.auditLog.no_data') }}</td>
            </tr>
            <tr v-else v-for="log in filteredLogs" :key="log.id" class="hover:bg-gray-50 transition-colors">
              <td class="px-6 py-4 text-sm text-gray-600 whitespace-nowrap">
                {{ formatRelativeTime(log.createdAt) }}
              </td>
              <td class="px-6 py-4 text-sm">
                <div class="font-medium text-gray-900">{{ log.userName || log.userEmail || '-' }}</div>
              </td>
              <td class="px-6 py-4">
                <span :class="actionBadgeClass(log.action)" class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold">
                  {{ formatAction(log.action) }}
                </span>
              </td>
              <td class="px-6 py-4 text-sm text-gray-600">
                {{ log.entityType || '-' }}
              </td>
              <td class="px-6 py-4 text-sm font-mono text-gray-600">
                {{ log.entityId ? log.entityId.substring(0, 8) + '...' : '-' }}
              </td>
              <td class="px-6 py-4 text-sm font-mono text-gray-400">
                {{ log.ipAddress || '-' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="pagination" class="px-6 py-4 border-t border-gray-200 flex items-center justify-between">
        <div class="text-sm text-gray-600">
          {{ t('admin.auditLog.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, pagination.total), total: pagination.total }) }}
        </div>
        <div class="flex items-center gap-2">
          <button @click="prevPage" :disabled="page <= 1" class="px-3 py-1.5 border border-gray-200 rounded-lg text-sm hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed">
            {{ t('admin.auditLog.previous') }}
          </button>
          <button @click="nextPage" :disabled="page >= pagination.totalPages" class="px-3 py-1.5 border border-gray-200 rounded-lg text-sm hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed">
            {{ t('admin.auditLog.next') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()
const { formatDate } = useDisplay()

const logs = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20

const searchQuery = ref('')
const entityTypeFilter = ref('')
const actionFilter = ref('')

const filteredLogs = computed(() => logs.value.filter(log => {
  const matchesSearch = !searchQuery.value ||
    log.userName?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    log.userEmail?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    log.entityType?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    log.entityId?.toLowerCase().includes(searchQuery.value.toLowerCase())
  return matchesSearch
}))

const fetchLogs = async () => {
  pending.value = true
  error.value = ''
  const params: Record<string, any> = { page: page.value, limit: pageSize }
  if (entityTypeFilter.value) params.entityType = entityTypeFilter.value
  if (actionFilter.value) params.action = actionFilter.value

  try {
    const res = await api.get<any>('/admin/audit-log', params)
    logs.value = res.data || []
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const resetAndFetch = () => {
  page.value = 1
  fetchLogs()
}

const actionBadgeClass = (action: string) => {
  if (action === 'create') return 'bg-green-100 text-green-800'
  if (action === 'update') return 'bg-orange-100 text-orange-800'
  if (action === 'delete') return 'bg-red-100 text-red-800'
  if (action === 'status_change') return 'bg-yellow-100 text-yellow-800'
  return 'bg-gray-100 text-gray-800'
}

const formatAction = (action: string) => {
  if (!action) return t('common.enum.order_status.unknown')
  return action.split('_').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ')
}

const formatRelativeTime = (dateStr: string) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffSec = Math.floor(diffMs / 1000)
  const diffMin = Math.floor(diffSec / 60)
  const diffHour = Math.floor(diffMin / 60)
  const diffDay = Math.floor(diffHour / 24)

  if (diffSec < 60) return t('admin.auditLog.just_now')
  if (diffMin < 60) return t('admin.auditLog.minutes_ago', { count: diffMin })
  if (diffHour < 24) return t('admin.auditLog.hours_ago', { count: diffHour })
  if (diffDay < 7) return t('admin.auditLog.days_ago', { count: diffDay })
  return formatDate(dateStr)
}

const prevPage = () => { if (page.value > 1) { page.value -= 1; fetchLogs() } }
const nextPage = () => { if (pagination.value && page.value < pagination.value.totalPages) { page.value += 1; fetchLogs() } }

onMounted(fetchLogs)
</script>
