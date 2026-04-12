<template>
  <div class="input-select" :class="{ 'input-select--error': error, 'input-select--disabled': disabled }">
    <label v-if="label" :for="id" class="input-select__label">
      {{ label }}
      <span v-if="required" class="input-select__required">*</span>
    </label>

    <div class="input-select__wrapper">
      <select
        :id="id"
        :value="modelValue"
        :disabled="disabled"
        :required="required"
        class="input-select__field"
        @change="handleChange"
        @blur="handleBlur"
        @focus="handleFocus"
      >
        <option v-if="placeholder" value="" disabled selected>
          {{ placeholder }}
        </option>
        <option
          v-for="option in options"
          :key="option.value"
          :value="option.value"
          :disabled="option.disabled"
        >
          {{ option.label }}
        </option>
      </select>

      <div class="input-select__icon">
        <Icon name="lucide:chevron-down" size="16" />
      </div>
    </div>

    <div v-if="$slots.hint || hint" class="input-select__hint">
      <slot name="hint">{{ hint }}</slot>
    </div>

    <div v-if="error" class="input-select__error">
      <Icon name="lucide:alert-circle" size="12" />
      <span>{{ error }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
interface SelectOption {
  value: string | number | boolean
  label: string
  disabled?: boolean
}

interface Props {
  id: string
  modelValue: string | number | boolean | undefined
  options: SelectOption[]
  label?: string
  placeholder?: string
  hint?: string
  error?: string
  disabled?: boolean
  required?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  required: false
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number | boolean]
  blur: []
  focus: []
}>()

const handleChange = (event: Event) => {
  const target = event.target as HTMLSelectElement

  // Determine the type based on the first option's value
  const firstOption = props.options[0]
  if (typeof firstOption?.value === 'boolean') {
    emit('update:modelValue', target.value === 'true')
  } else if (typeof firstOption?.value === 'number') {
    emit('update:modelValue', Number(target.value))
  } else {
    emit('update:modelValue', target.value)
  }
}

const handleBlur = () => {
  emit('blur')
}

const handleFocus = () => {
  emit('focus')
}
</script>

<style scoped>
.input-select {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.input-select__label {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
}

.input-select__required {
  color: var(--color-error);
  margin-left: 2px;
}

.input-select__wrapper {
  position: relative;
}

.input-select__field {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  padding-right: calc(var(--spacing-md) + 24px);
  font-size: var(--text-base);
  font-family: inherit;
  color: var(--color-text);
  background-color: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  appearance: none;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.input-select__field:hover:not(:disabled) {
  border-color: var(--color-accent);
}

.input-select__field:focus {
  outline: none;
  border-color: var(--color-highlight);
  box-shadow: 0 0 0 3px rgba(255, 107, 74, 0.1);
}

.input-select__field:disabled {
  background-color: var(--color-bg-alt);
  color: var(--color-text-light);
  cursor: not-allowed;
}

.input-select__icon {
  position: absolute;
  right: var(--spacing-md);
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
  color: var(--color-text-light);
}

.input-select__hint {
  font-size: var(--text-xs);
  color: var(--color-text-light);
}

.input-select__error {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: var(--text-xs);
  color: var(--color-error);
}

.input-select--error .input-select__field {
  border-color: var(--color-error);
}

.input-select--error .input-select__field:focus {
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
}
</style>
