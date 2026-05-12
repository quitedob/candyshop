<template>
  <div class="mx-auto max-w-3xl px-4 py-12">
    <h1 class="text-3xl font-bold text-gray-900">{{ t('legal.data_processing_title') }}</h1>
    <p class="mt-2 text-sm text-gray-500">{{ t('legal.last_updated', { date: '2026-04-20' }) }}</p>
    <p class="mt-4 text-gray-700">{{ t('legal.data_processing_subtitle') }}</p>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.data_processing_scope') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.data_processing_scope_body') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.data_processing_ai') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.data_processing_ai_body') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.data_processing_retention') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.data_processing_retention_body') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.data_processing_contact') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.data_processing_contact_body') }}</p>
    </section>

    <div class="mt-12 flex flex-wrap gap-4 text-sm">
      <NuxtLink to="/privacy" class="text-orange-600 hover:underline">{{ t('legal.privacy_title') }}</NuxtLink>
      <NuxtLink to="/terms" class="text-orange-600 hover:underline">{{ t('legal.terms_title') }}</NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
const { t } = useI18n()
definePageMeta({ layout: 'default' })
useHead({
  title: () => t('legal.data_processing_title'),
})
</script>
