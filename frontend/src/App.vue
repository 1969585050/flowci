<template>
  <div id="app-shell">
    <div class="titlebar">
      <div class="drag-zone" style="--wails-draggable: drag"></div>
      <div class="window-controls" style="--wails-draggable: no-drag">
        <button
          class="wc-btn"
          :class="{ 'wc-pin-active': alwaysOnTop }"
          :title="alwaysOnTop ? '取消置顶' : '置顶'"
          @click="onTogglePin"
        >
          <Pin :size="13" :stroke-width="1.75" />
        </button>
        <button class="wc-btn" title="最小化" @click="onMinimise">
          <Minus :size="13" :stroke-width="1.75" />
        </button>
        <button class="wc-btn" title="最大化 / 还原" @click="onToggleMax">
          <Minimize2 v-if="isMax" :size="12" :stroke-width="1.75" />
          <Maximize2 v-else :size="11" :stroke-width="1.75" />
        </button>
        <button class="wc-btn wc-close" title="关闭" @click="onClose">
          <X :size="13" :stroke-width="1.75" />
        </button>
      </div>
    </div>
    <div class="app-body">
      <aside class="sidebar">
        <nav class="nav">
          <router-link
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="nav-item"
            active-class="active"
            :title="item.label"
          >
            <component :is="item.icon" class="nav-icon" :size="24" :stroke-width="1.75" />
          </router-link>
        </nav>
        <div class="sidebar-actions">
          <button class="sa-btn" :title="isDark ? '亮色' : '暗色'" @click="toggleTheme">
            <Sun v-if="isDark" :size="14" :stroke-width="1.75" />
            <Moon v-else :size="14" :stroke-width="1.75" />
          </button>
        </div>
      </aside>

      <main class="content">
        <router-view v-slot="{ Component }">
          <transition name="route" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>

    <ToastHost />
    <ConfirmDialog />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, provide } from 'vue'
import {
  LayoutDashboard, Package, GitBranch, Hammer, Workflow,
  Rocket, Layers, Upload, Settings,
  Sun, Moon, Pin, Minus, Maximize2, Minimize2, X,
} from 'lucide-vue-next'
import ToastHost from './components/ToastHost.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'
import { useSettings } from './composables/useSettings'
import { useToast } from './composables/useToast'
import {
  WindowMinimise, WindowToggleMaximise, WindowIsMaximised, QuitApp,
  SetWindowAlwaysOnTop, GetWindowAlwaysOnTop,
} from './wailsjs/go/handler/App'

const toast = useToast()
provide('toast', {
  success: toast.success,
  error: toast.error,
  info: toast.info,
  warning: toast.warning,
})

const navItems = [
  { to: '/dashboard',    icon: LayoutDashboard, label: '仪表盘' },
  { to: '/projects',     icon: Package,         label: '项目' },
  { to: '/repositories', icon: GitBranch,       label: '仓库源' },
  { to: '/build',        icon: Hammer,          label: '构建' },
  { to: '/pipelines',    icon: Workflow,        label: '流水线' },
  { to: '/deploy',       icon: Rocket,          label: '部署' },
  { to: '/images',       icon: Layers,          label: '镜像' },
  { to: '/push',         icon: Upload,          label: '推送' },
  { to: '/settings',     icon: Settings,        label: '设置' },
] as const

const { theme, isDark, load, setTheme } = useSettings()

provide('theme', {
  current: theme,
  isDark,
  setTheme,
})

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
  } catch (e) { console.error(e) }
}

async function onClose() {
  try { await QuitApp() } catch (e) { console.error(e) }
}

let pollHandle: number | undefined
onMounted(() => {
  void load()
  void refreshState()
  pollHandle = window.setInterval(refreshState, 1500)
})
onUnmounted(() => {
  if (pollHandle) window.clearInterval(pollHandle)
})
</script>

<style scoped>
#app-shell {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
  position: relative;
}

.titlebar {
  display: flex;
  align-items: stretch;
  height: 28px;
  background: var(--bg-sidebar);
  flex-shrink: 0;
}

.drag-zone {
  flex: 1;
  -webkit-app-region: drag;
}

.app-body {
  display: flex;
  flex: 1;
  min-height: 0;
}

.sidebar {
  width: var(--sidebar-collapsed-width);
  flex-shrink: 0;
  background: var(--bg-sidebar);
  display: flex;
  flex-direction: column;
  padding: var(--space-5) 0;
  overflow: hidden;
}

.nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.nav-item {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0;
  padding: var(--space-3) var(--space-1);
  margin: 0 var(--space-1);
  border-radius: var(--radius-md);
  color: var(--text-nav);
  text-decoration: none;
  transition:
    background var(--transition-fast),
    color var(--transition-fast);
  cursor: pointer;
}
.nav-item:hover {
  background: var(--bg-titlebar-hover);
  color: var(--text-nav-hover);
}
.nav-item.active {
  background: var(--bg-nav-active);
  color: var(--text-nav-active);
  font-weight: var(--weight-medium);
}
.nav-icon {
  flex-shrink: 0;
}

.sidebar-actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-1);
  padding: var(--space-3) 0;
}

.sa-btn {
  width: 28px;
  height: 28px;
  border-radius: var(--radius-md);
  background: transparent;
  border: none;
  color: var(--text-titlebar-icon);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background var(--transition-fast), color var(--transition-fast);
}
.sa-btn:hover {
  background: var(--bg-titlebar-hover);
  color: var(--text-titlebar);
}

.window-controls {
  display: flex;
  -webkit-app-region: no-drag;
}
.wc-btn {
  width: 36px;
  height: 28px;
  background: transparent;
  border: none;
  color: #6b7280;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background var(--transition-fast), color var(--transition-fast);
}
.wc-btn:hover {
  background: rgba(255,255,255,0.08);
  color: #e6e8ef;
}
.wc-close:hover {
  background: #e81123;
  color: #ffffff;
}
.wc-pin-active {
  background: var(--brand-500);
  color: var(--text-on-brand);
}

.content {
  flex: 1;
  overflow-y: auto;
  scrollbar-gutter: stable;
  padding: var(--space-6);
}

.route-enter-active,
.route-leave-active {
  transition: opacity var(--transition-base), transform var(--transition-base);
}
.route-enter-from {
  opacity: 0;
  transform: translateY(6px);
}
.route-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
