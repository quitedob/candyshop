<template>
  <div class="mt-6">
    <div v-if="pending" class="text-center py-10">
      <p class="text-gray-500">{{ t('customer.dashboard.loading') }}</p>
    </div>
    
    <div v-else-if="error" class="bg-red-50 p-4 rounded-md">
      <div class="flex">
        <div class="flex-shrink-0">
          <Icon name="heroicons:x-circle" class="h-5 w-5 text-red-400" />
        </div>
        <div class="ml-3">
          <h3 class="text-sm font-medium text-red-800">{{ t('customer.dashboard.error') }}</h3>
          <div class="mt-2 text-sm text-red-700">
            <p>{{ error }}</p>
          </div>
        </div>
      </div>
    </div>
    
    <template v-else>
      <!-- Pending Approval Card -->
      <div v-if="isPending" class="pending-card">
        <div class="pending-card__icon">
          <Icon name="heroicons:clock" class="h-8 w-8" />
        </div>
        <div class="pending-card__content">
          <h3 class="pending-card__title">{{ t('customer.pending.dashboard_title') }}</h3>
          <p class="pending-card__message">{{ t('customer.pending.dashboard_message') }}</p>
          <div class="pending-card__steps">
            <div class="pending-card__step">
              <Icon name="heroicons:eye" class="h-5 w-5" />
              <span>{{ t('customer.pending.step_browse') }}</span>
            </div>
            <div class="pending-card__step">
              <Icon name="heroicons:shield-check" class="h-5 w-5" />
              <span>{{ t('customer.pending.step_review') }}</span>
            </div>
            <div class="pending-card__step">
              <Icon name="heroicons:shopping-bag" class="h-5 w-5" />
              <span>{{ t('customer.pending.step_order') }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white overflow-hidden shadow rounded-lg mb-8">
        <div class="px-4 py-5 sm:p-6">
          <div class="flex items-center">
            <div class="h-16 w-16 rounded-full bg-blue-100 flex items-center justify-center font-bold text-blue-700 text-2xl">
              {{ user?.firstName?.charAt(0) }}{{ user?.lastName?.charAt(0) }}
            </div>
            <div class="ml-5">
              <h2 class="text-lg leading-6 font-medium text-gray-900">
                {{ t('customer.dashboard.welcome', { name: user?.firstName }) }}
              </h2>
              <p class="text-sm text-gray-500">
                {{ t('customer.dashboard.subtitle') }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-2">
        <div class="bg-white overflow-hidden shadow rounded-lg">
          <div class="p-5">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <Icon name="heroicons:shopping-bag" class="h-6 w-6 text-gray-400" />
              </div>
              <div class="ml-5 w-0 flex-1">
                <dl>
                  <dt class="text-sm font-medium text-gray-500 truncate">{{ t('customer.dashboard.recent_orders') }}</dt>
                  <dd>
                    <div class="text-lg font-medium text-gray-900">{{ t('customer.dashboard.orders_count', { count: recentOrders.length }) }}</div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
          <div class="bg-gray-50 px-5 py-3">
            <div class="text-sm">
              <NuxtLink to="/customer/orders" class="font-medium text-blue-600 hover:text-blue-500">
                {{ t('customer.dashboard.view_all_orders') }} <span aria-hidden="true">&rarr;</span>
              </NuxtLink>
            </div>
          </div>
        </div>

        <div class="bg-white overflow-hidden shadow rounded-lg">
          <div class="p-5">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <Icon name="heroicons:user" class="h-6 w-6 text-gray-400" />
              </div>
              <div class="ml-5 w-0 flex-1">
                <dl>
                  <dt class="text-sm font-medium text-gray-500 truncate">{{ t('customer.dashboard.company_info') }}</dt>
                  <dd>
                    <div class="text-sm font-medium text-gray-900">{{ user?.company || t('customer.dashboard.no_company') }}</div>
                    <div class="text-sm text-gray-500">{{ user?.email }}</div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
          <div class="bg-gray-50 px-5 py-3">
            <div class="text-sm">
              <NuxtLink to="/customer/profile" class="font-medium text-blue-600 hover:text-blue-500">
                {{ t('customer.dashboard.update_profile') }} <span aria-hidden="true">&rarr;</span>
              </NuxtLink>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Recent Orders List -->
      <div class="mt-8">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">{{ t('customer.dashboard.latest_orders') }}</h3>
        <div v-if="recentOrders.length > 0" class="bg-white shadow overflow-hidden sm:rounded-md">
          <ul role="list" class="divide-y divide-gray-200">
            <li v-for="order in recentOrders" :key="order.id">
              <NuxtLink :to="`/customer/orders/${order.id}`" class="block hover:bg-gray-50">
                <div class="px-4 py-4 sm:px-6">
                  <div class="flex items-center justify-between">
                    <p class="text-sm font-medium text-blue-600 truncate">Order #{{ order.orderNumber || order.id.substring(0,8) }}</p>
                    <div class="ml-2 flex-shrink-0 flex">
                      <p class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
                        {{ order.status }}
                      </p>
                    </div>
                  </div>
                  <div class="mt-2 sm:flex sm:justify-between">
                    <div class="sm:flex">
                      <p class="flex items-center text-sm text-gray-500">
                        <Icon name="heroicons:currency-dollar" class="flex-shrink-0 mr-1.5 h-4 w-4 text-gray-400" />
                        {{ order.totalAmount?.toLocaleString() }}
                      </p>
                    </div>
                    <div class="mt-2 flex items-center text-sm text-gray-500 sm:mt-0">
                      <Icon name="heroicons:calendar" class="flex-shrink-0 mr-1.5 h-4 w-4 text-gray-400" />
                      <p>{{ t('customer.dashboard.placed_on') }} <time :datetime="order.createdAt">{{ new Date(order.createdAt).toLocaleDateString() }}</time></p>
                    </div>
                  </div>
                </div>
              </NuxtLink>
            </li>
          </ul>
        </div>
        <div v-else class="bg-white shadow sm:rounded-lg p-6 text-center text-gray-500">
          {{ t('customer.dashboard.no_orders') }}
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const { t } = useI18n()
const { user, isPending } = useAuth()
const api = useApi()
const tts = useTTS()

const recentOrders = ref<any[]>([])
const pending = ref(true)
const error = ref('')

onMounted(async () => {
  pending.value = true
  try {
    const res = await api.get<any>('/user/orders?limit=5')
    recentOrders.value = res.data || []
    tts.playWelcome()
  } catch (err: any) {
    error.value = err?.message || 'Failed to load dashboard data'
  } finally { pending.value = false }
})
</script>

<style scoped>
.pending-card {
  display: flex;
  gap: 1.5rem;
  padding: 1.5rem;
  margin-bottom: 2rem;
  background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%);
  border: 1px solid #f59e0b;
  border-radius: 0.75rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.pending-card__icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 3rem;
  height: 3rem;
  background: #f59e0b;
  color: white;
  border-radius: 9999px;
}

.pending-card__title {
  font-size: 1.125rem;
  font-weight: 600;
  color: #92400e;
  margin-bottom: 0.25rem;
}

.pending-card__message {
  font-size: 0.875rem;
  color: #a16207;
  margin-bottom: 1rem;
}

.pending-card__steps {
  display: flex;
  gap: 1.5rem;
  flex-wrap: wrap;
}

.pending-card__step {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.8125rem;
  font-weight: 500;
  color: #92400e;
}

.pending-card__step svg {
  color: #f59e0b;
}

@media (max-width: 640px) {
  .pending-card {
    flex-direction: column;
    text-align: center;
  }

  .pending-card__icon {
    margin: 0 auto;
  }

  .pending-card__steps {
    justify-content: center;
  }
}
</style>

