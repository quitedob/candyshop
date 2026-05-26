<template>
  <div class="admin-layout" :class="{ 'dark': isDark }">
    <!-- Mobile top bar -->
    <header class="admin-mobile-topbar">
      <button
        type="button"
        class="admin-mobile-topbar__btn"
        :aria-expanded="isMobileNavOpen ? 'true' : 'false'"
        aria-controls="admin-sidebar-nav"
        :aria-label="t('admin.a11y.openNav')"
        @click="isMobileNavOpen = !isMobileNavOpen"
      >
        <Icon name="material-symbols:menu" class="admin-mobile-topbar__icon" aria-hidden="true" />
      </button>
      <span class="admin-mobile-topbar__brand">{{ t('admin.brand') }}</span>
    </header>

    <!-- Mobile backdrop -->
    <div
      class="admin-sidebar-backdrop"
      :class="{ 'admin-sidebar-backdrop--visible': isMobileNavOpen }"
      aria-hidden="true"
      @click="isMobileNavOpen = false"
    />

    <!-- Sidebar — Alexandria primary-container -->
    <aside
      class="admin-sidebar"
      :class="{
        'admin-sidebar--collapsed': isCollapsed,
        'admin-sidebar--mobile-open': isMobileNavOpen
      }"
    >
      <!-- Header -->
      <div class="admin-sidebar__header">
        <NuxtLink to="/" class="admin-sidebar__logo" @click="isMobileNavOpen = false">
          <svg viewBox="0 0 180 40" fill="none" class="admin-sidebar__logo-svg">
            <circle cx="20" cy="20" r="16" fill="var(--color-highlight)" opacity="0.25"/>
            <path d="M14 20C14 16.6863 16.6863 14 20 14C23.3137 14 26 16.6863 26 20C26 23.3137 23.3137 26 20 26C16.6863 26 14 23.3137 14 20Z" stroke="var(--color-highlight)" stroke-width="2.5"/>
            <circle cx="17" cy="17" r="2" fill="var(--color-accent)"/>
            <circle cx="23" cy="23" r="2" fill="white"/>
            <text x="45" y="27" font-family="Noto Serif, Georgia, serif" font-size="18" font-weight="700" fill="white">CandyPro</text>
          </svg>
        </NuxtLink>
        <button
          type="button"
          class="admin-sidebar__close"
          :aria-label="t('admin.a11y.closeNav')"
          @click="isMobileNavOpen = false"
        >
          <Icon name="material-symbols:close" class="admin-sidebar__close-icon" aria-hidden="true" />
        </button>
        <button
          type="button"
          class="admin-sidebar__toggle"
          :aria-expanded="!isCollapsed ? 'true' : 'false'"
          :aria-label="t('admin.a11y.toggleSidebar')"
          @click="isCollapsed = !isCollapsed"
        >
          <Icon :name="isCollapsed ? 'material-symbols:chevron-right' : 'material-symbols:chevron-left'" class="admin-sidebar__toggle-icon" aria-hidden="true" />
        </button>
      </div>

      <!-- Navigation -->
      <nav id="admin-sidebar-nav" class="admin-sidebar__nav" :aria-label="t('admin.brand')">
        <!-- Group: Main -->
        <div class="admin-sidebar__group">
          <NuxtLink
            v-for="item in mainNav"
            :key="item.key"
            :to="localePath(item.href)"
            class="admin-sidebar__link"
            :class="{ 'admin-sidebar__link--active': isActive(item.href) }"
            :title="!showNavLabels ? t(item.key) : ''"
            @click="isMobileNavOpen = false"
          >
            <Icon :name="item.icon" class="admin-sidebar__link-icon" aria-hidden="true" />
            <span v-if="showNavLabels" class="admin-sidebar__link-text">{{ t(item.key) }}</span>
          </NuxtLink>
        </div>

        <div class="admin-sidebar__divider" />

        <!-- Group: Management -->
        <div v-if="showNavLabels" class="admin-sidebar__group-label">{{ t('admin.nav.management') }}</div>
        <NuxtLink
          v-for="item in managementNav"
          :key="item.key"
          :to="localePath(item.href)"
          class="admin-sidebar__link"
          :class="{ 'admin-sidebar__link--active': isActive(item.href) }"
          :title="!showNavLabels ? t(item.key) : ''"
          @click="isMobileNavOpen = false"
        >
          <Icon :name="item.icon" class="admin-sidebar__link-icon" aria-hidden="true" />
          <span v-if="showNavLabels" class="admin-sidebar__link-text">{{ t(item.key) }}</span>
        </NuxtLink>

        <!-- Super Admin -->
        <template v-if="isSuperAdmin">
          <div class="admin-sidebar__divider" />
          <div v-if="showNavLabels" class="admin-sidebar__group-label">{{ t('admin.nav.superadmin') }}</div>
          <NuxtLink
            v-for="item in superAdminNav"
            :key="item.key"
            :to="localePath(item.href)"
            class="admin-sidebar__link"
            :class="{ 'admin-sidebar__link--active': isActive(item.href) }"
            :title="!showNavLabels ? t(item.key) : ''"
            @click="isMobileNavOpen = false"
          >
            <Icon :name="item.icon" class="admin-sidebar__link-icon" aria-hidden="true" />
            <span v-if="showNavLabels" class="admin-sidebar__link-text">{{ t(item.key) }}</span>
          </NuxtLink>
        </template>
      </nav>

      <!-- CTA Button -->
      <div v-if="showNavLabels" class="admin-sidebar__cta">
        <NuxtLink :to="localePath('/admin/orders')" class="admin-sidebar__cta-btn">
          <Icon name="material-symbols:add" size="18" aria-hidden="true" />
          {{ t('admin.nav.orders') }}
        </NuxtLink>
      </div>

      <!-- Footer -->
      <div class="admin-sidebar__footer">
        <div class="admin-sidebar__lang">
          <LanguageSwitcher />
        </div>
        <button
          type="button"
          class="admin-sidebar__logout min-h-11 min-w-11"
          :aria-label="isDark ? t('admin.theme.light') : t('admin.theme.dark')"
          :title="isDark ? t('admin.theme.light') : t('admin.theme.dark')"
          @click="toggleDark"
        >
          <Icon :name="isDark ? 'material-symbols:light-mode' : 'material-symbols:dark-mode'" class="admin-sidebar__logout-icon" aria-hidden="true" />
          <span v-if="showNavLabels" class="admin-sidebar__logout-text">{{ isDark ? t('admin.theme.light') : t('admin.theme.dark') }}</span>
        </button>
        <div class="admin-sidebar__user" :class="{ 'admin-sidebar__user--collapsed': !showNavLabels }">
          <div class="admin-sidebar__user-avatar">{{ userInitials }}</div>
          <div v-if="showNavLabels" class="admin-sidebar__user-info">
            <p class="admin-sidebar__user-name">{{ user?.firstName }} {{ user?.lastName }}</p>
            <p class="admin-sidebar__user-role">{{ roleLabel(user?.role) }}</p>
          </div>
        </div>
        <button type="button" class="admin-sidebar__logout" :title="!showNavLabels ? t('admin.logout') : ''" :aria-label="t('admin.logout')" @click="handleLogout">
          <Icon name="material-symbols:logout" class="admin-sidebar__logout-icon" aria-hidden="true" />
          <span v-if="showNavLabels" class="admin-sidebar__logout-text">{{ t('admin.logout') }}</span>
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

