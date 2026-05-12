<template>
  <div class="input-file" :class="{ 'input-file--error': error, 'input-file--disabled': disabled }">
    <label v-if="label" :for="id" class="input-file__label">
      {{ label }}
      <span v-if="required" class="input-file__required">*</span>
    </label>

    <div class="input-file__dropzone" :class="{ 'input-file__dropzone--dragover': isDragover }">
      <input
        ref="inputRef"
        type="file"
        :id="id"
        :accept="accept"
        :multiple="maxFiles > 1"
        :disabled="disabled"
        class="input-file__input"
        @change="handleFileChange"
        @dragenter="isDragover = true"
        @dragleave="isDragover = false"
        @dragover.prevent
        @drop.prevent="handleDrop"
      />

      <label :for="id" class="input-file__label-content">
        <div class="input-file__icon">
          <Icon name="lucide:upload-cloud" size="32" />
        </div>

        <div class="input-file__text">
          <p class="input-file__text-primary">
            {{ t('input_file.upload') }}
          </p>
          <p class="input-file__text-secondary">
            {{ acceptText }}
          </p>
        </div>
      </label>
    </div>

    <div v-if="error" class="input-file__error">
      <Icon name="lucide:alert-circle" size="12" />
      <span>{{ error }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface Props {
  id: string
  label?: string
  error?: string
  disabled?: boolean
  required?: boolean
  accept?: string
  maxFiles?: number
  maxSize?: number // in MB
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  required: false,
  accept: 'image/jpeg,image/png,image/webp,application/pdf',
  maxFiles: 1,
  maxSize: 5
})

const emit = defineEmits<{
  'files-selected': [files: FileList]
  error: [message: string]
}>()

const { t } = useI18n()
const inputRef = ref<HTMLInputElement>()
const isDragover = ref(false)

const acceptText = computed(() => {
  const extensions = props.accept.split(',').map(ext => {
    const map: Record<string, string> = {
      'image/jpeg': 'JPG',
      'image/png': 'PNG',
      'image/webp': 'WebP',
      'application/pdf': 'PDF'
    }
    return map[ext] || ext
  })
  return `${t('input_file.max_size', { size: props.maxSize })}, ${extensions.join(', ')}`
})

const validateFile = (file: File): boolean => {
  // Check file size
  const maxSizeBytes = props.maxSize * 1024 * 1024
  if (file.size > maxSizeBytes) {
    emit('error', t('input_file.error_too_large', { name: file.name, max: props.maxSize }))
    return false
  }

  // Check file type
  if (props.accept) {
    const acceptedTypes = props.accept.split(',')
    if (!acceptedTypes.includes(file.type)) {
      emit('error', t('input_file.error_type', { name: file.name }))
      return false
    }
  }

  return true
}

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files) {
    processFiles(target.files)
  }
}

const handleDrop = (event: DragEvent) => {
  isDragover.value = false

  if (event.dataTransfer?.files) {
    processFiles(event.dataTransfer.files)
  }
}

const processFiles = (files: FileList) => {
  // Check file count
  if (files.length > props.maxFiles) {
    emit('error', t('input_file.error_max_files', { count: props.maxFiles }))
    return
  }

  // Validate each file
  const validFiles: File[] = []
  for (let i = 0; i < files.length; i++) {
    if (validateFile(files[i])) {
      validFiles.push(files[i])
    }
  }

  if (validFiles.length > 0) {
    emit('files-selected', files)
  }
}

const clear = () => {
  if (inputRef.value) {
    inputRef.value.value = ''
  }
}

defineExpose({
  clear
})
</script>

<style scoped>
.input-file {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.input-file__label {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
}

.input-file__required {
  color: var(--color-error);
  margin-left: 2px;
}

.input-file__dropzone {
  position: relative;
  border: 2px dashed var(--color-border);
  border-radius: var(--radius-lg);
  background-color: var(--color-bg-alt);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.input-file__dropzone--dragover {
  border-color: var(--color-highlight);
  background-color: rgba(var(--color-highlight-rgb), 0.05);
}

.input-file__input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
}

.input-file__label-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--spacing-xl);
  cursor: pointer;
}

.input-file__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  margin-bottom: var(--spacing-md);
  color: var(--color-text-light);
  background-color: white;
  border-radius: var(--radius-full);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.input-file__dropzone:hover .input-file__icon {
  color: var(--color-highlight);
  background-color: rgba(var(--color-highlight-rgb), 0.1);
}

.input-file__text {
  text-align: center;
}

.input-file__text-primary {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
  margin-bottom: var(--spacing-xs);
}

.input-file__text-secondary {
  font-size: var(--text-xs);
  color: var(--color-text-light);
  margin: 0;
}

.input-file__error {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: var(--text-xs);
  color: var(--color-error);
}

.input-file--error .input-file__dropzone {
  border-color: var(--color-error);
}
</style>
