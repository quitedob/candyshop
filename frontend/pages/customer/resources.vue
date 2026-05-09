<template>
  <div class="max-w-4xl mx-auto px-4 py-10">
    <h1 class="text-2xl font-bold text-gray-900">{{ t('customer.resources.title') }}</h1>
    <p class="mt-2 text-gray-600">{{ t('customer.resources.subtitle') }}</p>

    <div class="mt-8 grid gap-4 sm:grid-cols-2">
      <NuxtLink
        :to="localePath('/factory-quality')"
        class="block rounded-xl border border-gray-200 bg-white p-5 shadow-sm hover:border-orange-200 hover:shadow transition"
      >
        <h2 class="font-semibold text-gray-900">{{ t('customer.resources.factory') }}</h2>
        <p class="mt-1 text-sm text-gray-500">{{ t('customer.resources.factory_desc') }}</p>
      </NuxtLink>
      <NuxtLink
        :to="localePath('/oem-solutions')"
        class="block rounded-xl border border-gray-200 bg-white p-5 shadow-sm hover:border-orange-200 hover:shadow transition"
      >
        <h2 class="font-semibold text-gray-900">{{ t('customer.resources.oem') }}</h2>
        <p class="mt-1 text-sm text-gray-500">{{ t('customer.resources.oem_desc') }}</p>
      </NuxtLink>
      <NuxtLink
        :to="localePath('/cases-clients')"
        class="block rounded-xl border border-gray-200 bg-white p-5 shadow-sm hover:border-orange-200 hover:shadow transition"
      >
        <h2 class="font-semibold text-gray-900">{{ t('customer.resources.cases') }}</h2>
        <p class="mt-1 text-sm text-gray-500">{{ t('customer.resources.cases_desc') }}</p>
      </NuxtLink>
      <NuxtLink
        :to="localePath('/blog')"
        class="block rounded-xl border border-gray-200 bg-white p-5 shadow-sm hover:border-orange-200 hover:shadow transition"
      >
        <h2 class="font-semibold text-gray-900">{{ t('customer.resources.blog') }}</h2>
        <p class="mt-1 text-sm text-gray-500">{{ t('customer.resources.blog_desc') }}</p>
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'customer', middleware: ['auth'] })
const { t } = useI18n()
const localePath = useLocalePath()
</script>
