<template>
  <div class="input-textarea" :class="{ 'input-textarea--error': error, 'input-textarea--disabled': disabled }">
    <label v-if="label" :for="id" class="input-textarea__label">
      {{ label }}
      <span v-if="required" class="input-textarea__required">*</span>
    </label>

    <textarea
      :id="id"
      ref="textareaRef"
      :name="name"
      :value="modelValue"
      :placeholder="placeholder"
      :disabled="disabled"
      :readonly="readonly"
      :required="required"
      :rows="rows"
      :maxlength="maxlength"
      :minlength="minlength"
      :autocomplete="autocomplete"
      class="input-textarea__field"
      @input="handleInput"
      @blur="handleBlur"
      @focus="handleFocus"
    />

    <div v-if="maxlength" class="input-textarea__counter">
      {{ characterCount }} / {{ maxlength }}
    </div>

    <div v-if="error" class="input-textarea__error">
      <Icon name="lucide:alert-circle" size="12" aria-hidden="true" />
      <span>{{ error }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface Props {
  id: string
  modelValue: string
  name?: string
  label?: string
  placeholder?: string
  error?: string
  disabled?: boolean
  readonly?: boolean
  required?: boolean
  rows?: number
  maxlength?: number
  minlength?: number
  autocomplete?: string
  autoResize?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  rows: 4,
  disabled: false,
  readonly: false,
  required: false,
  autoResize: false
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  blur: []
  focus: []
}>()

const textareaRef = ref<HTMLTextAreaElement>()

const characterCount = computed(() => props.modelValue.length)

const handleInput = (event: Event) => {
  const target = event.target as HTMLTextAreaElement
  emit('update:modelValue', target.value)

  if (props.autoResize) {
    autoResize()
  }
}

const handleBlur = () => {
  emit('blur')
}

const handleFocus = () => {
  emit('focus')
}

const autoResize = () => {
  if (textareaRef.value) {
    textareaRef.value.style.height = 'auto'
    textareaRef.value.style.height = `${textareaRef.value.scrollHeight}px`
  }
}

defineExpose({
  textareaRef,
  focus: () => textareaRef.value?.focus()
})
</script>

<style scoped>
.input-textarea {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.input-textarea__label {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
}

.input-textarea__required {
  color: var(--color-error);
  margin-left: 2px;
}

.input-textarea__field {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-base);
  font-family: inherit;
  line-height: 1.6;
  color: var(--color-text);
  background-color: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  resize: vertical;
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.input-textarea__field::placeholder {
  color: var(--color-text-lighter);
}

.input-textarea__field:hover:not(:disabled) {
  border-color: var(--color-accent);
}

.input-textarea__field:focus {
  outline: none;
  border-color: var(--color-highlight);
  box-shadow: 0 0 0 3px rgba(var(--color-highlight-rgb), 0.1);
}

.input-textarea__field:disabled {
  background-color: var(--color-bg-alt);
  color: var(--color-text-light);
  cursor: not-allowed;
  resize: none;
}

.input-textarea__field:read-only {
  background-color: var(--color-bg-alt);
}

.input-textarea__counter {
  align-self: flex-end;
  font-size: var(--text-xs);
  color: var(--color-text-light);
}

.input-textarea__error {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: var(--text-xs);
  color: var(--color-error);
}

.input-textarea--error .input-textarea__field {
  border-color: var(--color-error);
}

.input-textarea--error .input-textarea__field:focus {
  box-shadow: 0 0 0 3px rgba(var(--color-error-rgb), 0.1);
}
</style>
