<template>
  <header
    ref="headerRef"
    :class="['header', { 'header--scrolled': isScrolled, 'header--menu-open': isMenuOpen }]"
  >
    <div class="container">
      <div class="header__inner">
        <!-- Logo -->
        <NuxtLink
          :to="localePath('/')"
          class="header__logo"
          aria-label="Go to homepage"
        >
          <svg
            viewBox="0 0 180 40"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            class="header__logo-svg"
          >
            <!-- Simple candy/swirl logo -->
            <circle cx="20" cy="20" r="16" fill="var(--color-highlight)" opacity="0.2"/>
            <path
              d="M14 20C14 16.6863 16.6863 14 20 14C23.3137 14 26 16.6863 26 20C26 23.3137 23.3137 26 20 26C16.6863 26 14 23.3137 14 20Z"
              stroke="var(--color-highlight)"
              stroke-width="2.5"
            />
            <circle cx="17" cy="17" r="2" fill="var(--color-accent)"/>
            <circle cx="23" cy="23" r="2" fill="var(--color-primary)"/>

            <!-- Brand name -->
            <text
              x="45"
              y="27"
              font-family="Inter, system-ui, sans-serif"
              font-size="18"
              font-weight="700"
              fill="var(--color-primary)"
            >
              CandyPro
            </text>
            <text
              x="145"
              y="27"
              font-family="Inter, system-ui, sans-serif"
              font-size="10"
              font-weight="500"
              fill="var(--color-accent)"
            >
              OEM
            </text>
          </svg>
        </NuxtLink>

        <!-- Desktop Navigation -->
        <nav class="header__nav hide-mobile" aria-label="Main navigation">
          <ul class="header__nav-list">
            <li v-for="item in navItems" :key="item.key" class="header__nav-item">
              <NuxtLink
                :to="localePath(item.to)"
                class="header__nav-link"
                :class="{ 'header__nav-link--active': isActive(item.to) }"
              >
                {{ $t(item.key) }}
                <svg
                  v-if="item.children"
                  class="header__nav-arrow"
                  width="8"
                  height="5"
                  viewBox="0 0 8 5"
                  fill="none"
                >
                  <path d="M1 1L4 4L7 1" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </NuxtLink>

              <!-- Dropdown Menu -->
              <div v-if="item.children" class="header__dropdown">
                <ul class="header__dropdown-list">
                  <li v-for="child in item.children" :key="child.to" class="header__dropdown-item">
                    <NuxtLink
                      :to="localePath(child.to)"
                      class="header__dropdown-link"
                    >
                      {{ $t(child.key) }}
                    </NuxtLink>
                  </li>
                </ul>
              </div>
            </li>
          </ul>
        </nav>

        <!-- Right Side Actions -->
        <div class="header__actions">
          <!-- Language Switcher -->
          <div class="header__lang hide-mobile">
            <button
              v-for="locale in availableLocales"
              :key="locale.code"
              :class="['header__lang-btn', { 'header__lang-btn--active': locale.code === currentLocale }]"
              @click="switchLocale(locale.code)"
            >
              {{ locale.code.toUpperCase() }}
            </button>
          </div>

          <!-- CTA Buttons -->
          <div class="header__cta hide-mobile">
            <a
              :href="whatsappUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-ghost btn-sm header__cta-btn"
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
                <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z"/>
              </svg>
              <span>{{ $t('whatsapp.us') }}</span>
            </a>
            <!-- Auth-conditional CTAs -->
            <template v-if="!isAuthenticated">
              <NuxtLink :to="localePath('/auth/login')" class="btn btn-ghost btn-sm">
                {{ $t('nav.sign_in') }}
              </NuxtLink>
              <NuxtLink :to="localePath('/auth/register')" class="btn btn-highlight btn-sm">
                {{ $t('nav.register') }}
              </NuxtLink>
            </template>
            <template v-else>
              <NuxtLink
                :to="localePath(isAdmin ? '/admin' : '/customer/dashboard')"
                class="btn btn-highlight btn-sm"
              >
                {{ $t(isAdmin ? 'nav.admin_panel' : 'nav.my_account') }}
              </NuxtLink>
            </template>
          </div>

          <!-- Mobile Menu Toggle -->
          <button
            ref="menuToggleRef"
            class="header__toggle hide-desktop"
            :class="{ 'header__toggle--active': isMenuOpen }"
            @click="toggleMenu"
            aria-label="Toggle menu"
            :aria-expanded="isMenuOpen ? 'true' : 'false'"
          >
            <span></span>
            <span></span>
            <span></span>
          </button>
        </div>
      </div>
    </div>

    <!-- Mobile Menu Backdrop -->
    <div 
      class="header__mobile-backdrop hide-desktop" 
      :class="{ 'header__mobile-backdrop--active': isMenuOpen }"
      @click="closeMenu"
      aria-hidden="true"
    ></div>

    <!-- Mobile Menu -->
    <div class="header__mobile" :class="{ 'header__mobile--open': isMenuOpen }">
      <!-- Mobile Close Button inside menu -->
      <div class="header__mobile-header">
        <span class="header__mobile-title">Menu</span>
        <button class="header__mobile-close" @click="closeMenu" aria-label="Close menu">
          <Icon name="lucide:x" size="24" />
        </button>
      </div>
      
      <nav class="header__mobile-nav" aria-label="Mobile navigation">
        <ul class="header__mobile-list">
          <li v-for="item in navItems" :key="item.key" class="header__mobile-item">
            <NuxtLink
              :to="localePath(item.to)"
              class="header__mobile-link"
              :class="{ 'header__mobile-link--active': isActive(item.to) }"
              @click="closeMenu"
            >
              {{ $t(item.key) }}
            </NuxtLink>

            <!-- Mobile Submenu -->
            <ul v-if="item.children" class="header__mobile-submenu">
              <li v-for="child in item.children" :key="child.to">
                <NuxtLink
                  :to="localePath(child.to)"
                  class="header__mobile-sublink"
                  @click="closeMenu"
                >
                  {{ $t(child.key) }}
                </NuxtLink>
              </li>
            </ul>
          </li>
        </ul>

        <!-- Mobile Language & CTA -->
        <div class="header__mobile-actions">
          <!-- Mobile Language Switcher -->
          <div class="header__mobile-lang-switcher">
            <button
              v-for="loc in availableLocales"
              :key="loc.code"
              :class="['header__mobile-lang-btn', { 'header__mobile-lang-btn--active': loc.code === currentLocale }]"
              @click="switchLocale(loc.code)"
            >
              {{ loc.code.toUpperCase() }}
            </button>
          </div>

          <a
            :href="whatsappUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-ghost btn-sm"
          >
            <WhatsAppIcon size="18" />
            <span>{{ $t('whatsapp.us') }}</span>
          </a>

          <template v-if="!isAuthenticated">
            <NuxtLink :to="localePath('/auth/login')" class="btn btn-ghost btn-sm" @click="closeMenu">
              {{ $t('nav.sign_in') }}
            </NuxtLink>
            <NuxtLink :to="localePath('/auth/register')" class="btn btn-highlight btn-sm" @click="closeMenu">
              {{ $t('nav.register') }}
            </NuxtLink>
          </template>
          <template v-else>
            <NuxtLink
              :to="localePath(isAdmin ? '/admin' : '/customer/dashboard')"
              class="btn btn-highlight btn-sm"
              @click="closeMenu"
            >
              {{ $t(isAdmin ? 'nav.admin_panel' : 'nav.my_account') }}
            </NuxtLink>
          </template>
        </div>
      </nav>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useI18n, useLocalePath, useSwitchLocalePath } from '#i18n'

