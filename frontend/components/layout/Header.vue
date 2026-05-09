<template>
  <header
    :class="['header', { 'header--scrolled': isScrolled, 'header--menu-open': isMenuOpen }]"
  >
    <div class="container">
      <div class="header__inner">
        <!-- Logo -->
        <NuxtLink :to="localePath('/')" class="header__logo" aria-label="Go to homepage">
          <svg viewBox="0 0 180 40" fill="none" xmlns="http://www.w3.org/2000/svg" class="header__logo-svg">
            <circle cx="20" cy="20" r="16" fill="var(--color-highlight)" opacity="0.2"/>
            <path d="M14 20C14 16.6863 16.6863 14 20 14C23.3137 14 26 16.6863 26 20C26 23.3137 23.3137 26 20 26C16.6863 26 14 23.3137 14 20Z" stroke="var(--color-highlight)" stroke-width="2.5"/>
            <circle cx="17" cy="17" r="2" fill="var(--color-accent)"/>
            <circle cx="23" cy="23" r="2" fill="var(--color-primary)"/>
            <text x="45" y="27" font-family="Inter, system-ui, sans-serif" font-size="18" font-weight="700" fill="var(--color-primary)">CandyPro</text>
            <text x="145" y="27" font-family="Inter, system-ui, sans-serif" font-size="10" font-weight="500" fill="var(--color-accent)">OEM</text>
          </svg>
        </NuxtLink>

        <!-- Desktop Navigation -->
        <nav class="header__nav hide-mobile" aria-label="Main navigation">
          <ul class="header__nav-list">
            <li
              v-for="item in navItems"
              :key="item.key"
              :class="['header__nav-item', { 'header__nav-item--has-dropdown': item.children }]"
            >
              <NuxtLink
                :to="localePath(item.to)"
                :class="['header__nav-link', { 'header__nav-link--active': isActive(item.to) }]"
              >
                {{ $t(item.key) }}
                <svg v-if="item.children" class="header__nav-arrow" width="8" height="5" viewBox="0 0 8 5" fill="none">
                  <path d="M1 1L4 4L7 1" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </NuxtLink>
              <div v-if="item.children" class="header__dropdown" role="menu">
                <NuxtLink
                  v-for="child in item.children"
                  :key="child.to"
                  :to="localePath(child.to)"
                  class="header__dropdown-link"
                  role="menuitem"
                >
                  {{ $t(child.key) }}
                </NuxtLink>
              </div>
            </li>
          </ul>
        </nav>

        <!-- Right Side Actions -->
        <div class="header__actions">
          <!-- Language Switcher -->
          <div class="header__lang hide-mobile" role="group" aria-label="Language">
            <button
              v-for="loc in locales"
              :key="loc"
              :class="['header__lang-btn', { 'header__lang-btn--active': locale === loc }]"
              :aria-pressed="locale === loc"
              @click="switchLocale(loc)"
            >
              {{ loc.toUpperCase() }}
            </button>
          </div>

          <!-- WhatsApp -->
          <a
            :href="whatsappUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="header__whatsapp hide-mobile"
            aria-label="Contact us on WhatsApp"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
              <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z"/>
            </svg>
          </a>

          <!-- Auth CTAs -->
          <div class="header__cta hide-mobile">
            <template v-if="!isAuthenticated">
              <NuxtLink :to="localePath('/auth/login')" class="btn btn-ghost btn-sm">
                {{ $t('nav.sign_in') }}
              </NuxtLink>
              <NuxtLink :to="localePath('/auth/register')" class="btn btn-highlight btn-sm">
                {{ $t('nav.register') }}
              </NuxtLink>
            </template>
            <NuxtLink
              v-else
              :to="localePath(isAdmin ? '/admin' : '/customer/dashboard')"
              class="btn btn-highlight btn-sm"
            >
              {{ $t(isAdmin ? 'nav.admin_panel' : 'nav.my_account') }}
            </NuxtLink>
          </div>

          <!-- Mobile Hamburger -->
          <button
            ref="menuToggleRef"
            class="header__toggle hide-desktop"
            :aria-expanded="isMenuOpen"
            aria-label="Toggle menu"
            @click="toggleMenu"
          >
            <span /><span /><span />
          </button>
        </div>
      </div>
    </div>

    <!-- Mobile Backdrop -->
    <Transition name="fade">
      <div
        v-if="isMenuOpen"
        class="header__backdrop hide-desktop"
        aria-hidden="true"
        @click="closeMenu"
      />
    </Transition>

    <!-- Mobile Drawer -->
    <Transition name="slide">
      <nav
        v-if="isMenuOpen"
        class="header__drawer hide-desktop"
        aria-label="Mobile navigation"
      >
        <div class="header__drawer-head">
          <span class="header__drawer-title">{{ $t('nav.home') }}</span>
          <button class="header__drawer-close" aria-label="Close menu" @click="closeMenu">
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <line x1="4" y1="4" x2="16" y2="16"/><line x1="16" y1="4" x2="4" y2="16"/>
            </svg>
          </button>
        </div>

        <ul class="header__drawer-list">
          <li v-for="item in navItems" :key="item.key" class="header__drawer-item">
            <template v-if="item.children">
              <button
                :class="['header__drawer-link', { 'header__drawer-link--open': openMobileSub === item.key }]"
                @click="openMobileSub = openMobileSub === item.key ? null : item.key"
              >
                {{ $t(item.key) }}
                <svg class="header__drawer-chevron" width="10" height="6" viewBox="0 0 10 6" fill="none">
                  <path d="M1 1L5 5L9 1" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </button>
              <ul v-show="openMobileSub === item.key" class="header__drawer-sub">
                <li v-for="child in item.children" :key="child.to">
                  <NuxtLink :to="localePath(child.to)" class="header__drawer-sublink" @click="closeMenu">
                    {{ $t(child.key) }}
                  </NuxtLink>
                </li>
              </ul>
            </template>
            <NuxtLink
              v-else
              :to="localePath(item.to)"
              :class="['header__drawer-link', { 'header__drawer-link--active': isActive(item.to) }]"
              @click="closeMenu"
            >
              {{ $t(item.key) }}
            </NuxtLink>
          </li>
        </ul>

        <div class="header__drawer-footer">
          <div class="header__drawer-lang" role="group" aria-label="Language">
            <button
              v-for="loc in locales"
              :key="loc"
              :class="['header__lang-btn', { 'header__lang-btn--active': locale === loc }]"
              @click="switchLocale(loc)"
            >
              {{ loc.toUpperCase() }}
            </button>
          </div>
          <template v-if="!isAuthenticated">
            <NuxtLink :to="localePath('/auth/login')" class="btn btn-ghost btn-sm" @click="closeMenu">
              {{ $t('nav.sign_in') }}
            </NuxtLink>
            <NuxtLink :to="localePath('/auth/register')" class="btn btn-highlight btn-sm" @click="closeMenu">
              {{ $t('nav.register') }}
            </NuxtLink>
          </template>
          <NuxtLink
            v-else
            :to="localePath(isAdmin ? '/admin' : '/customer/dashboard')"
            class="btn btn-highlight btn-sm"
            @click="closeMenu"
          >
            {{ $t(isAdmin ? 'nav.admin_panel' : 'nav.my_account') }}
          </NuxtLink>
        </div>
      </nav>
    </Transition>
  </header>
