<template>
  <div class="oem-options">
    <div class="oem-options__header">
      <h3 class="oem-options__title">{{ $t('oem.title') }}</h3>
      <p class="oem-options__subtitle">{{ subtitle }}</p>
    </div>

    <div class="oem-options__list">
      <div
        v-for="(option, index) in options"
        :key="index"
        class="oem-option"
        :class="{ 'oem-option--expanded': expandedIndex === index }"
      >
        <button
          class="oem-option__header"
          @click="toggleOption(index)"
          :aria-expanded="expandedIndex === index"
        >
          <div class="oem-option__icon">
            <Icon :name="option.icon" size="24" aria-hidden="true" />
          </div>
          <div class="oem-option__info">
            <h4 class="oem-option__title">{{ option.title }}</h4>
            <p class="oem-option__summary">{{ option.summary }}</p>
          </div>
          <Icon
            name="lucide:chevron-down"
            size="20"
            class="oem-option__chevron"
            :class="{ 'oem-option__chevron--open': expandedIndex === index }"
            aria-hidden="true"
          />
        </button>

        <div v-if="expandedIndex === index" class="oem-option__content">
          <div class="oem-option__body">
            <p v-if="option.description" class="oem-option__description">{{ option.description }}</p>

            <!-- Options grid -->
            <div v-if="option.choices" class="oem-option__choices">
              <div
                v-for="choice in option.choices"
                :key="choice.id"
                class="oem-choice"
                :class="{ 'oem-choice--selected': selectedChoices[index] === choice.id }"
                @click="selectChoice(index, choice.id)"
              >
                <div class="oem-choice__image">
                  <img
                    v-if="choice.image"
                    :src="choice.image"
                    :alt="choice.label"
                    loading="lazy"
                  />
                  <div v-else class="oem-choice__placeholder">
                    <Icon name="lucide:image" size="32" aria-hidden="true" />
                  </div>
                </div>
                <span class="oem-choice__label">{{ choice.label }}</span>
                <div v-if="selectedChoices[index] === choice.id" class="oem-choice__check">
                  <Icon name="lucide:check" size="16" aria-hidden="true" />
                </div>
              </div>
            </div>

            <!-- Pricing info -->
            <div v-if="option.pricing" class="oem-option__pricing">
              <span class="oem-option__pricing-label">{{ $t('oem.moq') }}:</span>
              <span class="oem-option__pricing-value">{{ option.pricing.moq }}</span>
            </div>
          </div>

          <!-- Action -->
          <div v-if="option.cta" class="oem-option__action">
            <Button
              :variant="option.cta.variant || 'highlight'"
              :size="option.cta.size || 'sm'"
              @click="handleCTA(option)"
            >
              {{ option.cta.label }}
            </Button>
          </div>
        </div>
      </div>
    </div>

    <!-- Summary -->
    <div v-if="showSummary" class="oem-options__summary">
      <h4 class="oem-options__summary-title">{{ $t('oem.title') }} {{ $t('form.summary') }}</h4>
      <ul class="oem-options__summary-list">
        <li v-for="(option, index) in options" :key="index">
          <span class="oem-options__summary-label">{{ option.title }}:</span>
          <span class="oem-options__summary-value">
            {{ getSelectedLabel(option, index) || $t('form.not_specified') }}
          </span>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface OEMChoice {
  id: string
  label: string
  image?: string
  description?: string
}

interface OEMOption {
  icon: string
  title: string
  summary: string
  description?: string
  choices?: OEMChoice[]
  pricing?: {
    moq: string
    leadTime?: string
  }
  cta?: {
    label: string
    variant?: string
    size?: string
    action?: string
  }
}

interface Props {
  options: OEMOption[]
  subtitle?: string
  showSummary?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  subtitle: 'Customize your order with these options',
  showSummary: false
})

const emit = defineEmits<{
  select: [optionIndex: number, choiceId: string]
  cta: [option: OEMOption]
}>()

const expandedIndex = ref(0)
const selectedChoices = ref<Record<number, string>>({})

const toggleOption = (index: number) => {
  expandedIndex.value = expandedIndex.value === index ? -1 : index
}

