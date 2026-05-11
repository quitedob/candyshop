<template>
  <div class="customer-layout">
    <!-- Pending Approval Banner -->
    <div v-if="isPending" class="pending-banner">
      <div class="container">
        <div class="pending-banner__inner">
          <Icon name="heroicons:clock" class="pending-banner__icon" />
          <span>{{ t('customer.pending.banner') }}</span>
        </div>
      </div>
    </div>

    <!-- Header：有待审核横幅时顶栏下移，避免与横幅同为 sticky top:0 重叠 -->
    <header
      class="customer-header"
      :class="{
        'customer-header--scrolled': isScrolled,
        'customer-header--with-pending': isPending
      }"
    >
      <div class="container">
        <div class="customer-header__inner">
          <!-- Logo -->
          <NuxtLink :to="localePath('/')" class="customer-header__logo">
            <svg viewBox="0 0 180 40" fill="none" class="customer-header__logo-svg">
              <circle cx="20" cy="20" r="16" fill="var(--color-highlight)" opacity="0.2"/>
              <path d="M14 20C14 16.6863 16.6863 14 20 14C23.3137 14 26 16.6863 26 20C26 23.3137 23.3137 26 20 26C16.6863 26 14 23.3137 14 20Z" stroke="var(--color-highlight)" stroke-width="2.5"/>
              <circle cx="17" cy="17" r="2" fill="var(--color-accent)"/>
              <circle cx="23" cy="23" r="2" fill="var(--color-primary)"/>
              <text x="45" y="27" font-family="Outfit, system-ui, sans-serif" font-size="18" font-weight="700" fill="var(--color-primary)">CandyPro</text>
              <text x="145" y="27" font-family="Outfit, system-ui, sans-serif" font-size="10" font-weight="600" fill="var(--color-accent)">OEM</text>
            </svg>
          </NuxtLink>

          <!-- Desktop / tablet: scroll when many items (avoids overflow on narrow widths) -->
          <nav class="customer-header__nav hide-mobile" :aria-label="t('customer.a11y.portalNav')">
            <ul class="customer-header__nav-list">
              <li v-for="item in navigation" :key="item.key" class="customer-header__nav-item">
                <NuxtLink
                  :to="localePath(item.href)"
                  class="customer-header__nav-link"
                  :class="{ 'customer-header__nav-link--active': isActive(item.href) }"
                >
                  {{ $t(item.key) }}
                </NuxtLink>
              </li>
            </ul>
          </nav>

          <!-- Right Side Actions -->
          <div class="customer-header__actions">
            <!-- User Menu -->
            <div class="customer-header__user">
              <div class="customer-header__user-avatar">
                {{ userInitials }}
              </div>
              <span class="customer-header__user-name hide-mobile">{{ user?.firstName }}</span>
            </div>

            <!-- Cart / Order Request -->
            <NuxtLink
              :to="localePath('/customer/cart')"
              class="customer-header__cart"
              :aria-label="t('customer.a11y.cart')"
            >
              <Icon name="heroicons:shopping-bag" class="customer-header__cart-icon" aria-hidden="true" />
              <span v-if="cartCount > 0" class="customer-header__cart-badge">{{ cartCount }}</span>
            </NuxtLink>

            <!-- Logout Button -->
            <button
              type="button"
              class="customer-header__logout hide-mobile"
              :title="t('customer.nav.logout')"
              :aria-label="t('customer.nav.logout')"
              @click="handleLogout"
            >
              <Icon name="heroicons:arrow-right-on-rectangle" class="h-5 w-5" aria-hidden="true" />
            </button>

            <!-- Mobile Menu Toggle -->
            <button
              type="button"
              class="customer-header__menu-toggle hide-desktop"
              :aria-expanded="isMenuOpen ? 'true' : 'false'"
              :aria-controls="customerMobileNavId"
              :aria-label="isMenuOpen ? t('customer.a11y.closeMenu') : t('customer.a11y.openMenu')"
              @click="isMenuOpen = !isMenuOpen"
            >
              <Icon :name="isMenuOpen ? 'heroicons:x-mark' : 'heroicons:bars-3'" class="customer-header__menu-icon" aria-hidden="true" />
            </button>
          </div>
        </div>
      </div>

      <!-- Mobile Navigation -->
      <Transition name="slide-down">
        <div
          v-if="isMenuOpen"
          :id="customerMobileNavId"
          class="customer-header__mobile-nav hide-desktop"
          role="navigation"
          :aria-label="t('customer.brand')"
        >
          <nav class="customer-header__mobile-nav-inner">
            <NuxtLink
              v-for="item in navigation"
              :key="item.key"
              :to="localePath(item.href)"
              class="customer-header__mobile-link"
              :class="{ 'customer-header__mobile-link--active': isActive(item.href) }"
              @click="isMenuOpen = false"
            >
              {{ $t(item.key) }}
            </NuxtLink>
            <button
              type="button"
              class="customer-header__mobile-link customer-header__mobile-link--logout"
              @click="handleLogout"
            >
              {{ $t('customer.nav.logout') }}
            </button>
          </nav>
        </div>
      </Transition>
    </header>

    <!-- 移动端菜单打开时遮罩，点击关闭并防止误触主内容 -->
    <div
      class="customer-header__backdrop hide-desktop"
      :class="{ 'customer-header__backdrop--visible': isMenuOpen }"
      aria-hidden="true"
      @click="isMenuOpen = false"
    />

    <!-- Page Header -->
    <div class="customer-content">
      <div class="container">
        <header class="customer-content__header">
          <h1 class="customer-content__title">{{ pageTitle }}</h1>
        </header>
        <main class="customer-content__main">
          <slot />
        </main>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'