const { t, locale } = useI18n()
const localePath = useLocalePath()
const switchLocalePath = useSwitchLocalePath()
const route = useRoute()
const config = useRuntimeConfig()
const { isAuthenticated, isAdmin, initAuth } = useAuth()

// State
const isScrolled = ref(false)
const isMenuOpen = ref(false)
const menuToggleRef = ref<HTMLButtonElement | null>(null)

// Initialise auth state and set up scroll listener + saved locale
onMounted(() => {
  initAuth()

  const savedLocale = getSavedLocale()
  if (savedLocale && savedLocale !== locale.value) {
    locale.value = savedLocale
  }

  window.addEventListener('scroll', handleScroll)
})

// Navigation items
const navItems = [
  { key: 'nav.home', to: '/' },
  {
    key: 'nav.products',
    to: '/products',
    children: [
      { key: 'product.categories.gummy_candy', to: '/products/gummy-candy' },
      { key: 'product.categories.hard_candy', to: '/products/hard-candy' },
      { key: 'product.categories.aerated_candy', to: '/products/aerated-candy' },
      { key: 'product.categories.toffee_candy', to: '/products/toffee-candy' },
      { key: 'product.categories.compound_chocolate', to: '/products/compound-chocolate' }
    ]
  },
  { key: 'nav.oem', to: '/oem-solutions' },
  { key: 'nav.factory', to: '/factory-quality' },
  { key: 'nav.cases', to: '/cases-clients' },
  { key: 'nav.blog', to: '/blog' },
  { key: 'nav.about', to: '/about' },
  { key: 'nav.contact', to: '/contact' }
]

