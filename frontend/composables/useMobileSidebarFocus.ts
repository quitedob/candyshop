import { nextTick, onBeforeUnmount, ref, watch, type Ref } from 'vue'

/** Keyboard focus follows the mobile navigation while desktop navigation stays inline. */
export function useMobileSidebarFocus(isOpen: Ref<boolean>, isMobileViewport: Ref<boolean>) {
  const mobileSidebarRef = ref<HTMLElement | null>(null)
  const mobileMenuButtonRef = ref<HTMLButtonElement | null>(null)
  const mobileSidebarCloseButtonRef = ref<HTMLButtonElement | null>(null)

  const getFocusableElements = () => Array.from(
    mobileSidebarRef.value?.querySelectorAll<HTMLElement>(
      'a[href], button, input, select, textarea, [tabindex]',
    ) || [],
  ).filter(element => element.tabIndex >= 0
    && !element.matches(':disabled')
    && !element.closest('[inert]')
    && element.getClientRects().length > 0)

  const handleSidebarKeydown = (event: KeyboardEvent) => {
    if (!isOpen.value || !isMobileViewport.value || event.defaultPrevented) return
    if (event.key === 'Escape') {
      event.preventDefault()
      isOpen.value = false
      return
    }
    if (event.key !== 'Tab') return

    const focusableElements = getFocusableElements()
    const firstFocusableElement = focusableElements[0]
    const lastFocusableElement = focusableElements.at(-1)
    if (!firstFocusableElement || !lastFocusableElement) {
      event.preventDefault()
      mobileSidebarRef.value?.focus()
      return
    }
    const focusIsOutsideSidebar = !mobileSidebarRef.value?.contains(document.activeElement)
    if (event.shiftKey && (document.activeElement === firstFocusableElement || focusIsOutsideSidebar)) {
      event.preventDefault()
      lastFocusableElement.focus()
    } else if (!event.shiftKey && (document.activeElement === lastFocusableElement || focusIsOutsideSidebar)) {
      event.preventDefault()
      firstFocusableElement.focus()
    }
  }

  watch([isOpen, isMobileViewport], async ([open, mobileViewport], [wasOpen]) => {
    if (!import.meta.client) return
    document.removeEventListener('keydown', handleSidebarKeydown)
    await nextTick()
    if (open && mobileViewport && isOpen.value && isMobileViewport.value) {
      document.addEventListener('keydown', handleSidebarKeydown)
      mobileSidebarCloseButtonRef.value?.focus()
    } else if (!open && wasOpen && mobileViewport && !isOpen.value) {
      mobileMenuButtonRef.value?.focus()
    }
  }, { flush: 'post' })

  onBeforeUnmount(() => {
    if (import.meta.client) document.removeEventListener('keydown', handleSidebarKeydown)
  })

  return { mobileSidebarRef, mobileMenuButtonRef, mobileSidebarCloseButtonRef }
}
