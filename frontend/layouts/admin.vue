<template>
  <div class="admin-layout">
    <!-- Sidebar -->
    <aside class="admin-sidebar" :class="{ 'admin-sidebar--collapsed': isCollapsed }">
      <!-- Logo -->
      <div class="admin-sidebar__header">
        <NuxtLink to="/" class="admin-sidebar__logo">
          <svg viewBox="0 0 180 40" fill="none" class="admin-sidebar__logo-svg">
            <circle cx="20" cy="20" r="16" fill="var(--color-highlight)" opacity="0.3"/>
            <path d="M14 20C14 16.6863 16.6863 14 20 14C23.3137 14 26 16.6863 26 20C26 23.3137 23.3137 26 20 26C16.6863 26 14 23.3137 14 20Z" stroke="var(--color-highlight)" stroke-width="2.5"/>
            <circle cx="17" cy="17" r="2" fill="var(--color-accent)"/>
            <circle cx="23" cy="23" r="2" fill="white"/>
            <text x="45" y="27" font-family="Outfit, system-ui, sans-serif" font-size="18" font-weight="700" fill="white">CandyPro</text>
          </svg>
        </NuxtLink>
        <button class="admin-sidebar__toggle" @click="isCollapsed = !isCollapsed">
          <Icon :name="isCollapsed ? 'heroicons:chevron-right' : 'heroicons:chevron-left'" class="admin-sidebar__toggle-icon" />
        </button>
      </div>

      <!-- Navigation -->
      <nav class="admin-sidebar__nav">
        <!-- Group: Main -->
        <div class="admin-sidebar__group">
          <template v-for="item in mainNav" :key="item.key">
            <NuxtLink
              :to="item.href"
              class="admin-sidebar__link"
              :class="{ 'admin-sidebar__link--active': isActive(item.href) }"
              :title="isCollapsed ? t(item.key) : ''"
            >
              <Icon :name="item.icon" class="admin-sidebar__link-icon" />
              <span v-if="!isCollapsed" class="admin-sidebar__link-text">{{ t(item.key) }}</span>
            </NuxtLink>
          </template>
        </div>

        <!-- Divider -->
        <div class="admin-sidebar__divider"></div>

        <!-- Group: Management -->
        <div v-if="!isCollapsed" class="admin-sidebar__group-label">{{ t('admin.nav.management') }}</div>
        <template v-for="item in managementNav" :key="item.key">
          <NuxtLink
            :to="item.href"
            class="admin-sidebar__link"
            :class="{ 'admin-sidebar__link--active': isActive(item.href) }"
            :title="isCollapsed ? t(item.key) : ''"
          >
            <Icon :name="item.icon" class="admin-sidebar__link-icon" />
            <span v-if="!isCollapsed" class="admin-sidebar__link-text">{{ t(item.key) }}</span>
          </NuxtLink>
        </template>

        <!-- Divider -->
        <div class="admin-sidebar__divider"></div>

        <!-- Group: Super Admin -->
        <template v-if="isSuperAdmin">
          <div v-if="!isCollapsed" class="admin-sidebar__group-label">{{ t('admin.nav.superadmin') }}</div>
          <template v-for="item in superAdminNav" :key="item.key">
            <NuxtLink
              :to="item.href"
              class="admin-sidebar__link"
              :class="{ 'admin-sidebar__link--active': isActive(item.href) }"
              :title="isCollapsed ? t(item.key) : ''"
            >
              <Icon :name="item.icon" class="admin-sidebar__link-icon" />
              <span v-if="!isCollapsed" class="admin-sidebar__link-text">{{ t(item.key) }}</span>
            </NuxtLink>
          </template>
        </template>
      </nav>

      <!-- User Profile -->
      <div class="admin-sidebar__footer">
        <!-- Language Switcher -->
        <div v-if="!isCollapsed" class="admin-sidebar__lang">
          <button
            :class="['admin-sidebar__lang-btn', { 'admin-sidebar__lang-btn--active': locale === 'zh' }]"
            @click="switchLocale('zh')"
          >中文</button>
          <button
            :class="['admin-sidebar__lang-btn', { 'admin-sidebar__lang-btn--active': locale === 'en' }]"
            @click="switchLocale('en')"
          >EN</button>
        </div>
        <div class="admin-sidebar__user">
          <div class="admin-sidebar__user-avatar">{{ userInitials }}</div>
          <div v-if="!isCollapsed" class="admin-sidebar__user-info">
            <p class="admin-sidebar__user-name">{{ user?.firstName }} {{ user?.lastName }}</p>
            <p class="admin-sidebar__user-role">{{ user?.role }}</p>
          </div>
        </div>
        <button v-if="!isCollapsed" @click="handleLogout" class="admin-sidebar__logout">
          <Icon name="heroicons:arrow-right-start-on-rectangle" class="admin-sidebar__logout-icon" />
          <span>{{ t('admin.logout') }}</span>
        </button>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="admin-main">
      <div class="admin-main__content">
        <slot />
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'

