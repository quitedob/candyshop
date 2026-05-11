<template>
  <div class="auth-layout">
    <main class="auth-layout__main">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
useSeo({ noindex: true })
</script>

<style scoped>
/* 认证区专用令牌：羊皮纸底、陶土强调、无障碍焦点蓝（uxdesign） */
.auth-layout {
  --auth-focus-ring: #3898ec;
  --color-bg: #f5f4ed;
  --color-bg-alt: #faf9f5;
  --color-bg-dark: #eceae0;
  --color-primary: #141413;
  --color-text: #141413;
  --color-text-light: #5e5d59;
  --color-text-lighter: #87867f;
  --color-border: #e8e6dc;
  --color-border-light: #f0eee6;
  --color-highlight: #c96442;
  --color-highlight-hover: #b25436;
  --color-highlight-light: #d97757;
  --color-highlight-rgb: 201, 100, 66;
  --color-accent: #d97757;
  --color-accent-dark: #c96442;
  --color-error: #b53333;

  min-height: 100vh;
  min-height: 100dvh;
  background-color: var(--color-bg);
  position: relative;
  overflow-x: hidden;
}

.auth-layout::before {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: radial-gradient(ellipse 85% 55% at 100% 0%, rgba(201, 100, 66, 0.06) 0%, transparent 55%),
    radial-gradient(ellipse 70% 50% at 0% 100%, rgba(93, 92, 88, 0.04) 0%, transparent 50%);
}

.auth-layout__main {
  position: relative;
  z-index: 1;
}
</style>
