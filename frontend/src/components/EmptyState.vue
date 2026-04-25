<template>
  <div class="empty-state">
    <div class="icon">{{ icon }}</div>
    <h3 class="title">{{ title }}</h3>
    <p v-if="description" class="description">{{ description }}</p>
    <slot name="actions" />
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  icon?: string
  title: string
  description?: string
}>(), {
  icon: '📦',
  description: undefined,
})
</script>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-12) var(--space-5);
  gap: var(--space-4);
  text-align: center;
  color: var(--text-secondary);
}

.icon {
  font-size: 56px;
  line-height: 1;
  opacity: 0.7;
}

.title {
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.description {
  color: var(--text-muted);
  max-width: 360px;
  margin: 0;
  font-size: var(--text-sm);
  line-height: 1.6;
}
</style>
