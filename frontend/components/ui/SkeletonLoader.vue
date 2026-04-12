<template>
  <div
    class="skeleton"
    :class="[`skeleton-${type}`, { 'animate': animate }]"
    :style="{ width: width || undefined, height: height || undefined }"
  ></div>
</template>

<script setup lang="ts">
defineProps<{
  type?: 'text' | 'title' | 'image' | 'card' | 'circle'
  animate?: boolean
  width?: string
  height?: string
}>()
</script>

<style scoped>
.skeleton {
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  border-radius: 4px;
}

.skeleton.animate {
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.skeleton-text {
  height: 1rem;
  width: 100%;
}

.skeleton-title {
  height: 1.5rem;
  width: 60%;
}

.skeleton-image {
  height: 200px;
  width: 100%;
  border-radius: 8px;
}

.skeleton-card {
  height: 300px;
  width: 100%;
  border-radius: 12px;
}

.skeleton-circle {
  width: 48px;
  height: 48px;
  border-radius: 50%;
}
</style>
