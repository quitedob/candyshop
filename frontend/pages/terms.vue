<template>
  <div class="legal-page">
    <div class="container pt-8">
      <Breadcrumb :items="[{ label: t('legal.terms_title') }]" />
    </div>

    <div class="container section">
      <h1 class="text-4xl font-bold mb-4">{{ t('legal.terms_title') }}</h1>
      <p class="text-light mb-8">{{ t('legal.last_updated', { date: '2026-08-13' }) }}</p>

      <div class="legal-content prose max-w-none">
        <h2>{{ t('legal.terms.accept_title') }}</h2>
        <p>{{ t('legal.terms.accept_desc') }}</p>

        <h2>{{ t('legal.terms.company_title') }}</h2>
        <p>{{ t('legal.terms.company_desc') }}</p>

        <h2>{{ t('legal.terms.b2b_title') }}</h2>
        <p>{{ t('legal.terms.b2b_desc') }}</p>

        <h2>{{ t('legal.terms.registration_title') }}</h2>
        <p>{{ t('legal.terms.registration_desc') }}</p>

        <h2>{{ t('legal.terms.license_title') }}</h2>
        <p>{{ t('legal.terms.license_desc') }}</p>

        <h2>{{ t('legal.terms.ip_title') }}</h2>
        <p>{{ t('legal.terms.ip_desc') }}</p>

        <h2>{{ t('legal.terms.payment_title') }}</h2>
        <p>{{ t('legal.terms.payment_desc') }}</p>

        <h2>{{ t('legal.terms.delivery_title') }}</h2>
        <p>{{ t('legal.terms.delivery_desc') }}</p>

        <h2>{{ t('legal.terms.limitations_title') }}</h2>
        <p>{{ t('legal.terms.limitations_desc') }}</p>

        <h2>{{ t('legal.terms.law_title') }}</h2>
        <p>{{ t('legal.terms.law_desc') }}</p>

        <h2>{{ t('legal.terms.revisions_title') }}</h2>
        <p>{{ t('legal.terms.revisions_desc') }}</p>

        <h2>{{ t('legal.terms.termination_title') }}</h2>
        <p>{{ t('legal.terms.termination_desc') }}</p>

        <h2>{{ t('legal.export_control_title') }}</h2>
        <p>{{ t('legal.export_control_desc') }}</p>
        <p>{{ t('legal.sanctions_notice') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#i18n'

const { t } = useI18n()

usePageOgImage({
  title: `Terms of Service | ${t('seo.default_title')}`,
  description: t('legal.terms.accept_desc'),
  schema: {
    '@context': 'https://schema.org',
    '@type': 'WebPage',
    name: t('legal.terms_title'),
    description: t('legal.terms.accept_desc'),
  },
})
</script>

<style scoped>
.legal-page {
  padding-top: calc(var(--header-height) + var(--spacing-md));
}
.legal-content h2 {
  font-size: 1.5rem;
  margin-top: 2rem;
  margin-bottom: 1rem;
  color: var(--color-primary);
}
.legal-content p, .legal-content li {
  color: var(--color-text-light);
  line-height: 1.6;
  margin-bottom: 0.5rem;
}
.legal-content ul {
  list-style-type: disc;
  padding-left: 1.5rem;
  margin-bottom: 1rem;
}
</style>
