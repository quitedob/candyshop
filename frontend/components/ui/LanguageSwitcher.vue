<template>
  <div class="lang-switcher" :class="{ 'lang-switcher--open': isOpen }">
    <!-- Trigger Button -->
    <button
      class="lang-switcher__trigger"
      :aria-expanded="isOpen"
      aria-haspopup="listbox"
      @click="toggle"
    >
      <span class="lang-switcher__current">
        {{ currentLocaleFlag }} {{ currentLocale?.name }}
      </span>
      <Icon
        name="lucide:chevron-down"
        size="14"
        class="lang-switcher__chevron"
        :class="{ 'lang-switcher__chevron--open': isOpen }"
        aria-hidden="true"
      />
    </button>

    <!-- Dropdown Menu -->
    <Teleport to="body">
      <div
        v-if="isOpen"
        class="lang-switcher__dropdown"
        @click.outside="close"
      >
        <ul class="lang-switcher__list" role="listbox">
          <li
            v-for="locale in availableLocales"
            :key="locale.code"
            class="lang-switcher__item"
            role="option"
            :aria-selected="locale.code === currentLocale?.code"
          >
            <button
              class="lang-switcher__option"
              :class="{ 'lang-switcher__option--active': locale.code === currentLocale?.code }"
              @click="selectLocale(locale.code)"
            >
              <span class="lang-switcher__flag">{{ getLocaleFlag(locale.code) }}</span>
              <span class="lang-switcher__name">{{ locale.name }}</span>
              <Icon
                v-if="locale.code === currentLocale?.code"
                name="lucide:check"
                size="16"
                class="lang-switcher__check"
                aria-hidden="true"
              />
            </button>
          </li>
        </ul>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

interface Locale {
  code: string
  name: string
  flag?: string
}

interface Props {
  locales?: Locale[]
}

const props = withDefaults(defineProps<Props>(), {
  locales: () => [
    { code: 'en', name: 'English', flag: '🇺🇸' },
    { code: 'zh', name: '中文', flag: '🇨🇳' }
  ]
})

const { locale, setLocale, availableLocales: i18nLocales } = useI18n()
const localePath = useLocalePath()
const route = useRoute()

const isOpen = ref(false)

const getLocaleFlag = (code: string): string => {
  const flags: Record<string, string> = {
    en: '🇺🇸', zh: '🇨🇳', es: '🇪🇸', fr: '🇫🇷', de: '🇩🇪',
    ja: '🇯🇵', ko: '🇰🇷', pt: '🇵🇹', ru: '🇷🇺', ar: '🇸🇦',
    it: '🇮🇹', nl: '🇳🇱', pl: '🇵🇱', tr: '🇹🇷',
    vi: '🇻🇳', th: '🇹🇭', id: '🇮🇩', ms: '🇲🇾'
  }
  return flags[code] || '🌐'
}

const availableLocales = computed(() =>
  props.locales.length > 0 ? props.locales : i18nLocales.value.map(l => ({
    code: l.code,
    name: l.name || l.code,
    flag: getLocaleFlag(l.code)
  }))
)

const currentLocale = computed(() =>
  availableLocales.value.find(l => l.code === locale.value)
)

const currentLocaleFlag = computed(() =>
  currentLocale.value?.flag || getLocaleFlag(locale.value)
)

const toggle = () => {
  isOpen.value = !isOpen.value
}

const close = () => {
  isOpen.value = false
}

const selectLocale = async (code: string) => {
  if (code === locale.value) {
    close()
    return
  }

  // Get the current path without locale prefix
  const currentPath = route.path.replace(`/${locale.value}`, '').replace(/^\/?/, '/')

  // Navigate to the new locale
  await setLocale(code)
  await navigateTo(localePath(currentPath, code))

  close()
}

// Handle click outside
const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  const switcher = document.querySelector('.lang-switcher')

  if (switcher && !switcher.contains(target)) {
    close()
  }
}

// Handle escape key
const handleEscape = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    close()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleEscape)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleEscape)
})
</script>

<style scoped>
.lang-switcher {
  position: relative;
  display: inline-block;
}

.lang-switcher__trigger {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
  background-color: var(--color-bg-alt);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.lang-switcher__trigger:hover {
  border-color: var(--color-accent);
  background-color: var(--color-bg);
}

.lang-switcher__current {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.lang-switcher__chevron {
  color: var(--color-text-light);
  transition: transform var(--transition-fast);
}

.lang-switcher__chevron--open {
  transform: rotate(180deg);
}

/* Dropdown */
.lang-switcher__dropdown {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  z-index: var(--z-dropdown);
  min-width: 160px;
  background-color: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  animation: dropdownIn 0.2s ease;
}

@keyframes dropdownIn {
  from {
    opacity: 0;
    transform: translateY(-8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.lang-switcher__list {
  list-style: none;
  margin: 0;
  padding: var(--spacing-xs);
}

.lang-switcher__item {
  border-radius: var(--radius-sm);
}

.lang-switcher__option {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-sm);
  color: var(--color-text);
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
  text-align: left;
}

.lang-switcher__option:hover {
  background-color: var(--color-bg-alt);
}

.lang-switcher__option--active {
  background-color: var(--color-primary);
  color: var(--color-text-on-primary);
}

.lang-switcher__flag {
  font-size: var(--text-lg);
  line-height: 1;
}

.lang-switcher__name {
  flex: 1;
}

.lang-switcher__check {
  margin-left: auto;
}

/* Compact variant */
.lang-switcher--compact .lang-switcher__trigger {
  padding: var(--spacing-xs);
  border-radius: var(--radius-full);
}

.lang-switcher--compact .lang-switcher__current span:last-child {
  display: none;
}

@media (max-width: 767px) {
  .lang-switcher__dropdown {
    position: fixed;
    top: auto;
    bottom: 0;
    right: 0;
    left: 0;
    border-radius: var(--radius-lg) var(--radius-lg) 0 0;
    box-shadow: var(--shadow-2xl);
  }

  .lang-switcher__list {
    padding: var(--spacing-sm);
  }

  .lang-switcher__option {
    padding: var(--spacing-md);
    font-size: var(--text-base);
  }
}
</style>
