<template>
  <div class="process-flow" :class="`process-flow--${variant}`">
    <div
      v-for="(step, index) in steps"
      :key="step.id"
      class="process-step"
      :class="{ 'process-step--active': index === activeStep }"
    >
      <!-- Connector -->
      <div v-if="index > 0" class="process-step__connector">
        <div class="process-step__line"></div>
        <div class="process-step__arrow">
          <Icon name="lucide:chevron-down" size="16" />
        </div>
      </div>

      <!-- Icon/Number -->
      <div class="process-step__indicator">
        <div v-if="step.icon" class="process-step__icon">
          <Icon :name="step.icon" size="24" />
        </div>
        <div v-else class="process-step__number">{{ index + 1 }}</div>
      </div>

      <!-- Content -->
      <div class="process-step__content">
        <h3 class="process-step__title">{{ step.title }}</h3>
        <p v-if="step.description" class="process-step__description">{{ step.description }}</p>

        <!-- Details -->
        <div v-if="step.details" class="process-step__details">
          <div v-for="detail in step.details" :key="detail.label" class="process-step__detail">
            <span class="process-step__detail-label">{{ detail.label }}:</span>
            <span class="process-step__detail-value">{{ detail.value }}</span>
          </div>
        </div>

        <!-- Status -->
        <div v-if="step.status" class="process-step__status">
          <span
            class="process-step__status-badge"
            :class="`process-step__status-badge--${step.status}`"
          >
            {{ step.status }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface ProcessDetail {
  label: string
  value: string
}

interface ProcessStep {
  id: string
  title: string
  description?: string
  icon?: string
  details?: ProcessDetail[]
  status?: 'pending' | 'in-progress' | 'completed'
}

interface Props {
  steps: ProcessStep[]
  variant?: 'default' | 'compact' | 'horizontal' | 'vertical'
  activeStep?: number
}

defineProps<Props>()
</script>

<style scoped>
.process-flow {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

/* Default/Vertical variant */
.process-flow--default,
.process-flow--vertical {
  flex-direction: column;
}

.process-step {
  position: relative;
  display: flex;
  gap: var(--spacing-md);
}

.process-step__connector {
  position: absolute;
  top: 0;
  left: 20px;
  width: 40px;
  height: 100%;
  pointer-events: none;
}

.process-step__line {
  position: absolute;
  left: 50%;
  top: 40px;
  bottom: -20px;
  width: 2px;
  background: linear-gradient(to bottom, var(--color-border) 0%, var(--color-accent) 100%);
}

.process-step__arrow {
  position: absolute;
  bottom: -5px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  background-color: var(--color-bg);
  border: 2px solid var(--color-accent);
  border-radius: var(--radius-full);
  color: var(--color-accent);
}

.process-step:last-child .process-step__connector {
  display: none;
}

.process-step__indicator {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
}

.process-step__icon,
.process-step__number {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background-color: var(--color-primary);
  color: var(--color-text-on-primary);
  border-radius: var(--radius-full);
}

.process-step__number {
  font-weight: 600;
}

.process-step--active .process-step__icon,
.process-step--active .process-step__number {
  background-color: var(--color-highlight);
  box-shadow: 0 0 0 4px rgba(var(--color-highlight-rgb), 0.2);
}

.process-step__content {
  flex: 1;
  padding: var(--spacing-lg);
  background-color: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.process-step--active .process-step__content {
  border-color: var(--color-highlight);
  box-shadow: var(--shadow-md);
}

.process-step__title {
  font-size: var(--text-lg);
  font-weight: 600;
  margin-bottom: var(--spacing-xs);
}

.process-step__description {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin-bottom: var(--spacing-md);
}

.process-step__details {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-md);
}

.process-step__detail {
  display: flex;
  gap: var(--spacing-xs);
  font-size: var(--text-sm);
}

.process-step__detail-label {
  color: var(--color-text-light);
}

.process-step__detail-value {
  font-weight: 500;
  color: var(--color-primary);
}

.process-step__status-badge {
  display: inline-block;
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: var(--text-xs);
  font-weight: 600;
  border-radius: var(--radius-sm);
}

.process-step__status-badge--pending {
  background-color: var(--color-bg-alt);
  color: var(--color-text-light);
}

.process-step__status-badge--in-progress {
  background-color: rgba(var(--color-warning-rgb), 0.1);
  color: var(--color-warning);
}

.process-step__status-badge--completed {
  background-color: rgba(var(--color-success-rgb), 0.1);
  color: var(--color-success);
}

/* Horizontal variant */
.process-flow--horizontal {
  flex-direction: row;
  overflow-x: auto;
  padding-bottom: var(--spacing-lg);
}

.process-flow--horizontal .process-step {
  flex-direction: column;
  text-align: center;
  min-width: 150px;
}

.process-flow--horizontal .process-step__connector {
  top: 20px;
  left: 100%;
  width: 40px;
  height: 40px;
}

.process-flow--horizontal .process-step__line {
  top: 50%;
  left: 0;
  right: 0;
  bottom: auto;
  width: 100%;
  height: 2px;
  transform: translateY(-50%);
}

.process-flow--horizontal .process-step__arrow {
  top: 50%;
  left: auto;
  right: -10px;
  bottom: auto;
  transform: translateY(-50%);
}

.process-flow--horizontal .process-step__content {
  padding: var(--spacing-md);
}

/* Compact variant */
.process-flow--compact .process-step__indicator {
  width: 32px;
  height: 32px;
}

.process-flow--compact .process-step__content {
  padding: var(--spacing-sm) var(--spacing-md);
}

.process-flow--compact .process-step__title {
  font-size: var(--text-base);
}

/* Connected variant (with line behind) */
.process-flow--connected {
  position: relative;
  padding-left: 40px;
}

.process-flow--connected::before {
  content: '';
  position: absolute;
  left: 20px;
  top: 20px;
  bottom: 20px;
  width: 2px;
  background: linear-gradient(to bottom, var(--color-accent) 0%, var(--color-highlight) 100%);
}

.process-flow--connected .process-step {
  gap: var(--spacing-lg);
}

.process-flow--connected .process-step__connector {
  display: none;
}
</style>
