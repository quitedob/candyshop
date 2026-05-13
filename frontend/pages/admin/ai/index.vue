<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-2">{{ t('admin.ai.title') }}</h1>
    <p class="text-gray-600">{{ t('admin.ai.description') }}</p>
    <div class="mt-8 p-8 bg-white rounded-2xl border border-orange-100 text-center">
      <Icon name="heroicons:sparkles" class="h-16 w-16 text-orange-300 mx-auto mb-4" />
      <p class="text-gray-500">{{ t('admin.ai.coming_soon') }}</p>
    </div>
  </div>
</template>
<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: ['auth'] })
const { t } = useI18n()
</script>
