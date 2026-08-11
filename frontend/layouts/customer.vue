<template>
  <div class="customer-layout">
    <!-- Pending Approval Banner -->
    <div v-if="isPending" class="pending-banner">
      <div class="pending-banner__inner">
        <Icon name="material-symbols:schedule" class="pending-banner__icon" aria-hidden="true" />
        <span>{{ t('customer.pending.banner') }}</span>
      </div>
    </div>

    <!-- Mobile Top Bar -->
    <header class="customer-mobile-topbar">
      <button
        type="button"
        class="customer-mobile-topbar__btn"
        :aria-expanded="isMobileNavOpen ? 'true' : 'false'"
        :aria-controls="customerSidebarId"
        :aria-label="t('customer.a11y.openMenu')"
        @click="isMobileNavOpen = !isMobileNavOpen"
      >
        <Icon name="material-symbols:menu" size="22" aria-hidden="true" />
      </button>
      <span class="customer-mobile-topbar__brand">{{ t('customer.brand') }}</span>
      <NuxtLink :to="localePath('/customer/cart')" class="customer-mobile-topbar__cart">
        <Icon name="material-symbols:shopping-bag" size="22" aria-hidden="true" />
        <span v-if="cartCount > 0" class="customer-mobile-topbar__cart-badge">{{ cartCount }}</span>
      </NuxtLink>
    </header>

    <!-- Mobile Backdrop -->
    <div
      class="customer-sidebar-backdrop"
      :class="{ 'customer-sidebar-backdrop--visible': isMobileNavOpen }"
      aria-hidden="true"
      @click="isMobileNavOpen = false"
    />

    <!-- Sidebar -->
    <aside
      :id="customerSidebarId"
      class="customer-sidebar"
      :class="{
        'customer-sidebar--collapsed': isCollapsed,
        'customer-sidebar--mobile-open': isMobileNavOpen
      }"
    >
      <!-- Header -->
      <div class="customer-sidebar__header">
        <NuxtLink :to="localePath('/')" class="customer-sidebar__logo" @click="isMobileNavOpen = false">
          <svg viewBox="0 0 180 40" fill="none" class="customer-sidebar__logo-svg" :class="{ 'customer-sidebar__logo-svg--compact': !showNavLabels }">
            <circle cx="20" cy="20" r="16" fill="var(--color-highlight)" opacity="0.2"/>
            <path d="M14 20C14 16.6863 16.6863 14 20 14C23.3137 14 26 16.6863 26 20C26 23.3137 23.3137 26 20 26C16.6863 26 14 23.3137 14 20Z" stroke="var(--color-highlight)" stroke-width="2.5"/>
            <circle cx="17" cy="17" r="2" fill="var(--color-accent)"/>
            <circle cx="23" cy="23" r="2" fill="white"/>
            <text v-if="showNavLabels" x="45" y="27" font-family="Noto Serif, Georgia, serif" font-size="18" font-weight="700" fill="white">CandyPro</text>
          </svg>
        </NuxtLink>
        <button
          type="button"
          class="customer-sidebar__toggle"
          :aria-label="isCollapsed ? t('customer.a11y.expandSidebar') : t('customer.a11y.collapseSidebar')"
          @click="toggleSidebar"
        >
          <Icon :name="isCollapsed ? 'material-symbols:chevron-right' : 'material-symbols:chevron-left'" class="customer-sidebar__toggle-icon" aria-hidden="true" />
        </button>
        <button
          type="button"
          class="customer-sidebar__close"
          :aria-label="t('customer.a11y.closeMenu')"
          @click="isMobileNavOpen = false"
        >
          <Icon name="material-symbols:close" size="22" aria-hidden="true" />
        </button>
      </div>

      <!-- User Profile -->
      <div class="customer-sidebar__user" :class="{ 'customer-sidebar__user--collapsed': !showNavLabels }">
        <div class="customer-sidebar__user-avatar">{{ userInitials }}</div>
        <div v-if="showNavLabels" class="customer-sidebar__user-info">
          <p class="customer-sidebar__user-name">{{ user?.firstName }} {{ user?.lastName }}</p>
          <p class="customer-sidebar__user-role">{{ user?.company || user?.email }}</p>
        </div>
      </div>

      <!-- Navigation -->
      <nav class="customer-sidebar__nav" :aria-label="t('customer.brand')">
        <!-- Group: Main -->
        <div v-if="showNavLabels" class="customer-sidebar__group-label">{{ t('customer.nav.main_group') }}</div>
        <NuxtLink
          v-for="item in mainNav"
          :key="item.key"
          :to="localePath(item.href)"
          class="customer-sidebar__link"
          :class="{ 'customer-sidebar__link--active': isActive(item.href) }"
          :title="!showNavLabels ? t(item.key) : undefined"
          @click="isMobileNavOpen = false"
        >
          <Icon :name="item.icon" class="customer-sidebar__link-icon" aria-hidden="true" />
          <span v-if="showNavLabels" class="customer-sidebar__link-text">{{ t(item.key) }}</span>
          <span v-if="showNavLabels && item.key === 'customer.nav.cart' && cartCount > 0" class="customer-sidebar__badge">{{ cartCount }}</span>
          <span v-if="!showNavLabels && item.key === 'customer.nav.cart' && cartCount > 0" class="customer-sidebar__dot-badge" />
        </NuxtLink>

        <div class="customer-sidebar__divider" />

        <!-- Group: Orders & Trade -->
        <div v-if="showNavLabels" class="customer-sidebar__group-label">{{ t('customer.nav.orders_group') }}</div>
        <NuxtLink
          v-for="item in orderNav"
          :key="item.key"
          :to="localePath(item.href)"
          class="customer-sidebar__link"
          :class="{ 'customer-sidebar__link--active': isActive(item.href) }"
          :title="!showNavLabels ? t(item.key) : undefined"
          @click="isMobileNavOpen = false"
        >
          <Icon :name="item.icon" class="customer-sidebar__link-icon" aria-hidden="true" />
          <span v-if="showNavLabels" class="customer-sidebar__link-text">{{ t(item.key) }}</span>
        </NuxtLink>

        <div class="customer-sidebar__divider" />

        <!-- Group: Account -->
        <div v-if="showNavLabels" class="customer-sidebar__group-label">{{ t('customer.nav.account_group') }}</div>
        <NuxtLink
          v-for="item in accountNav"
          :key="item.key"
          :to="localePath(item.href)"
          class="customer-sidebar__link"
          :class="{ 'customer-sidebar__link--active': isActive(item.href) }"
          :title="!showNavLabels ? t(item.key) : undefined"
          @click="isMobileNavOpen = false"
        >
          <Icon :name="item.icon" class="customer-sidebar__link-icon" aria-hidden="true" />
          <span v-if="showNavLabels" class="customer-sidebar__link-text">{{ t(item.key) }}</span>
        </NuxtLink>
      </nav>

      <!-- Footer -->
      <div class="customer-sidebar__footer">
        <div v-if="showNavLabels" class="customer-sidebar__lang">
          <LanguageSwitcher />
        </div>
        <button
          type="button"
          class="customer-sidebar__logout"
          :title="!showNavLabels ? t('customer.nav.logout') : undefined"
          :aria-label="t('customer.nav.logout')"
          @click="handleLogout"
        >
          <Icon name="material-symbols:logout" size="18" aria-hidden="true" />
          <span v-if="showNavLabels">{{ t('customer.nav.logout') }}</span>
        </button>
      </div>
    </aside>

    <!-- Main Content -->
    <div class="customer-main">
      <div class="customer-main__content">
        <!-- Breadcrumb -->
        <div v-if="breadcrumbItems.length > 1" class="customer-breadcrumb">
          <NuxtLink
            v-for="(crumb, idx) in breadcrumbItems"
            :key="idx"
            :to="crumb.to"
            class="customer-breadcrumb__link"
            :class="{ 'customer-breadcrumb__link--current': idx === breadcrumbItems.length - 1 }"
          >
            <span v-if="idx > 0" class="customer-breadcrumb__sep">/</span>
            {{ crumb.label }}
          </NuxtLink>
        </div>

        <slot />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
