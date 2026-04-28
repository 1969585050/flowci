<template>
  <div id="app-shell">
    <!-- 顶部自绘标题栏（Frameless 模式必需） -->
    <TitleBar />

    <div class="app-body">
      <aside class="sidebar" :class="{ collapsed: sidebarCollapsed }">
        <nav class="nav">
          <router-link
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="nav-item"
            active-class="active"
          >
            <component :is="item.icon" class="nav-icon" :size="18" :stroke-width="1.75" />
            <span>{{ item.label }}</span>
          </router-link>
        </nav>
        <div class="sidebar-footer">
          <button class="footer-btn collapse-btn" :title="sidebarCollapsed ? '展开侧边栏' : '折叠侧边栏'" @click="toggleSidebar">
            <ChevronRight v-if="sidebarCollapsed" :size="16" :stroke-width="1.75" />
            <ChevronLeft v-else :size="16" :stroke-width="1.75" />
          </button>
          <button class="footer-btn" :title="themeTooltip" @click="toggleTheme">
            <component :is="themeIcon" :size="16" :stroke-width="1.75" />
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
import {
  LayoutDashboard, Package, GitBranch, Hammer, Workflow,
  Rocket, Layers, Upload, Settings,
  Sun, Moon, Monitor,
  ChevronLeft, ChevronRight,
} from 'lucide-vue-next'
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

const { theme, isDark, load, setTheme, watchSystemTheme, sidebarCollapsed, toggleSidebar } = useSettings()

provide('theme', {
  current: theme,
  isDark,
  setTheme,
})

const themeIcon = computed(() => {
  if (theme.value === 'system') return Monitor
  return isDark.value ? Moon : Sun
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
  width: var(--sidebar-width);
  transition: width var(--transition-base);
  flex-shrink: 0;
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border-sidebar);
  display: flex;
  flex-direction: column;
  padding: var(--space-5) 0;
}

.sidebar.collapsed {
  width: var(--sidebar-collapsed-width);
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
  padding: var(--space-2) var(--space-4);
  margin: 0 var(--space-2);
  border-radius: var(--radius-md);
  color: var(--text-nav);
  text-decoration: none;
  font-size: var(--text-base);
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
.nav-item span {
  transition: opacity var(--transition-fast);
}

.nav-icon {
  flex-shrink: 0;
}

.sidebar.collapsed .nav-item {
  justify-content: center;
  padding: var(--space-2) var(--space-1);
  margin: 0 var(--space-1);
  gap: 0;
}
.sidebar.collapsed .nav-item span {
  opacity: 0;
  width: 0;
  overflow: hidden;
  pointer-events: none;
}

.sidebar-footer {
  padding: var(--space-3) var(--space-4);
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
}

.sidebar.collapsed .sidebar-footer {
  flex-direction: column;
  align-items: center;
}
.footer-btn {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md);
  background: transparent;
  border: 1px solid var(--border-sidebar);
  color: var(--text-nav);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition:
    background var(--transition-fast),
    color var(--transition-fast),
    border-color var(--transition-fast);
}
.footer-btn:hover {
  background: var(--bg-titlebar-hover);
  color: var(--text-nav-hover);
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
