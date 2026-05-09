<template>
  <nav class="breadcrumb" :aria-label="$t('common.a11y.breadcrumb')">
    <ol class="breadcrumb__list" itemscope itemtype="https://schema.org/BreadcrumbList">
      <!-- Home -->
      <li
        class="breadcrumb__item"
        itemprop="itemListElement"
        itemscope
        itemtype="https://schema.org/ListItem"
      >
        <NuxtLink
          :to="localePath('/')"
          class="breadcrumb__link breadcrumb__link--home"
          itemprop="item"
        >
          <Icon name="lucide:home" size="16" />
          <span itemprop="name">{{ $t('nav.home') }}</span>
        </NuxtLink>
        <meta itemprop="position" :content="1" />
      </li>

      <!-- Breadcrumb items -->
      <li
        v-for="(item, index) in items"
        :key="index"
        class="breadcrumb__item"
        itemprop="itemListElement"
        itemscope
        itemtype="https://schema.org/ListItem"
      >
        <span class="breadcrumb__separator" aria-hidden="true">
          <Icon name="lucide:chevron-right" size="14" />
        </span>

        <NuxtLink
          v-if="item.to && !isLast(index)"
          :to="localePath(item.to)"
          class="breadcrumb__link"
          itemprop="item"
        >
          <span itemprop="name">{{ item.label }}</span>
        </NuxtLink>

        <span v-else class="breadcrumb__current" itemprop="name">
          {{ item.label }}
        </span>

        <meta itemprop="position" :content="String(index + 2)" />
      </li>
    </ol>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useLocalePath } from '#i18n'

interface BreadcrumbItem {
  label: string
  to?: string
}

interface Props {
  items?: BreadcrumbItem[]
}

const props = withDefaults(defineProps<Props>(), {
  items: () => []
})

const localePath = useLocalePath()

const isLast = (index: number) => {
  return index === props.items.length - 1
}
</script>

<style scoped>
.breadcrumb {
  padding: var(--spacing-md) 0;
}

.breadcrumb__list {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  list-style: none;
  margin: 0;
  padding: 0;
}

.breadcrumb__item {
  display: flex;
  align-items: center;
}

.breadcrumb__link {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: var(--text-sm);
  color: var(--color-text-light);
  transition: color var(--transition-fast);
}

.breadcrumb__link:hover {
  color: var(--color-highlight);
}

.breadcrumb__link--home {
  gap: var(--spacing-xs);
}

.breadcrumb__separator {
  display: flex;
  align-items: center;
  margin: 0 var(--spacing-sm);
  color: var(--color-border);
}

.breadcrumb__current {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
}

/* Rich snippets for SEO */
.breadcrumb__link[itemprop="item"],
.breadcrumb__current[itemprop="name"] {
  display: inline-flex;
}

@media (max-width: 640px) {
  .breadcrumb__list {
    font-size: var(--text-xs);
  }

  .breadcrumb__link {
    font-size: var(--text-xs);
  }

  .breadcrumb__current {
    font-size: var(--text-xs);
  }

  .breadcrumb__separator {
    margin: 0 var(--spacing-xs);
  }
}
</style>
