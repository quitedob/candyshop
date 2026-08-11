<template>
  <div ref="switcherRef" class="lang-switcher" :class="{ 'lang-switcher--open': isOpen }">
    <!-- Trigger Button -->
    <button
      class="lang-switcher__trigger"
      :aria-expanded="isOpen"
      aria-haspopup="listbox"
      @click.stop="toggle"
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
        ref="dropdownRef"
        class="lang-switcher__dropdown"
        :style="dropdownStyle"
        @click.stop
      >
        <ul class="lang-switcher__list" role="listbox">
          <li
            v-for="loc in availableLocales"
            :key="loc.code"
            class="lang-switcher__item"
            role="option"
            :aria-selected="loc.code === currentLocale?.code"
          >
            <button
              class="lang-switcher__option"
              :class="{ 'lang-switcher__option--active': loc.code === currentLocale?.code }"
              @click="selectLocale(loc.code)"
            >
              <span class="lang-switcher__flag">{{ getLocaleFlag(loc.code) }}</span>
              <span class="lang-switcher__name">{{ loc.name }}</span>
              <Icon
                v-if="loc.code === currentLocale?.code"
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
import { ref, computed, onMounted, onUnmounted, nextTick, watch, unref } from 'vue'
import { LOCALE_META } from '~/composables/useLocales'

interface Locale {
  code: string
  name: string
  flag?: string
}

interface Props {
  locales?: Locale[]
}

const props = withDefaults(defineProps<Props>(), {
  locales: () => []
})

const { locale, setLocale, locales: i18nLocales } = useI18n()
const switchLocalePath = useSwitchLocalePath()

const isOpen = ref(false)
const switcherRef = ref<HTMLElement | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)

const getLocaleFlag = (code: string): string => {
  const flags: Record<string, string> = {
    en: '🇺🇸', zh: '🇨🇳',
    ja: '🇯🇵', ko: '🇰🇷', ar: '🇸🇦',
    vi: '🇻🇳', th: '🇹🇭', id: '🇮🇩', ms: '🇲🇾'
  }
  return flags[code] || '🌐'
}

/** 解析可用语言：优先 props，其次 i18n 配置，最后回退到 LOCALE_META */
const resolveConfiguredLocales = (): Locale[] => {
  const fromI18n = unref(i18nLocales)
  const list = Array.isArray(fromI18n) ? fromI18n : []
  const source = list.length > 0 ? list : LOCALE_META
  return source.map((l) => ({
    code: l.code,
    name: l.name || l.code,
    flag: getLocaleFlag(l.code),
  }))
}

const availableLocales = computed(() =>
  props.locales.length > 0 ? props.locales : resolveConfiguredLocales()
)

const currentLocale = computed(() =>
  availableLocales.value.find(l => l.code === locale.value)
)

const currentLocaleFlag = computed(() =>
  currentLocale.value?.flag || getLocaleFlag(locale.value)
)

const dropdownStyle = ref<Record<string, string>>({})

const DROPDOWN_GAP = 4
const VIEWPORT_MARGIN = 8
const MOBILE_BREAKPOINT = 768

const isMobileViewport = () =>
  import.meta.client && window.innerWidth < MOBILE_BREAKPOINT

const updateDropdownPosition = async () => {
  if (!switcherRef.value) return
  await nextTick()

  if (isMobileViewport()) {
    dropdownStyle.value = {
      position: 'fixed',
      bottom: '0',
      left: '0',
      right: '0',
      top: 'auto',
      maxHeight: '70vh',
      overflowY: 'auto',
      zIndex: '9999',
    }
    return
  }

  const rect = switcherRef.value.getBoundingClientRect()
  const dropdownHeight = dropdownRef.value?.offsetHeight ?? 280
  const minWidth = Math.max(rect.width, 160)

  const spaceBelow = window.innerHeight - rect.bottom - VIEWPORT_MARGIN
  const spaceAbove = rect.top - VIEWPORT_MARGIN
  const openUp = spaceBelow < dropdownHeight && spaceAbove > spaceBelow

  let top = openUp
    ? rect.top - dropdownHeight - DROPDOWN_GAP
    : rect.bottom + DROPDOWN_GAP
  top = Math.max(
    VIEWPORT_MARGIN,
    Math.min(top, window.innerHeight - dropdownHeight - VIEWPORT_MARGIN),
  )

  let left = rect.left
  if (left + minWidth > window.innerWidth - VIEWPORT_MARGIN) {
    left = window.innerWidth - minWidth - VIEWPORT_MARGIN
  }
  left = Math.max(VIEWPORT_MARGIN, left)

  dropdownStyle.value = {
    position: 'fixed',
    top: `${top}px`,
    left: `${left}px`,
    minWidth: `${minWidth}px`,
    maxHeight: `${window.innerHeight - VIEWPORT_MARGIN * 2}px`,
    overflowY: 'auto',
    zIndex: '9999',
  }
}

const toggle = async () => {
  if (!isOpen.value) {
    isOpen.value = true
    await updateDropdownPosition()
  } else {
    isOpen.value = false
  }
}

const close = () => {
  isOpen.value = false
}

const selectLocale = async (code: string) => {
  if (code === locale.value) {
    close()
    return
  }

  if (import.meta.client) {
    localStorage.setItem('user-locale', code)
    document.cookie = `user-locale=${code}; path=/; max-age=${60 * 60 * 24 * 365}; SameSite=Lax`
  }

  await setLocale(code as Parameters<typeof setLocale>[0])
  await navigateTo(switchLocalePath(code as Parameters<typeof switchLocalePath>[0]))

  close()
}

const handleClickOutside = (event: MouseEvent) => {
  if (!isOpen.value) return
  const target = event.target as HTMLElement
  if (switcherRef.value?.contains(target)) return
  if (dropdownRef.value?.contains(target)) return
  close()
}

const handleEscape = (event: KeyboardEvent) => {
  if (event.key === 'Escape') close()
}

const onViewportChange = () => {
  if (isOpen.value) updateDropdownPosition()
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside, true)
  document.addEventListener('keydown', handleEscape)
  window.addEventListener('resize', onViewportChange)
  window.addEventListener('scroll', onViewportChange, true)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside, true)
  document.removeEventListener('keydown', handleEscape)
  window.removeEventListener('resize', onViewportChange)
  window.removeEventListener('scroll', onViewportChange, true)
})

watch(isOpen, (open) => {
  if (open) updateDropdownPosition()
})
</script>

<style scoped>
.lang-switcher {
  position: relative;
  display: inline-block;
  min-width: 0;
  max-width: 100%;
}

.lang-switcher__trigger {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  max-width: 100%;
  box-sizing: border-box;
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
  min-width: 0;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.lang-switcher__chevron {
  color: var(--color-text-light);
  transition: transform var(--transition-fast);
  flex-shrink: 0;
}

.lang-switcher__chevron--open {
  transform: rotate(180deg);
}

.lang-switcher__dropdown {
  background-color: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  animation: dropdownIn 0.2s ease;
}

@keyframes dropdownIn {
  from { opacity: 0; transform: translateY(-8px); }
  to { opacity: 1; transform: translateY(0); }
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

@media (max-width: 767px) {
  .lang-switcher__dropdown {
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
