<template>
  <div class="inquiry-floating">
    <!-- WhatsApp Button -->
    <a
      :href="whatsappUrl"
      target="_blank"
      rel="noopener noreferrer"
      class="inquiry-floating__whatsapp"
      :aria-label="$t('whatsapp.us')"
    >
      <WhatsAppIcon size="28" />
      <span class="inquiry-floating__tooltip">{{ $t('whatsapp.us') }}</span>
    </a>

    <!-- Quick Inquiry Button -->
    <button
      class="inquiry-floating__inquire"
      :aria-label="$t('form.submit')"
      @click="openInquiry"
    >
      <Icon name="lucide:mail" size="24" />
      <span class="inquiry-floating__tooltip">{{ $t('product.inquire_now') }}</span>
    </button>

    <!-- Scroll to Top -->
    <button
      v-show="showScrollTop"
      class="inquiry-floating__top"
      :aria-label="t('common.a11y.scroll_to_top')"
      @click="scrollToTop"
    >
      <Icon name="lucide:chevron-up" size="20" />
    </button>

    <!-- Inquiry Modal (if needed) -->
    <Teleport to="body">
      <div
        v-if="isInquiryOpen"
        class="inquiry-modal"
        @click.self="closeInquiry"
      >
        <div class="inquiry-modal__content">
          <div class="inquiry-modal__header">
            <h3>{{ $t('form.title') }}</h3>
            <button
              class="inquiry-modal__close"
              :aria-label="t('common.a11y.close')"
              @click="closeInquiry"
            >
              <Icon name="lucide:x" size="24" />
            </button>
          </div>
          <div class="inquiry-modal__body">
            <InquiryForm @success="closeInquiry" />
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

const route = useRoute()
const config = useRuntimeConfig()
const { t } = useI18n()

// State
const showScrollTop = ref(false)
const isInquiryOpen = ref(false)

// WhatsApp URL
const whatsappUrl = computed(() => {
  const number = config.public.whatsappNumber
  const message = encodeURIComponent(t('whatsapp.message'))
  return `https://wa.me/${number}?text=${message}`
})

// Handle scroll
const handleScroll = () => {
  showScrollTop.value = window.scrollY > 500
}

// Scroll to top
const scrollToTop = () => {
  const prefersReduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  window.scrollTo({ top: 0, behavior: prefersReduced ? 'instant' : 'smooth' })
}

// Open inquiry modal
const openInquiry = () => {
  isInquiryOpen.value = true
  document.body.style.overflow = 'hidden'
}

// Close inquiry modal
const closeInquiry = () => {
  isInquiryOpen.value = false
  document.body.style.overflow = ''
}

onMounted(() => {
  window.addEventListener('scroll', handleScroll)
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<style scoped>
.inquiry-floating {
  position: fixed;
  bottom: var(--spacing-xl);
  right: var(--spacing-xl);
  z-index: var(--z-fixed);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

/* WhatsApp Button */
.inquiry-floating__whatsapp {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  background: linear-gradient(135deg, #25D366 0%, #128C7E 100%);
  border-radius: var(--radius-full);
  color: white;
  box-shadow: var(--shadow-lg);
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
}

.inquiry-floating__whatsapp:hover {
  transform: scale(1.1);
  box-shadow: var(--shadow-xl);
}

.inquiry-floating__whatsapp:active {
  transform: scale(1.05);
}

/* Inquiry Button */
.inquiry-floating__inquire {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  background-color: var(--color-highlight);
  border-radius: var(--radius-full);
  color: white;
  box-shadow: var(--shadow-lg);
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
}

.inquiry-floating__inquire:hover {
  background-color: var(--color-highlight-hover);
  transform: scale(1.1);
  box-shadow: var(--shadow-xl);
}

.inquiry-floating__inquire:active {
  transform: scale(1.05);
}

/* Scroll to Top Button */
.inquiry-floating__top {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  background-color: var(--color-primary);
  border-radius: var(--radius-full);
  color: white;
  box-shadow: var(--shadow-md);
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
  animation: fadeInUp 0.3s ease;
}

.inquiry-floating__top:hover {
  background-color: var(--color-primary-light);
  transform: translateY(-2px);
}

/* Tooltip */
.inquiry-floating__tooltip {
  position: absolute;
  right: calc(100% + var(--spacing-sm));
  padding: var(--spacing-xs) var(--spacing-sm);
  background-color: var(--color-primary);
  color: white;
  font-size: var(--text-sm);
  font-weight: 500;
  white-space: nowrap;
  border-radius: var(--radius-md);
  opacity: 0;
  visibility: hidden;
  transform: translateX(10px);
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
}

.inquiry-floating__tooltip::after {
  content: '';
  position: absolute;
  top: 50%;
  right: -4px;
  transform: translateY(-50%);
  border: 4px solid transparent;
  border-left-color: var(--color-primary);
}

.inquiry-floating__whatsapp:hover .inquiry-floating__tooltip,
.inquiry-floating__inquire:hover .inquiry-floating__tooltip {
  opacity: 1;
  visibility: visible;
  transform: translateX(0);
}

/* Modal */
.inquiry-modal {
  position: fixed;
  inset: 0;
  z-index: var(--z-modal);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-md);
  background-color: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  animation: fadeIn 0.2s ease;
}

.inquiry-modal__content {
  width: 100%;
  max-width: 600px;
  max-height: 90vh;
  background-color: white;
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-2xl);
  display: flex;
  flex-direction: column;
  animation: scaleIn 0.3s ease;
}

.inquiry-modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-light);
}

.inquiry-modal__header h3 {
  margin: 0;
  font-size: var(--text-xl);
  color: var(--color-primary);
}

.inquiry-modal__close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  color: var(--color-text-light);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.inquiry-modal__close:hover {
  background-color: var(--color-bg-alt);
  color: var(--color-text);
}

.inquiry-modal__body {
  padding: var(--spacing-lg);
  overflow-y: auto;
}

/* Mobile */
@media (max-width: 767px) {
  .inquiry-floating {
    bottom: var(--spacing-md);
    right: var(--spacing-md);
  }

  .inquiry-floating__whatsapp,
  .inquiry-floating__inquire {
    width: 48px;
    height: 48px;
  }

  .inquiry-floating__top {
    width: 40px;
    height: 40px;
  }

  .inquiry-modal__content {
    max-height: calc(100vh - var(--spacing-lg));
  }
}

/* Animations */
@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes scaleIn {
  from {
    opacity: 0;
    transform: scale(0.9);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