// Available locales
const availableLocales = [
  { code: 'en', name: 'English' },
  { code: 'zh', name: '中文' }
]

// Current locale
const currentLocale = computed(() => locale.value)

// WhatsApp URL
const whatsappUrl = computed(() => {
  const number = config.public.whatsappNumber
  const message = encodeURIComponent(t('whatsapp.message'))
  return `https://wa.me/${number}?text=${message}`
})

// Check if route is active
const isActive = (to: string) => {
  if (to === '/') {
    return route.path === '/' || route.path === '' || route.path.startsWith(`/${locale.value}`)
  }
  return route.path.startsWith(to) || route.path.startsWith(`/${locale.value}${to}`)
}

// Storage key for locale
const LOCALE_STORAGE_KEY = 'user-locale'

// Get saved locale from localStorage
const getSavedLocale = (): string | null => {
  if (import.meta.client) {
    return localStorage.getItem(LOCALE_STORAGE_KEY)
  }
  return null
}

// Save locale to localStorage
const saveLocale = (newLocale: string) => {
  if (import.meta.client) {
    localStorage.setItem(LOCALE_STORAGE_KEY, newLocale)
  }
}

// Switch locale with persistence
const switchLocale = async (newLocale: string) => {
  saveLocale(newLocale)
  // Navigate to the localized path
  const path = switchLocalePath(newLocale)
  await navigateTo(path)
}

// Toggle mobile menu
const toggleMenu = () => {
  isMenuOpen.value = !isMenuOpen.value
  if (isMenuOpen.value) {
    document.body.classList.add('body-lock')
  } else {
    document.body.classList.remove('body-lock')
  }
}

const closeMenu = () => {
  isMenuOpen.value = false
  document.body.classList.remove('body-lock')
  nextTick(() => menuToggleRef.value?.focus())
}

// Handle scroll
const handleScroll = () => {
  isScrolled.value = window.scrollY > 50
}

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<style scoped>
.header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: var(--z-sticky);
  background-color: transparent;
  transition: background-color var(--transition-base), box-shadow var(--transition-base);
}

.header--scrolled,
.header--menu-open {
  background-color: rgba(255, 255, 255, 0.98);
  backdrop-filter: blur(10px);
  box-shadow: var(--shadow-sm);
}

.header__inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: var(--header-height);
  transition: height var(--transition-base);
}

.header--scrolled .header__inner {
  height: var(--header-height-scrolled);
}

/* Logo */
.header__logo {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.header__logo-svg {
  width: 150px;
  height: 40px;
  transition: transform var(--transition-base);
}

.header__logo:hover .header__logo-svg {
  transform: scale(1.02);
}

/* Desktop Navigation */
.header__nav {
  flex: 1;
  display: flex;
  justify-content: center;
}

.header__nav-list {
  display: flex;
  align-items: center;
  gap: var(--spacing-xl);
}

.header__nav-item {
  position: relative;
}

.header__nav-link {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
  padding: var(--spacing-sm) 0;
  position: relative;
  transition: color var(--transition-fast);
}

.header__nav-link::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  width: 0;
  height: 2px;
  background-color: var(--color-highlight);
  transition: width var(--transition-base);
}

.header__nav-link:hover,
.header__nav-link--active {
  color: var(--color-highlight);
}

.header__nav-link:hover::after,
.header__nav-link--active::after {
  width: 100%;
}

.header__nav-arrow {
  transition: transform var(--transition-fast);
}

.header__nav-item:hover .header__nav-arrow {
  transform: rotate(180deg);
}

/* Dropdown */
.header__dropdown {
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%) translateY(10px);
  min-width: 200px;
  padding: var(--spacing-sm) 0;
  background-color: white;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  opacity: 0;
  visibility: hidden;
  transition: all var(--transition-base);
}

.header__nav-item:hover .header__dropdown {
  opacity: 1;
  visibility: visible;
  transform: translateX(-50%) translateY(0);
}

.header__dropdown-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.header__dropdown-item {
  border-bottom: 1px solid var(--color-border-light);
}

.header__dropdown-item:last-child {
  border-bottom: none;
}

.header__dropdown-link {
  display: block;
  padding: var(--spacing-sm) var(--spacing-lg);
  font-size: var(--text-sm);
  color: var(--color-text);
  white-space: nowrap;
  transition: all var(--transition-fast);
}

