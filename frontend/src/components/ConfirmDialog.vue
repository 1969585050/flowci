<template>
  <transition name="dialog">
    <div v-if="pending" class="overlay" @click.self="answer(false)">
      <div class="dialog" :class="variantClass" role="dialog" :aria-labelledby="titleId">
        <h3 :id="titleId" class="title">{{ pending.options.title }}</h3>
        <p class="message">{{ pending.options.message }}</p>
        <div class="actions">
          <button class="btn btn-cancel" @click="answer(false)">
            {{ pending.options.cancelText ?? '取消' }}
          </button>
          <button class="btn btn-confirm" :class="variantClass" @click="answer(true)">
            {{ pending.options.confirmText ?? '确定' }}
          </button>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useConfirm } from '../composables/useConfirm'

const { current, _answer } = useConfirm()
const pending = current

const titleId = `confirm-title-${Math.random().toString(36).slice(2, 8)}`

const variantClass = computed(() => {
  return pending.value?.options.variant === 'danger' ? 'danger' : 'default'
})

function answer(ok: boolean) {
  _answer(ok)
}

function onKey(e: KeyboardEvent) {
  if (!pending.value) return
  if (e.key === 'Escape') answer(false)
  if (e.key === 'Enter') answer(true)
}

onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))
</script>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
  backdrop-filter: blur(4px);
}

.dialog {
  background: var(--card-bg);
  border-radius: var(--radius-xl);
  padding: var(--space-6);
  width: 420px;
  max-width: calc(100vw - var(--space-8));
  box-shadow: var(--shadow-lg);
  border-top: 4px solid var(--brand-start);
}
.dialog.danger {
  border-top-color: var(--danger-fg);
}

.title {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: var(--space-3);
}

.message {
  color: var(--text-secondary);
  font-size: var(--text-base);
  margin-bottom: var(--space-6);
  word-break: break-word;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
}

.btn {
  padding: var(--space-3) var(--space-5);
  border: none;
  border-radius: var(--radius-md);
  font-size: var(--text-base);
  font-weight: 500;
  cursor: pointer;
  transition: transform var(--transition-fast), box-shadow var(--transition-fast), background var(--transition-fast);
}

.btn-cancel {
  background: var(--bg-primary);
  color: var(--text-primary);
}
.btn-cancel:hover {
  background: var(--border-color);
}

.btn-confirm.default {
  background: linear-gradient(135deg, var(--brand-start), var(--brand-end));
  color: #fff;
}
.btn-confirm.default:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px var(--brand-soft);
}

.btn-confirm.danger {
  background: var(--danger-fg);
  color: #fff;
}
.btn-confirm.danger:hover {
  background: #7f1d1d;
}

.dialog-enter-active,
.dialog-leave-active {
  transition: opacity var(--transition-base);
}
.dialog-enter-active .dialog,
.dialog-leave-active .dialog {
  transition: transform var(--transition-base), opacity var(--transition-base);
}
.dialog-enter-from,
.dialog-leave-to {
  opacity: 0;
}
.dialog-enter-from .dialog,
.dialog-leave-to .dialog {
  opacity: 0;
  transform: scale(0.96) translateY(-8px);
}
</style>
