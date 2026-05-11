<template>
  <footer class="footer">
    <!-- Newsletter CTA Banner -->
    <div class="footer__newsletter">
      <div class="container">
        <div class="footer__newsletter-inner">
          <div class="footer__newsletter-content">
            <h3>{{ $t('form.title') }}</h3>
            <p>{{ $t('form.subtitle') }}</p>
          </div>
          <div class="footer__newsletter-actions">
            <NuxtLink :to="localePath('/contact')" class="footer__newsletter-btn footer__newsletter-btn--inquiry">
              <Icon name="lucide:mail" size="20" />
              {{ $t('contact.send_inquiry') }}
            </NuxtLink>
            <a :href="whatsappUrl" target="_blank" rel="noopener noreferrer" class="footer__newsletter-btn footer__newsletter-btn--whatsapp">
              <WhatsAppIcon size="20" />
              {{ $t('whatsapp.us') }}
            </a>
          </div>
        </div>
      </div>
    </div>

    <!-- Main Footer -->
    <div class="footer__main">
      <div class="container">
        <div class="footer__grid">
          <!-- Brand Column -->
          <div class="footer__col footer__col--brand">
            <div class="footer__logo">
              <svg viewBox="0 0 180 40" fill="none" xmlns="http://www.w3.org/2000/svg">
                <circle cx="20" cy="20" r="16" fill="var(--color-highlight)" opacity="0.2"/>
                <path d="M14 20C14 16.6863 16.6863 14 20 14C23.3137 14 26 16.6863 26 20C26 23.3137 23.3137 26 20 26C16.6863 26 14 23.3137 14 20Z" stroke="var(--color-highlight)" stroke-width="2.5"/>
                <text x="45" y="27" font-family="Inter, system-ui, sans-serif" font-size="18" font-weight="700" fill="currentColor">CandyPro</text>
                <text x="145" y="27" font-family="Inter, system-ui, sans-serif" font-size="10" font-weight="500" fill="var(--color-accent)">OEM</text>
              </svg>
            </div>
            <p class="footer__tagline">{{ $t('footer.tagline') }}</p>
            <div class="footer__certs">
              <span v-for="cert in certifications" :key="cert" class="footer__cert">{{ cert }}</span>
            </div>
          </div>

          <!-- Products Column -->
          <div class="footer__col">
            <h4 class="footer__title">{{ $t('footer.products') }}</h4>
            <ul class="footer__links">
              <li><NuxtLink :to="localePath('/products/gummy-candy')">{{ $t('product.categories.gummy_candy') }}</NuxtLink></li>
              <li><NuxtLink :to="localePath('/products/hard-candy')">{{ $t('product.categories.hard_candy') }}</NuxtLink></li>
              <li><NuxtLink :to="localePath('/products/aerated-candy')">{{ $t('product.categories.aerated_candy') }}</NuxtLink></li>
              <li><NuxtLink :to="localePath('/products/toffee-candy')">{{ $t('product.categories.toffee_candy') }}</NuxtLink></li>
              <li><NuxtLink :to="localePath('/products/compound-chocolate')">{{ $t('product.categories.compound_chocolate') }}</NuxtLink></li>
              <li><NuxtLink :to="localePath('/products/licorice')">{{ $t('product.categories.licorice') }}</NuxtLink></li>
              <li><NuxtLink :to="localePath('/products/sour-candies')">{{ $t('product.categories.sour_candies') }}</NuxtLink></li>
            </ul>
          </div>

          <!-- Company Column -->
          <div class="footer__col">
            <h4 class="footer__title">{{ $t('footer.services') }}</h4>
            <ul class="footer__links">
              <li><NuxtLink :to="localePath('/about')">{{ $t('nav.about') }}</NuxtLink></li>
              <li><NuxtLink :to="localePath('/oem-solutions')">{{ $t('nav.oem') }}</NuxtLink></li>
              <li><NuxtLink :to="localePath('/factory-quality')">{{ $t('nav.factory') }}</NuxtLink></li>
              <li><NuxtLink :to="localePath('/cases-clients')">{{ $t('nav.cases') }}</NuxtLink></li>
              <li><NuxtLink :to="localePath('/blog')">{{ $t('nav.blog') }}</NuxtLink></li>
              <li><NuxtLink :to="localePath('/contact')">{{ $t('nav.contact') }}</NuxtLink></li>
              <li><ClientOnly><NuxtLink :to="portalLink">{{ portalLabel }}</NuxtLink><template #fallback><NuxtLink :to="localePath('/auth/login')">{{ $t('footer.sign_in') }}</NuxtLink></template></ClientOnly></li>
            </ul>
          </div>

          <!-- Contact Column -->
          <div class="footer__col">
            <h4 class="footer__title">{{ $t('contact.title') }}</h4>
            <ul class="footer__contact">
              <li class="footer__contact-item">
                <Icon name="lucide:map-pin" size="16" />
                <span>No. 88 Shipin Road, Jinshan District, Shanghai 201500, China</span>
              </li>
              <li class="footer__contact-item">
                <Icon name="lucide:phone" size="16" />
                <a href="tel:+862167310088">+86 21 6731 0088</a>
              </li>
              <li class="footer__contact-item">
                <Icon name="lucide:mail" size="16" />
                <a href="mailto:sales@candypro.com">sales@candypro.com</a>
              </li>
              <li class="footer__contact-item">
                <Icon name="lucide:message-circle" size="16" />
                <a :href="whatsappUrl" target="_blank" rel="noopener noreferrer">WhatsApp: +86 21 6731 0088</a>
              </li>
              <li class="footer__contact-item">
                <Icon name="lucide:clock" size="16" />
                <span>{{ $t('contact.working_hours') }}</span>
              </li>
            </ul>

            <!-- Social Links -->
            <div class="footer__social">
              <a href="https://facebook.com/candypro" target="_blank" rel="noopener noreferrer" :aria-label="t('common.a11y.facebook')" class="footer__social-link footer__social-link--facebook">
                <Icon name="lucide:facebook" size="18" />
              </a>
              <a href="https://linkedin.com/company/candypro" target="_blank" rel="noopener noreferrer" :aria-label="t('common.a11y.linkedin')" class="footer__social-link footer__social-link--linkedin">
                <Icon name="lucide:linkedin" size="18" />
              </a>
              <a href="https://instagram.com/candypro" target="_blank" rel="noopener noreferrer" :aria-label="t('common.a11y.instagram')" class="footer__social-link footer__social-link--instagram">
                <Icon name="lucide:instagram" size="18" />
              </a>
              <a href="https://youtube.com/@candypro" target="_blank" rel="noopener noreferrer" :aria-label="t('common.a11y.youtube')" class="footer__social-link footer__social-link--youtube">
                <Icon name="lucide:youtube" size="18" />
              </a>
              <a href="https://twitter.com/candypro" target="_blank" rel="noopener noreferrer" :aria-label="t('common.a11y.twitter')" class="footer__social-link footer__social-link--twitter">
                <Icon name="lucide:twitter" size="18" />
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Bottom Bar -->
    <div class="footer__bottom">
      <div class="container">
        <div class="footer__bottom-inner">
          <p class="footer__copyright">
            {{ $t('footer.copyright').replace('{year}', String(currentYear)) }}
          </p>
          <ul class="footer__legal">
            <li><NuxtLink :to="localePath('/privacy')">{{ $t('footer.privacy_policy') }}</NuxtLink></li>
            <li><NuxtLink :to="localePath('/terms')">{{ $t('footer.terms_of_service') }}</NuxtLink></li>
          </ul>
          <div class="footer__lang">
            <button
              v-for="loc in availableLocales"
              :key="loc.code"
              :class="['footer__lang-btn', { 'footer__lang-btn--active': loc.code === currentLocale }]"
              @click="switchLocale(loc.code)"
            >
              {{ loc.code === 'en' ? $t('languages.en_short') : $t('languages.zh') }}
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

