<template>
  <div id="app-shell">
    <!-- Titlebar -->
    <div class="titlebar" style="--wails-draggable:drag">
      <div class="titlebar-logo">
        <div class="brand-dot" />
        <span>FlowCI</span>
      </div>
      <div class="titlebar-spacer" />
      <div class="titlebar-actions" style="--wails-draggable:no-drag">
        <button class="wc-btn" :class="{ 'wc-pin-active': alwaysOnTop }" :title="alwaysOnTop ? '取消置顶' : '置顶'" @click="onTogglePin"><Pin :size="12" :stroke-width="1.75" /></button>
        <button class="wc-btn" title="主题" @click="toggleTheme"><Sun v-if="isDark" :size="12" :stroke-width="1.75" /><Moon v-else :size="12" :stroke-width="1.75" /></button>
        <button class="wc-btn" title="最小化" @click="onMinimise"><Minus :size="12" :stroke-width="1.75" /></button>
        <button class="wc-btn" title="最大化" @click="onToggleMax"><Minimize2 v-if="isMax" :size="11" :stroke-width="1.75" /><Maximize2 v-else :size="11" :stroke-width="1.75" /></button>
        <button class="wc-btn wc-close" title="关闭" @click="onClose"><X :size="12" :stroke-width="1.75" /></button>
      </div>
    </div>

    <!-- Body -->
    <div class="app-body">
      <!-- Sidebar -->
      <aside class="sidebar" :class="{ collapsed: sidebarCollapsed }">
        <div class="sb-header">
          <h4>FlowCI</h4>
          <button class="sb-toggle" @click="sidebarCollapsed = !sidebarCollapsed" :title="sidebarCollapsed ? '展开' : '折叠'">
            <ChevronLeft v-if="!sidebarCollapsed" :size="12" :stroke-width="1.75" />
            <ChevronRight v-else :size="12" :stroke-width="1.75" />
          </button>
        </div>
        <div class="sb-body">
          <div class="sb-section">
            <router-link to="/dashboard" class="sb-item" title="仪表盘"><LayoutDashboard :size="15" :stroke-width="1.5" /><span class="sb-label">仪表盘</span></router-link>
            <router-link to="/projects" class="sb-item" title="项目"><Package :size="15" :stroke-width="1.5" /><span class="sb-label">项目</span></router-link>
            <router-link to="/repositories" class="sb-item" title="仓库源"><GitBranch :size="15" :stroke-width="1.5" /><span class="sb-label">仓库源</span></router-link>
            <router-link to="/build" class="sb-item" title="构建"><Hammer :size="15" :stroke-width="1.5" /><span class="sb-label">构建</span></router-link>
            <router-link to="/pipelines" class="sb-item" title="流水线"><Workflow :size="15" :stroke-width="1.5" /><span class="sb-label">流水线</span></router-link>
            <router-link to="/images" class="sb-item" title="镜像"><Layers :size="15" :stroke-width="1.5" /><span class="sb-label">镜像</span></router-link>
            <router-link to="/deploy" class="sb-item" title="部署"><Rocket :size="15" :stroke-width="1.5" /><span class="sb-label">部署</span></router-link>
            <router-link to="/push" class="sb-item" title="推送"><Upload :size="15" :stroke-width="1.5" /><span class="sb-label">推送</span></router-link>
          </div>
          
        </div>
      </aside>

      <!-- Main Area -->
      <div class="main-area">
        <!-- Editor Surface -->
        <div class="editor-surface">
          <div class="editor-inner">
            <router-view v-slot="{ Component }">
              <transition name="route" mode="out-in">
                <component :is="Component" />
              </transition>
            </router-view>
          </div>
        </div>
      </div>
    </div>

    <!-- Status Bar -->
    <div class="status-bar">
      <span class="st-item"><span class="st-dot" :class="{ online: dockerConnected }" />{{ dockerConnected ? 'Docker 已连接' : 'Docker 未连接' }}</span>
      <span class="st-sep">|</span>
      <span class="st-item"><GitBranch :size="11" :stroke-width="1.75" /> {{ currentBranch }}</span>
      <span class="st-spacer" />
      <span class="st-item">队列: {{ buildQueue }}</span>
      <span class="st-sep">|</span>
      <span class="st-item">FlowCI v2.0.0</span>
    </div>

    <ToastHost />
    <ConfirmDialog />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, provide } from 'vue'

