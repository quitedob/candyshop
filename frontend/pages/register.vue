<script setup lang="ts">
import { useLocalePath } from '#i18n'

const localePath = useLocalePath()
await navigateTo(localePath('/auth/register'), { replace: true })
</script>

<template>
  <div />
</template>