useSeo({ noindex: true })

const { user, logout, isPending } = useAuth()
const route = useRoute()
const localePath = useLocalePath()
const { t } = useI18n()

const isScrolled = ref(false)
const isMenuOpen = ref(false)
const cartCount = ref(0)
const customerMobileNavId = 'customer-mobile-nav'

const api = useApi()
const fetchCartCount = async () => {
  try {
    const res = await api.getCart()
    cartCount.value = res?.items?.length ?? 0
  } catch {
    cartCount.value = 0
  }
}

// Close mobile menu and refresh cart on route change
watch(() => route.path, (newPath) => {
  isMenuOpen.value = false
  if (newPath.includes('/cart') || newPath.includes('/products') || newPath.includes('/orders')) {
    fetchCartCount()
  }
})

watch(isMenuOpen, (open) => {
  if (import.meta.client) {
    document.body.classList.toggle('body-lock', open)
  }
})

const onEscape = (e) => {
  if (e.key === 'Escape') {
    isMenuOpen.value = false
  }
}

// Navigation with i18n keys (matching customer.json)
const navigation = [
  { key: 'customer.nav.dashboard', href: '/customer/dashboard' },
  { key: 'customer.nav.products', href: '/customer/products' },
  { key: 'customer.nav.inquiries', href: '/customer/inquiries' },
  { key: 'customer.nav.orders', href: '/customer/orders' },
  { key: 'customer.nav.cart', href: '/customer/cart' },
  { key: 'customer.nav.oemProjects', href: '/customer/oem-projects' },
  { key: 'customer.nav.quotes', href: '/customer/quotes' },
  { key: 'customer.nav.trades', href: '/customer/trades' },
  { key: 'customer.nav.invoices', href: '/customer/invoices' },
  { key: 'customer.nav.shipments', href: '/customer/shipments' },
  { key: 'customer.nav.pricing', href: '/customer/pricing' },
  { key: 'customer.nav.company', href: '/customer/company' },
  { key: 'customer.nav.notifications', href: '/customer/notifications' },
  { key: 'customer.nav.profile', href: '/customer/profile' },
  { key: 'customer.nav.help', href: '/customer/help' },
  { key: 'customer.nav.resources', href: '/customer/resources' },
]

// Page title based on route
const pageTitle = computed(() => {
  const path = route.path
  if (path.includes('/orders')) return t('customer.page_titles.orders')
  if (path.includes('/inquiries')) return t('customer.page_titles.inquiries')
  if (path.includes('/quotes')) return t('customer.page_titles.quotes')
  if (path.includes('/notifications')) return t('customer.page_titles.notifications')
  if (path.includes('/trades')) return t('customer.page_titles.trades')
  if (path.includes('/invoices')) return t('customer.page_titles.invoices')
  if (path.includes('/shipments')) return t('customer.page_titles.shipments')
  if (path.includes('/company')) return t('customer.page_titles.company')
  if (path.includes('/pricing')) return t('customer.page_titles.pricing')
  if (path.includes('/profile')) return t('customer.page_titles.profile')
  if (path.includes('/help')) return t('customer.page_titles.help')
  if (path.includes('/resources')) return t('customer.page_titles.resources')
  if (path.includes('/cart')) return t('customer.cart.title')
  if (path.includes('/products')) return t('customer.products.title')
  if (path.includes('/oem-projects')) return t('customer.oemProjects.title')
  return t('customer.page_titles.dashboard')
})

// User initials for avatar
const userInitials = computed(() => {
  if (!user.value) return 'U'
  return `${user.value.firstName?.charAt(0) || ''}${user.value.lastName?.charAt(0) || ''}`
})

// Check if nav item is active
const isActive = (href) => {
  const fullPath = localePath(href)
  return route.path.startsWith(fullPath) || route.path === fullPath
}

// Handle scroll for header shadow
const handleScroll = () => {
  isScrolled.value = window.scrollY > 10
}

const handleLogout = async () => {
  await logout()
}

onMounted(() => {
  window.addEventListener('scroll', handleScroll)
  window.addEventListener('keydown', onEscape)
  fetchCartCount()
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
  window.removeEventListener('keydown', onEscape)
  if (import.meta.client) {
    document.body.classList.remove('body-lock')
  }
})
</script>

