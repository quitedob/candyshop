<template>
  <div class="mx-auto max-w-3xl px-4 py-12">
    <h1 class="text-3xl font-bold text-gray-900">{{ t('legal.returns_title') }}</h1>
    <p class="mt-2 text-sm text-gray-500">{{ t('legal.last_updated', { date: '2026-05-22' }) }}</p>
    <p class="mt-4 text-gray-700">{{ t('legal.returns_intro') }}</p>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.returns_conditions_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.returns_conditions_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.returns_timeframe_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.returns_timeframe_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.returns_process_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.returns_process_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.returns_refund_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.returns_refund_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.returns_exceptions_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.returns_exceptions_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.returns_shipping_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.returns_shipping_desc') }}</p>
    </section>

    <div class="mt-12 flex flex-wrap gap-4 text-sm">
      <NuxtLink to="/terms" class="text-orange-600 hover:underline">{{ t('legal.terms_title') }}</NuxtLink>
      <NuxtLink to="/legal/shipping" class="text-orange-600 hover:underline">{{ t('legal.shipping_title') }}</NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
const { t } = useI18n()
definePageMeta({ layout: 'default' })
usePageOgImage({
  title: t('legal.returns_title'),
  description: t('legal.returns_intro'),
  schema: {
    '@context': 'https://schema.org',
    '@type': 'WebPage',
    name: t('legal.returns_title'),
    description: t('legal.returns_intro'),
  },
})
</script>
