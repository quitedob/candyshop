<template>
  <div class="faq-page">
    <!-- Hero Section -->
    <section class="faq-hero section">
      <div class="container">
        <h1 class="faq-hero__title">{{ $t('faq.title') }}</h1>
        <p class="faq-hero__description">{{ $t('faq.subtitle') }}</p>
      </div>
    </section>

    <!-- FAQ Categories -->
    <section class="faq-content section section-lg">
      <div class="container">
        <!-- Category Tabs -->
        <div class="faq-tabs">
          <button
            v-for="cat in categories"
            :key="cat.key"
            :class="['faq-tab', { 'faq-tab--active': activeCategory === cat.key }]"
            @click="activeCategory = cat.key"
          >
            <Icon :name="cat.icon" size="18" />
            {{ cat.label }}
          </button>
        </div>

        <!-- FAQ List -->
        <div class="faq-list">
          <div
            v-for="(item, index) in filteredFaqs"
            :key="index"
            class="faq-item"
            :class="{ 'faq-item--open': openIndex === index }"
          >
            <button class="faq-item__question" @click="toggleFaq(index)">
              <span>{{ item.question }}</span>
              <Icon
                name="lucide:chevron-down"
                size="20"
                class="faq-item__icon"
                :class="{ 'faq-item__icon--open': openIndex === index }"
              />
            </button>
            <Transition name="collapse">
              <div v-if="openIndex === index" class="faq-item__answer">
                <p>{{ item.answer }}</p>
              </div>
            </Transition>
          </div>
        </div>
      </div>
    </section>

    <!-- Still Have Questions -->
    <section class="faq-cta section bg-alt">
      <div class="container">
        <div class="cta-content">
          <h2>{{ $t('faq.still_questions') }}</h2>
          <p>{{ $t('faq.contact_description') }}</p>
          <div class="cta-actions">
            <NuxtLink :to="localePath('/contact')" class="btn btn-highlight btn-lg">
              {{ $t('contact.send_inquiry') }}
            </NuxtLink>
            <a v-if="whatsappUrl" :href="whatsappUrl" target="_blank" rel="noopener noreferrer" class="btn btn-outline btn-lg">
              <WhatsAppIcon size="20" />
              {{ $t('whatsapp.us') }}
            </a>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n, useLocalePath } from '#i18n'

const { t, tm, rt } = useI18n()
const localePath = useLocalePath()
const config = useRuntimeConfig()

const activeCategory = ref('general')
const openIndex = ref<number | null>(0)

const categories = computed(() => [
  { key: 'general', label: t('faq.categories.general'), icon: 'lucide:help-circle' },
  { key: 'ordering', label: t('faq.categories.ordering'), icon: 'lucide:shopping-cart' },
  { key: 'oem', label: t('faq.categories.oem'), icon: 'lucide:package' },
  { key: 'shipping', label: t('faq.categories.shipping'), icon: 'lucide:truck' }
])

const allFaqs = computed(() => {
  const items = tm('faq.items') as Array<any>
  const faqs = []
  if (items && items.length === 15) {
    const cats = ['general', 'general', 'general', 'general', 'ordering', 'ordering', 'ordering', 'oem', 'oem', 'oem', 'oem', 'shipping', 'shipping', 'shipping', 'shipping']
    for (let i = 0; i < 15; i++) {
      faqs.push({
        category: cats[i],
        question: rt(items[i].question),
        answer: rt(items[i].answer)
      })
    }
  }
  return faqs
})

const filteredFaqs = computed(() => {
  return allFaqs.value.filter(faq => faq.category === activeCategory.value)
})

const whatsappUrl = computed(() => {
  const number = config.public.whatsappNumber
  if (!number) return ''
  const message = encodeURIComponent(t('whatsapp.message'))
  return `https://wa.me/${number}?text=${message}`
})

const toggleFaq = (index: number) => {
  openIndex.value = openIndex.value === index ? null : index
}

// Reset open index when category changes
watch(activeCategory, () => {
  openIndex.value = 0
})

// SEO
usePageOgImage({
  title: t('seo.faq_title'),
  description: t('seo.faq_description'),
  ogType: 'website',
  schema: useFAQSchema(allFaqs.value.map(faq => ({
    question: faq.question,
    answer: faq.answer,
  }))),
})
</script>

<style scoped>
.faq-hero {
  padding-top: calc(var(--header-height) + var(--spacing-2xl));
  text-align: center;
  background: linear-gradient(135deg, #fff5f0 0%, #fef3e2 100%);
}

.faq-hero__title {
  font-size: clamp(2rem, 4vw, 3rem);
  margin-bottom: var(--spacing-md);
  color: var(--color-primary);
}

.faq-hero__description {
  font-size: var(--text-lg);
  color: var(--color-text-light);
  max-width: 600px;
  margin: 0 auto;
}

/* Tabs */
.faq-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-xl);
  justify-content: center;
}

.faq-tab {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-sm) var(--spacing-lg);
  background-color: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
  cursor: pointer;
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.faq-tab:hover {
  border-color: var(--color-highlight);
  color: var(--color-highlight);
}

.faq-tab--active {
  background-color: var(--color-highlight);
  border-color: var(--color-highlight);
  color: white;
}

/* FAQ List */
.faq-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.faq-item {
  background-color: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.faq-item--open {
  border-color: var(--color-highlight);
  box-shadow: var(--shadow-md);
}

.faq-item__question {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: var(--spacing-lg);
  font-size: var(--text-base);
  font-weight: 600;
  text-align: left;
  background: none;
  border: none;
  cursor: pointer;
  transition: background-color var(--transition-fast);
}

.faq-item__question:hover {
  background-color: var(--color-bg-alt);
}

.faq-item__icon {
  flex-shrink: 0;
  color: var(--color-text-light);
  transition: transform var(--transition-fast);
}

.faq-item__icon--open {
  transform: rotate(180deg);
  color: var(--color-highlight);
}

.faq-item__answer {
  padding: 0 var(--spacing-lg) var(--spacing-lg);
}

.faq-item__answer p {
  color: var(--color-text-light);
  line-height: 1.8;
  margin: 0;
}

/* Collapse transition */
.collapse-enter-active,
.collapse-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
  overflow: hidden;
}

.collapse-enter-from,
.collapse-leave-to {
  opacity: 0;
  max-height: 0;
  padding-top: 0;
  padding-bottom: 0;
}

.collapse-enter-to,
.collapse-leave-from {
  opacity: 1;
  max-height: 500px;
}

/* CTA */
.faq-cta {
  text-align: center;
}

.cta-content {
  max-width: 700px;
  margin: 0 auto;
}

.cta-content h2 {
  margin-bottom: var(--spacing-md);
}

.cta-content p {
  color: var(--color-text-light);
  margin-bottom: var(--spacing-xl);
}

.cta-actions {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: var(--spacing-md);
}
</style>
