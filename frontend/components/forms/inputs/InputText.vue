<template>
  <div class="input-text" :class="{ 'input-text--error': error, 'input-text--disabled': disabled }">
    <label v-if="label" :for="id" class="input-text__label">
      {{ label }}
      <span v-if="required" class="input-text__required">*</span>
    </label>

    <div class="input-text__wrapper">
      <input
        :id="id"
        ref="inputRef"
        :type="type"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :readonly="readonly"
        :required="required"
        :maxlength="maxlength"
        :minlength="minlength"
        :pattern="pattern"
        :inputmode="inputmode"
        class="input-text__field"
        @input="handleInput"
        @blur="handleBlur"
        @focus="handleFocus"
      />

      <button
        v-if="clearable && modelValue"
        type="button"
        class="input-text__clear"
        @click="handleClear"
      >
        <Icon name="lucide:x" size="14" />
      </button>
    </div>

    <div v-if="$slots.hint || hint" class="input-text__hint">
      <slot name="hint">{{ hint }}</slot>
    </div>

    <div v-if="error" class="input-text__error">
      <Icon name="lucide:alert-circle" size="12" />
      <span>{{ error }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface Props {
  id: string
  modelValue: string
  type?: string
  label?: string
  placeholder?: string
  hint?: string
  error?: string
  disabled?: boolean
  readonly?: boolean
  required?: boolean
  clearable?: boolean
  maxlength?: number
  minlength?: number
  pattern?: string
  inputmode?: 'text' | 'email' | 'tel' | 'numeric' | 'decimal' | 'search'
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  disabled: false,
  readonly: false,
  required: false,
  clearable: false
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  blur: []
  focus: []
}>()

const inputRef = ref<HTMLInputElement>()

const handleInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', target.value)
}

const handleBlur = () => {
  emit('blur')
}

const handleFocus = () => {
  emit('focus')
}

const handleClear = () => {
  emit('update:modelValue', '')
  inputRef.value?.focus()
}

defineExpose({
  inputRef,
  focus: () => inputRef.value?.focus()
})
</script>

<style scoped>
.input-text {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.input-text__label {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
}

.input-text__required {
  color: var(--color-error);
  margin-left: 2px;
}

.input-text__wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-text__field {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-base);
  font-family: inherit;
  color: var(--color-text);
  background-color: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.input-text__field::placeholder {
  color: var(--color-text-lighter);
}

.input-text__field:hover:not(:disabled) {
  border-color: var(--color-accent);
}

.input-text__field:focus {
  outline: none;
  border-color: var(--color-highlight);
  box-shadow: 0 0 0 3px rgba(255, 107, 74, 0.1);
}

.input-text__field:disabled {
  background-color: var(--color-bg-alt);
  color: var(--color-text-light);
  cursor: not-allowed;
}

.input-text__field:read-only {
  background-color: var(--color-bg-alt);
}

.input-text__clear {
  position: absolute;
  right: var(--spacing-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  color: var(--color-text-light);
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}

.input-text__clear:hover {
  color: var(--color-text);
  background-color: var(--color-border-dark);
}

.input-text__hint {
  font-size: var(--text-xs);
  color: var(--color-text-light);
}

.input-text__error {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: var(--text-xs);
  color: var(--color-error);
}

.input-text--error .input-text__field {
  border-color: var(--color-error);
}

.input-text--error .input-text__field:focus {
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
}
</style>