useHead({ meta: [{ name: 'robots', content: 'noindex, nofollow' }] })

const { user, logout, isPending, isAuthenticated } = useAuth()
const route = useRoute()
const localePath = useLocalePath()
const { t, locale } = useI18n()

const isMobileNavOpen = ref(false)
const isCollapsed = ref(false)
const cartCount = ref(0)
const customerSidebarId = 'customer-sidebar-nav'

const showNavLabels = computed(() => !isCollapsed.value || isMobileNavOpen.value)

const toggleSidebar = () => {
  isCollapsed.value = !isCollapsed.value
}

const api = useApi()
const isCustomer = computed(() =>
  isAuthenticated.value && user?.value?.role === 'customer'
)
const fetchCartCount = async () => {
  if (!isCustomer.value) {
    cartCount.value = 0
    return
  }
  try {
    const res = await api.getCart()
    cartCount.value = res?.items?.length ?? 0
  } catch { cartCount.value = 0 }
}

// Navigation groups
const mainNav = [
  { key: 'customer.nav.dashboard', href: '/customer/dashboard', icon: 'material-symbols:home' },
  { key: 'customer.nav.products', href: '/customer/products', icon: 'material-symbols:package-2' },
  { key: 'customer.nav.cart', href: '/customer/cart', icon: 'material-symbols:shopping-bag' },
]

