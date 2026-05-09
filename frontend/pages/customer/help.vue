<template>
  <div class="max-w-3xl mx-auto px-4 py-10">
    <h1 class="text-2xl font-bold text-gray-900">{{ t('customer.help.title') }}</h1>
    <p class="mt-2 text-gray-600">{{ t('customer.help.subtitle') }}</p>

    <ul class="mt-8 space-y-4">
      <li>
        <NuxtLink :to="localePath('/customer/cart')" class="text-orange-600 hover:underline font-medium">
          {{ t('customer.help.link_cart') }}
        </NuxtLink>
        <p class="text-sm text-gray-500 mt-1">{{ t('customer.help.desc_cart') }}</p>
      </li>
      <li>
        <NuxtLink :to="localePath('/customer/orders')" class="text-orange-600 hover:underline font-medium">
          {{ t('customer.help.link_orders') }}
        </NuxtLink>
        <p class="text-sm text-gray-500 mt-1">{{ t('customer.help.desc_orders') }}</p>
      </li>
      <li>
        <NuxtLink :to="localePath('/customer/inquiries/new')" class="text-orange-600 hover:underline font-medium">
          {{ t('customer.help.link_inquiry') }}
        </NuxtLink>
        <p class="text-sm text-gray-500 mt-1">{{ t('customer.help.desc_inquiry') }}</p>
      </li>
      <li>
        <NuxtLink :to="localePath('/faq')" class="text-orange-600 hover:underline font-medium">
          {{ t('customer.help.link_faq') }}
        </NuxtLink>
      </li>
      <li>
        <NuxtLink :to="localePath('/contact')" class="text-orange-600 hover:underline font-medium">
          {{ t('customer.help.link_contact') }}
        </NuxtLink>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'customer', middleware: ['auth'] })
const { t } = useI18n()
const localePath = useLocalePath()
</script>
