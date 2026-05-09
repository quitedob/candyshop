<template>
  <div class="bg-white shadow overflow-hidden sm:rounded-lg">
    <div class="px-4 py-5 sm:px-6 border-b border-gray-200 flex items-center justify-between">
      <div>
        <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.notifications.title') }}</h3>
        <p class="mt-1 text-sm text-gray-500">{{ t('customer.notifications.subtitle') }}</p>
      </div>
      <button v-if="unreadCount > 0" @click="markAllRead" class="text-sm text-orange-600 hover:text-orange-800 font-medium">
        {{ t('customer.notifications.mark_all_read') }}
      </button>
    </div>

    <div v-if="pending" class="text-center py-10 text-gray-500">{{ t('customer.notifications.loading') }}</div>
    <div v-else-if="error" class="mx-4 my-4 rounded-md bg-red-50 p-4 text-sm text-red-700">{{ error }}</div>
    <div v-else-if="notifications.length === 0" class="text-center py-10 text-gray-500">
      {{ t('customer.notifications.no_notifications') }}
    </div>

    <ul v-else class="divide-y divide-gray-200">
      <li v-for="item in notifications" :key="`${item.type}-${item.reference}`"
        :class="['px-4 py-4 sm:px-6 transition-colors', item.isRead ? 'bg-white' : 'bg-orange-50']"
        @click="markRead(item)">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span v-if="!item.isRead" class="w-2 h-2 rounded-full bg-orange-500 flex-shrink-0"></span>
            <p class="text-sm font-medium text-gray-900">{{ item.title }}</p>
          </div>
          <span class="text-xs text-gray-500">{{ formatDate(item.createdAt) }}</span>
        </div>
        <p class="mt-1 text-sm text-gray-600 ml-4">{{ item.message }}</p>
        <div class="mt-1 ml-4">
          <NuxtLink v-if="item.type === 'order'" :to="localePath(`/customer/orders/${item.reference}`)"
            class="text-xs text-orange-600 hover:underline">
            {{ t('customer.notifications.view_order') }}
          </NuxtLink>
          <NuxtLink v-else-if="item.type === 'inquiry'" :to="localePath(`/customer/inquiries/${item.reference}`)"
            class="text-xs text-orange-600 hover:underline">
            {{ t('customer.notifications.view_inquiry') }}
          </NuxtLink>
        </div>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()
const { formatDate: fmtDate } = useDisplay()
const localePath = useLocalePath()

const notifications = ref<any[]>([])
const pending = ref(true)
const error = ref('')

const unreadCount = computed(() => notifications.value.filter(n => !n.isRead).length)

const fetchNotifications = async () => {
  pending.value = true
  error.value = ''
  try {
    const res = await api.get<any>('/user/notifications')
    notifications.value = res.data || []
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const markRead = async (item: any) => {
  if (item.isRead) return
  try {
    await api.put(`/user/notifications/${item.id}/read`, {})
    item.isRead = true
  } catch {
    // Silently fail — UI already updated optimistically
    item.isRead = true
  }
}

const markAllRead = async () => {
  try {
    await api.post('/user/notifications/mark-all-read', {})
    notifications.value.forEach(n => { n.isRead = true })
  } catch {
    notifications.value.forEach(n => { n.isRead = true })
  }
}

const formatDate = (value: string | null | undefined) => {
  if (!value) return t('customer.common.na')
  const dt = new Date(value)
  if (Number.isNaN(dt.getTime())) return t('customer.common.na')
  return fmtDate(dt.toISOString(), { dateStyle: 'medium', timeStyle: 'short' })
}

onMounted(fetchNotifications)
</script>
