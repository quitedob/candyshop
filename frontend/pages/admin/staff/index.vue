<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('admin.staff.title') }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ t('admin.staff.description') }}</p>
      </div>
    </div>

    <!-- Staff Table -->
    <div class="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.staff.col_name') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.staff.col_email') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.staff.col_role') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.staff.col_status') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.staff.col_last_login') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 bg-white">
            <tr v-if="pending">
              <td colspan="5" class="px-6 py-10 text-center text-sm text-gray-500">{{ t('admin.staff.loading') }}</td>
            </tr>
            <tr v-else-if="error">
              <td colspan="5" class="px-6 py-10 text-center text-sm text-red-600">{{ error }}</td>
            </tr>
            <tr v-else-if="staff.length === 0">
              <td colspan="5" class="px-6 py-10 text-center text-sm text-gray-500">{{ t('admin.staff.no_data') }}</td>
            </tr>
            <template v-else v-for="user in staff" :key="user.id">
              <tr
                class="hover:bg-gray-50 transition-colors cursor-pointer"
                role="button"
                tabindex="0"
                :aria-expanded="expandedUserId === user.id"
                :aria-label="t('admin.staff.expand_user')"
                @click="toggleExpand(user.id)"
                @keydown.enter.prevent="toggleExpand(user.id)"
                @keydown.space.prevent="toggleExpand(user.id)"
              >
                <td class="px-6 py-4">
                  <div class="flex items-center gap-3">
                    <div class="h-9 w-9 rounded-full bg-gray-200 flex items-center justify-center font-bold text-gray-600 text-sm">
                      {{ user.firstName?.charAt(0) || '' }}{{ user.lastName?.charAt(0) || '' }}
                    </div>
                    <div class="font-medium text-gray-900">{{ user.firstName }} {{ user.lastName }}</div>
                  </div>
                </td>
                <td class="px-6 py-4 text-sm text-gray-600">{{ user.email }}</td>
                <td class="px-6 py-4">
                  <span :class="roleBadgeClass(user.role)" class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold">
                    {{ formatRole(user.role) }}
                  </span>
                </td>
                <td class="px-6 py-4">
                  <span :class="statusBadgeClass(user.status)" class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold">
                    {{ formatStatus(user.status) }}
                  </span>
                </td>
                <td class="px-6 py-4 text-sm text-gray-600">
                  {{ user.lastLoginAt ? formatRelativeTime(user.lastLoginAt) : t('admin.staff.never') }}
                </td>
              </tr>
              <!-- Expanded Activity Row -->
              <tr v-if="expandedUserId === user.id">
                <td colspan="5" class="bg-gray-50 px-6 py-4">
                  <div class="ml-12">
                    <h4 class="text-sm font-semibold text-gray-700 mb-3">{{ t('admin.staff.recent_activity') }}</h4>
                    <div v-if="activityLoading" class="text-sm text-gray-500 py-2">{{ t('admin.staff.loading_activity') }}</div>
                    <div v-else-if="activityError" class="text-sm text-red-600 py-2">{{ activityError }}</div>
                    <div v-else-if="activity.length === 0" class="text-sm text-gray-500 py-2">{{ t('admin.staff.no_activity') }}</div>
                    <table v-else class="min-w-full divide-y divide-gray-200">
                      <thead>
                        <tr>
                          <th class="py-2 pr-4 text-left text-xs font-semibold text-gray-500">{{ t('admin.staff.col_time') }}</th>
                          <th class="py-2 pr-4 text-left text-xs font-semibold text-gray-500">{{ t('admin.staff.col_action') }}</th>
                          <th class="py-2 pr-4 text-left text-xs font-semibold text-gray-500">{{ t('admin.staff.col_entity_type') }}</th>
                          <th class="py-2 text-left text-xs font-semibold text-gray-500">{{ t('admin.staff.col_details') }}</th>
                        </tr>
                      </thead>
                      <tbody class="divide-y divide-gray-200">
                        <tr v-for="entry in activity" :key="entry.id" class="hover:bg-gray-100">
                          <td class="py-2 pr-4 text-xs text-gray-600 whitespace-nowrap">{{ formatRelativeTime(entry.createdAt) }}</td>
                          <td class="py-2 pr-4 text-xs">
                            <span :class="actionBadgeClass(entry.action)" class="inline-flex rounded px-1.5 py-0.5 text-xs font-semibold">
                              {{ entry.action }}
                            </span>
                          </td>
                          <td class="py-2 pr-4 text-xs text-gray-600">{{ entry.entityType }}</td>
                          <td class="py-2 text-xs text-gray-600 max-w-xs truncate">{{ entry.details }}</td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="pagination" class="px-6 py-4 border-t border-gray-200 flex items-center justify-between">
        <div class="text-sm text-gray-600">
          {{ t('admin.staff.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, pagination.total), total: pagination.total }) }}
        </div>
        <div class="flex items-center gap-2">
          <button @click="prevPage" :disabled="page <= 1" class="px-3 py-1.5 border border-gray-200 rounded-lg text-sm hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed">
            {{ t('admin.staff.previous') }}
          </button>
          <button @click="nextPage" :disabled="page >= pagination.totalPages" class="px-3 py-1.5 border border-gray-200 rounded-lg text-sm hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed">
            {{ t('admin.staff.next') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()
const { formatDate } = useDisplay()

const staff = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20

const expandedUserId = ref<string | null>(null)
const activity = ref<any[]>([])
const activityLoading = ref(false)
const activityError = ref('')

const fetchStaff = async () => {
  pending.value = true
  error.value = ''
  try {
    const res = await api.get<any>('/admin/staff', { page: page.value, limit: pageSize })
    staff.value = res.data || []
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const toggleExpand = async (userId: string) => {
  if (expandedUserId.value === userId) {
    expandedUserId.value = null
    activity.value = []
    return
  }
  expandedUserId.value = userId
  activity.value = []
  activityLoading.value = true
  activityError.value = ''
  try {
    const res = await api.get<any>(`/admin/staff/${userId}/activity`, { page: 1, limit: 10 })
    activity.value = res.data || []
  } catch (err: any) {
    activityError.value = err?.message || t('errors.api.load_failed')
  } finally {
    activityLoading.value = false
  }
}

const roleBadgeClass = (role: any) => {
  const name = typeof role === 'string' ? role : role?.name
  if (name === 'superadmin') return 'bg-orange-100 text-orange-800'
  if (name === 'admin') return 'bg-orange-100 text-orange-800'
  return 'bg-gray-100 text-gray-800'
}

const statusBadgeClass = (status: string) => {
  if (status === 'active') return 'bg-green-100 text-green-800'
  if (status === 'pending') return 'bg-yellow-100 text-yellow-800'
  if (status === 'suspended') return 'bg-red-100 text-red-800'
  return 'bg-gray-100 text-gray-800'
}

const actionBadgeClass = (action: string) => {
  if (action === 'create') return 'bg-green-100 text-green-800'
  if (action === 'update') return 'bg-orange-100 text-orange-800'
  if (action === 'delete') return 'bg-red-100 text-red-800'
  if (action === 'status_change') return 'bg-yellow-100 text-yellow-800'
  return 'bg-gray-100 text-gray-800'
}

const formatRole = (role: any) => {
  if (!role) return t('enum.order_status.unknown')
  const name = typeof role === 'string' ? role : role.name
  if (!name) return t('enum.order_status.unknown')
  return name.charAt(0).toUpperCase() + name.slice(1)
}

const formatStatus = (status: string) => {
  if (!status) return t('enum.order_status.unknown')
  return status.charAt(0).toUpperCase() + status.slice(1)
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

  if (diffSec < 60) return t('admin.staff.just_now')
  if (diffMin < 60) return t('admin.staff.minutes_ago', { count: diffMin })
  if (diffHour < 24) return t('admin.staff.hours_ago', { count: diffHour })
  if (diffDay < 7) return t('admin.staff.days_ago', { count: diffDay })
  return formatDate(dateStr)
}

const prevPage = () => { if (page.value > 1) { page.value -= 1; fetchStaff() } }
const nextPage = () => { if (pagination.value && page.value < pagination.value.totalPages) { page.value += 1; fetchStaff() } }

onMounted(fetchStaff)
</script>