.header__dropdown-link:hover {
  background-color: var(--color-bg-alt);
  color: var(--color-highlight);
  padding-left: calc(var(--spacing-lg) + var(--spacing-xs));
}

/* Actions */
.header__actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.header__lang {
  display: flex;
  align-items: center;
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-full);
  padding: 2px;
}

.header__lang-btn {
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-text-light);
  border-radius: var(--radius-full);
  transition: all var(--transition-fast);
}

.header__lang-btn:hover {
  color: var(--color-text);
}

.header__lang-btn--active {
  background-color: white;
  color: var(--color-primary);
  box-shadow: var(--shadow-xs);
}

.header__cta {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.header__cta-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

/* Mobile Toggle */
.header__toggle {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 5px;
  width: 28px;
  height: 28px;
  padding: 0;
}

.header__toggle span {
  display: block;
  width: 100%;
  height: 2px;
  background-color: var(--color-primary);
  border-radius: 2px;
  transition: all var(--transition-base);
  transform-origin: center;
}

.header__toggle--active span:nth-child(1) {
  transform: translateY(7px) rotate(45deg);
}

.header__toggle--active span:nth-child(2) {
  opacity: 0;
}

.header__toggle--active span:nth-child(3) {
  transform: translateY(-7px) rotate(-45deg);
}

/* Mobile Menu Backdrop */
.header__mobile-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  z-index: calc(var(--z-sticky) - 2);
  opacity: 0;
  visibility: hidden;
  transition: all var(--transition-base);
}

.header__mobile-backdrop--active {
  opacity: 1;
  visibility: visible;
}

/* Mobile Menu */
.header__mobile {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: 80%;
  max-width: 400px;
  background-color: white;
  z-index: calc(var(--z-sticky) - 1);
  overflow-y: auto;
  transform: translateX(100%);
  transition: transform var(--transition-base) cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: -4px 0 24px rgba(0, 0, 0, 0.1);
}

.header--menu-open .header__mobile {
  transform: translateX(0);
}

.header__mobile-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-lg) var(--spacing-lg) var(--spacing-md);
  border-bottom: 1px solid var(--color-border-light);
}

.header__mobile-title {
  font-weight: 700;
  font-size: var(--text-lg);
  color: var(--color-primary);
}

.header__mobile-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: var(--radius-full);
  background-color: var(--color-bg-alt);
  color: var(--color-text);
  border: none;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.header__mobile-close:hover {
  background-color: #fce7f3;
  color: var(--color-error);
}

.header__mobile-nav {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  padding: var(--spacing-lg);
}

.header__mobile-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.header__mobile-item {
  border-bottom: 1px solid var(--color-border-light);
}

.header__mobile-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-md) 0;
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text);
}

.header__mobile-link--active {
  color: var(--color-highlight);
}

.header__mobile-submenu {
  display: flex;
  flex-direction: column;
  padding: var(--spacing-sm) 0 var(--spacing-md) var(--spacing-lg);
}

.header__mobile-sublink {
  padding: var(--spacing-sm) 0;
  font-size: var(--text-base);
  color: var(--color-text-light);
}

.header__mobile-actions {
  margin-top: auto;
  padding-top: var(--spacing-xl);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.header__mobile-lang-switcher {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-full);
  padding: 4px;
}

.header__mobile-lang-btn {
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-text-light);
  border-radius: var(--radius-full);
  transition: all var(--transition-fast);
}

.header__mobile-lang-btn:hover {
  color: var(--color-text);
}

.header__mobile-lang-btn--active {
  background-color: white;
  color: var(--color-primary);
  box-shadow: var(--shadow-xs);
}

.header__mobile-lang {
  display: flex;
  gap: var(--spacing-sm);
}

.header__mobile-lang-btn {
  flex: 1;
  padding: var(--spacing-sm);
  font-size: var(--text-sm);
  font-weight: 600;
  border: 2px solid var(--color-border);
  border-radius: var(--radius-md);
  background: transparent;
  transition: all var(--transition-fast);
}

.header__mobile-lang-btn--active {
  border-color: var(--color-primary);
  background-color: var(--color-primary);
  color: white;
}

:global(body.body-lock) {
  overflow: hidden;
  /* Prevent scroll jumping on iOS */
  touch-action: none;
}

@media (max-width: 1023px) {
  .header__nav-list {
    gap: var(--spacing-md);
  }

  .header__nav-link {
    font-size: var(--text-xs);
  }
}

@media (max-width: 767px) {
  .header__logo-svg {
    width: 120px;
  }
}
</style>
