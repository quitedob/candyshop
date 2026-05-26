<template>
  <div class="contact-page">
    <!-- Breadcrumb -->
    <div class="container">
      <Breadcrumb :items="[{ label: t('nav.contact') }]" />
    </div>

    <!-- Hero -->
    <section class="contact-hero section">
      <div class="container">
        <div class="contact-hero__inner">
          <h1 class="contact-hero__title">{{ t('contact.title') }}</h1>
          <p class="contact-hero__subtitle">{{ t('contact.subtitle') }}</p>
        </div>
      </div>
    </section>

    <!-- Contact Content -->
    <section class="contact-content section">
      <div class="container">
        <div class="contact-content__inner">
          <!-- Inquiry Form -->
          <div class="contact-content__form">
            <div class="form-card">
              <h2>{{ t('form.title') }}</h2>
              <p>{{ t('form.subtitle') }}</p>

              <ClientOnly>
                <InquiryForm
                  :product-name="productFromQuery"
                  :category="categoryFromQuery"
                  :user-id="currentUserId || undefined"
                />
                <template #fallback>
                  <div class="form-loading">
                    <div class="animate-spin rounded-full h-8 w-8 border-4 border-primary border-t-transparent mx-auto" />
                    <p style="margin-top: 12px; color: var(--color-text-light);">{{ t('contact.loading_form') }}</p>
                  </div>
                </template>
              </ClientOnly>
            </div>
          </div>

          <!-- Contact Info -->
          <div class="contact-content__info">
            <!-- Quick Contact -->
            <div class="info-card">
              <h3>{{ t('contact.quick_contact') }}</h3>

              <div class="info-item">
                <div class="info-item__icon">
                  <Icon name="lucide:map-pin" size="20" />
                </div>
                <div>
                  <strong>{{ t('contact.address') }}</strong>
                  <p>No. 88 Shipin Road, Jinshan District<br>Shanghai 201500, China</p>
                </div>
              </div>

              <div class="info-item">
                <div class="info-item__icon">
                  <Icon name="lucide:phone" size="20" />
                </div>
                <div>
                  <strong>{{ t('contact.phone') }}</strong>
                  <p><a href="tel:+862167310088">+86 21 6731 0088</a></p>
                </div>
              </div>

              <div class="info-item">
                <div class="info-item__icon">
                  <Icon name="lucide:mail" size="20" />
                </div>
                <div>
                  <strong>{{ t('contact.email') }}</strong>
                  <p><a href="mailto:sales@candypro.com">sales@candypro.com</a></p>
                </div>
              </div>

              <div class="info-item">
                <div class="info-item__icon">
                  <WhatsAppIcon size="20" />
                </div>
                <div>
                  <strong>{{ t('contact.whatsapp') }}</strong>
                  <p><a :href="whatsappUrl" target="_blank" rel="noopener noreferrer">+86 21 6731 0088</a></p>
                </div>
              </div>

              <div class="info-item">
                <div class="info-item__icon">
                  <Icon name="lucide:clock" size="20" />
                </div>
                <div>
                  <strong>{{ t('contact.hours') }}</strong>
                  <p>{{ t('contact.working_hours') }}</p>
                </div>
              </div>
            </div>

            <!-- Map -->
            <div class="map-card">
              <iframe
                :src="mapSrc"
                width="100%"
                height="300"
                style="border:0;"
                allowfullscreen
                loading="lazy"
                referrerpolicy="no-referrer-when-downgrade"
                :title="t('contact.factory_location')"
              ></iframe>
            </div>

            <!-- Social Media -->
            <div class="social-card">
              <h3>{{ t('contact.social_media') }}</h3>
              <div class="social-links">
                <a
                  href="https://facebook.com/candypro"
                  target="_blank"
                  rel="noopener noreferrer"
                  :aria-label="t('a11y.facebook')"
                  class="social-link social-link--facebook"
                >
                  <Icon name="lucide:facebook" size="24" />
                </a>
                <a
                  href="https://linkedin.com/company/candypro"
                  target="_blank"
                  rel="noopener noreferrer"
                  :aria-label="t('a11y.linkedin')"
                  class="social-link social-link--linkedin"
                >
                  <Icon name="lucide:linkedin" size="24" />
                </a>
                <a
                  href="https://instagram.com/candypro"
                  target="_blank"
                  rel="noopener noreferrer"
                  :aria-label="t('a11y.instagram')"
                  class="social-link social-link--instagram"
                >
                  <Icon name="lucide:instagram" size="24" />
                </a>
                <a
                  href="https://youtube.com/@candypro"
                  target="_blank"
                  rel="noopener noreferrer"
                  :aria-label="t('a11y.youtube')"
                  class="social-link social-link--youtube"
                >
                  <Icon name="lucide:youtube" size="24" />
                </a>
                <a
                  href="https://twitter.com/candypro"
                  target="_blank"
                  rel="noopener noreferrer"
                  :aria-label="t('a11y.twitter')"
                  class="social-link social-link--twitter"
                >
                  <Icon name="lucide:twitter" size="24" />
                </a>
              </div>
            </div>

            <!-- WhatsApp CTA -->
            <a :href="whatsappUrl" target="_blank" rel="noopener noreferrer" class="whatsapp-card">
              <WhatsAppIcon size="48" />
              <div class="whatsapp-card__content">
                <h3>{{ t('whatsapp.us') }}</h3>
                <p>{{ t('contact.whatsapp_desc') }}</p>
              </div>
            </a>
          </div>
        </div>
      </div>
    </section>

    <!-- Global Offices -->
    <section class="offices section bg-alt">
      <div class="container">
        <div class="section-header">
          <h2>{{ t('contact.global_presence_title') }}</h2>
          <p>{{ t('contact.global_presence_subtitle') }}</p>
        </div>

        <div class="offices__grid">
          <div v-for="office in offices" :key="office.id" class="office-card">
            <h4>{{ office.country }}</h4>
            <p>{{ office.address }}</p>
            <a :href="`tel:${office.phone}`">{{ office.phone }}</a>
            <a :href="`mailto:${office.email}`">{{ office.email }}</a>
          </div>
        </div>
      </div>
    </section>

    <!-- FAQ Preview -->
    <section class="faq section">
      <div class="container container-narrow">
        <div class="section-header">
          <h2>{{ t('contact.faq_title') }}</h2>
        </div>

        <div class="faq__list">
          <div
            v-for="(faq, index) in faqs"
            :key="index"
            class="faq-item"
            :class="{ 'faq-item--open': openFaq === index }"
          >
            <button class="faq-item__question" @click="toggleFaq(index)">
              <span>{{ faq.question }}</span>
              <Icon name="lucide:chevron-down" size="18" />
            </button>
            <div v-if="openFaq === index" class="faq-item__answer">
              {{ faq.answer }}
            </div>
          </div>
        </div>

        <div class="faq__actions">
          <NuxtLink :to="localePath('/faq')" class="btn btn-ghost">
            {{ t('home.faq.view_all') }}
            <Icon name="lucide:arrow-right" size="16" />
          </NuxtLink>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n, useLocalePath } from '#i18n'

