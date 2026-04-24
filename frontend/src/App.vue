<template>
  <div id="app" :class="themeClass">
    <div class="sidebar">
      <div class="logo">
        <span class="logo-icon">🚀</span>
        <span class="logo-text">FlowCI</span>
      </div>
      <div class="nav">
        <router-link to="/projects" class="nav-item" active-class="active">
          <span class="nav-icon">📦</span>
          <span>项目</span>
        </router-link>
        <router-link to="/build" class="nav-item" active-class="active">
          <span class="nav-icon">🔨</span>
          <span>构建</span>
        </router-link>
        <router-link to="/pipelines" class="nav-item" active-class="active">
          <span class="nav-icon">🔧</span>
          <span>流水线</span>
        </router-link>
        <router-link to="/deploy" class="nav-item" active-class="active">
          <span class="nav-icon">🌐</span>
          <span>部署</span>
        </router-link>
        <router-link to="/images" class="nav-item" active-class="active">
          <span class="nav-icon">🗃️</span>
          <span>镜像</span>
        </router-link>
        <router-link to="/push" class="nav-item" active-class="active">
          <span class="nav-icon">📤</span>
          <span>推送</span>
        </router-link>
        <router-link to="/settings" class="nav-item" active-class="active">
          <span class="nav-icon">⚙️</span>
          <span>设置</span>
        </router-link>
      </div>
    </div>
    <div class="content">
      <router-view />
    </div>
    <Toast ref="toastRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, provide, onMounted, computed, watch } from 'vue'
import Toast from './components/Toast.vue'
import { GetSettings, SaveSettings } from './wailsjs/go/main/App'

const toastRef = ref<InstanceType<typeof Toast>>()
const currentTheme = ref('dark')

provide('toast', {
  success(msg: string) { toastRef.value?.addToast('success', msg) },
  error(msg: string) { toastRef.value?.addToast('error', msg) },
  info(msg: string) { toastRef.value?.addToast('info', msg) }
})

provide('theme', {
  current: currentTheme,
  isDark: computed(() => currentTheme.value === 'dark'),
  setTheme: (theme: string) => {
    currentTheme.value = theme
    document.documentElement.setAttribute('data-theme', theme)
  }
})

const themeClass = computed(() => `theme-${currentTheme.value}`)

function getSystemTheme(): string {
  if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    return 'dark'
  }
  return 'light'
}

function applyTheme(theme: string) {
  const actualTheme = theme === 'system' ? getSystemTheme() : theme
  currentTheme.value = actualTheme
  document.documentElement.setAttribute('data-theme', actualTheme)
}

async function loadTheme() {
  try {
    const settings = await GetSettings(null)
    if (settings.theme) {
      applyTheme(settings.theme)
    }
  } catch (e) {
    console.error('Failed to load theme:', e)
    applyTheme('dark')
  }
}

function setupSystemThemeListener() {
  if (window.matchMedia) {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    mediaQuery.addEventListener('change', () => {
      const settings_val = localStorage.getItem('flowci_theme')
      if (settings_val === 'system') {
        applyTheme('system')
      }
    })
  }
}

onMounted(() => {
  loadTheme()
  setupSystemThemeListener()
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

:root {
  --bg-primary: #f5f5f5;
  --bg-secondary: #ffffff;
  --bg-sidebar: linear-gradient(180deg, #1e1e2e 0%, #2d2d44 100%);
  --text-primary: #333333;
  --text-secondary: #666666;
  --text-nav: #a0a0b0;
  --text-nav-hover: #ffffff;
  --text-nav-active: #667eea;
  --border-color: #e0e0e0;
  --shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
  --card-bg: #ffffff;
}

[data-theme="light"] {
  --bg-primary: #f5f7fa;
  --bg-secondary: #ffffff;
  --bg-sidebar: linear-gradient(180deg, #1a1a2e 0%, #2d2d44 100%);
  --text-primary: #1a1a2e;
  --text-secondary: #4a5568;
  --text-nav: #a0a0b0;
  --text-nav-hover: #ffffff;
  --text-nav-active: #667eea;
  --border-color: #e2e8f0;
  --shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  --card-bg: #ffffff;
}

[data-theme="dark"] {
  --bg-primary: #0f0f1a;
  --bg-secondary: #1a1a2e;
  --bg-sidebar: linear-gradient(180deg, #1e1e2e 0%, #2d2d44 100%);
  --text-primary: #e2e8f0;
  --text-secondary: #a0a0b0;
  --text-nav: #a0a0b0;
  --text-nav-hover: #ffffff;
  --text-nav-active: #667eea;
  --border-color: #2d3748;
  --shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
  --card-bg: #1a1a2e;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background: var(--bg-primary);
  color: var(--text-primary);
}

#app {
  display: flex;
  height: 100vh;
}

.theme-dark {
  --bg-primary: #0f0f1a;
  --bg-secondary: #1a1a2e;
  --text-primary: #e2e8f0;
  --text-secondary: #a0a0b0;
  --border-color: #2d3748;
  --shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
  --card-bg: #1a1a2e;
}

.theme-light {
  --bg-primary: #f5f7fa;
  --bg-secondary: #ffffff;
  --text-primary: #1a1a2e;
  --text-secondary: #4a5568;
  --border-color: #e2e8f0;
  --shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  --card-bg: #ffffff;
}

.sidebar {
  width: 240px;
  background: var(--bg-sidebar);
  color: #fff;
  display: flex;
  flex-direction: column;
  padding: 20px 0;
}

.logo {
  padding: 0 20px 30px;
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 24px;
  font-weight: bold;
}

.logo-icon {
  font-size: 32px;
}

.logo-text {
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.nav {
  flex: 1;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 24px;
  color: #a0a0b0;
  text-decoration: none;
  transition: all 0.2s;
  cursor: pointer;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #fff;
}

.nav-item.active {
  background: rgba(102, 126, 234, 0.2);
  color: #667eea;
  border-left: 3px solid #667eea;
}

.nav-icon {
  font-size: 20px;
}

.content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}
</style>