const orderNav = [
  { key: 'customer.nav.orders', href: '/customer/orders', icon: 'material-symbols:receipt-long' },
  { key: 'customer.nav.quick_order', href: '/customer/orders/quick', icon: 'material-symbols:bolt' },
  { key: 'customer.nav.inquiries', href: '/customer/inquiries', icon: 'material-symbols:inbox' },
  { key: 'customer.nav.quotes', href: '/customer/quotes', icon: 'material-symbols:description' },
  { key: 'customer.nav.trades', href: '/customer/trades', icon: 'material-symbols:attach-money' },
]

const accountNav = [
  { key: 'customer.nav.invoices', href: '/customer/invoices', icon: 'material-symbols:description' },
  { key: 'customer.nav.shipments', href: '/customer/shipments', icon: 'material-symbols:local-shipping' },
  { key: 'customer.nav.returns', href: '/customer/returns', icon: 'material-symbols:assignment-return' },
  { key: 'customer.nav.oemProjects', href: '/customer/oem-projects', icon: 'material-symbols:auto-awesome' },
  { key: 'customer.nav.pricing', href: '/customer/pricing', icon: 'material-symbols:sell' },
  { key: 'customer.nav.company', href: '/customer/company', icon: 'material-symbols:apartment' },
  { key: 'customer.nav.notifications', href: '/customer/notifications', icon: 'material-symbols:notifications' },
  { key: 'customer.nav.profile', href: '/customer/profile', icon: 'material-symbols:person' },
  { key: 'customer.nav.resources', href: '/customer/resources', icon: 'material-symbols:menu-book' },
]

const userInitials = computed(() => {
  const current = user?.value
  if (!current) return 'U'
  return `${current.firstName?.charAt(0) || ''}${current.lastName?.charAt(0) || ''}`
})

const allNavItems = [...mainNav, ...orderNav, ...accountNav]

const isActive = (href: string) => {
  const path = route?.path
  if (!path) return false
  const fullPath = localePath(href)
  if (href === '/customer/dashboard') return path === fullPath

  // 最长前缀匹配：子路由（如 /orders/quick）只激活更具体的导航项
  const matchingPaths = allNavItems
    .map((item) => localePath(item.href))
    .filter((p) => path === p || path.startsWith(`${p}/`))
    .sort((a, b) => b.length - a.length)

  return matchingPaths[0] === fullPath
}

