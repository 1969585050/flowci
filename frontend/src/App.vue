<template>
  <div id="app-shell">
    <aside class="sidebar">
      <router-link to="/dashboard" class="logo" title="返回仪表盘">
        <span class="logo-icon">🚀</span>
        <span class="logo-text">FlowCI</span>
      </router-link>
      <nav class="nav">
        <router-link
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="nav-item"
          active-class="active"
        >
          <span class="nav-icon">{{ item.icon }}</span>
          <span>{{ item.label }}</span>
        </router-link>
      </nav>
      <div class="sidebar-footer">
        <button
          class="footer-btn"
          :class="{ active: alwaysOnTop }"
          :title="alwaysOnTop ? '取消窗口置顶' : '窗口始终置顶'"
          @click="toggleAlwaysOnTop"
        >📌</button>
        <button class="footer-btn" :title="themeTooltip" @click="toggleTheme">
          {{ themeIcon }}
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

    <ToastHost />
    <ConfirmDialog />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, provide, ref } from 'vue'
import ToastHost from './components/ToastHost.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'
import { useSettings } from './composables/useSettings'
import { useToast } from './composables/useToast'
import { SetWindowAlwaysOnTop, GetWindowAlwaysOnTop } from './wailsjs/go/handler/App'

// 为已存在的 view（仍然用 inject('toast') / inject('theme')）提供兼容适配。
// 新代码应直接 import { useToast } / { useSettings } 而非 inject。
const toast = useToast()
provide('toast', {
  success: toast.success,
  error: toast.error,
  info: toast.info,
  warning: toast.warning,
})

const navItems = [
  { to: '/projects', icon: '📦', label: '项目' },
  { to: '/repositories', icon: '🌿', label: '仓库源' },
  { to: '/build', icon: '🔨', label: '构建' },
  { to: '/pipelines', icon: '🔧', label: '流水线' },
  { to: '/deploy', icon: '🌐', label: '部署' },
  { to: '/images', icon: '🗃️', label: '镜像' },
  { to: '/push', icon: '📤', label: '推送' },
  { to: '/settings', icon: '⚙️', label: '设置' },
] as const

const { theme, isDark, load, setTheme, watchSystemTheme } = useSettings()

// 兼容旧 inject('theme') 接口（SettingsView 等用到）
provide('theme', {
  current: theme,
  isDark,
  setTheme,
})

const themeIcon = computed(() => {
  if (theme.value === 'system') return '🖥'
  return isDark.value ? '🌙' : '☀️'
})

const themeTooltip = computed(() => {
  return `主题：${theme.value}（点击切换）`
})

function toggleTheme() {
  const next = theme.value === 'dark' ? 'light' : theme.value === 'light' ? 'system' : 'dark'
  void setTheme(next)
}

// 窗口始终置顶切换
const alwaysOnTop = ref(false)
async function loadAlwaysOnTop() {
  try { alwaysOnTop.value = await GetWindowAlwaysOnTop() } catch { /* ignore */ }
}
async function toggleAlwaysOnTop() {
  const next = !alwaysOnTop.value
  try {
    await SetWindowAlwaysOnTop(next)
    alwaysOnTop.value = next
    toast.info(next ? '窗口已置顶' : '已取消窗口置顶')
  } catch (e) {
    toast.error(`切换失败: ${e instanceof Error ? e.message : String(e)}`)
  }
}

onMounted(() => {
  void load()
  watchSystemTheme()
  void loadAlwaysOnTop()
})
</script>

<style scoped>
#app-shell {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

.sidebar {
  width: 220px;
  flex-shrink: 0;
  background: var(--bg-sidebar);
  color: #fff;
  display: flex;
  flex-direction: column;
  padding: var(--space-5) 0;
}

.logo {
  padding: 0 var(--space-5) var(--space-6);
  display: flex;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--text-2xl);
  font-weight: bold;
  text-decoration: none;
  cursor: pointer;
  transition: opacity 0.15s;
}
.logo:hover { opacity: 0.85; }

.logo-icon {
  font-size: 30px;
}

.logo-text {
  background: linear-gradient(90deg, var(--brand-start) 0%, var(--brand-end) 100%);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
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
  gap: var(--space-3);
  padding: var(--space-3) var(--space-5);
  margin: 0 var(--space-2);
  border-radius: var(--radius-md);
  color: var(--text-nav);
  text-decoration: none;
  transition: background var(--transition-fast), color var(--transition-fast);
  cursor: pointer;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-nav-hover);
}

.nav-item.active {
  background: var(--bg-nav-active);
  color: var(--text-nav-active);
  font-weight: 500;
}

.nav-icon {
  font-size: var(--text-xl);
  width: 24px;
  text-align: center;
}

.sidebar-footer {
  padding: var(--space-3) var(--space-5);
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
}

.footer-btn {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.06);
  border: none;
  color: #fff;
  font-size: var(--text-lg);
  cursor: pointer;
  transition: background var(--transition-fast), transform var(--transition-fast);
}
.footer-btn:hover {
  background: rgba(255, 255, 255, 0.12);
}
.footer-btn.active {
  background: linear-gradient(135deg, var(--brand-start), var(--brand-end));
  box-shadow: 0 0 12px var(--brand-soft);
}

.content {
  flex: 1;
  overflow-y: auto;
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
