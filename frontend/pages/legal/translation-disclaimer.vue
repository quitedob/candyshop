<template>
  <div class="mx-auto max-w-3xl px-4 py-12">
    <h1 class="text-3xl font-bold text-gray-900">{{ t('legal.translation_disclaimer_title') }}</h1>
    <p class="mt-2 text-sm text-gray-500">{{ t('legal.last_updated', { date: '2026-05-22' }) }}</p>
    <p class="mt-4 text-gray-700">{{ t('legal.translation_disclaimer_intro') }}</p>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.translation_accuracy_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.translation_accuracy_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.translation_authoritative_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.translation_authoritative_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.translation_legal_effect_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.translation_legal_effect_desc') }}</p>
    </section>

    <div class="mt-12 flex flex-wrap gap-4 text-sm">
      <NuxtLink :to="localePath('/privacy')" class="text-orange-600 hover:underline">{{ t('legal.privacy_title') }}</NuxtLink>
      <NuxtLink :to="localePath('/terms')" class="text-orange-600 hover:underline">{{ t('legal.terms_title') }}</NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
const { t } = useI18n()
const localePath = useLocalePath()
definePageMeta({ layout: 'default' })
usePageOgImage({
  title: t('legal.translation_disclaimer_title'),
  description: t('legal.translation_disclaimer_intro'),
  schema: {
    '@context': 'https://schema.org',
    '@type': 'WebPage',
    name: t('legal.translation_disclaimer_title'),
    description: t('legal.translation_disclaimer_intro'),
  },
})
</script>
