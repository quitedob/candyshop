<!-- OgImage/Default.vue — 品牌默认 OG 动态模板（nuxt-og-image / Satori） -->
<template>
  <div
    style="
      display: flex;
      flex-direction: column;
      justify-content: center;
      align-items: flex-start;
      width: 100%;
      height: 100%;
      background: linear-gradient(135deg, #1c1917 0%, #c2410c 100%);
      padding: 64px;
      font-family: 'Work Sans', sans-serif;
    "
  >
    <div
      style="
        font-size: 28px;
        font-weight: 600;
        color: #fed7aa;
        margin-bottom: 24px;
        letter-spacing: 0.05em;
      "
    >
      CandyPro OEM
    </div>
    <h1
      style="
        font-size: 56px;
        font-weight: 700;
        color: #ffffff;
        line-height: 1.2;
        margin: 0 0 24px 0;
        max-width: 900px;
      "
    >
      {{ title }}
    </h1>
    <p
      v-if="description"
      style="
        font-size: 28px;
        color: #fef3c7;
        line-height: 1.4;
        margin: 0;
        max-width: 800px;
      "
    >
      {{ description }}
    </p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title?: string
  description?: string
}>()
</script>
