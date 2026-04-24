<template>
  <div class="toast-container" aria-live="polite" aria-atomic="true">
    <transition-group name="toast" tag="div" class="toast-stack">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="toast-item"
        :class="t.type"
        role="status"
      >
        <span class="toast-icon">{{ iconFor(t.type) }}</span>
        <span class="toast-message">{{ t.message }}</span>
        <button class="toast-close" :aria-label="'关闭'" @click="dismiss(t.id)">
          &times;
        </button>
      </div>
    </transition-group>
  </div>
</template>

<script setup lang="ts">
import { useToast } from '../composables/useToast'
import type { ToastItem } from '../types'

const { toasts, dismiss } = useToast()

function iconFor(type: ToastItem['type']): string {
  switch (type) {
    case 'success': return '✓'
    case 'error':   return '✕'
    case 'warning': return '⚠'
    case 'info':    return 'ℹ'
  }
}
</script>

<style scoped>
.toast-container {
  position: fixed;
  top: var(--space-5);
  right: var(--space-5);
  z-index: 9999;
  pointer-events: none;
}

.toast-stack {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.toast-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-md);
  min-width: 280px;
  max-width: 440px;
  font-size: var(--text-base);
  pointer-events: auto;
  border-left: 3px solid transparent;
}

.toast-item.success {
  background: var(--success-bg);
  color: var(--success-fg);
  border-left-color: var(--success-fg);
}
.toast-item.error {
  background: var(--danger-bg);
  color: var(--danger-fg);
  border-left-color: var(--danger-fg);
}
.toast-item.warning {
  background: var(--warning-bg);
  color: var(--warning-fg);
  border-left-color: var(--warning-fg);
}
.toast-item.info {
  background: var(--info-bg);
  color: var(--info-fg);
  border-left-color: var(--info-fg);
}

.toast-icon {
  flex-shrink: 0;
  font-weight: bold;
  font-size: var(--text-lg);
  line-height: 1;
}

.toast-message {
  flex: 1;
  word-break: break-word;
}

.toast-close {
  background: none;
  border: none;
  font-size: var(--text-xl);
  color: inherit;
  opacity: 0.5;
  cursor: pointer;
  line-height: 1;
  padding: 0 var(--space-1);
}
.toast-close:hover {
  opacity: 1;
}

.toast-enter-active,
.toast-leave-active {
  transition: all var(--transition-slow);
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(120%);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(120%);
}
.toast-leave-active {
  position: absolute;
  right: 0;
}
</style>
