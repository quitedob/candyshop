<template>
  <div class="quality-timeline" :class="`quality-timeline--${variant}`">
    <div
      v-for="(item, index) in items"
      :key="item.year"
      class="timeline-item"
      :class="{ 'timeline-item--featured': item.featured }"
    >
      <!-- Year marker -->
      <div class="timeline-item__year">
        {{ item.year }}
      </div>

      <!-- Content -->
      <div class="timeline-item__content">
        <h3 class="timeline-item__title">{{ item.milestone }}</h3>
        <p v-if="item.description" class="timeline-item__description">
          {{ item.description }}
        </p>

        <!-- Image if available -->
        <div v-if="item.image" class="timeline-item__image">
          <img :src="item.image" :alt="item.milestone" />
        </div>

        <!-- Tags -->
        <div v-if="item.tags" class="timeline-item__tags">
          <span
            v-for="tag in item.tags"
            :key="tag"
            class="timeline-item__tag"
          >
            {{ tag }}
          </span>
        </div>
      </div>

      <!-- Icon (optional) -->
      <div v-if="item.icon" class="timeline-item__icon">
        <Icon :name="item.icon" size="20" />
      </div>

      <!-- Connector (not for last item) -->
      <div v-if="index < items.length - 1" class="timeline-item__connector"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface TimelineItem {
  year: string | number
  milestone: string
  description?: string
  icon?: string
  image?: string
  tags?: string[]
  featured?: boolean
}

interface Props {
  items: TimelineItem[]
  variant?: 'default' | 'compact' | 'horizontal'
}

defineProps<Props>()
</script>

<style scoped>
.quality-timeline {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

/* Default vertical timeline */
.timeline-item {
  position: relative;
  display: grid;
  grid-template-columns: 100px 1fr;
  gap: var(--spacing-md);
}

.timeline-item__year {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 50px;
  font-size: var(--text-xl);
  font-weight: 700;
  color: white;
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-accent) 100%);
  border-radius: var(--radius-lg);
}

.timeline-item--featured .timeline-item__year {
  background: linear-gradient(135deg, var(--color-highlight) 0%, var(--color-accent) 100%);
  box-shadow: var(--shadow-md);
}

.timeline-item__content {
  padding: var(--spacing-lg);
  background-color: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.timeline-item--featured .timeline-item__content {
  border-color: var(--color-highlight);
  box-shadow: var(--shadow-md);
}

.timeline-item__title {
  font-size: var(--text-lg);
  font-weight: 600;
  margin-bottom: var(--spacing-sm);
}

.timeline-item__description {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  line-height: 1.6;
  margin-bottom: var(--spacing-md);
}

.timeline-item__image {
  margin-top: var(--spacing-md);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.timeline-item__image img {
  width: 100%;
  height: auto;
}

.timeline-item__tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-xs);
  margin-top: var(--spacing-md);
}

.timeline-item__tag {
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: var(--text-xs);
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-sm);
  color: var(--color-text-light);
}

.timeline-item__icon {
  position: absolute;
  right: calc(100% + var(--spacing-md));
  top: var(--spacing-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background-color: var(--color-highlight);
  color: white;
  border-radius: var(--radius-full);
}

.timeline-item__connector {
  position: absolute;
  left: 50px;
  top: 50px;
  bottom: calc(0px - var(--spacing-xl) + 50px);
  width: 2px;
  background: linear-gradient(to bottom, var(--color-accent) 0%, var(--color-border) 100%);
}

/* Compact variant */
.quality-timeline--compact .timeline-item {
  grid-template-columns: 70px 1fr;
  gap: var(--spacing-sm);
}

.quality-timeline--compact .timeline-item__year {
  height: 40px;
  font-size: var(--text-base);
}

.quality-timeline--compact .timeline-item__content {
  padding: var(--spacing-md);
}

.quality-timeline--compact .timeline-item__connector {
  left: 35px;
}

/* Horizontal variant */
.quality-timeline--horizontal {
  flex-direction: row;
  overflow-x: auto;
  padding-bottom: var(--spacing-xl);
  gap: var(--spacing-xl);
}

.quality-timeline--horizontal .timeline-item {
  grid-template-columns: 1fr;
  grid-template-rows: auto 1fr;
  min-width: 250px;
}

.quality-timeline--horizontal .timeline-item__year {
  width: 80px;
  height: 40px;
  margin: 0 auto var(--spacing-md);
}

.quality-timeline--horizontal .timeline-item__connector {
  left: auto;
  right: calc(0px - var(--spacing-xl) / 2);
  top: 20px;
  bottom: auto;
  width: calc(var(--spacing-xl) - 2px);
  height: 2px;
  background: linear-gradient(to right, var(--color-accent) 0%, var(--color-border) 100%);
}

.quality-timeline--horizontal .timeline-item__icon {
  position: static;
  margin-bottom: var(--spacing-sm);
}

/* Alternating variant (can be added via extending) */
@media (min-width: 768px) {
  .quality-timeline--alternating .timeline-item:nth-child(even) {
    grid-template-columns: 1fr 100px;
  }

  .quality-timeline--alternating .timeline-item:nth-child(even) .timeline-item__year {
    order: 2;
  }

  .quality-timeline--alternating .timeline-item:nth-child(even) .timeline-item__content {
    order: 1;
  }

  .quality-timeline--alternating .timeline-item:nth-child(even) .timeline-item__connector {
    left: auto;
    right: 50px;
  }
}
</style>