const { user, logout } = useAuth()
const route = useRoute()
const { t, locale, setLocale } = useI18n()

const isCollapsed = ref(false)
const isSuperAdmin = computed(() => user.value?.role === 'superadmin')

// User initials
const userInitials = computed(() => {
  if (!user.value) return 'U'
  return `${user.value.firstName?.charAt(0) || ''}${user.value.lastName?.charAt(0) || ''}`
})

// Main Navigation
const mainNav = [
  { key: 'admin.nav.dashboard', href: '/admin', icon: 'heroicons:home' },
  { key: 'admin.nav.inquiries', href: '/admin/inquiries', icon: 'heroicons:inbox-in' },
  { key: 'admin.nav.orders', href: '/admin/orders', icon: 'heroicons:shopping-cart' },
]

// Management Navigation
const managementNav = [
  { key: 'admin.nav.analytics', href: '/admin/analytics', icon: 'heroicons:chart-bar' },
  { key: 'admin.nav.financial', href: '/admin/financial', icon: 'heroicons:banknotes' },
  { key: 'admin.nav.inventory', href: '/admin/inventory', icon: 'heroicons:archive-box' },
  { key: 'admin.nav.shipments', href: '/admin/shipments', icon: 'heroicons:truck' },
  { key: 'admin.nav.invoices', href: '/admin/invoices', icon: 'heroicons:document-text' },
  { key: 'admin.nav.products', href: '/admin/products', icon: 'heroicons:cube' },
  { key: 'admin.nav.pricing', href: '/admin/pricing', icon: 'heroicons:tag' },
  { key: 'admin.nav.companies', href: '/admin/companies', icon: 'heroicons:building-office' },
  { key: 'admin.nav.trades', href: '/admin/trades', icon: 'heroicons:currency-dollar' },
  { key: 'admin.nav.oemProjects', href: '/admin/oem-projects', icon: 'heroicons:sparkles' },
]

// Super Admin Navigation
const superAdminNav = [
  { key: 'admin.nav.users', href: '/admin/users', icon: 'heroicons:users' },
  { key: 'admin.nav.staff', href: '/admin/staff', icon: 'heroicons:user-group' },
  { key: 'admin.nav.certifications', href: '/admin/certifications', icon: 'heroicons:shield-check' },
  { key: 'admin.nav.auditLog', href: '/admin/audit-log', icon: 'heroicons:clock' },
  { key: 'admin.nav.settings', href: '/admin/settings', icon: 'heroicons:cog-6-tooth' },
  { key: 'admin.nav.content', href: '/admin/content', icon: 'heroicons:newspaper' },
]

// Check if nav item is active
const isActive = (href) => {
  if (href === '/admin') {
    return route.path === '/admin'
  }
  return route.path.startsWith(href)
}

const handleLogout = async () => {
  await logout()
}

const switchLocale = (code) => {
  setLocale(code)
}

// Persist collapse state
onMounted(() => {
  const saved = localStorage.getItem('admin-sidebar-collapsed')
  if (saved) {
    isCollapsed.value = saved === 'true'
  }
})

watch(isCollapsed, (val) => {
  localStorage.setItem('admin-sidebar-collapsed', val.toString())
})
</script>

<style scoped>
.admin-layout {
  display: flex;
  min-height: 100vh;
  background: var(--color-bg);
}

/* Sidebar */
.admin-sidebar {
  width: 260px;
  height: 100vh;
  position: sticky;
  top: 0;
  background: linear-gradient(180deg, var(--color-primary) 0%, var(--color-primary-dark) 100%);
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  overflow: hidden;
}