// /contact is a public marketing page; it must not require authentication.
// Previously this declared `middleware: ['auth']`, and /contact was also in
// the auth middleware's protectedRoutes set, so visiting /ja/contact (or any
// locale-prefixed contact page) without a session redirected to /auth/login.
const { t, tm, rt } = useI18n()
const localePath = useLocalePath()
const config = useRuntimeConfig()
const route = useRoute()
const { isAuthenticated, user } = useAuth()

const productFromQuery = computed(() => (route.query.product as string) || undefined)
const categoryFromQuery = computed(() => (route.query.category as string) || undefined)
const currentUserId = computed(() => isAuthenticated.value ? (user.value?.id || '') : '')

const openFaq = ref(0)

const offices = computed(() => [
  {
    id: 'cn',
    country: t('contact.offices.headquarters_country'),
    address: t('contact.offices.headquarters_address'),
    phone: '+86 21 6731 0088',
    email: 'hq@candypro.com'
  },
  {
    id: 'us',
    country: t('contact.offices.us_country'),
    address: t('contact.offices.us_address'),
    phone: '+1 626 888 0168',
    email: 'us@candypro.com'
  },
  {
    id: 'uk',
    country: t('contact.offices.uk_country'),
    address: t('contact.offices.uk_address'),
    phone: '+44 20 7946 0958',
    email: 'eu@candypro.com'
  },
  {
    id: 'ae',
    country: t('contact.offices.uae_country'),
    address: t('contact.offices.uae_address'),
    phone: '+971 4 355 6800',
    email: 'mena@candypro.com'
  }
])

const faqs = computed(() => {
  const items = tm('faq.items') as Array<any>
  if (items && items.length >= 4) {
    return [
      { question: rt(items[4].question), answer: rt(items[4].answer) }, // How do I place an order?
      { question: t('contact.faq.response_time_q'), answer: t('contact.faq.response_time_a') },
      { question: rt(items[2].question), answer: rt(items[2].answer) }, // Can I get samples before placing an order?
      { question: t('contact.faq.visit_factory_q'), answer: t('contact.faq.visit_factory_a') }
    ]
  }
  return []
})

const toggleFaq = (index: number) => {
  openFaq.value = openFaq.value === index ? -1 : index
}