</template>

<script setup lang="ts">
const LOCALE_KEY = 'user-locale'

const { t, locale, setLocale } = useI18n()
const localePath = useLocalePath()
const route = useRoute()
const config = useRuntimeConfig()
const { isAuthenticated, isAdmin, initAuth } = useAuth()

const locales = ['en', 'zh'] as const
const isScrolled = ref(false)
const isMenuOpen = ref(false)
const openMobileSub = ref<string | null>(null)
const menuToggleRef = ref<HTMLButtonElement | null>(null)

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
      { key: 'product.categories.compound_chocolate', to: '/products/compound-chocolate' },
    ],
  },
  { key: 'nav.oem', to: '/oem-solutions' },
  { key: 'nav.factory', to: '/factory-quality' },
  { key: 'nav.cases', to: '/cases-clients' },
  { key: 'nav.blog', to: '/blog' },
  { key: 'nav.about', to: '/about' },
  { key: 'nav.contact', to: '/contact' },
]

const whatsappUrl = computed(() => {
  const num = config.public.whatsappNumber
  return `https://wa.me/${num}?text=${encodeURIComponent(t('whatsapp.message'))}`
})

const isActive = (to: string): boolean => {
  const path = route.path
  if (to === '/') return path === '/' || path === `/${locale.value}`
  return path.startsWith(to) || path.startsWith(`/${locale.value}${to}`)
}

