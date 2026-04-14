<template>
  <footer class="footer">
    <!-- Newsletter Section -->
    <div class="footer__newsletter">
      <div class="container">
        <div class="footer__newsletter-inner">
          <div class="footer__newsletter-content">
            <h3>{{ $t('form.title') }}</h3>
            <p>{{ $t('form.subtitle') }}</p>
          </div>
          <a :href="whatsappUrl" target="_blank" rel="noopener noreferrer" class="btn btn-highlight">
            <WhatsAppIcon size="20" />
            {{ $t('whatsapp.us') }}
          </a>
        </div>
      </div>
    </div>

    <!-- Main Footer -->
    <div class="footer__main">
      <div class="container">
        <div class="footer__grid">
          <!-- Company Info -->
          <div class="footer__section footer__section--brand">
            <div class="footer__logo">
              <svg viewBox="0 0 180 40" fill="none" xmlns="http://www.w3.org/2000/svg">
                <circle cx="20" cy="20" r="16" fill="var(--color-highlight)" opacity="0.2"/>
                <path
                  d="M14 20C14 16.6863 16.6863 14 20 14C23.3137 14 26 16.6863 26 20C26 23.3137 23.3137 26 20 26C16.6863 26 14 23.3137 14 20Z"
                  stroke="var(--color-highlight)"
                  stroke-width="2.5"
                />
                <text
                  x="45"
                  y="27"
                  font-family="Inter, system-ui, sans-serif"
                  font-size="18"
                  font-weight="700"
                  fill="currentColor"
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
            </div>
            <p class="footer__tagline">{{ $t('footer.tagline') }}</p>
            <div class="footer__certs">
              <div v-for="cert in certifications" :key="cert" class="footer__cert" :title="cert">
                {{ cert }}
              </div>
            </div>
          </div>

          <!-- Products -->
          <div class="footer__section">
            <h4 class="footer__title">{{ $t('footer.products') }}</h4>
            <ul class="footer__links">
              <li>
                <NuxtLink :to="localePath('/products/gummy-candy')">
                  {{ $t('product.categories.gummy_candy') }}
                </NuxtLink>
              </li>
              <li>
                <NuxtLink :to="localePath('/products/hard-candy')">
                  {{ $t('product.categories.hard_candy') }}
                </NuxtLink>
              </li>
              <li>
                <NuxtLink :to="localePath('/products/aerated-candy')">
                  {{ $t('product.categories.aerated_candy') }}
                </NuxtLink>
              </li>
              <li>
                <NuxtLink :to="localePath('/products/toffee-candy')">
                  {{ $t('product.categories.toffee_candy') }}
                </NuxtLink>
              </li>
              <li>
                <NuxtLink :to="localePath('/products/compound-chocolate')">
                  {{ $t('product.categories.compound_chocolate') }}
                </NuxtLink>
              </li>
            </ul>
          </div>

          <!-- Services -->
          <div class="footer__section">
            <h4 class="footer__title">{{ $t('footer.services') }}</h4>
            <ul class="footer__links">
              <li>
                <NuxtLink :to="localePath('/oem-solutions')">{{ $t('nav.oem') }}</NuxtLink>
              </li>
              <li>
                <NuxtLink :to="localePath('/factory-quality')">{{ $t('nav.factory') }}</NuxtLink>
              </li>
              <li>
                <NuxtLink :to="localePath('/cases-clients')">{{ $t('nav.cases') }}</NuxtLink>
              </li>
              <li>
                <NuxtLink :to="localePath('/blog')">{{ $t('nav.blog') }}</NuxtLink>
              </li>
              <li>
                <NuxtLink :to="portalLink">{{ portalLabel }}</NuxtLink>
              </li>
            </ul>
          </div>

          <!-- Contact -->
          <div class="footer__section">
            <h4 class="footer__title">{{ $t('contact.title') }}</h4>
            <ul class="footer__contact">
              <li class="footer__contact-item">
                <Icon name="lucide:map-pin" size="16" />
                <span>{{ $t('contact.offices.headquarters_address') }}, {{ $t('contact.offices.headquarters_country') }}</span>
              </li>
              <li class="footer__contact-item">
                <Icon name="lucide:phone" size="16" />
                <a href="tel:+861234567890">+86 123 4567 890</a>
              </li>
              <li class="footer__contact-item">
                <Icon name="lucide:mail" size="16" />
                <a href="mailto:info@candypro.com">info@candypro.com</a>
              </li>
              <li class="footer__contact-item">
                <Icon name="lucide:clock" size="16" />
                <span>{{ $t('contact.working_hours') }}</span>
              </li>
            </ul>

            <!-- Social Links -->
            <div class="footer__social">
              <a
                href="https://facebook.com"
                target="_blank"
                rel="noopener noreferrer"
                aria-label="Facebook"
                class="footer__social-link"
              >
                <Icon name="lucide:facebook" size="20" />
              </a>
              <a
                href="https://linkedin.com"
                target="_blank"
                rel="noopener noreferrer"
                aria-label="LinkedIn"
                class="footer__social-link"
              >
                <Icon name="lucide:linkedin" size="20" />
              </a>
              <a
                href="https://instagram.com"
                target="_blank"
                rel="noopener noreferrer"
                aria-label="Instagram"
                class="footer__social-link"
              >
                <Icon name="lucide:instagram" size="20" />
              </a>
              <a
                href="https://youtube.com"
                target="_blank"
                rel="noopener noreferrer"
                aria-label="YouTube"
                class="footer__social-link"
              >
                <Icon name="lucide:youtube" size="20" />
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Bottom Footer -->
    <div class="footer__bottom">
      <div class="container">
        <div class="footer__bottom-inner">
          <p class="footer__copyright">
            {{ $t('footer.copyright').replace('{year}', currentYear) }}
          </p>

          <ul class="footer__legal">
            <li>
              <NuxtLink :to="localePath('/privacy')">{{ $t('footer.privacy_policy') }}</NuxtLink>
            </li>
            <li>
              <NuxtLink :to="localePath('/terms')">{{ $t('footer.terms_of_service') }}</NuxtLink>
            </li>
          </ul>

          <!-- Language Switcher -->
          <div class="footer__lang">
            <button
              v-for="loc in availableLocales"
              :key="loc.code"
              :class="['footer__lang-btn', { 'footer__lang-btn--active': loc.code === currentLocale }]"
              @click="switchLocale(loc.code)"
            >
              {{ loc.code.toUpperCase() }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n, useLocalePath } from '#i18n'

