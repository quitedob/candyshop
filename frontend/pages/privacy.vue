<template>
  <div class="legal-page">
    <!-- Breadcrumb -->
    <div class="container pt-8">
      <Breadcrumb :items="[{ label: t('legal.privacy_title') }]" />
    </div>

    <div class="container section">
      <h1 class="text-4xl font-bold mb-4">{{ t('legal.privacy_title') }}</h1>
      <p class="text-gray-500 mb-8">{{ t('legal.last_updated', { date: formatDate(new Date().toISOString()) }) }}</p>

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
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#i18n'

const { t } = useI18n()
const { formatDate } = useDisplay()

useSeo({
  title: `Privacy Policy | ${t('seo.default_title')}`
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