const { t, locale, setLocale } = useI18n()
const localePath = useLocalePath()
const config = useRuntimeConfig()
const { isAuthenticated, isAdmin } = useAuth()

const currentYear = computed(() => new Date().getFullYear())
const certifications = ['HACCP', 'ISO22000', 'BRC', 'HALAL', 'FDA']
const hydrated = ref(false)

const availableLocales = [
  { code: 'en', name: 'English' },
  { code: 'zh', name: '中文' }
]

const currentLocale = computed(() => locale.value)

const whatsappUrl = computed(() => {
  const number = config.public.whatsappNumber
  const message = encodeURIComponent(t('whatsapp.message'))
  return `https://wa.me/${number}?text=${message}`
})

const portalLink = computed(() => {
  if (hydrated.value && isAuthenticated.value) {
    return localePath(isAdmin.value ? '/admin' : '/customer/products')
  }
  return localePath('/auth/login')
})

const portalLabel = computed(() => {
  return (hydrated.value && isAuthenticated.value) ? t('footer.customer_portal') : t('footer.sign_in')
})

onMounted(() => { hydrated.value = true })

const LOCALE_STORAGE_KEY = 'user-locale'

const switchLocale = async (newLocale: string) => {
  if (import.meta.client) {
    localStorage.setItem(LOCALE_STORAGE_KEY, newLocale)
  }
  await setLocale(newLocale)
}
</script>