const { t, locale } = useI18n()
const localePath = useLocalePath()
const config = useRuntimeConfig()
const { isAuthenticated, isAdmin } = useAuth()

// Current year for copyright
const currentYear = computed(() => new Date().getFullYear())

// Certifications
const certifications = ['HACCP', 'ISO22000', 'BRC', 'HALAL']

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

const portalLink = computed(() => {
  if (isAuthenticated.value) {
    return localePath(isAdmin.value ? '/admin' : '/customer/products')
  }
  return localePath('/auth/login')
})

const portalLabel = computed(() => {
  if (isAuthenticated.value) {
    return t('footer.customer_portal')
  }
  return t('footer.sign_in')
})

// Switch locale
const switchLocale = (newLocale: string) => {
  locale.value = newLocale
}
</script>

<style scoped>
.footer {
  background-color: var(--color-primary);
  color: white;
}

/* Newsletter */
.footer__newsletter {
  padding: var(--spacing-3xl) 0;
  background: linear-gradient(135deg, var(--color-highlight) 0%, var(--color-accent) 100%);
}

.footer__newsletter-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-xl);
}

.footer__newsletter-content h3 {
  color: white;
  margin-bottom: var(--spacing-xs);
}

.footer__newsletter-content p {
  opacity: 0.9;
  margin-bottom: 0;
}

/* Main Footer */
.footer__main {
  padding: var(--spacing-4xl) 0 var(--spacing-3xl);
}

.footer__grid {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1.5fr;
  gap: var(--spacing-3xl);
}

@media (max-width: 1024px) {
  .footer__grid {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--spacing-2xl);
  }
}

@media (max-width: 640px) {
  .footer__grid {
    grid-template-columns: 1fr;
    gap: var(--spacing-xl);
  }

  .footer__newsletter-inner {
    flex-direction: column;
    text-align: center;
  }
}

/* Section */
.footer__section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.footer__section--brand {
  gap: var(--spacing-lg);
}

.footer__logo {
  width: 150px;
  color: white;
}

.footer__tagline {
  color: rgba(255, 255, 255, 0.8);
  font-size: var(--text-sm);
  line-height: 1.6;
  margin-bottom: 0;
}

.footer__certs {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
}

.footer__cert {
  padding: var(--spacing-xs) var(--spacing-sm);
  background-color: rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-sm);
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-accent);
}

.footer__title {
  color: white;
  font-size: var(--text-sm);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0;
}

.footer__links {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.footer__links a {
  color: rgba(255, 255, 255, 0.8);
  font-size: var(--text-sm);
  transition: color var(--transition-fast);
}

.footer__links a:hover {
  color: var(--color-accent);
}

.footer__contact {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.footer__contact-item {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-sm);
  color: rgba(255, 255, 255, 0.8);
  font-size: var(--text-sm);
}

.footer__contact-item svg {
  flex-shrink: 0;
  margin-top: 2px;
}

.footer__contact-item a {
  color: inherit;
  transition: color var(--transition-fast);
}

.footer__contact-item a:hover {
  color: var(--color-accent);
}

.footer__social {
  display: flex;
  gap: var(--spacing-sm);
  margin-top: var(--spacing-sm);
}

.footer__social-link {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background-color: rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-full);
  color: white;
  transition: all var(--transition-fast);
}

.footer__social-link:hover {
  background-color: var(--color-highlight);
  transform: translateY(-2px);
}

/* Bottom Footer */
.footer__bottom {
  padding: var(--spacing-lg) 0;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.footer__bottom-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-lg);
  flex-wrap: wrap;
}

.footer__copyright {
  color: rgba(255, 255, 255, 0.6);
  font-size: var(--text-sm);
  margin-bottom: 0;
}

.footer__legal {
  display: flex;
  gap: var(--spacing-lg);
}

.footer__legal a {
  color: rgba(255, 255, 255, 0.6);
  font-size: var(--text-sm);
  transition: color var(--transition-fast);
}

.footer__legal a:hover {
  color: white;
}

.footer__lang {
  display: flex;
  background-color: rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-full);
  padding: 2px;
}

.footer__lang-btn {
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: var(--text-xs);
  font-weight: 600;
  color: rgba(255, 255, 255, 0.6);
  border-radius: var(--radius-full);
  transition: all var(--transition-fast);
}

.footer__lang-btn:hover {
  color: white;
}

.footer__lang-btn--active {
  background-color: white;
  color: var(--color-primary);
}

@media (max-width: 640px) {
  .footer__bottom-inner {
    flex-direction: column;
    text-align: center;
  }

  .footer__legal {
    flex-direction: column;
    gap: var(--spacing-sm);
  }
}
</style>
