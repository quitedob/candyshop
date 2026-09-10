<template>
  <div class="legal-page">
    <div class="container pt-8">
      <Breadcrumb :items="[{ label: t('legal.privacy_title') }]" />
    </div>

    <div class="container section">
      <h1 class="text-4xl font-bold mb-4">{{ t('legal.privacy_title') }}</h1>
      <p class="text-light mb-8">{{ t('legal.last_updated', { date: '2026-08-13' }) }}</p>

      <div class="legal-content prose max-w-none">
        <h2>{{ t('legal.privacy.intro_title') }}</h2>
        <p>{{ t('legal.privacy.intro') }}</p>

        <h2>{{ t('legal.privacy.data_collect_title') }}</h2>
        <p>{{ t('legal.privacy.data_collect_desc') }}</p>
        <ul>
          <li><strong>{{ t('legal.privacy.identity_data') }}</strong></li>
          <li><strong>{{ t('legal.privacy.contact_data') }}</strong></li>
        </ul>

        <h2>{{ t('legal.privacy.use_data_title') }}</h2>
        <p>{{ t('legal.privacy.use_data_desc') }}</p>
        <ul>
          <li>{{ t('legal.privacy.use_contract') }}</li>
          <li>{{ t('legal.privacy.use_legitimate') }}</li>
          <li>{{ t('legal.privacy.use_legal') }}</li>
        </ul>

        <h2>{{ t('legal.privacy.data_retention_title') }}</h2>
        <p>{{ t('legal.privacy.data_retention_desc') }}</p>

        <h2>{{ t('legal.privacy.third_party_title') }}</h2>
        <p>{{ t('legal.privacy.third_party_desc') }}</p>

        <h2>{{ t('legal.privacy.user_rights_title') }}</h2>
        <p>{{ t('legal.privacy.user_rights_desc') }}</p>
        <ul>
          <li>{{ t('legal.privacy.user_rights_access') }}</li>
          <li>{{ t('legal.privacy.user_rights_correction') }}</li>
          <li>{{ t('legal.privacy.user_rights_deletion') }}</li>
          <li>{{ t('legal.privacy.user_rights_portability') }}</li>
          <li>{{ t('legal.privacy.user_rights_restriction') }}</li>
          <li>{{ t('legal.privacy.user_rights_objection') }}</li>
        </ul>

        <h2>{{ t('legal.privacy.cross_border_title') }}</h2>
        <p>{{ t('legal.privacy.cross_border_desc') }}</p>

        <h2>{{ t('legal.privacy.cookies_title') }}</h2>
        <p>{{ t('legal.privacy.cookies_desc') }}</p>

        <h2>{{ t('legal.privacy.dpo_contact_title') }}</h2>
        <p>{{ t('legal.privacy.dpo_contact_desc') }}</p>

        <h2>{{ t('legal.privacy.children_title') }}</h2>
        <p>{{ t('legal.privacy.children_desc') }}</p>

        <h2>{{ t('legal.privacy.changes_title') }}</h2>
        <p>{{ t('legal.privacy.changes_desc') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#i18n'

const { t } = useI18n()

usePageOgImage({
  title: `Privacy Policy | ${t('seo.default_title')}`,
  description: t('legal.privacy.intro'),
  schema: {
    '@context': 'https://schema.org',
    '@type': 'WebPage',
    name: t('legal.privacy_title'),
    description: t('legal.privacy.intro'),
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
