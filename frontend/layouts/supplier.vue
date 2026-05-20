<template>
  <div class="min-h-screen bg-stone-50">
    <header class="bg-white border-b border-gray-200">
      <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div class="flex h-16 items-center justify-between">
          <div class="flex items-center gap-3">
            <NuxtLink :to="localePath('/supplier')" class="text-lg font-bold text-gray-900">
              CandyPro<span class="text-orange-500">Supplier</span>
            </NuxtLink>
          </div>
          <nav class="flex items-center gap-6">
            <NuxtLink to="/supplier" class="text-sm font-medium text-gray-600 hover:text-gray-900">{{ t('supplier.nav_dashboard') }}</NuxtLink>
            <NuxtLink to="/supplier/profile" class="text-sm font-medium text-gray-600 hover:text-gray-900">{{ t('supplier.nav_profile') }}</NuxtLink>
            <NuxtLink to="/supplier/purchase-orders" class="text-sm font-medium text-gray-600 hover:text-gray-900">{{ t('supplier.nav_pos') }}</NuxtLink>
          </nav>
        </div>
      </div>
    </header>
    <main class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-8">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
const localePath = useLocalePath()
const { t } = useI18n()
</script>