const selectChoice = (optionIndex: number, choiceId: string) => {
  selectedChoices.value[optionIndex] = choiceId
  emit('select', optionIndex, choiceId)
}

const handleCTA = (option: OEMOption) => {
  emit('cta', option)
}

const getSelectedLabel = (option: OEMOption, index: number) => {
  const choiceId = selectedChoices.value[index]
  if (!option.choices) return null
  const choice = option.choices.find(c => c.id === choiceId)
  return choice?.label || null
}

// Expand first option by default
onMounted(() => {
  if (props.options.length > 0) {
    expandedIndex.value = 0
  }
})
</script>

<style scoped>
.oem-options {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  overflow: hidden;
}

.oem-options__header {
  padding: var(--spacing-xl);
  background-color: var(--color-bg-alt);
  border-bottom: 1px solid var(--color-border);
}

.oem-options__title {
  font-size: var(--text-2xl);
  font-weight: 600;
  color: var(--color-primary);
  margin-bottom: var(--spacing-xs);
}

.oem-options__subtitle {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin: 0;
}

.oem-options__list {
  padding: var(--spacing-lg);
}

.oem-option {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  margin-bottom: var(--spacing-md);
  overflow: hidden;
}

.oem-option:last-child {
  margin-bottom: 0;
}

.oem-option__header {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  width: 100%;
  padding: var(--spacing-md);
  background: none;
  border: none;
  cursor: pointer;
  transition: background-color var(--transition-fast);
}

.oem-option__header:hover {
  background-color: var(--color-bg-alt);
}

.oem-option__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-md);
  color: var(--color-primary);
  flex-shrink: 0;
}

.oem-option__info {
  flex: 1;
  text-align: left;
}

.oem-option__title {
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-primary);
  margin-bottom: var(--spacing-xs);
}

.oem-option__summary {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin: 0;
}

.oem-option__chevron {
  flex-shrink: 0;
  color: var(--color-text-light);
  transition: transform var(--transition-fast);
}

.oem-option__chevron--open {
  transform: rotate(180deg);
}

.oem-option__content {
  padding: 0 var(--spacing-md) var(--spacing-md);
  border-top: 1px solid var(--color-border-light);
}

.oem-option__body {
  padding: var(--spacing-md) 0;
}

.oem-option__description {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  line-height: 1.6;
  margin-bottom: var(--spacing-lg);
}

.oem-option__choices {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-lg);
}

.oem-choice {
  position: relative;
  cursor: pointer;
}

.oem-choice__image {
  aspect-ratio: 1/1;
  border-radius: var(--radius-md);
  overflow: hidden;
  background-color: var(--color-bg-alt);
  border: 2px solid var(--color-border);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.oem-choice:hover .oem-choice__image,
.oem-choice--selected .oem-choice__image {
  border-color: var(--color-highlight);
}

.oem-choice__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.oem-choice__placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-lighter);
}

.oem-choice__label {
  display: block;
  margin-top: var(--spacing-sm);
  font-size: var(--text-xs);
  font-weight: 500;
  text-align: center;
  color: var(--color-text);
}

.oem-choice__check {
  position: absolute;
  top: var(--spacing-xs);
  right: var(--spacing-xs);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background-color: var(--color-highlight);
  border-radius: var(--radius-full);
  color: white;
}

.oem-option__pricing {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-md);
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-md);
}

.oem-option__pricing-label {
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

.oem-option__pricing-value {
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-primary);
}

.oem-option__action {
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--color-border-light);
}

/* Summary */
.oem-options__summary {
  padding: var(--spacing-lg);
  background-color: var(--color-bg-alt);
  border-top: 1px solid var(--color-border);
}

.oem-options__summary-title {
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-primary);
  margin-bottom: var(--spacing-md);
}

.oem-options__summary-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.oem-options__summary-list li {
  display: flex;
  justify-content: space-between;
  padding: var(--spacing-xs) 0;
  font-size: var(--text-sm);
}

.oem-options__summary-label {
  color: var(--color-text-light);
}

.oem-options__summary-value {
  font-weight: 500;
  color: var(--color-primary);
}

@media (max-width: 640px) {
  .oem-option__choices {
    grid-template-columns: repeat(3, 1fr);
  }

  .oem-choice__label {
    font-size: 10px;
  }
}
</style>
