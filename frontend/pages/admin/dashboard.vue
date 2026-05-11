<template>
  <div />
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: ['auth'] })

await navigateTo('/admin', { replace: true })
</script>