<script setup lang="ts">
// 门户页不索引；避免 useSeo 在 layout 中调用 useRoute（CSR/HMR 时 router 可能未就绪）
useHead({ meta: [{ name: 'robots', content: 'noindex, nofollow' }] })

const { user, logout } = useAuth()
const { isDark, toggle: toggleDark } = useDarkMode()
const route = useRoute()
const localePath = useLocalePath()
const { t, te, locale, loadLocaleMessages } = useI18n()

await useAsyncData(
  () => `admin-locale-messages-${locale.value}`,
  () => loadLocaleMessages(locale.value),
  { watch: [locale] },
)

const isCollapsed = ref(false)
const isMobileNavOpen = ref(false)
const isSuperAdmin = computed(() => user?.value?.role === 'superadmin')
const showNavLabels = computed(() => !isCollapsed.value || isMobileNavOpen.value)

const userInitials = computed(() => {
  const current = user?.value
  if (!current) return 'U'
  return `${current.firstName?.charAt(0) || ''}${current.lastName?.charAt(0) || ''}`
})

const roleLabel = (role?: string | null) => {
  if (!role) return ''
  const key = `roles.${role}`
  return te(key) ? t(key) : role
}

const mainNav = [
  { key: 'admin.nav.dashboard', href: '/admin', icon: 'material-symbols:home' },
  { key: 'admin.nav.inquiries', href: '/admin/inquiries', icon: 'material-symbols:inbox' },
  { key: 'admin.nav.orders', href: '/admin/orders', icon: 'material-symbols:shopping-cart' },
  { key: 'admin.nav.returns', href: '/admin/returns', icon: 'material-symbols:assignment-return' },
]