// Breadcrumb from route
const breadcrumbItems = computed(() => {
  const items = [{ label: t('customer.nav.dashboard'), to: localePath('/customer/dashboard') }]
  const path = route?.path
  if (!path) return items

  const segments = [
    { match: '/orders', label: t('customer.nav.orders'), to: localePath('/customer/orders') },
    { match: '/inquiries', label: t('customer.nav.inquiries'), to: localePath('/customer/inquiries') },
    { match: '/quotes', label: t('customer.nav.quotes'), to: localePath('/customer/quotes') },
    { match: '/trades', label: t('customer.nav.trades'), to: localePath('/customer/trades') },
    { match: '/invoices', label: t('customer.nav.invoices'), to: localePath('/customer/invoices') },
    { match: '/shipments', label: t('customer.nav.shipments'), to: localePath('/customer/shipments') },
    { match: '/returns', label: t('customer.nav.returns'), to: localePath('/customer/returns') },
    { match: '/oem-projects', label: t('customer.nav.oemProjects'), to: localePath('/customer/oem-projects') },
    { match: '/pricing', label: t('customer.nav.pricing'), to: localePath('/customer/pricing') },
    { match: '/company', label: t('customer.nav.company'), to: localePath('/customer/company') },
    { match: '/notifications', label: t('customer.nav.notifications'), to: localePath('/customer/notifications') },
    { match: '/profile', label: t('customer.nav.profile'), to: localePath('/customer/profile') },
    { match: '/resources', label: t('customer.nav.resources'), to: localePath('/customer/resources') },
    { match: '/help', label: t('customer.nav.help'), to: localePath('/customer/help') },
    { match: '/products', label: t('customer.nav.products'), to: localePath('/customer/products') },
    { match: '/cart', label: t('customer.nav.cart'), to: localePath('/customer/cart') },
  ]

  for (const seg of segments) {
    if (path.includes(seg.match) && seg.to !== items[0]?.to) {
      items.push(seg)
      break
    }
  }
  return items
})

const handleLogout = async () => {
  isMobileNavOpen.value = false
  await logout()
}

let mediaQuery
const closeMobileIfDesktop = () => {
  if (mediaQuery && !mediaQuery.matches) isMobileNavOpen.value = false
}

watch(isCustomer, (ok) => {
  if (ok) fetchCartCount()
})

watch(route, (newRoute) => {
  isMobileNavOpen.value = false
  const path = newRoute.path
  if (path.includes('/cart') || path.includes('/products') || path.includes('/orders')) {
    fetchCartCount()
  }
})

watch(isMobileNavOpen, (open) => {
  if (import.meta.client) document.body.classList.toggle('body-lock', open)
})

const onEscape = (e) => { if (e.key === 'Escape') isMobileNavOpen.value = false }

onMounted(() => {
  const saved = localStorage.getItem('customer-sidebar-collapsed')
  if (saved) isCollapsed.value = saved === 'true'
  mediaQuery = window.matchMedia('(min-width: 768px)')
  mediaQuery.addEventListener('change', closeMobileIfDesktop)
  window.addEventListener('keydown', onEscape)
  if (isCustomer.value) fetchCartCount()
})

watch(isCollapsed, (val) => {
  if (import.meta.client) localStorage.setItem('customer-sidebar-collapsed', String(val))
})

onUnmounted(() => {
  if (mediaQuery) mediaQuery.removeEventListener('change', closeMobileIfDesktop)
  window.removeEventListener('keydown', onEscape)
  if (import.meta.client) document.body.classList.remove('body-lock')
})
</script>

<style scoped>
/* Pending Banner */
.pending-banner {
  position: sticky;
  top: 0;
  z-index: calc(var(--z-sticky) + 1);
  background: rgba(var(--color-warning-rgb), 0.12);
  border-bottom: 1px solid rgba(var(--color-warning-rgb), 0.25);
}

.pending-banner__inner {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-warning-hover);
}

.pending-banner__icon { width: 18px; height: 18px; flex-shrink: 0; }

/* Layout */
.customer-layout {
  display: flex;
  min-height: 100vh;
  background: var(--color-bg);
}

/* Mobile Top Bar */
.customer-mobile-topbar {
  display: none;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-sm) var(--spacing-md);
  padding-top: max(var(--spacing-sm), env(safe-area-inset-top, 0px));
  background: var(--color-bg);
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  z-index: 570;
}

.customer-mobile-topbar__btn {
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

.customer-mobile-topbar__brand {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-text);
}

.customer-mobile-topbar__cart {
  position: relative;
  min-width: 44px;
  min-height: 44px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-light);
}

