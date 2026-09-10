<template>
  <div class="mx-auto max-w-3xl px-4 py-12">
    <h1 class="text-3xl font-bold text-gray-900">{{ t('legal.shipping_title') }}</h1>
    <p class="mt-2 text-sm text-gray-500">{{ t('legal.last_updated', { date: '2026-05-22' }) }}</p>
    <p class="mt-4 text-gray-700">{{ t('legal.shipping_intro') }}</p>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.shipping_methods_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.shipping_methods_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.shipping_incoterms_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.shipping_incoterms_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.shipping_customs_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.shipping_customs_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.shipping_timeline_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.shipping_timeline_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.shipping_insurance_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.shipping_insurance_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.shipping_restrictions_title') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.shipping_restrictions_desc') }}</p>
    </section>

    <div class="mt-12 flex flex-wrap gap-4 text-sm">
      <NuxtLink :to="localePath('/terms')" class="text-orange-600 hover:underline">{{ t('legal.terms_title') }}</NuxtLink>
      <NuxtLink :to="localePath('/legal/returns')" class="text-orange-600 hover:underline">{{ t('legal.returns_title') }}</NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
const { t } = useI18n()
const localePath = useLocalePath()
definePageMeta({ layout: 'default' })
usePageOgImage({
  title: t('legal.shipping_title'),
  description: t('legal.shipping_intro'),
  schema: {
    '@context': 'https://schema.org',
    '@type': 'WebPage',
    name: t('legal.shipping_title'),
    description: t('legal.shipping_intro'),
  },
})
</script>