const switchLocale = async (code: string) => {
  if (import.meta.client) localStorage.setItem(LOCALE_KEY, code)
  await setLocale(code as 'en' | 'zh')
}

const handleScroll = () => { isScrolled.value = window.scrollY > 50 }

const toggleMenu = () => {
  isMenuOpen.value = !isMenuOpen.value
  document.body.classList.toggle('body-lock', isMenuOpen.value)
}

const closeMenu = () => {
  isMenuOpen.value = false
  openMobileSub.value = null
  document.body.classList.remove('body-lock')
  nextTick(() => menuToggleRef.value?.focus())
}

onMounted(() => {
  initAuth()
  const saved = import.meta.client && localStorage.getItem(LOCALE_KEY)
  if (saved && saved !== locale.value) setLocale(saved as 'en' | 'zh')
  window.addEventListener('scroll', handleScroll, { passive: true })
})

onUnmounted(() => { window.removeEventListener('scroll', handleScroll) })

watch(() => route.path, closeMenu)
</script>

<style scoped>
/* ── Base ── */
.header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: var(--z-sticky);
  background: transparent;
  transition: background-color var(--transition-base), box-shadow var(--transition-base);
}

.header--scrolled,
.header--menu-open {
  background: rgba(255, 255, 255, 0.98);
  backdrop-filter: blur(8px);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
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

/* ── Logo ── */
.header__logo {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.header__logo-svg {
  width: 150px;
  height: 40px;
}

/* ── Desktop Nav ── */
.header__nav {
  flex: 1;
  display: flex;
  justify-content: center;
}

.header__nav-list {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
}

.header__nav-item {
  position: relative;
}

.header__nav-link {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: var(--spacing-sm) 0;
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-primary);
  transition: color var(--transition-fast);
}

.header__nav-link::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  width: 0;
  height: 2px;
  background: var(--color-highlight);
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

/* ── Dropdown (hover + focus-within + click) ── */
.header__dropdown {
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%) translateY(8px);
  min-width: 200px;
  padding: var(--spacing-xs) 0;
  background: #fff;
  border-radius: var(--radius-lg);
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
  opacity: 0;
  visibility: hidden;
  pointer-events: none;
  transition: opacity var(--transition-base), transform var(--transition-base), visibility var(--transition-base);
}

.header__nav-item:hover > .header__dropdown,
.header__nav-item:focus-within > .header__dropdown {
  opacity: 1;
  visibility: visible;
  pointer-events: auto;
  transform: translateX(-50%) translateY(0);
}

.header__nav-item:hover > .header__nav-link .header__nav-arrow,
.header__nav-item:focus-within > .header__nav-link .header__nav-arrow {
  transform: rotate(180deg);
}

.header__dropdown-link {
  display: block;
  padding: var(--spacing-sm) var(--spacing-lg);
  font-size: var(--text-sm);
  color: var(--color-primary);
  white-space: nowrap;
  transition: background var(--transition-fast), color var(--transition-fast);
}

.header__dropdown-link:hover {
  background: var(--color-bg-alt);
  color: var(--color-highlight);
}