.customer-mobile-topbar__cart-badge {
  position: absolute;
  top: 4px;
  right: 4px;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  background: var(--color-highlight);
  color: white;
  font-size: 10px;
  font-weight: 600;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Backdrop */
.customer-sidebar-backdrop { display: none; }
.customer-sidebar-backdrop--visible {
  display: block;
  position: fixed;
  inset: 0;
  z-index: 550;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(2px);
}

@media (min-width: 768px) {
  .customer-sidebar-backdrop--visible { display: none !important; }
}

/* Sidebar */
.customer-sidebar {
  width: 250px;
  height: 100vh;
  position: sticky;
  top: 0;
  background: var(--color-primary-container);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  z-index: 560;
  border-right: 1px solid rgba(255, 255, 255, 0.06);
  transition: width 0.22s ease;
  flex-shrink: 0;
}

/* H4: the sidebar is authored for a dark surface (white text, white logo,
   translucent-white overlays). Dark mode flips --color-primary-container to a
   light #e6e1e0, which would render every sidebar label invisible. Keep the
   sidebar on the dark surface in both themes by re-scoping the token here. */
.dark .customer-sidebar {
  --color-primary-container: #1e1b19;
  --color-primary-container-rgb: 30, 27, 25;
  --color-on-primary-container: #888380;
}

.customer-sidebar--collapsed {
  width: 72px;
}

/* Sidebar Header */
.customer-sidebar__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-xs);
  padding: var(--spacing-lg) var(--spacing-md);
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.customer-sidebar__logo { display: block; flex: 1; min-width: 0; }
.customer-sidebar__logo-svg { height: 34px; width: auto; transition: height 0.22s ease; }
.customer-sidebar__logo-svg--compact { height: 40px; width: 40px; }

.customer-sidebar__toggle {
  display: none;
  min-width: 32px;
  min-height: 32px;
  align-items: center;
  justify-content: center;
  border: none;
  background: rgba(255, 255, 255, 0.06);
  cursor: pointer;
  border-radius: var(--radius-md);
  color: rgba(255, 255, 255, 0.55);
  flex-shrink: 0;
  padding: 0;
}

.customer-sidebar__toggle:hover {
  background: rgba(255, 255, 255, 0.12);
  color: white;
}

.customer-sidebar__toggle-icon { width: 18px; height: 18px; }

@media (min-width: 768px) {
  .customer-sidebar__toggle { display: inline-flex; }
}

.customer-sidebar__close {
  display: none;
  min-width: 44px;
  min-height: 44px;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  cursor: pointer;
  border-radius: var(--radius-md);
  color: rgba(255, 255, 255, 0.5);
}

.customer-sidebar__close:hover {
  background: rgba(255, 255, 255, 0.08);
  color: white;
}

/* User Profile */
.customer-sidebar__user {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-md);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.customer-sidebar__user-avatar {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-full);
  background: linear-gradient(135deg, var(--color-highlight), var(--color-accent));
  color: white;
  font-size: var(--text-base);
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.customer-sidebar__user--collapsed {
  justify-content: center;
  padding-inline: var(--spacing-sm);
}

.customer-sidebar__user-info { overflow: hidden; }

.customer-sidebar__user-name {
  font-size: var(--text-sm);
  font-weight: 500;
  color: rgba(255, 255, 255, 0.85);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.customer-sidebar__user-role {
  font-size: var(--text-xs);
  color: rgba(255, 255, 255, 0.35);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.customer-sidebar__dot-badge {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 8px;
  height: 8px;
  border-radius: var(--radius-full);
  background: var(--color-highlight);
}

.customer-sidebar--collapsed .customer-sidebar__link {
  justify-content: center;
  padding-inline: var(--spacing-sm);
  transform: none;
}

.customer-sidebar--collapsed .customer-sidebar__link--active {
  transform: none;
}

.customer-sidebar--collapsed .customer-sidebar__link {
  position: relative;
}

.customer-sidebar--collapsed .customer-sidebar__footer {
  padding-inline: var(--spacing-sm);
}

.customer-sidebar--collapsed .customer-sidebar__logout {
  justify-content: center;
  padding-inline: var(--spacing-sm);
}

[dir="rtl"] .customer-sidebar--collapsed .customer-sidebar__link--active {
  transform: none;
}

@media (min-width: 768px) {
  .customer-sidebar--collapsed .customer-sidebar__close { display: none; }
}

/* Navigation */
.customer-sidebar__nav {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-md) 0;
}

.customer-sidebar__nav::-webkit-scrollbar { width: 4px; }
.customer-sidebar__nav::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.12);
  border-radius: 2px;
}

.customer-sidebar__group-label {
  padding: var(--spacing-xs) var(--spacing-md);
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: rgba(255, 255, 255, 0.3);
}

