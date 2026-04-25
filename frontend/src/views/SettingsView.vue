<template>
  <div class="settings-view">
    <h1>设置</h1>

    <div class="card">
      <h3>Docker 连接</h3>
      <div class="status-badge" :class="dockerStatus.status">
        <span class="status-dot"></span>
        {{ dockerStatus.text }}
      </div>
      <div v-if="dockerStatus.version" class="version-info">
        <span>Docker 版本: {{ dockerStatus.version }}</span>
      </div>

      <div class="form-group" style="margin-top: 20px;">
        <label>Docker Host（远程 daemon 地址，留空使用本地）</label>
        <input
          v-model="settings.dockerHost"
          type="text"
          placeholder="tcp://192.168.1.100:2375 或 ssh://user@host"
        />
        <p class="hint">
          支持 tcp:// / ssh:// / npipe:// / unix:// 协议，等同于设置 DOCKER_HOST 环境变量。
          保存后立即生效，无需重启。
        </p>
      </div>

      <div style="display: flex; gap: 12px; margin-top: 16px;">
        <button class="btn-outline" @click="checkDocker">检查连接</button>
        <button class="btn-primary" @click="saveSettings">保存</button>
      </div>
    </div>

    <div class="card">
      <h3>主题设置</h3>
      <div class="theme-toggle">
        <button
          class="theme-btn"
          :class="{ active: settings.theme === 'system' }"
          @click="setTheme('system')"
        >
          💻 跟随系统
        </button>
        <button
          class="theme-btn"
          :class="{ active: settings.theme === 'dark' }"
          @click="setTheme('dark')"
        >
          🌙 深色
        </button>
        <button
          class="theme-btn"
          :class="{ active: settings.theme === 'light' }"
          @click="setTheme('light')"
        >
          ☀️ 浅色
        </button>
      </div>
    </div>

    <div class="card">
      <h3>默认配置</h3>
      <div class="form-group">
        <label>默认镜像仓库</label>
        <input v-model="settings.defaultRegistry" type="text" placeholder="docker.io" />
      </div>
      <div class="form-group">
        <label>默认工作目录</label>
        <input v-model="settings.defaultWorkdir" type="text" placeholder="/workspace" />
      </div>
      <button class="btn-primary" @click="saveSettings">保存设置</button>
    </div>

    <div class="card">
      <h3>关于</h3>
      <div class="about-info">
        <p><strong>FlowCI</strong> - 轻量级 Docker 构建部署工具</p>
        <p>版本: 0.1.0</p>
        <p>技术栈: Wails + Go + Vue 3</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, inject, onMounted } from 'vue'
import { CheckDocker, GetSettings, SaveSettings } from '../wailsjs/go/handler/App'

const toast = inject('toast') as { success: (msg: string) => void; error: (msg: string) => void; info: (msg: string) => void }
const themeContext = inject('theme') as { current: { value: string }; setTheme: (theme: string) => void }

const dockerStatus = ref({
  status: 'checking',
  text: '检查中...',
  version: ''
})

const settings = ref({
  defaultRegistry: 'docker.io',
  defaultWorkdir: '/workspace',
  theme: 'system',
  dockerHost: '',
})

async function loadSettings() {
  try {
    const result = await GetSettings()
    if (result.defaultRegistry) settings.value.defaultRegistry = result.defaultRegistry
    if (result.defaultWorkdir) settings.value.defaultWorkdir = result.defaultWorkdir
    if (result.dockerHost) settings.value.dockerHost = result.dockerHost
    if (result.theme) {
      settings.value.theme = result.theme
      themeContext?.setTheme(result.theme)
    }
  } catch (e) {
    console.error('Failed to load settings:', e)
  }
}

function setTheme(theme: string) {
  settings.value.theme = theme
  themeContext?.setTheme(theme)
  saveSettings()
}

async function checkDocker() {
  dockerStatus.value = {
    status: 'checking',
    text: '检查中...',
    version: ''
  }
  
  try {
    const result = await CheckDocker()
    dockerStatus.value = {
      status: result.connected ? 'connected' : 'disconnected',
      text: result.connected ? '已连接' : '未连接',
      version: result.version || ''
    }
  } catch (e) {
    dockerStatus.value = {
      status: 'error',
      text: '检查失败',
      version: ''
    }
  }
}

async function saveSettings() {
  try {
    await SaveSettings({
      settings: {
        defaultRegistry: settings.value.defaultRegistry,
        defaultWorkdir: settings.value.defaultWorkdir,
        theme: settings.value.theme,
        dockerHost: settings.value.dockerHost,
      },
    })
    toast?.success('设置已保存')
    // dockerHost 改了后立刻重测连接
    void checkDocker()
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast?.error(`保存失败: ${msg}`)
  }
}

onMounted(() => {
  checkDocker()
  loadSettings()
})
</script>

<style scoped>
.settings-view {
  max-width: 700px;
}

h1 {
  font-size: 28px;
  margin-bottom: 24px;
}

.card {
  background: var(--card-bg, #fff);
  border-radius: 12px;
  padding: 24px;
  box-shadow: var(--shadow, 0 2px 12px rgba(0, 0, 0, 0.05));
  margin-bottom: 20px;
}

.theme-toggle {
  display: flex;
  gap: 12px;
}

.theme-btn {
  flex: 1;
  padding: 12px 20px;
  border: 2px solid var(--border-color, #e0e0e0);
  border-radius: 8px;
  background: transparent;
  color: var(--text-secondary, #666);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.theme-btn:hover {
  border-color: #667eea;
  color: #667eea;
}

.theme-btn.active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-color: transparent;
  color: white;
}

.card {
  background: var(--card-bg, #fff);
  border-radius: 12px;
  padding: 24px;
  box-shadow: var(--shadow, 0 2px 12px rgba(0, 0, 0, 0.05));
  margin-bottom: 20px;
}

.card h3 {
  font-size: 18px;
  color: var(--text-primary, #1a1a2e);
  margin-bottom: 16px;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border-radius: 8px;
  font-weight: 600;
}

.status-badge.checking {
  background: #fef3c7;
  color: #d97706;
}

.status-badge.connected {
  background: #dcfce7;
  color: #16a34a;
}

.status-badge.disconnected,
.status-badge.error {
  background: #fee2e2;
  color: #dc2626;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: currentColor;
}

.version-info {
  margin-top: 12px;
  color: var(--text-secondary, #666);
  font-size: 14px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}

.form-group label {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary, #333);
}

.form-group input {
  padding: 12px;
  border: 2px solid var(--border-color, #e0e0e0);
  border-radius: 8px;
  font-size: 14px;
  background: var(--card-bg, #fff);
  color: var(--text-primary, #333);
  transition: border-color 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: #667eea;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary:hover {
  transform: translateY(-2px);
}

.btn-outline {
  background: transparent;
  color: #667eea;
  border: 2px solid #667eea;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-outline:hover {
  background: #667eea;
  color: white;
}

.about-info p {
  color: var(--text-secondary, #666);
  margin-bottom: 8px;
}
</style>
