<template>
  <div class="customer-dashboard">
    <!-- Loading: skeleton cards -->
    <div v-if="pending" class="dashboard-grid">
      <MetricCard v-for="i in 4" :key="i" :loading="true" />
    </div>

    <!-- Error -->
    <ErrorState v-else-if="error" :message="error" :action-text="t('errors.tryAgain')" @action="loadData" />

    <template v-else>
      <!-- Pending Approval Card -->
      <div v-if="isPending" class="pending-card">
        <div class="pending-card__icon">
          <Icon name="heroicons:clock" size="32" aria-hidden="true" />
        </div>
        <div class="pending-card__content">
          <h3 class="pending-card__title">{{ t('customer.pending.dashboard_title') }}</h3>
          <p class="pending-card__message">{{ t('customer.pending.dashboard_message') }}</p>
          <div class="pending-card__steps">
            <div class="pending-card__step">
              <Icon name="heroicons:eye" size="20" aria-hidden="true" />
              <span>{{ t('customer.pending.step_browse') }}</span>
            </div>
            <div class="pending-card__step">
              <Icon name="heroicons:shield-check" size="20" aria-hidden="true" />
              <span>{{ t('customer.pending.step_review') }}</span>
            </div>
            <div class="pending-card__step">
              <Icon name="heroicons:shopping-bag" size="20" aria-hidden="true" />
              <span>{{ t('customer.pending.step_order') }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Welcome Card -->
      <div class="welcome-card curator-lift">
        <div class="welcome-card__avatar">
          {{ userInitials }}
        </div>
        <div class="welcome-card__text">
          <h2 class="welcome-card__greeting">{{ t('customer.dashboard.welcome', { name: user?.firstName }) }}</h2>
          <p class="welcome-card__sub">{{ user?.company || user?.email }}</p>
        </div>
      </div>

      <!-- KPI Bento Grid -->
      <div class="dashboard-grid">
        <MetricCard
          :title="t('customer.dashboard.recent_orders')"
          :value="recentOrders.length"
          icon="heroicons:receipt-long"
          icon-bg="accent"
          :to="localePath('/customer/orders')"
        />
        <MetricCard
          :title="t('customer.dashboard.company_info')"
          :value="user?.company || t('customer.dashboard.no_company')"
          icon="heroicons:building-office"
          icon-bg="info"
        />
        <MetricCard
          :title="t('customer.nav.quotes')"
          :value="'—'"
          icon="heroicons:document-text"
          icon-bg="warning"
        />
        <MetricCard
          :title="t('customer.nav.notifications')"
          :value="'—'"
          icon="heroicons:bell"
          icon-bg="default"
        />
      </div>

      <!-- Recent Orders -->
      <div class="mt-8">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold text-gray-900">{{ t('customer.dashboard.latest_orders') }}</h3>
          <NuxtLink :to="localePath('/customer/orders')" class="text-sm font-medium text-accent hover:text-highlight transition-colors">
            {{ t('customer.dashboard.view_all_orders') }} &rarr;
          </NuxtLink>
        </div>

        <!-- Orders List -->
        <div v-if="recentOrders.length > 0" class="orders-list">
          <NuxtLink
            v-for="order in recentOrders"
            :key="order.id"
            :to="localePath(`/customer/orders/${order.id}`)"
            class="orders-list__item group"
          >
            <div class="orders-list__info">
              <span class="orders-list__id">#{{ order.orderNumber || order.id?.substring(0, 8) }}</span>
              <StatusBadge :status="order.status" type="order" />
            </div>
            <div class="orders-list__meta">
              <span class="orders-list__amount">{{ formatNumber(order.totalAmount) }}</span>
              <span class="orders-list__date">
                <Icon name="heroicons:calendar" size="14" aria-hidden="true" />
                {{ formatDate(order.createdAt) }}
              </span>
            </div>
          </NuxtLink>
        </div>

        <!-- Empty state -->
        <EmptyState
          v-else
          icon="heroicons:receipt-long"
          :title="t('customer.dashboard.no_orders')"
          :description="t('customer.dashboard.no_orders_desc')"
          :action-text="t('customer.nav.products')"
          :action-to="localePath('/customer/products')"
        />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const { t } = useI18n()
const { formatNumber, formatDate } = useDisplay()
const localePath = useLocalePath()
const { user, isPending } = useAuth()
const api = useApi()

const recentOrders = ref<any[]>([])
const pending = ref(true)
const error = ref('')

const userInitials = computed(() => {
  if (!user.value) return 'U'
  return `${user.value.firstName?.charAt(0) || ''}${user.value.lastName?.charAt(0) || ''}`
})

const loadData = async () => {
  pending.value = true
  error.value = ''
  try {
    const res = await api.get<any>('/user/orders?limit=5')
    recentOrders.value = res.data || []
  } catch (err: any) {
    error.value = err?.message || t('customer.dashboard.error')
  } finally { pending.value = false }
}

onMounted(loadData)
</script>

<style scoped>
.customer-dashboard {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(1, 1fr);
  gap: var(--spacing-lg);
}

@media (min-width: 640px) {
  .dashboard-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1024px) {
  .dashboard-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

/* Welcome Card */
.welcome-card {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  padding: var(--spacing-xl);
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
}

.welcome-card__avatar {
  width: 64px;
  height: 64px;
  border-radius: var(--radius-full);
  background: var(--color-highlight-light);
  color: var(--color-accent);
  font-family: var(--font-display);
  font-size: var(--text-2xl);
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border: 1px solid var(--color-border-light);
}

.welcome-card__greeting {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--color-text);
  margin: 0;
}

.welcome-card__sub {
  font-size: var(--text-sm);
  color: var(--color-text-lighter);
  margin-top: var(--spacing-xs);
}

/* Pending Card */
.pending-card {
  display: flex;
  gap: var(--spacing-lg);
  padding: var(--spacing-xl);
  background: rgba(var(--color-warning-rgb), 0.06);
  border: 1px solid rgba(var(--color-warning-rgb), 0.2);
  border-radius: var(--radius-lg);
}

.pending-card__icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  background: var(--color-warning);
  color: white;
  border-radius: var(--radius-full);
}

.pending-card__title {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: var(--spacing-xs);
}

.pending-card__message {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin-bottom: var(--spacing-md);
}

.pending-card__steps {
  display: flex;
  gap: var(--spacing-lg);
  flex-wrap: wrap;
}

.pending-card__step {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
}

/* Orders List */
.orders-list {
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.orders-list__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-md) var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-light);
  transition: background-color var(--transition-fast);
  color: inherit;
  text-decoration: none;
}

.orders-list__item:last-child {
  border-bottom: none;
}

.orders-list__item:hover {
  background: var(--color-bg-alt);
}

.orders-list__info {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.orders-list__id {
  font-family: var(--font-mono);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
}

.orders-list__meta {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
}

.orders-list__amount {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-text);
}

.orders-list__date {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: var(--text-xs);
  color: var(--color-text-lighter);
}

@media (max-width: 640px) {
  .pending-card {
    flex-direction: column;
    text-align: center;
  }
  .pending-card__icon { margin: 0 auto; }
  .pending-card__steps { justify-content: center; }
  .orders-list__item { flex-direction: column; align-items: flex-start; gap: var(--spacing-sm); }
}
</style>