import {
  LayoutDashboard, Package, GitBranch, Hammer, Workflow, Rocket,
  Layers, Upload, Settings, Sun, Moon, Pin, Minus, Maximize2,
  Minimize2, X, Folder, ChevronLeft, ChevronRight,
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
provide('toast', { success: toast.success, error: toast.error, info: toast.info, warning: toast.warning })

const { theme, isDark, load, setTheme } = useSettings()
provide('theme', { current: theme, isDark, setTheme })
function toggleTheme() { void setTheme(theme.value === 'dark' ? 'light' : 'dark') }

/* --- Activity Bar --- */
/* --- Demo data --- */
const projectList = ref([
  { id: '1', name: 'my-app', buildCount: 8 },
  { id: '2', name: 'backend-api', buildCount: 5 },
  { id: '3', name: 'frontend-web', buildCount: 3 },
])
const sidebarCollapsed = ref(false)
const dockerConnected = ref(true)
const currentBranch = ref('master')
const buildQueue = ref(2)

/* --- Window controls --- */
const alwaysOnTop = ref(false)
const isMax = ref(false)
async function refreshState() {
  try { alwaysOnTop.value = await GetWindowAlwaysOnTop(); isMax.value = await WindowIsMaximised() } catch {}
}
async function onTogglePin() { try { await SetWindowAlwaysOnTop(!alwaysOnTop.value); alwaysOnTop.value = !alwaysOnTop.value } catch {} }
async function onMinimise() { try { await WindowMinimise() } catch {} }
async function onToggleMax() { try { await WindowToggleMaximise(); setTimeout(refreshState, 50) } catch {} }
async function onClose() { try { await QuitApp() } catch {} }

let poll: number | undefined
onMounted(() => { void load(); void refreshState(); poll = window.setInterval(refreshState, 1500) })
onUnmounted(() => { if (poll) window.clearInterval(poll) })
</script>

<style scoped>
#app-shell { display:flex; flex-direction:column; height:100vh; overflow:hidden; background:var(--bg-canvas) }

/* Titlebar */
.titlebar { display:flex; align-items:center; height:36px; background:var(--bg-elevated); border-bottom:1px solid var(--border-subtle); flex-shrink:0; padding:0 var(--space-2); user-select:none; gap:var(--space-2) }
.titlebar-logo { display:flex; align-items:center; gap:var(--space-2); font-weight:var(--weight-semibold); font-size:var(--text-sm); color:var(--text-primary) }
.brand-dot { width:8px; height:8px; border-radius:2px; background:var(--brand-500) }
.titlebar-spacer { flex:1 }
.titlebar-actions { display:flex; align-items:center; gap:1px }
.wc-btn { width:32px; height:26px; display:flex; align-items:center; justify-content:center; border:none; background:transparent; color:var(--text-titlebar-icon); cursor:pointer; border-radius:var(--radius-sm); transition:all var(--transition-fast) }
.wc-btn:hover { background:var(--bg-titlebar-hover); color:var(--text-titlebar) }
.wc-btn.wc-pin-active { background:var(--brand-500); color:var(--text-on-brand) }
.wc-close:hover { background:#E05559; color:#fff }

/* Body */
.app-body { display:flex; flex:1; min-height:0; overflow:hidden; background:var(--bg-canvas); gap:4px; padding:4px }

/* Sidebar */
.sidebar { width:var(--sidebar-w); flex-shrink:0; min-height:0; background:var(--bg-panel); display:flex; flex-direction:column; overflow-y:auto; border:1px solid var(--border-strong); border-radius:var(--radius-md); padding:var(--space-2) 0; transition:width var(--transition-base) }
.sidebar.collapsed { width:48px }
.sidebar.collapsed .sb-header h4 { display:none }
.sidebar.collapsed .sb-label { display:none }
.sidebar.collapsed .sb-badge { display:none }
.sidebar.collapsed .sb-section-label { display:none }
.sidebar.collapsed .sb-item { justify-content:center; padding:var(--space-3) var(--space-1) }
.sidebar.collapsed .sb-header { justify-content:center }
.sidebar.collapsed .sb-toggle { margin:0 auto }
.sb-header { display:flex; align-items:center; justify-content:space-between; padding:var(--space-2) var(--space-4); border-bottom:1px solid var(--border-subtle); flex-shrink:0 }
.sb-header h4 { font-size:10px; font-weight:var(--weight-semibold); color:var(--text-ghost); text-transform:uppercase; letter-spacing:.08em; margin:0 }
.sb-toggle { width:20px; height:20px; display:flex; align-items:center; justify-content:center; border:none; background:transparent; color:var(--text-muted); cursor:pointer; border-radius:var(--radius-sm) }
.sb-toggle:hover { background:var(--bg-titlebar-hover); color:var(--text-primary) }
.sb-close { width:20px; height:20px; display:flex; align-items:center; justify-content:center; border:none; background:transparent; color:var(--text-muted); cursor:pointer; border-radius:var(--radius-sm) }
.sb-close:hover { background:var(--bg-titlebar-hover); color:var(--text-primary) }
.sb-body { flex:1; overflow-y:auto; padding:var(--space-1) 0 }
.sb-section { margin-bottom:var(--space-1) }
.sb-section + .sb-section { border-top:1px solid var(--border-subtle); padding-top:var(--space-1) }
.sb-section-label { padding:var(--space-2) var(--space-4) var(--space-1); font-size:var(--text-xs); font-weight:var(--weight-semibold); color:var(--text-ghost); text-transform:uppercase; letter-spacing:.06em }
.sb-item { display:flex; align-items:center; gap:var(--space-2); padding:var(--space-2) var(--space-4); cursor:pointer; font-size:var(--text-base); color:var(--text-nav); text-decoration:none; transition:all var(--transition-fast); white-space:nowrap }
.sb-item:hover { background:var(--bg-titlebar-hover); color:var(--text-nav-hover) }
.sb-item.router-link-exact-active { background:var(--bg-nav-active); color:var(--text-nav-active); font-weight:var(--weight-medium) }
.sb-label { flex:1; overflow:hidden; text-overflow:ellipsis }
.sb-badge { font-size:10px; color:var(--text-muted); font-family:var(--font-mono) }

/* Main Area */
.main-area { flex:1; display:flex; flex-direction:column; min-width:0; min-height:0; overflow:hidden; background:var(--bg-panel); border:1px solid var(--border-strong); border-radius:var(--radius-md) }

/* Tab Bar */
.tab-bar { display:flex; align-items:stretch; height:32px; background:var(--bg-elevated); border-bottom:1px solid var(--border-subtle); flex-shrink:0; overflow-x:auto; border-radius:var(--radius-md) var(--radius-md) 0 0 }
.tab-bar::-webkit-scrollbar { height:0 }
.tab { display:flex; align-items:center; gap:var(--space-2); padding:0 var(--space-4); font-size:var(--text-sm); color:var(--text-secondary); cursor:pointer; border-bottom:2px solid transparent; white-space:nowrap; text-decoration:none; transition:all var(--transition-fast); flex-shrink:0 }
.tab:hover { color:var(--text-primary); background:var(--bg-titlebar-hover) }
.tab.active { color:var(--brand-500); border-bottom-color:var(--brand-500) }
.editor-surface { flex:1; overflow-y:auto; display:flex; flex-direction:column; background:var(--bg-editor); border-top:1px solid var(--border-subtle); box-shadow:inset 0 1px 0 rgba(255,255,255,.02) }
.editor-inner { max-width:1100px; width:100%; margin:0 auto; padding:var(--space-5) var(--space-6) }

/* Status Bar */
.status-bar { display:flex; align-items:center; height:22px; background:var(--bg-elevated); border-top:1px solid var(--border-default); flex-shrink:0; padding:0 var(--space-3); font-size:10px; color:var(--text-muted); gap:var(--space-2); user-select:none }
.st-item { display:flex; align-items:center; gap:4px; white-space:nowrap }
.st-dot { width:6px; height:6px; border-radius:50% }
.st-dot.online { background:var(--success-fg); box-shadow:0 0 4px var(--success-fg) }
.st-sep { color:var(--border-strong) }
.st-spacer { flex:1 }

/* Route transitions */
.route-enter-active, .route-leave-active { transition:opacity var(--transition-base),transform var(--transition-base) }
.route-enter-from { opacity:0; transform:translateY(6px) }
.route-leave-to { opacity:0; transform:translateY(-6px) }
</style>

