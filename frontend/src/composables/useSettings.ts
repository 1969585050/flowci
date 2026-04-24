import { ref, computed, readonly } from 'vue'
import type { ThemeMode } from '../types'
// NOTE: 这些 import 在 `wails generate module` 之前可能指向旧路径。
// 本仓阶段 2 后端把 App 从 main 包迁到了 handler 包；在本机运行一次
// `wails generate module` 即可让 wailsjs/go/handler/App 就绪。
// 详细迁移步骤见 docs/standards/FRONTEND_MIGRATION.md。
// @ts-expect-error wailsjs 绑定需要本地 `wails generate module` 后生效
import { GetSettings, SaveSettings } from '../wailsjs/go/handler/App'

/**
 * 全局设置状态。启动阶段调一次 load() 从 SQLite 读；改动时写回 SQLite。
 * 不使用 localStorage 持久化（见 frontend-spec.md §6.3）。
 */

const theme = ref<ThemeMode>('dark')
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
    isDark: computed(() => {
      const actual = theme.value === 'system' ? resolveSystemTheme() : theme.value
      return actual === 'dark'
    }),
    load,
    setTheme,
    watchSystemTheme,
  }
}
