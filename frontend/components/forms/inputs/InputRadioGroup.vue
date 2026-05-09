<template>
  <div class="input-radio-group" :class="{ 'input-radio-group--error': error }">
    <label v-if="label" class="input-radio-group__label">
      {{ label }}
      <span v-if="required" class="input-radio-group__required">*</span>
    </label>

    <div class="input-radio-group__options">
      <label
        v-for="option in options"
        :key="String(option.value)"
        class="input-radio-group__option"
        :class="{ 'input-radio-group__option--checked': modelValue === option.value }"
      >
        <input
          type="radio"
          :name="name"
          :value="option.value"
          :checked="modelValue === option.value"
          :disabled="option.disabled || disabled"
          @change="handleChange"
        />

        <span class="input-radio-group__radio">
          <span class="input-radio-group__radio-dot"></span>
        </span>

        <span class="input-radio-group__label-text">{{ option.label }}</span>
      </label>
    </div>

    <div v-if="error" class="input-radio-group__error">
      <Icon name="lucide:alert-circle" size="12" />
      <span>{{ error }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
interface RadioOption {
  value: string | number | boolean
  label: string
  disabled?: boolean
}

interface Props {
  name: string
  modelValue: string | number | boolean | undefined
  options: RadioOption[]
  label?: string
  error?: string
  disabled?: boolean
  required?: boolean
}

defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: string | number | boolean]
}>()

const handleChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  const value = target.value

  // Determine the type based on the first option's value
  // This is passed from the parent component
  if (value === 'true') {
    emit('update:modelValue', true)
  } else if (value === 'false') {
    emit('update:modelValue', false)
  } else {
    emit('update:modelValue', value)
  }
}
</script>

<style scoped>
.input-radio-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.input-radio-group__label {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
}

.input-radio-group__required {
  color: var(--color-error);
  margin-left: 2px;
}

.input-radio-group__options {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-md);
}

.input-radio-group__option {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-sm);
  color: var(--color-text);
  background-color: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.input-radio-group__option:hover:not(.input-radio-group__option--disabled) {
  border-color: var(--color-accent);
  background-color: var(--color-bg-alt);
}

.input-radio-group__option--checked {
  border-color: var(--color-highlight);
  background-color: rgba(var(--color-highlight-rgb), 0.05);
}

.input-radio-group__option input[type="radio"] {
  position: absolute;
  opacity: 0;
  pointer-events: none;
}

.input-radio-group__radio {
  position: relative;
  width: 18px;
  height: 18px;
  border: 2px solid var(--color-border);
  border-radius: var(--radius-full);
  transition: all var(--transition-fast);
  flex-shrink: 0;
}

.input-radio-group__option:hover .input-radio-group__radio {
  border-color: var(--color-accent);
}

.input-radio-group__option--checked .input-radio-group__radio {
  border-color: var(--color-highlight);
}

.input-radio-group__radio-dot {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) scale(0);
  width: 8px;
  height: 8px;
  background-color: var(--color-highlight);
  border-radius: var(--radius-full);
  transition: transform var(--transition-fast);
}

.input-radio-group__option--checked .input-radio-group__radio-dot {
  transform: translate(-50%, -50%) scale(1);
}

.input-radio-group__label-text {
  font-size: var(--text-sm);
}

.input-radio-group__error {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: var(--text-xs);
  color: var(--color-error);
}
</style>