<style scoped>
.footer {
  background-color: var(--color-primary);
  color: var(--color-text-on-primary);
}

/* ── Newsletter CTA ── */
.footer__newsletter {
  padding: var(--spacing-3xl) 0;
  background: linear-gradient(135deg, var(--color-highlight-hover) 0%, var(--color-accent) 100%);
}

.footer__newsletter-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-xl);
}

.footer__newsletter-content h3 {
  color: var(--color-text-on-primary);
  font-size: var(--text-2xl);
  margin-bottom: var(--spacing-xs);
}

.footer__newsletter-content p {
  color: rgba(255, 255, 255, 0.9);
  margin-bottom: 0;
}

.footer__newsletter-actions {
  display: flex;
  gap: var(--spacing-md);
  flex-shrink: 0;
}

.footer__newsletter-btn {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-xl);
  border-radius: var(--radius-md);
  font-weight: 600;
  font-size: var(--text-sm);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
  white-space: nowrap;
}

.footer__newsletter-btn--inquiry {
  background-color: var(--color-bg);
  color: var(--color-highlight);
}

.footer__newsletter-btn--inquiry:hover {
  background-color: var(--color-bg-alt);
  transform: translateY(-1px);
}

.footer__newsletter-btn--whatsapp {
  background-color: rgba(0, 0, 0, 0.2);
  color: var(--color-text-on-primary);
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.footer__newsletter-btn--whatsapp:hover {
  background-color: rgba(0, 0, 0, 0.35);
  transform: translateY(-1px);
}

/* ── Main Grid ── */
.footer__main {
  padding: var(--spacing-4xl) 0 var(--spacing-3xl);
}

.footer__grid {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1.5fr;
  gap: var(--spacing-3xl);
}

/* ── Columns ── */
.footer__col {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.footer__col--brand {
  gap: var(--spacing-lg);
}

.footer__logo {
  width: 150px;
  color: var(--color-text-on-primary);
}

.footer__tagline {
  color: rgba(255, 255, 255, 0.7);
  font-size: var(--text-sm);
  line-height: 1.6;
  margin: 0;
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
  color: var(--color-text-on-primary);
  font-size: var(--text-sm);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: var(--spacing-xs);
}

.footer__links {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.footer__links a {
  color: rgba(255, 255, 255, 0.7);
  font-size: var(--text-sm);
  transition: color var(--transition-fast);
}

.footer__links a:hover {
  color: var(--color-accent);
}

/* ── Contact Info ── */
.footer__contact {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.footer__contact-item {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-sm);
  color: rgba(255, 255, 255, 0.7);
  font-size: var(--text-sm);
  line-height: 1.5;
}

.footer__contact-item svg {
  flex-shrink: 0;
  margin-top: 3px;
  color: var(--color-accent);
}

.footer__contact-item a {
  color: inherit;
  transition: color var(--transition-fast);
}

.footer__contact-item a:hover {
  color: var(--color-accent);
}

/* ── Social Links ── */
.footer__social {
  display: flex;
  gap: var(--spacing-sm);
  margin-top: var(--spacing-md);
}

.footer__social-link {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background-color: rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-full);
  color: var(--color-text-on-primary);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.footer__social-link:hover {
  transform: translateY(-2px);
}

.footer__social-link--facebook:hover {
  background-color: #1877f2;
}

.footer__social-link--linkedin:hover {
  background-color: #0a66c2;
}

.footer__social-link--instagram:hover {
  background: linear-gradient(45deg, #f09433, #e6683c, #dc2743, #cc2366, #bc1888);
}

.footer__social-link--youtube:hover {
  background-color: #ff0000;
}

.footer__social-link--twitter:hover {
  background-color: #1da1f2;
}

/* ── Bottom Bar ── */
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
  color: rgba(255, 255, 255, 0.5);
  font-size: var(--text-sm);
  margin: 0;
}

.footer__legal {
  display: flex;
  gap: var(--spacing-lg);
}

.footer__legal a {
  color: rgba(255, 255, 255, 0.5);
  font-size: var(--text-sm);
  transition: color var(--transition-fast);
}

.footer__legal a:hover {
  color: var(--color-text-on-primary);
}

/* ── Language Switcher ── */
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
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.footer__lang-btn:hover {
  color: var(--color-text-on-primary);
}

.footer__lang-btn--active {
  background-color: var(--color-text-on-primary);
  color: var(--color-primary);
}

/* ── Responsive ── */
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

  .footer__newsletter-actions {
    flex-direction: column;
    width: 100%;
  }

  .footer__newsletter-btn {
    justify-content: center;
  }

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