const managementNav = [
  { key: 'admin.nav.analytics', href: '/admin/analytics', icon: 'material-symbols:bar-chart' },
  { key: 'admin.nav.financial', href: '/admin/financial', icon: 'material-symbols:payments' },
  { key: 'admin.nav.inventory', href: '/admin/inventory', icon: 'material-symbols:inventory-2' },
  { key: 'admin.nav.xlsx', href: '/admin/xlsx', icon: 'material-symbols:table-chart' },
  { key: 'admin.nav.shipments', href: '/admin/shipments', icon: 'material-symbols:local-shipping' },
  { key: 'admin.nav.invoices', href: '/admin/invoices', icon: 'material-symbols:description' },
  { key: 'admin.nav.products', href: '/admin/products', icon: 'material-symbols:package-2' },
  { key: 'admin.nav.categories', href: '/admin/categories', icon: 'material-symbols:category' },
  { key: 'admin.nav.pricing', href: '/admin/pricing', icon: 'material-symbols:sell' },
  { key: 'admin.nav.coupons', href: '/admin/coupons', icon: 'material-symbols:confirmation-number' },
  { key: 'admin.nav.shipping_rates', href: '/admin/shipping-rates', icon: 'material-symbols:local-shipping' },
  { key: 'admin.nav.tax_rates', href: '/admin/tax-rates', icon: 'material-symbols:percent' },
  { key: 'admin.nav.companies', href: '/admin/companies', icon: 'material-symbols:apartment' },
  { key: 'admin.nav.trades', href: '/admin/trades', icon: 'material-symbols:attach-money' },
  { key: 'admin.nav.oemProjects', href: '/admin/oem-projects', icon: 'material-symbols:factory' },
  { key: 'admin.nav.ai', href: '/admin/ai', icon: 'material-symbols:psychology' },
]

const superAdminNav = [
  { key: 'admin.nav.users', href: '/admin/users', icon: 'material-symbols:group' },
  { key: 'admin.nav.organizations', href: '/admin/organizations', icon: 'material-symbols:corporate-fare' },
  { key: 'admin.nav.channels', href: '/admin/channels', icon: 'material-symbols:hub' },
  { key: 'admin.nav.webhooks', href: '/admin/webhooks', icon: 'material-symbols:webhook' },
  { key: 'admin.nav.hooks', href: '/admin/hooks', icon: 'material-symbols:extension' },
  { key: 'admin.nav.staff', href: '/admin/staff', icon: 'material-symbols:groups' },
  { key: 'admin.nav.certifications', href: '/admin/certifications', icon: 'material-symbols:verified-user' },
  { key: 'admin.nav.auditLog', href: '/admin/audit-log', icon: 'material-symbols:schedule' },
  { key: 'admin.nav.settings', href: '/admin/settings', icon: 'material-symbols:settings' },
  { key: 'admin.nav.translations', href: '/admin/translations', icon: 'material-symbols:translate' },
  { key: 'admin.nav.content', href: '/admin/content', icon: 'material-symbols:newspaper' },
]

const isActive = (href: string) => {
  const path = route?.path
  if (!path) return false
  const fullPath = localePath(href)
  if (href === '/admin') return path === fullPath
  return path.startsWith(fullPath)
}

const handleLogout = async () => {
  isMobileNavOpen.value = false
  await logout()
}

let mediaQuery: MediaQueryList | undefined
const closeMobileIfDesktop = () => {
  if (mediaQuery && !mediaQuery.matches) isMobileNavOpen.value = false
}

onMounted(async () => {
  const saved = localStorage.getItem('admin-sidebar-collapsed')
  if (saved) isCollapsed.value = saved === 'true'
  mediaQuery = window.matchMedia('(min-width: 768px)')
  mediaQuery.addEventListener('change', closeMobileIfDesktop)
})

watch(isCollapsed, (val) => {
  localStorage.setItem('admin-sidebar-collapsed', val.toString())
})

watch(() => route?.path, () => { isMobileNavOpen.value = false })

watch(isMobileNavOpen, (open) => {
  if (import.meta.client) document.body.classList.toggle('body-lock', open)
})

