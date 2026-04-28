import { ref, computed, readonly } from 'vue'
import type { ThemeMode } from '../types'
// wails CLI 启动时会自动 generate 出 wailsjs/go/handler/App（对应 handler.App）
// 若本地尚未跑过 `wails dev` / `wails generate module`，该路径不存在，TS 会报错
// 一旦跑过任一命令，类型即恢复。`@ts-ignore` 使本文件在两种状态下都可编译。
// @ts-ignore
import { GetSettings, SaveSettings } from '../wailsjs/go/handler/App'

/**
 * 全局设置状态。启动阶段调一次 load() 从 SQLite 读；改动时写回 SQLite。
 * 不使用 localStorage 持久化（见 frontend-spec.md §6.3）。
 */

const theme = ref<ThemeMode>('dark')
const sidebarCollapsed = ref(false)
const loaded = ref(false)

function resolveSystemTheme(): 'dark' | 'light' {
  if (window.matchMedia?.('(prefers-color-scheme: dark)').matches) return 'dark'
  return 'light'
}

function applyTheme(mode: ThemeMode) {
  const actual = mode === 'system' ? resolveSystemTheme() : mode
  document.documentElement.setAttribute('data-theme', actual)
}

async function load() {
  try {
    const settings = await GetSettings()
    if (settings?.theme) {
      theme.value = settings.theme as ThemeMode
    }
    if (settings?.sidebar_collapsed === 'true') {
      sidebarCollapsed.value = true
    }
  } catch (e) {
    // 读失败保持默认 dark；不抛给调用方
    console.error('load settings failed:', e)
  }
  applyTheme(theme.value)
  loaded.value = true
}

async function setTheme(mode: ThemeMode) {
  theme.value = mode
  applyTheme(mode)
  try {
    await SaveSettings({ settings: { theme: mode } })
  } catch (e) {
    console.error('save theme failed:', e)
  }
}

async function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
  try {
    await SaveSettings({ settings: { sidebar_collapsed: String(sidebarCollapsed.value) } })
  } catch (e) {
    console.error('save sidebar_collapsed failed:', e)
  }
}

function watchSystemTheme() {
  const mq = window.matchMedia?.('(prefers-color-scheme: dark)')
  if (!mq) return
  mq.addEventListener('change', () => {
    if (theme.value === 'system') applyTheme('system')
  })
}

export function useSettings() {
  return {
    theme: readonly(theme),
    loaded: readonly(loaded),
    sidebarCollapsed: readonly(sidebarCollapsed),
    toggleSidebar,
    isDark: computed(() => {
      const actual = theme.value === 'system' ? resolveSystemTheme() : theme.value
      return actual === 'dark'
    }),
    load,
    setTheme,
    watchSystemTheme,
  }
}