.admin-sidebar--collapsed {
  width: 72px;
}

/* Header */
.admin-sidebar__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-md);
  border-bottom: 1px solid rgba(255,255,255,0.1);
}

.admin-sidebar__logo {
  display: block;
  flex: 1;
}

.admin-sidebar__logo-svg {
  height: 36px;
  width: auto;
}

.admin-sidebar__toggle {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  color: rgba(255,255,255,0.6);
  transition: all var(--transition-fast);
}

.admin-sidebar__toggle:hover {
  background: rgba(255,255,255,0.1);
  color: white;
}

.admin-sidebar__toggle-icon {
  width: 18px;
  height: 18px;
}

/* Navigation */
.admin-sidebar__nav {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-md) 0;
}

.admin-sidebar__nav::-webkit-scrollbar {
  width: 4px;
}

.admin-sidebar__nav::-webkit-scrollbar-thumb {
  background: rgba(255,255,255,0.2);
  border-radius: 2px;
}

.admin-sidebar__group {
  margin-bottom: var(--spacing-sm);
}

.admin-sidebar__group-label {
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: rgba(255,255,255,0.4);
}

.admin-sidebar__divider {
  height: 1px;
  background: rgba(255,255,255,0.1);
  margin: var(--spacing-md) 0;
}

.admin-sidebar__link {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  margin: 2px var(--spacing-sm);
  border-radius: var(--radius-md);
  color: rgba(255,255,255,0.7);
  font-size: var(--text-sm);
  font-weight: 500;
  transition: all var(--transition-fast);
}

.admin-sidebar__link:hover {
  background: rgba(255,255,255,0.1);
  color: white;
}

.admin-sidebar__link--active {
  background: rgba(255, 107, 74, 0.2);
  color: var(--color-highlight);
}

.admin-sidebar__link--active:hover {
  background: rgba(255, 107, 74, 0.25);
}

.admin-sidebar__link-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.admin-sidebar__link-text {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.admin-sidebar--collapsed .admin-sidebar__link {
  justify-content: center;
  padding: var(--spacing-sm);
}

/* Footer */
.admin-sidebar__footer {
  padding: var(--spacing-md);
  border-top: 1px solid rgba(255,255,255,0.1);
}

/* Language Switcher */
.admin-sidebar__lang {
  display: flex;
  gap: 4px;
  margin-bottom: var(--spacing-sm);
  background: rgba(255,255,255,0.08);
  border-radius: var(--radius-md);
  padding: 3px;
}

.admin-sidebar__lang-btn {
  flex: 1;
  padding: 4px 0;
  border-radius: calc(var(--radius-md) - 2px);
  font-size: var(--text-xs);
  font-weight: 600;
  color: rgba(255,255,255,0.5);
  transition: all var(--transition-fast);
  text-align: center;
}

.admin-sidebar__lang-btn:hover {
  color: rgba(255,255,255,0.8);
}

.admin-sidebar__lang-btn--active {
  background: rgba(255,255,255,0.15);
  color: white;
}

.admin-sidebar__user {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
}

.admin-sidebar__user-avatar {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  background: linear-gradient(135deg, var(--color-highlight), var(--color-accent));
  color: white;
  font-size: var(--text-sm);
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.admin-sidebar__user-info {
  overflow: hidden;
}

.admin-sidebar__user-name {
  font-size: var(--text-sm);
  font-weight: 500;
  color: white;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.admin-sidebar__user-role {
  font-size: var(--text-xs);
  color: rgba(255,255,255,0.5);
  text-transform: capitalize;
}

.admin-sidebar__logout {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  width: 100%;
  padding: var(--spacing-sm);
  border-radius: var(--radius-md);
  color: rgba(255,255,255,0.5);
  font-size: var(--text-sm);
  transition: all var(--transition-fast);
}

.admin-sidebar__logout:hover {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.admin-sidebar__logout-icon {
  width: 18px;
  height: 18px;
}

.admin-sidebar--collapsed .admin-sidebar__footer {
  display: none;
}

/* Main Content */
.admin-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.admin-main__content {
  flex: 1;
  padding: var(--spacing-xl);
  animation: fadeInUp 0.4s ease forwards;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