onUnmounted(() => {
  if (mediaQuery) mediaQuery.removeEventListener('change', closeMobileIfDesktop)
  if (import.meta.client) document.body.classList.remove('body-lock')
})
</script>

<style scoped>
.admin-layout {
  display: flex;
  min-height: 100vh;
  background: var(--color-bg);
}

/* Mobile top bar */
.admin-mobile-topbar {
  display: none;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-sm) var(--spacing-md);
  padding-top: max(var(--spacing-sm), env(safe-area-inset-top, 0px));
  background: var(--color-bg);
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  z-index: 570;
}

.admin-mobile-topbar__btn {
  min-width: 44px;
  min-height: 44px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  background: var(--color-bg-alt);
  color: var(--color-primary);
  border: none;
  padding: 0;
  cursor: pointer;
}

.admin-mobile-topbar__icon { width: 22px; height: 22px; }

.admin-mobile-topbar__brand {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-text);
}

.admin-sidebar-backdrop { display: none; }
.admin-sidebar-backdrop--visible {
  display: block;
  position: fixed;
  inset: 0;
  z-index: 550;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(2px);
}

@media (min-width: 768px) {
  .admin-sidebar-backdrop--visible { display: none !important; }
}

/* Sidebar — Alexandria primary-container */
.admin-sidebar {
  width: 280px;
  height: 100vh;
  position: sticky;
  top: 0;
  background: var(--color-primary-container);
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease, transform 0.3s ease;
  overflow: hidden;
  z-index: 560;
  border-right: 1px solid rgba(255, 255, 255, 0.06);
}

.admin-sidebar--collapsed {
  width: 72px;
}

/* Header */
.admin-sidebar__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-xs);
  padding: var(--spacing-lg) var(--spacing-md);
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.admin-sidebar__logo { display: block; flex: 1; min-width: 0; }
.admin-sidebar__logo-svg { height: 36px; width: auto; }

.admin-sidebar__toggle {
  min-width: 44px;
  min-height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  padding: 0;
  background: transparent;
  cursor: pointer;
  border-radius: var(--radius-md);
  color: rgba(255, 255, 255, 0.4);
  transition: background-color var(--transition-fast), color var(--transition-fast);
}

.admin-sidebar__toggle:hover {
  background: rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.8);
}

.admin-sidebar__toggle-icon { width: 18px; height: 18px; }

.admin-sidebar__close {
  display: none;
  min-width: 44px;
  min-height: 44px;
  align-items: center;
  justify-content: center;
  border: none;
  padding: 0;
  background: transparent;
  cursor: pointer;
  border-radius: var(--radius-md);
  color: rgba(255, 255, 255, 0.5);
}

.admin-sidebar__close:hover { background: rgba(255, 255, 255, 0.08); color: white; }
.admin-sidebar__close-icon { width: 22px; height: 22px; }

/* Navigation */
.admin-sidebar__nav {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-md) 0;
}

.admin-sidebar__nav::-webkit-scrollbar { width: 4px; }
.admin-sidebar__nav::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.12);
  border-radius: 2px;
}

.admin-sidebar__group { margin-bottom: var(--spacing-xs); }

.admin-sidebar__group-label {
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: rgba(255, 255, 255, 0.3);
}

.admin-sidebar__divider {
  height: 1px;
  background: rgba(255, 255, 255, 0.06);
  margin: var(--spacing-sm) var(--spacing-md);
}

.admin-sidebar__link {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  margin: 1px var(--spacing-sm);
  border-radius: var(--radius-md);
  color: rgba(255, 255, 255, 0.55);
  font-size: var(--text-sm);
  font-weight: 500;
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast);
  min-height: 44px;
  box-sizing: border-box;
}

.admin-sidebar__link:hover {
  background: rgba(255, 255, 255, 0.06);
  color: rgba(255, 255, 255, 0.85);
}

.admin-sidebar__link--active {
  background: var(--color-highlight);
  color: white;
  font-weight: 600;
  transform: translateX(4px);
}

.admin-sidebar__link--active:hover {
  background: var(--color-highlight);
  color: white;
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

.admin-sidebar--collapsed .admin-sidebar__link--active {
  transform: none;
}

/* CTA */
.admin-sidebar__cta {
  padding: var(--spacing-md);
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.admin-sidebar__cta-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--radius-md);
  background: var(--color-highlight);
  color: white;
  font-size: var(--text-sm);
  font-weight: 600;
  text-decoration: none;
  transition: opacity var(--transition-fast);
  min-height: 44px;
}