// Map URL (Google Maps with key if available, otherwise OpenStreetMap)
const mapSrc = computed(() => {
  const key = config.public.googleMapsApiKey
  if (key) {
    return `https://www.google.com/maps/embed/v1/place?key=${key}&q=No.88+Shipin+Road+Jinshan+District+Shanghai+China&zoom=15`
  }
  return 'https://www.openstreetmap.org/export/embed.html?bbox=121.22%2C30.68%2C121.46%2C30.92&layer=mapnik&marker=30.88%2C121.34'
})

// WhatsApp URL
const whatsappUrl = computed(() => {
  const number = config.public.whatsappNumber
  const message = encodeURIComponent(t('whatsapp.message'))
  return `https://wa.me/${number}?text=${message}`
})

// SEO
usePageOgImage({
  title: `${t('nav.contact')} | ${t('seo.default_title')}`,
  description: t('seo.contact_description'),
  ogType: 'website'
})
</script>

<style scoped>
.contact-hero__inner {
  text-align: center;
  max-width: 700px;
  margin: 0 auto;
}

.contact-hero__title {
  font-size: clamp(2rem, 4vw, 3rem);
  margin-bottom: var(--spacing-md);
}

.contact-hero__subtitle {
  font-size: var(--text-xl);
  color: var(--color-text-light);
}

/* Contact Content */
.contact-content__inner {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--spacing-3xl);
}

@media (min-width: 1024px) {
  .contact-content__inner {
    grid-template-columns: 1.5fr 1fr;
  }
}

.form-card {
  padding: var(--spacing-2xl);
  background-color: white;
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-md);
}

.form-card h2 {
  font-size: var(--text-2xl);
  margin-bottom: var(--spacing-sm);
}

.form-card p {
  color: var(--color-text-light);
  margin-bottom: var(--spacing-xl);
}

/* Info Cards */
.info-card,
.map-card,
.social-card,
.whatsapp-card {
  padding: var(--spacing-xl);
  background-color: white;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  margin-bottom: var(--spacing-lg);
}

.info-card h3,
.social-card h3 {
  font-size: var(--text-lg);
  margin-bottom: var(--spacing-md);
}

.info-item {
  display: flex;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-lg);
}

.info-item__icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-md);
  color: var(--color-primary);
}

.info-item strong {
  display: block;
  font-size: var(--text-sm);
  color: var(--color-primary);
  margin-bottom: var(--spacing-xs);
}

.info-item p,
.info-item a {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin: 0;
}

.info-item a:hover {
  color: var(--color-highlight);
}

/* Map */
.map-card {
  padding: 0;
  overflow: hidden;
}

.map-card iframe {
  display: block;
  border-radius: var(--radius-lg);
}

/* Social */
.social-links {
  display: flex;
  gap: var(--spacing-sm);
}

.social-link {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-full);
  color: var(--color-text);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.social-link:hover {
  background-color: var(--color-primary);
  color: white;
  transform: translateY(-2px);
}

.social-link--facebook:hover {
  background-color: #1877F2;
}

.social-link--linkedin:hover {
  background-color: #0A66C2;
}

.social-link--instagram:hover {
  background-color: #E4405F;
}

.social-link--youtube:hover {
  background-color: #FF0000;
}

.social-link--twitter:hover {
  background-color: #1DA1F2;
}

/* WhatsApp Card */
.whatsapp-card {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  background: linear-gradient(135deg, #25D366 0%, #128C7E 100%);
  color: white;
  text-decoration: none;
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
}

.whatsapp-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

.whatsapp-card svg {
  flex-shrink: 0;
}

.whatsapp-card__content h3 {
  font-size: var(--text-lg);
  margin-bottom: var(--spacing-xs);
}

.whatsapp-card__content p {
  font-size: var(--text-sm);
  opacity: 0.9;
  margin: 0;
}

/* Offices */
.offices__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: var(--spacing-xl);
}

.office-card {
  padding: var(--spacing-xl);
  background-color: white;
  border-radius: var(--radius-lg);
}

.office-card h4 {
  font-size: var(--text-lg);
  margin-bottom: var(--spacing-sm);
}

.office-card p {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin-bottom: var(--spacing-xs);
}

.office-card a {
  display: block;
  font-size: var(--text-sm);
  color: var(--color-primary);
  margin-bottom: var(--spacing-xs);
}

/* FAQ */
.faq__list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-xl);
}

.faq-item {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.faq-item__question {
  display: flex;
  justify-content: space-between;
  width: 100%;
  padding: var(--spacing-lg);
  font-weight: 500;
  background: none;
  border: none;
  cursor: pointer;
}

.faq-item__answer {
  padding: 0 var(--spacing-lg) var(--spacing-lg);
  color: var(--color-text-light);
  line-height: 1.6;
}

.faq__actions {
  text-align: center;
}
</style>