<style scoped>
/* Pending Banner */
.pending-banner {
  position: sticky;
  top: 0;
  z-index: calc(var(--z-sticky) + 1);
  background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
  border-bottom: 1px solid #f59e0b;
}

.pending-banner__inner {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-sm);
  font-weight: 500;
  color: #92400e;
}

.pending-banner__icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.customer-layout {
  min-height: 100vh;
  background-color: var(--color-bg);
}

/* Header */
.customer-header {
  position: sticky;
  top: 0;
  z-index: var(--z-sticky);
  background: var(--color-bg);
  border-bottom: 1px solid var(--color-border-light);
  transition: box-shadow var(--transition-base);
}

.customer-header--scrolled {
  box-shadow: var(--shadow-md);
}

.customer-header--with-pending {
  top: 2.75rem;
}

.customer-header__inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 72px;
}

.customer-header__logo {
  display: block;
}

.customer-header__logo-svg {
  height: 40px;
  width: auto;
}

/* Navigation */
.customer-header__nav {
  flex: 1 1 auto;
  min-width: 0;
  margin: 0 var(--spacing-sm);
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: thin;
}

.customer-header__nav-list {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  flex-wrap: nowrap;
  width: max-content;
  min-height: 44px;
}

.customer-header__nav-link {
  display: block;
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text-light);
  border-radius: var(--radius-md);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.customer-header__nav-link:hover {
  color: var(--color-primary);
  background: var(--color-bg-alt);
}

.customer-header__nav-link--active {
  color: var(--color-highlight);
  background: rgba(var(--color-highlight-rgb), 0.08);
}

/* Actions */
.customer-header__actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.customer-header__user {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.customer-header__user-avatar {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  background: linear-gradient(135deg, var(--color-accent), var(--color-accent-dark));
  color: white;
  font-size: var(--text-sm);
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}

.customer-header__user-name {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
}

.customer-header__cart {
  position: relative;
  min-width: 44px;
  min-height: 44px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-sm);
  color: var(--color-text-light);
  transition: color var(--transition-fast);
  box-sizing: border-box;
}

.customer-header__cart:hover {
  color: var(--color-highlight);
}

.customer-header__cart-icon {
  width: 24px;
  height: 24px;
}

.customer-header__cart-badge {
  position: absolute;
  top: 2px;
  right: 2px;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  background: var(--color-highlight);
  color: white;
  font-size: 11px;
  font-weight: 600;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
}

.customer-header__menu-toggle {
  min-width: 44px;
  min-height: 44px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-sm);
  color: var(--color-text);
}

.customer-header__logout {
  min-width: 44px;
  min-height: 44px;
  padding: var(--spacing-sm);
  color: var(--color-text-light);
  transition: color var(--transition-fast);
  background: none;
  border: none;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
}

.customer-header__logout:hover {
  color: var(--color-error, #ef4444);
}

.customer-header__menu-icon {
  width: 24px;
  height: 24px;
}

/* Mobile Navigation */
.customer-header__mobile-nav {
  background: var(--color-bg);
  border-top: 1px solid var(--color-border-light);
  padding: var(--spacing-md) 0;
}

.customer-header__mobile-nav-inner {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.customer-header__mobile-link {
  display: flex;
  align-items: center;
  min-height: 44px;
  line-height: 1.3;
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-base);
  font-weight: 500;
  color: var(--color-text-light);
  border-radius: var(--radius-md);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
  box-sizing: border-box;
}

.customer-header__mobile-link:hover,
.customer-header__mobile-link--active {
  color: var(--color-highlight);
  background: rgba(var(--color-highlight-rgb), 0.08);
}

.customer-header__mobile-link--logout {
  margin-top: var(--spacing-xs);
  border: none;
  width: 100%;
  cursor: pointer;
  background: transparent;
  font: inherit;
  justify-content: flex-start;
  color: var(--color-error, #ef4444);
}

.customer-header__backdrop {
  display: none;
  position: fixed;
  inset: 0;
  z-index: calc(var(--z-sticky) - 1);
  background: rgba(0, 0, 0, 0.35);
}

.customer-header__backdrop--visible {
  display: block;
}

@media (min-width: 768px) {
  .customer-header__backdrop--visible {
    display: none !important;
  }
}

/* Content Area */
.customer-content {
  padding: var(--spacing-xl) 0 var(--spacing-3xl);
}

.customer-content__header {
  margin-bottom: var(--spacing-xl);
}

.customer-content__title {
  font-family: var(--font-display);
  font-size: var(--text-3xl);
  font-weight: 700;
  color: var(--color-primary);
  margin: 0;
}

.customer-content__main {
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

/* Slide Transition */
.slide-down-enter-active,
.slide-down-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.slide-down-enter-from,
.slide-down-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

/* Responsive */
@media (max-width: 768px) {
  .customer-header__inner {
    height: 64px;
  }

  .customer-content {
    padding: var(--spacing-lg) 0 var(--spacing-2xl);
  }

  .customer-content__title {
    font-size: var(--text-2xl);
  }
}
</style>
