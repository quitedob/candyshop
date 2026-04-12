<template>
  <div class="auth-layout">
    <main class="auth-layout__main">
      <slot />
    </main>
  </div>
</template>

<style scoped>
.auth-layout {
  min-height: 100vh;
  background: linear-gradient(135deg, var(--color-bg) 0%, var(--color-bg-alt) 50%, var(--color-bg-dark) 100%);
  position: relative;
  overflow: hidden;
}

.auth-layout::before {
  content: '';
  position: absolute;
  top: -50%;
  right: -20%;
  width: 80%;
  height: 200%;
  background: radial-gradient(ellipse at center, rgba(212, 165, 116, 0.08) 0%, transparent 70%);
  pointer-events: none;
}

.auth-layout::after {
  content: '';
  position: absolute;
  bottom: -30%;
  left: -10%;
  width: 60%;
  height: 150%;
  background: radial-gradient(ellipse at center, rgba(255, 107, 74, 0.05) 0%, transparent 60%);
  pointer-events: none;
}

.auth-layout__main {
  position: relative;
  z-index: 1;
}
</style>
