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

    <!-- Header -->
    <header class="customer-header" :class="{ 'customer-header--scrolled': isScrolled }">
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

          <!-- Desktop Navigation -->
          <nav class="customer-header__nav hide-mobile">
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

            <!-- Cart Badge -->
            <NuxtLink :to="localePath('/customer/cart')" class="customer-header__cart">
              <Icon name="heroicons:shopping-bag" class="customer-header__cart-icon" />
              <span v-if="cartCount > 0" class="customer-header__cart-badge">{{ cartCount }}</span>
            </NuxtLink>

            <!-- Logout Button -->
            <button @click="handleLogout" class="customer-header__logout hide-mobile" :title="t('customer.nav.logout')">
              <Icon name="heroicons:arrow-right-on-rectangle" class="h-5 w-5" />
            </button>

            <!-- Mobile Menu Toggle -->
            <button
              class="customer-header__menu-toggle hide-desktop"
              @click="isMenuOpen = !isMenuOpen"
              aria-label="Toggle menu"
            >
              <Icon :name="isMenuOpen ? 'heroicons:x-mark' : 'heroicons:bars-3'" class="customer-header__menu-icon" />
            </button>
          </div>
        </div>
      </div>

      <!-- Mobile Navigation -->
      <Transition name="slide-down">
        <div v-if="isMenuOpen" class="customer-header__mobile-nav hide-desktop">
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
            <button @click="handleLogout" class="customer-header__mobile-link text-left text-red-600">
              {{ $t('customer.nav.logout') }}
            </button>
          </nav>
        </div>
      </Transition>
    </header>

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

const { user, logout, isPending } = useAuth()
const route = useRoute()
const localePath = useLocalePath()
const { t } = useI18n()

const isScrolled = ref(false)
const isMenuOpen = ref(false)
const cartCount = ref(0)

// Fetch cart count from API
const api = useApi()
const fetchCartCount = async () => {
  try {
    const res = await api.getCart()
    cartCount.value = res?.itemCount ?? 0
  } catch {
    cartCount.value = 0
  }
}

// Refresh cart count and close mobile menu on route change
watch(() => route.path, (newPath) => {
  isMenuOpen.value = false
  if (newPath.includes('/cart') || newPath.includes('/products') || newPath.includes('/orders')) {
    fetchCartCount()
  }
})

// Navigation with i18n keys (matching customer.json)
const navigation = [
  { key: 'customer.nav.dashboard', href: '/customer/dashboard' },
  { key: 'customer.nav.products', href: '/customer/products' },
  { key: 'customer.nav.cart', href: '/customer/cart' },
  { key: 'customer.nav.orders', href: '/customer/orders' },
  { key: 'customer.nav.inquiries', href: '/customer/inquiries' },
  { key: 'customer.nav.oemProjects', href: '/customer/oem-projects' },
  { key: 'customer.nav.quotes', href: '/customer/quotes' },
  { key: 'customer.nav.trades', href: '/customer/trades' },
  { key: 'customer.nav.invoices', href: '/customer/invoices' },
  { key: 'customer.nav.shipments', href: '/customer/shipments' },
  { key: 'customer.nav.pricing', href: '/customer/pricing' },
  { key: 'customer.nav.company', href: '/customer/company' },
  { key: 'customer.nav.notifications', href: '/customer/notifications' },
  { key: 'customer.nav.profile', href: '/customer/profile' },
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
  fetchCartCount()
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
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
.customer-header__nav-list {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.customer-header__nav-link {
  display: block;
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text-light);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.customer-header__nav-link:hover {
  color: var(--color-primary);
  background: var(--color-bg-alt);
}

.customer-header__nav-link--active {
  color: var(--color-highlight);
  background: rgba(255, 107, 74, 0.08);
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
  padding: var(--spacing-sm);
  color: var(--color-text-light);
  transition: color var(--transition-fast);
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
  padding: var(--spacing-sm);
  color: var(--color-text);
}

.customer-header__logout {
  padding: var(--spacing-sm);
  color: var(--color-text-light);
  transition: color var(--transition-fast);
  background: none;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
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
  display: block;
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-base);
  font-weight: 500;
  color: var(--color-text-light);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.customer-header__mobile-link:hover,
.customer-header__mobile-link--active {
  color: var(--color-highlight);
  background: rgba(255, 107, 74, 0.08);
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
  transition: all 0.3s ease;
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
