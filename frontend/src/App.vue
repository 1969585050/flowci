<template>
  <div id="app-shell">
    <!-- 顶部自绘标题栏（Frameless 模式必需） -->
    <TitleBar />

    <div class="app-body">
      <aside class="sidebar">
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
    </div>

    <ToastHost />
    <ConfirmDialog />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, provide } from 'vue'
import TitleBar from './components/TitleBar.vue'
import ToastHost from './components/ToastHost.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'
import { useSettings } from './composables/useSettings'
import { useToast } from './composables/useToast'

// 兼容旧 view 的 inject('toast') / inject('theme')
const toast = useToast()
provide('toast', {
  success: toast.success,
  error: toast.error,
  info: toast.info,
  warning: toast.warning,
})

const navItems = [
  { to: '/dashboard', icon: '🏠', label: '仪表盘' },
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

provide('theme', {
  current: theme,
  isDark,
  setTheme,
})

const themeIcon = computed(() => {
  if (theme.value === 'system') return '🖥'
  return isDark.value ? '🌙' : '☀️'
})
const themeTooltip = computed(() => `主题：${theme.value}（点击切换）`)

function toggleTheme() {
  const next = theme.value === 'dark' ? 'light' : theme.value === 'light' ? 'system' : 'dark'
  void setTheme(next)
}

onMounted(() => {
  void load()
  watchSystemTheme()
})
</script>

<style scoped>
#app-shell {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.app-body {
  display: flex;
  flex: 1;
  min-height: 0;
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
  transition: background var(--transition-fast);
}
.footer-btn:hover {
  background: rgba(255, 255, 255, 0.12);
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
