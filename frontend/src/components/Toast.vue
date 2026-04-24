<template>
  <div class="toast-container">
    <transition-group name="toast">
      <div v-for="toast in toasts" :key="toast.id" class="toast-item" :class="toast.type">
        <span class="toast-icon">{{ iconMap[toast.type] }}</span>
        <span class="toast-message">{{ toast.message }}</span>
        <button class="toast-close" @click="removeToast(toast.id)">&times;</button>
      </div>
    </transition-group>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface Toast {
  id: number
  type: 'success' | 'error' | 'info'
  message: string
}

const iconMap: Record<string, string> = {
  success: '✅',
  error: '❌',
  info: 'ℹ️'
}

const toasts = ref<Toast[]>([])
let nextId = 0

function addToast(type: 'success' | 'error' | 'info', message: string, duration = 3000) {
  const id = nextId++
  toasts.value.push({ id, type, message })
  setTimeout(() => removeToast(id), duration)
}

function removeToast(id: number) {
  toasts.value = toasts.value.filter(t => t.id !== id)
}

defineExpose({ addToast })
</script>

<style scoped>
.toast-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 8px;
  pointer-events: none;
}

.toast-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 20px;
  border-radius: 10px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  min-width: 300px;
  max-width: 450px;
  pointer-events: auto;
  font-size: 14px;
}

.toast-item.success {
  background: #f0fdf4;
  color: #166534;
  border: 1px solid #bbf7d0;
}

.toast-item.error {
  background: #fef2f2;
  color: #991b1b;
  border: 1px solid #fecaca;
}

.toast-item.info {
  background: #eff6ff;
  color: #1e40af;
  border: 1px solid #bfdbfe;
}

.toast-icon {
  font-size: 18px;
  flex-shrink: 0;
}

.toast-message {
  flex: 1;
}

.toast-close {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  opacity: 0.5;
  padding: 0 4px;
  color: inherit;
}

.toast-close:hover {
  opacity: 1;
}

.toast-enter-active {
  transition: all 0.3s ease;
}

.toast-leave-active {
  transition: all 0.2s ease;
}

.toast-enter-from {
  opacity: 0;
  transform: translateX(100%);
}

.toast-leave-to {
  opacity: 0;
  transform: translateX(100%);
}
</style>