.customer-sidebar__divider {
  height: 1px;
  background: rgba(255, 255, 255, 0.06);
  margin: var(--spacing-sm) var(--spacing-md);
}

.customer-sidebar__link {
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
  text-decoration: none;
}

.customer-sidebar__link:hover {
  background: rgba(255, 255, 255, 0.06);
  color: rgba(255, 255, 255, 0.85);
}

.customer-sidebar__link--active {
  background: var(--color-highlight);
  color: white;
  font-weight: 600;
  transform: translateX(4px);
}

.customer-sidebar__link--active:hover {
  background: var(--color-highlight);
  color: white;
}

.customer-sidebar__link-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.customer-sidebar__link-text {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.customer-sidebar__badge {
  margin-left: auto;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  background: var(--color-highlight);
  color: white;
  font-size: 11px;
  font-weight: 600;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
}

.customer-sidebar__link--active .customer-sidebar__badge {
  background: rgba(255, 255, 255, 0.25);
}

/* Sidebar Footer */
.customer-sidebar__footer {
  padding: var(--spacing-md);
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.customer-sidebar__lang {
  display: block;
  margin-bottom: var(--spacing-sm);
  background: rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-md);
  padding: 3px;
  overflow: hidden;
  min-width: 0;
}

.customer-sidebar__lang :deep(.lang-switcher) {
  display: block;
  width: 100%;
}

.customer-sidebar__lang :deep(.lang-switcher__trigger) {
  width: 100%;
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.85);
  padding: 8px 10px;
}

.customer-sidebar__lang :deep(.lang-switcher__trigger:hover) {
  background: rgba(255, 255, 255, 0.08);
  border-color: transparent;
}

.customer-sidebar__lang :deep(.lang-switcher__chevron) {
  color: rgba(255, 255, 255, 0.45);
}

.customer-sidebar__lang-btn {
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

.customer-sidebar__lang-btn:hover { color: rgba(255, 255, 255, 0.7); }
.customer-sidebar__lang-btn--active {
  background: rgba(255, 255, 255, 0.12);
  color: white;
}

.customer-sidebar__logout {
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

.customer-sidebar__logout:hover {
  background: rgba(186, 26, 26, 0.15);
  color: var(--color-error);
}

/* Main Content Area */
.customer-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.customer-main__content {
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
  .customer-main__content { animation: none; }
}

/* Breadcrumb */
.customer-breadcrumb {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  margin-bottom: var(--spacing-lg);
  font-size: var(--text-sm);
  color: var(--color-text-lighter);
}

.customer-breadcrumb__link {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  color: var(--color-text-lighter);
  text-decoration: none;
  transition: color var(--transition-fast);
}

.customer-breadcrumb__link:hover {
  color: var(--color-highlight);
}

.customer-breadcrumb__link--current {
  color: var(--color-text);
  font-weight: 500;
}

.customer-breadcrumb__sep {
  color: var(--color-border);
}

/* RTL */
[dir="rtl"] .customer-sidebar {
  border-right: none;
  border-left: 1px solid rgba(255, 255, 255, 0.06);
}

[dir="rtl"] .customer-sidebar__link--active {
  transform: translateX(-4px);
}

/* Responsive */
@media (max-width: 767px) {
  .customer-layout { flex-direction: column; }
  .customer-mobile-topbar { display: flex; }

  .customer-sidebar {
    position: fixed;
    left: 0;
    top: 0;
    bottom: 0;
    width: min(300px, 88vw);
    height: 100dvh;
    transform: translateX(-100%);
    box-shadow: none;
  }

  [dir="rtl"] .customer-sidebar {
    left: auto;
    right: 0;
    transform: translateX(100%);
  }

  .customer-sidebar--mobile-open {
    transform: translateX(0);
    box-shadow: 8px 0 32px rgba(0, 0, 0, 0.25);
  }

  [dir="rtl"] .customer-sidebar--mobile-open {
    transform: translateX(0);
    box-shadow: -8px 0 32px rgba(0, 0, 0, 0.25);
  }

  .customer-sidebar__close { display: flex; }

  .customer-main__content {
    padding: var(--spacing-md);
    padding-bottom: max(var(--spacing-xl), env(safe-area-inset-bottom, 0px));
  }
}
</style>
