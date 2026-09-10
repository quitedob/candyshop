<template>
  <div class="mx-auto max-w-3xl px-4 py-12">
    <h1 class="text-3xl font-bold text-gray-900">{{ t('legal.cookie_title') }}</h1>
    <p class="mt-2 text-sm text-gray-500">{{ t('legal.last_updated', { date: '2026-05-22' }) }}</p>
    <p class="mt-4 text-gray-700">{{ t('legal.cookie_intro') }}</p>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.cookie_what_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.cookie_what_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.cookie_types_title') }}</h2>

      <h3 class="text-lg font-semibold text-gray-800 mt-6">{{ t('legal.cookie_essential_title') }}</h3>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.cookie_essential_desc') }}</p>

      <h3 class="text-lg font-semibold text-gray-800 mt-6">{{ t('legal.cookie_functional_title') }}</h3>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.cookie_functional_desc') }}</p>

      <h3 class="text-lg font-semibold text-gray-800 mt-6">{{ t('legal.cookie_analytics_title') }}</h3>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.cookie_analytics_desc') }}</p>

      <h3 class="text-lg font-semibold text-gray-800 mt-6">{{ t('legal.cookie_marketing_title') }}</h3>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.cookie_marketing_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.cookie_third_party_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.cookie_third_party_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.cookie_manage_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.cookie_manage_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.cookie_consent_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.cookie_consent_desc') }}</p>
    </section>

    <div class="mt-12 flex flex-wrap gap-4 text-sm">
      <NuxtLink :to="localePath('/privacy')" class="text-orange-600 hover:underline">{{ t('legal.privacy_title') }}</NuxtLink>
      <NuxtLink :to="localePath('/terms')" class="text-orange-600 hover:underline">{{ t('legal.terms_title') }}</NuxtLink>
      <NuxtLink :to="localePath('/legal/data-processing')" class="text-orange-600 hover:underline">{{ t('legal.data_processing_title') }}</NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
const { t } = useI18n()
const localePath = useLocalePath()
definePageMeta({ layout: 'default' })
usePageOgImage({
  title: t('legal.cookie_title'),
  description: t('legal.cookie_intro'),
  schema: {
    '@context': 'https://schema.org',
    '@type': 'WebPage',
    name: t('legal.cookie_title'),
    description: t('legal.cookie_intro'),
  },
})
</script>