/* ── Actions ── */
.header__actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.header__lang {
  display: flex;
  background: var(--color-bg-alt);
  border-radius: var(--radius-full);
  padding: 2px;
}

.header__lang-btn {
  padding: 4px 10px;
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-text-light);
  border-radius: var(--radius-full);
  transition: all var(--transition-fast);
  cursor: pointer;
}

.header__lang-btn:hover {
  color: var(--color-primary);
}

.header__lang-btn--active {
  background: #fff;
  color: var(--color-primary);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.08);
}

.header__whatsapp {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  color: #25d366;
  transition: background var(--transition-fast);
}

.header__whatsapp:hover {
  background: rgba(37, 211, 102, 0.1);
}

.header__cta {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

/* ── Mobile Toggle ── */
.header__toggle {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 5px;
  width: 44px;
  height: 44px;
  padding: 0;
  cursor: pointer;
}

.header__toggle span {
  display: block;
  width: 22px;
  height: 2px;
  background: var(--color-primary);
  border-radius: 2px;
  transition: all var(--transition-base);
  transform-origin: center;
}

.header--menu-open .header__toggle span:nth-child(1) {
  transform: translateY(7px) rotate(45deg);
}

.header--menu-open .header__toggle span:nth-child(2) {
  opacity: 0;
}

.header--menu-open .header__toggle span:nth-child(3) {
  transform: translateY(-7px) rotate(-45deg);
}

/* ── Mobile Backdrop ── */
.header__backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(4px);
  z-index: calc(var(--z-sticky) + 1);
}

/* ── Mobile Drawer ── */
.header__drawer {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: min(85vw, 380px);
  background: #fff;
  z-index: calc(var(--z-sticky) + 2);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  box-shadow: -4px 0 24px rgba(0, 0, 0, 0.1);
}

.header__drawer-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-light);
}

.header__drawer-title {
  font-weight: 700;
  font-size: var(--text-lg);
  color: var(--color-primary);
}

.header__drawer-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  color: var(--color-text-light);
  cursor: pointer;
  transition: background var(--transition-fast);
}

.header__drawer-close:hover {
  background: var(--color-bg-alt);
}

.header__drawer-list {
  flex: 1;
  padding: var(--spacing-sm) var(--spacing-lg);
}

.header__drawer-item {
  border-bottom: 1px solid var(--color-border-light);
}

.header__drawer-item:last-child {
  border-bottom: none;
}

.header__drawer-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: var(--spacing-md) 0;
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-primary);
  background: none;
  border: none;
  cursor: pointer;
  text-align: left;
}

.header__drawer-link--active {
  color: var(--color-highlight);
}

.header__drawer-chevron {
  transition: transform var(--transition-fast);
}

.header__drawer-link--open .header__drawer-chevron {
  transform: rotate(180deg);
}

.header__drawer-sub {
  padding: 0 0 var(--spacing-sm) var(--spacing-md);
}

.header__drawer-sublink {
  display: block;
  padding: var(--spacing-xs) 0;
  font-size: var(--text-sm);
  color: var(--color-text-light);
  transition: color var(--transition-fast);
}

.header__drawer-sublink:hover {
  color: var(--color-highlight);
}

.header__drawer-footer {
  padding: var(--spacing-lg);
  border-top: 1px solid var(--color-border-light);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.header__drawer-lang {
  display: flex;
  background: var(--color-bg-alt);
  border-radius: var(--radius-full);
  padding: 2px;
  align-self: flex-start;
}

/* ── Transitions ── */
.fade-enter-active,
.fade-leave-active {
  transition: opacity var(--transition-base);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.slide-enter-active,
.slide-leave-active {
  transition: transform var(--transition-base);
}

.slide-enter-from,
.slide-leave-to {
  transform: translateX(100%);
}

/* ── Global body lock ── */
:global(body.body-lock) {
  overflow: hidden;
  touch-action: none;
}

/* ── Responsive ── */
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
