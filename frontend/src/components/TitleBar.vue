<template>
  <div class="titlebar" @dblclick="onTitleBarDblClick">
    <!-- 左侧 logo / 标题（点击回 dashboard） -->
    <router-link to="/dashboard" class="title-brand" @click.stop>
      <span class="brand-icon">🚀</span>
      <span class="brand-name">FlowCI</span>
    </router-link>

    <!-- 中间拖拽区（占位，wails CSS API 让此区可拖窗口） -->
    <div class="drag-zone" style="--wails-draggable: drag"></div>

    <!-- 右侧按钮组 -->
    <div class="title-actions" style="--wails-draggable: no-drag">
      <button
        class="ti-btn pin-btn"
        :class="{ active: alwaysOnTop }"
        :title="alwaysOnTop ? '取消窗口置顶' : '窗口始终置顶'"
        @click="onTogglePin"
      >📌</button>
      <button class="ti-btn" title="最小化" @click="onMinimise">
        <svg width="10" height="10" viewBox="0 0 10 10"><path d="M0 5h10" stroke="currentColor" stroke-width="1.2"/></svg>
      </button>
      <button class="ti-btn" title="最大化 / 还原" @click="onToggleMax">
        <!-- 还原 vs 最大化 SVG -->
        <svg v-if="isMax" width="10" height="10" viewBox="0 0 10 10">
          <path d="M2 0v2H0v8h8V8h2V0H2zm5 9H1V3h6v6zm2-2H8V2H3V1h6v6z" fill="currentColor"/>
        </svg>
        <svg v-else width="10" height="10" viewBox="0 0 10 10">
          <rect x="0.5" y="0.5" width="9" height="9" fill="none" stroke="currentColor"/>
        </svg>
      </button>
      <button class="ti-btn close-btn" title="关闭" @click="onClose">
        <svg width="10" height="10" viewBox="0 0 10 10"><path d="M1 1l8 8M9 1l-8 8" stroke="currentColor" stroke-width="1.2"/></svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, inject, onMounted, onUnmounted } from 'vue'
import {
  WindowMinimise, WindowToggleMaximise, WindowIsMaximised, QuitApp,
  SetWindowAlwaysOnTop, GetWindowAlwaysOnTop,
} from '../wailsjs/go/handler/App'

const toast = inject('toast') as { success: (m: string) => void; error: (m: string) => void; info?: (m: string) => void }

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
    toast?.info?.(next ? '窗口已置顶' : '已取消窗口置顶')
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
    // 给 webview 一点时间反映新状态
    setTimeout(refreshState, 50)
  } catch (e) {
    console.error(e)
  }
}

async function onClose() {
  try { await QuitApp() } catch (e) { console.error(e) }
}

async function onTitleBarDblClick(e: MouseEvent) {
  // 点到按钮区不切换最大化
  const target = e.target as HTMLElement
  if (target.closest('.title-actions, .title-brand')) return
  await onToggleMax()
}

// 监听窗口最大化状态变化（用户拖窗口 / 系统手势）
let pollHandle: number | undefined
onMounted(() => {
  void refreshState()
  // wails runtime 没有 onResize 事件直接订阅；轮询低频检查
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
  color: #fff;
  user-select: none;
  flex-shrink: 0;
  font-size: 12px;
}

/* 左侧 logo */
.title-brand {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 14px;
  text-decoration: none;
  color: #fff;
  cursor: pointer;
  transition: background 0.12s;
}
.title-brand:hover { background: rgba(255, 255, 255, 0.06); }
.brand-icon { font-size: 14px; }
.brand-name {
  font-weight: 600;
  background: linear-gradient(90deg, var(--brand-start), var(--brand-end));
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}

/* 中间拖拽 */
.drag-zone {
  flex: 1;
  -webkit-app-region: drag;  /* 双保险：另一种 wails 拖拽语法 */
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
  color: #cbd5e1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  transition: background 0.12s, color 0.12s;
  -webkit-app-region: no-drag;
}
.ti-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
}
.ti-btn.pin-btn.active {
  background: linear-gradient(135deg, var(--brand-start), var(--brand-end));
  color: #fff;
  box-shadow: inset 0 -2px 0 rgba(255, 255, 255, 0.4);
}
.ti-btn.close-btn:hover {
  background: #e81123;  /* Windows 标准关闭红 */
  color: #fff;
}
</style>