.admin-sidebar__cta-btn:hover {
  opacity: 0.9;
}

/* Footer */
.admin-sidebar__footer {
  padding: var(--spacing-md);
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.admin-sidebar__lang {
  display: flex;
  gap: 4px;
  margin-bottom: var(--spacing-sm);
  background: rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-md);
  padding: 3px;
}

.admin-sidebar__lang--icon { margin-bottom: var(--spacing-sm); }

.admin-sidebar__lang-btn {
  flex: 1;
  padding: 6px 0;
  border: none;
  cursor: pointer;
  border-radius: calc(var(--radius-md) - 2px);
  font-size: var(--text-xs);
  font-weight: 600;
  color: rgba(255, 255, 255, 0.35);
  transition: background-color var(--transition-fast), color var(--transition-fast);
  text-align: center;
  background: transparent;
}

.admin-sidebar__lang-btn:hover { color: rgba(255, 255, 255, 0.7); }

.admin-sidebar__lang-btn--active {
  background: rgba(255, 255, 255, 0.12);
  color: white;
}

.admin-sidebar__user {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
}

.admin-sidebar__user--collapsed {
  justify-content: center;
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

.admin-sidebar__user-info { overflow: hidden; }

.admin-sidebar__user-name {
  font-size: var(--text-sm);
  font-weight: 500;
  color: rgba(255, 255, 255, 0.85);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.admin-sidebar__user-role {
  font-size: var(--text-xs);
  color: rgba(255, 255, 255, 0.35);
  text-transform: capitalize;
}

.admin-sidebar__logout {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  width: 100%;
  min-height: 44px;
  padding: var(--spacing-sm);
  border: none;
  cursor: pointer;
  border-radius: var(--radius-md);
  color: rgba(255, 255, 255, 0.35);
  font-size: var(--text-sm);
  transition: background-color var(--transition-fast), color var(--transition-fast);
  background: transparent;
}

.admin-sidebar__logout:hover {
  background: rgba(186, 26, 26, 0.15);
  color: var(--color-error);
}

.admin-sidebar__logout-icon { width: 18px; height: 18px; flex-shrink: 0; }
.admin-sidebar__logout-text { overflow: hidden; text-overflow: ellipsis; }

/* Main Content */
.admin-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.admin-main__content {
  flex: 1;
  min-width: 0;
  padding: var(--spacing-xl);
  animation: fadeInUp 0.4s ease forwards;
}

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

@media (prefers-reduced-motion: reduce) {
  .admin-main__content { animation: none; }
}

/* RTL */
[dir="rtl"] .admin-layout {
  flex-direction: row-reverse;
}

[dir="rtl"] .admin-sidebar {
  border-right: none;
  border-left: 1px solid rgba(255, 255, 255, 0.06);
}

[dir="rtl"] .admin-sidebar__toggle-icon {
  transform: scaleX(-1);
}

[dir="rtl"] .admin-sidebar__link--active {
  transform: translateX(-4px);
}

[dir="rtl"] .admin-sidebar--collapsed .admin-sidebar__link--active {
  transform: none;
}

/* Responsive */
@media (max-width: 767px) {
  .admin-layout { flex-direction: column; }
  .admin-mobile-topbar { display: flex; }

  .admin-sidebar {
    position: fixed;
    left: 0;
    top: 0;
    bottom: 0;
    width: min(300px, 88vw) !important;
    height: 100dvh;
    max-height: 100dvh;
    transform: translateX(-100%);
    box-shadow: none;
  }

  [dir="rtl"] .admin-sidebar {
    left: auto;
    right: 0;
    transform: translateX(100%);
  }

  .admin-sidebar--collapsed { width: min(300px, 88vw) !important; }

  .admin-sidebar--mobile-open {
    transform: translateX(0);
    box-shadow: 8px 0 32px rgba(0, 0, 0, 0.25);
  }

  [dir="rtl"] .admin-sidebar--mobile-open {
    transform: translateX(0);
    box-shadow: -8px 0 32px rgba(0, 0, 0, 0.25);
  }

  .admin-sidebar__toggle { display: none; }
  .admin-sidebar__close { display: flex; }

  .admin-main { width: 100%; }

  .admin-main__content {
    padding: var(--spacing-md);
    padding-bottom: max(var(--spacing-xl), env(safe-area-inset-bottom, 0px));
  }
}
</style>
