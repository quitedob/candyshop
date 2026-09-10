<template>
  <div class="cert-list" :class="`cert-list--${variant}`">
    <div
      v-for="cert in certifications"
      :key="cert.id"
      class="cert-item"
      :class="{ 'cert-item--featured': cert.featured }"
    >
      <!-- Badge -->
      <div class="cert-item__badge">
        <img
          v-if="cert.badgeUrl"
          :src="cert.badgeUrl"
          :alt="cert.name"
          class="cert-item__badge-img"
        />
        <div v-else class="cert-item__badge-placeholder">
          <Icon name="lucide:award" size="24" />
          <span>{{ getAbbreviation(cert.name) }}</span>
        </div>
      </div>

      <!-- Content -->
      <div class="cert-item__content">
        <h3 class="cert-item__name">{{ cert.name }}</h3>
        <p class="cert-item__description">{{ cert.description }}</p>

        <div class="cert-item__meta">
          <span class="cert-item__issuer">
            <Icon name="lucide:building" size="14" />
            {{ cert.issuer }}
          </span>
          <span class="cert-item__valid">
            <Icon name="lucide:calendar" size="14" />
            {{ t('factory.valid_until') }} {{ cert.validUntil }}
          </span>
        </div>
      </div>

      <!-- Download -->
      <div class="cert-item__actions">
        <a
          v-if="cert.certificateUrl"
          :href="cert.certificateUrl"
          :download="`certificate-${cert.id}.pdf`"
          class="cert-item__download"
        >
          <Icon name="lucide:download" size="16" />
          {{ t('factory.download') }}
        </a>
        <NuxtLink v-else :to="localePath('/contact')" class="cert-item__download">
          <Icon name="lucide:mail" size="16" />
          {{ t('nav.contact') }}
        </NuxtLink>
        <button
          v-if="cert.verifyUrl"
          @click="openVerify(cert)"
          class="cert-item__verify"
        >
          <Icon name="lucide:external-link" size="16" />
          {{ t('factory.verify') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Certification {
  id: string
  name: string
  abbreviation: string
  description: string
  issuer: string
  validUntil: string
  badgeUrl?: string
  certificateUrl?: string
  verifyUrl?: string
  featured?: boolean
}

interface Props {
  certifications: Certification[]
  variant?: 'default' | 'compact' | 'grid'
}

defineProps<Props>()

const { t } = useI18n()
const localePath = useLocalePath()

const getAbbreviation = (name: string) => {
  return name.split(' ').map(word => word[0]).join('').substring(0, 4)
}

const openVerify = (cert: Certification) => {
  if (cert.verifyUrl) {
    window.open(cert.verifyUrl, '_blank', 'noopener,noreferrer')
  }
}
</script>

<style scoped>
.cert-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.cert-list--grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: var(--spacing-xl);
}

.cert-item {
  display: flex;
  gap: var(--spacing-md);
  padding: var(--spacing-lg);
  background-color: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
}

.cert-item:hover {
  border-color: var(--color-accent);
  box-shadow: var(--shadow-md);
}

.cert-item--featured {
  border-color: var(--color-highlight);
  background: linear-gradient(to right, #fff5f2 0%, white 20%);
}

.cert-item__badge {
  flex-shrink: 0;
  width: 80px;
  height: 80px;
}

.cert-item__badge-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.cert-item__badge-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-md);
  color: var(--color-accent);
  font-size: var(--text-xs);
  font-weight: 600;
}

.cert-item__badge-placeholder span {
  margin-top: var(--spacing-xs);
}

.cert-item__content {
  flex: 1;
}

.cert-item__name {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-primary);
  margin-bottom: var(--spacing-xs);
}

.cert-item__description {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  line-height: 1.5;
  margin-bottom: var(--spacing-sm);
}

.cert-item__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-md);
}

.cert-item__issuer,
.cert-item__valid {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: var(--text-xs);
  color: var(--color-text-light);
}

.cert-item__actions {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  align-self: center;
}

.cert-item__download,
.cert-item__verify {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: var(--text-xs);
  background: none;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  cursor: pointer;
  text-decoration: none;
  color: var(--color-text);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
  white-space: nowrap;
}

.cert-item__download:hover,
.cert-item__verify:hover {
  border-color: var(--color-highlight);
  color: var(--color-highlight);
}

/* Compact variant */
.cert-list--compact .cert-item {
  padding: var(--spacing-md);
}

.cert-list--compact .cert-item__badge {
  width: 50px;
  height: 50px;
}

.cert-list--compact .cert-item__name {
  font-size: var(--text-base);
}

.cert-list--compact .cert-item__actions {
  flex-direction: row;
}

/* Grid variant */
.cert-list--grid .cert-item {
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.cert-list--grid .cert-item__badge {
  width: 100px;
  height: 100px;
}

.cert-list--grid .cert-item__meta {
  justify-content: center;
}

.cert-list--grid .cert-item__actions {
  flex-direction: row;
  width: 100%;
  justify-content: center;
}
</style>
