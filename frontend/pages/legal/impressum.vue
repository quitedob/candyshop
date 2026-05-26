<template>
  <div class="mx-auto max-w-3xl px-4 py-12">
    <h1 class="text-3xl font-bold text-gray-900">{{ t('legal.impressum_title') }}</h1>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.impressum_company_name') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.impressum_company_name_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.impressum_address') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.impressum_address_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.impressum_representative') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.impressum_representative_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.impressum_register') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.impressum_register_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.impressum_vat') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.impressum_vat_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.impressum_contact') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.impressum_contact_desc') }}</p>
    </section>

    <section class="mt-10 space-y-4">
      <h2 class="text-xl font-semibold text-gray-900">{{ t('legal.impressum_disclaimer') }}</h2>
      <p class="text-gray-700 leading-relaxed">{{ t('legal.impressum_disclaimer_desc') }}</p>
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
usePageOgImage({
  title: t('legal.impressum_title'),
  description: t('legal.impressum_company_name_desc'),
  schema: {
    '@context': 'https://schema.org',
    '@type': 'WebPage',
    name: t('legal.impressum_title'),
    description: t('legal.impressum_company_name_desc'),
  },
})
</script>
