<template>
  <div class="loader" :class="size">
    <div class="spinner" :style="{ borderTopColor: color }"></div>
    <span v-if="label" class="label">{{ label }}</span>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  size?: 'sm' | 'md' | 'lg'
  label?: string
  color?: string
}>(), {
  size: 'md',
  color: 'var(--brand-start)',
})
</script>

<style scoped>
.loader {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  padding: var(--space-8) var(--space-4);
}

.spinner {
  border: 3px solid var(--border-color);
  border-top-color: var(--brand-start);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.sm .spinner { width: 18px; height: 18px; border-width: 2px; }
.md .spinner { width: 32px; height: 32px; border-width: 3px; }
.lg .spinner { width: 48px; height: 48px; border-width: 4px; }

.label {
  color: var(--text-secondary);
  font-size: var(--text-sm);
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
