<template>
  <div />
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: ['auth'] })

const localePath = useLocalePath()
await navigateTo(localePath('/admin'), { replace: true })
</script>
