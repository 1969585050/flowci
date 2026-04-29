<template>
  <div class="titlebar" @dblclick="onTitleBarDblClick">
    <!-- 中间拖拽区 -->
    <div class="drag-zone" style="--wails-draggable: drag"></div>

    <!-- 右侧按钮组 -->
    <div class="title-actions" style="--wails-draggable: no-drag">
      <button class="ti-btn" :title="isDark ? '切换亮色主题' : '切换暗色主题'" @click="toggleTheme">
        <Sun v-if="isDark" :size="13" :stroke-width="1.75" />
        <Moon v-else :size="13" :stroke-width="1.75" />
      </button>
      <button
        class="ti-btn pin-btn"
        :class="{ active: alwaysOnTop }"
        :title="alwaysOnTop ? '取消窗口置顶' : '窗口始终置顶'"
        @click="onTogglePin"
      >
        <Pin :size="13" :stroke-width="1.75" />
      </button>
      <button class="ti-btn" title="最小化" @click="onMinimise">
        <Minus :size="13" :stroke-width="1.75" />
      </button>
      <button class="ti-btn" title="最大化 / 还原" @click="onToggleMax">
        <Minimize2 v-if="isMax" :size="12" :stroke-width="1.75" />
        <Maximize2 v-else :size="11" :stroke-width="1.75" />
      </button>
      <button class="ti-btn close-btn" title="关闭" @click="onClose">
        <X :size="13" :stroke-width="1.75" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, inject, onMounted, onUnmounted } from 'vue'
import { Pin, Minus, Maximize2, Minimize2, X, Sun, Moon } from 'lucide-vue-next'
import { useSettings } from '../composables/useSettings'
import {
  WindowMinimise, WindowToggleMaximise, WindowIsMaximised, QuitApp,
  SetWindowAlwaysOnTop, GetWindowAlwaysOnTop,
} from '../wailsjs/go/handler/App'

const toast = inject('toast') as { success: (m: string) => void; error: (m: string) => void; info?: (m: string) => void }

const { theme, isDark, setTheme } = useSettings()

function toggleTheme() {
  void setTheme(theme.value === 'dark' ? 'light' : 'dark')
}

const alwaysOnTop = ref(false)
const isMax = ref(false)

async function refreshState() {
  try {
    alwaysOnTop.value = await GetWindowAlwaysOnTop()
    isMax.value = await WindowIsMaximised()
  } catch { /* ignore */ }
}

async function onTogglePin() {
  const next = !alwaysOnTop.value
  try {
    await SetWindowAlwaysOnTop(next)
    alwaysOnTop.value = next
  } catch (e) {
    toast?.error(`切换失败: ${e instanceof Error ? e.message : String(e)}`)
  }
}

async function onMinimise() {
  try { await WindowMinimise() } catch (e) { console.error(e) }
}

async function onToggleMax() {
  try {
    await WindowToggleMaximise()
    setTimeout(refreshState, 50)
  } catch (e) {
    console.error(e)
  }
}

async function onClose() {
  try { await QuitApp() } catch (e) { console.error(e) }
}

async function onTitleBarDblClick(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (target.closest('.title-actions')) return
  await onToggleMax()
}

let pollHandle: number | undefined
onMounted(() => {
  void refreshState()
  pollHandle = window.setInterval(refreshState, 1500)
})
onUnmounted(() => {
  if (pollHandle) window.clearInterval(pollHandle)
})
</script>

<style scoped>
.titlebar {
  display: flex;
  align-items: stretch;
  height: 36px;
  background: var(--bg-sidebar);
  border-bottom: 1px solid var(--border-sidebar);
  color: var(--text-titlebar);
  user-select: none;
  flex-shrink: 0;
  font-size: var(--text-sm);
}

/* 中间拖拽 */
.drag-zone {
  flex: 1;
  -webkit-app-region: drag;
}

/* 右侧按钮 */
.title-actions {
  display: flex;
  align-items: stretch;
  -webkit-app-region: no-drag;
}
.ti-btn {
  width: 46px;
  background: transparent;
  border: none;
  color: var(--text-titlebar-icon);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background var(--transition-fast), color var(--transition-fast);
  -webkit-app-region: no-drag;
}
.ti-btn:hover {
  background: var(--bg-titlebar-hover);
  color: var(--text-titlebar);
}
.ti-btn.pin-btn.active {
  background: var(--brand-500);
  color: var(--text-on-brand);
}
.ti-btn.pin-btn.active:hover {
  background: var(--brand-600);
}
.ti-btn.close-btn:hover {
  background: #e81123;
  color: #ffffff;
}
</style>
