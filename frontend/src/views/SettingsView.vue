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
        <button class="btn-outline" @click="detectEnv" :disabled="detecting">
          {{ detecting ? '检测中…' : '🔍 检测环境' }}
        </button>
        <button class="btn-primary" @click="saveSettings">保存</button>
      </div>

      <div v-if="envReport" class="env-report">
        <div class="env-row" :class="envReport.connected ? 'ok' : 'fail'">
          <span class="dot"></span>
          <span class="env-label">Docker daemon</span>
          <span class="env-value">
            {{ envReport.connected
              ? `已连接 · server ${envReport.serverVersion} (${envReport.serverOS}/${envReport.serverArch})`
              : envReport.message || '不可达' }}
          </span>
        </div>
        <div v-if="envReport.clientVersion" class="env-row ok">
          <span class="dot"></span>
          <span class="env-label">Docker CLI</span>
          <span class="env-value">v{{ envReport.clientVersion }}</span>
        </div>
        <div class="env-row" :class="envReport.hasBuildx ? 'ok' : 'warn'">
          <span class="dot"></span>
          <span class="env-label">buildx 插件</span>
          <span class="env-value">{{ envReport.hasBuildx ? '可用' : '不可用（构建镜像将失败）' }}</span>
        </div>
        <div class="env-row" :class="envReport.hasCompose ? 'ok' : 'warn'">
          <span class="dot"></span>
          <span class="env-label">compose 插件</span>
          <span class="env-value">{{ envReport.hasCompose ? '可用' : '不可用（不能用 compose 部署）' }}</span>
        </div>
      </div>
    </div>

    <div class="card">
      <h3>Git 环境</h3>
      <div v-if="gitReport === null" class="env-row" style="padding: 0;">
        <span class="env-label">检测中…</span>
      </div>
      <template v-else>
        <div class="env-row" :class="gitReport.installed ? 'ok' : 'fail'">
          <span class="dot"></span>
          <span class="env-label">git CLI</span>
          <span class="env-value">
            {{ gitReport.installed ? gitReport.version : (gitReport.message || '未安装') }}
          </span>
        </div>
        <div v-if="gitReport.path" class="env-row ok">
          <span class="dot"></span>
          <span class="env-label">安装位置</span>
          <span class="env-value mono">{{ gitReport.path }}</span>
        </div>

        <div v-if="!gitReport.installed && gitReport.installHints?.length" class="install-hints">
          <p class="install-title">推荐安装方式：</p>
          <div v-for="(hint, idx) in gitReport.installHints" :key="idx" class="hint-row">
            <div class="hint-label">{{ hint.label }}</div>
            <div v-if="hint.command" class="hint-cmd">
              <code>{{ hint.command }}</code>
              <button class="btn-copy" @click="copyHintCmd(hint.command)">复制</button>
            </div>
            <a v-if="hint.url" :href="hint.url" target="_blank" rel="noopener" class="hint-link">
              {{ hint.command ? '官方文档 →' : '打开下载页 →' }}
            </a>
          </div>
        </div>
      </template>

      <button class="btn-outline" style="margin-top: 16px;" @click="checkGit">
        重新检测
      </button>
    </div>

    <div class="card">
      <h3>🦊 Gitea 集成</h3>
      <p class="hint" style="margin-top: 0;">
        配置 Gitea 实例 URL 和 Personal Access Token，
        即可自动扫描你能访问的全部仓库并一键导入为 FlowCI 项目（自动 clone 到本地）。
      </p>

      <div class="form-group">
        <label>Gitea 实例 URL</label>
        <input
          v-model="gitea.baseURL"
          type="text"
          placeholder="https://gitea.example.com 或 http://localhost:3000"
        />
      </div>

      <div class="form-group">
        <label>
          Access Token
          <span v-if="giteaStatus?.hasToken" class="ai-key-status">✓ 已配置</span>
          <span v-else class="ai-key-status missing">未配置</span>
        </label>
        <input
          v-model="gitea.tokenInput"
          type="password"
          :placeholder="giteaStatus?.hasToken ? '已保存（留空不修改；填新值覆盖）' : '从 Gitea 用户设置 → 应用 → 生成新令牌'"
        />
      </div>

      <details class="gitea-help">
        <summary>📘 如何生成 Token？</summary>
        <ol>
          <li>登录你的 Gitea 实例</li>
          <li>右上角头像 → <strong>设置</strong> → 左侧 <strong>应用</strong> 标签</li>
          <li>“管理访问令牌”区域 → 输入名称（如 <code>FlowCI</code>）→ 选择 scope <code>read:repository</code></li>
          <li>点击 <strong>生成令牌</strong> → 复制显示的 token（只显示一次！）粘贴到上方输入框</li>
          <li>保存配置 → 点 <strong>验证连接</strong> 看到用户名即成功</li>
        </ol>
        <a v-if="giteaStatus?.tokenSettingsUrl"
           :href="giteaStatus.tokenSettingsUrl"
           target="_blank" rel="noopener"
           class="hint-link">
          🔗 直接打开 Token 设置页 →
        </a>
      </details>

      <div style="display: flex; gap: 12px; margin-top: 16px;">
        <button class="btn-primary" @click="saveGitea" :disabled="savingGitea">
          {{ savingGitea ? '保存中…' : '保存配置' }}
        </button>
        <button class="btn-outline" @click="verifyGitea" :disabled="verifying || !giteaStatus?.hasToken">
          {{ verifying ? '验证中…' : '验证连接' }}
        </button>
      </div>

      <div v-if="giteaUser" class="env-row ok" style="margin-top: 12px;">
        <span class="dot"></span>
        <span class="env-label">当前用户</span>
        <span class="env-value">{{ giteaUser.username }} ({{ giteaUser.email || '无邮箱' }})</span>
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
      <h3>🤖 AI 助手</h3>
      <p class="hint" style="margin-bottom: 16px; margin-top: 0;">
        配置 OpenAI 兼容的 LLM API（OpenAI / DeepSeek / 月之暗面 / 本地 ollama 等），
        构建失败时可一键 AI 诊断日志。API key 存 OS keyring，不入数据库。
      </p>

      <div class="form-group">
        <label>Base URL</label>
        <input
          v-model="settings.aiBaseURL"
          type="text"
          placeholder="https://api.openai.com 或 http://localhost:11434"
        />
      </div>
      <div class="form-group">
        <label>Model</label>
        <input
          v-model="settings.aiModel"
          type="text"
          placeholder="gpt-4o-mini / deepseek-chat / qwen2.5:7b 等"
        />
      </div>
      <div class="form-group">
        <label>API Key
          <span v-if="aiKeyConfigured" class="ai-key-status">✓ 已配置</span>
          <span v-else class="ai-key-status missing">未配置</span>
        </label>
        <input
          v-model="aiKeyInput"
          type="password"
          :placeholder="aiKeyConfigured ? '已保存（留空不修改；填新值覆盖；填空格清除）' : 'sk-...'"
        />
      </div>

      <div style="display: flex; gap: 12px;">
        <button class="btn-primary" @click="saveAISettings">保存 AI 配置</button>
      </div>
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
import {
  CheckDocker, GetSettings, SaveSettings, DetectDockerEnv,
  GetAIKeyStatus, SaveAIKey, DetectGitEnv,
  SaveGiteaConfig, GetGiteaStatus, VerifyGitea,
} from '../wailsjs/go/handler/App'

const toast = inject('toast') as { success: (msg: string) => void; error: (msg: string) => void; info: (msg: string) => void }
const themeContext = inject('theme') as { current: { value: string }; setTheme: (theme: string) => void }

const dockerStatus = ref({
  status: 'checking',
  text: '检查中...',
  version: ''
})

const detecting = ref(false)
const envReport = ref<{
  host: string
  connected: boolean
  clientVersion: string
  serverVersion: string
  serverOS: string
  serverArch: string
  hasBuildx: boolean
  hasCompose: boolean
  message: string
} | null>(null)

async function detectEnv() {
  detecting.value = true
  try {
    envReport.value = await DetectDockerEnv({ host: settings.value.dockerHost })
    if (envReport.value?.connected) {
      toast?.success('环境检测完成')
    } else {
      toast?.warning?.(envReport.value?.message || '环境检测发现问题')
    }
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast?.error(`检测失败: ${msg}`)
  }
  detecting.value = false
}

const settings = ref({
  defaultRegistry: 'docker.io',
  defaultWorkdir: '/workspace',
  theme: 'system',
  dockerHost: '',
  aiBaseURL: '',
  aiModel: '',
})

const aiKeyInput = ref('')
const aiKeyConfigured = ref(false)

interface GitInstallHint {
  method: string
  label: string
  command: string
  url: string
}
interface GitReport {
  installed: boolean
  version: string
  path: string
  message: string
  installHints: GitInstallHint[]
}
const gitReport = ref<GitReport | null>(null)

async function checkGit() {
  gitReport.value = null
  try {
    gitReport.value = await DetectGitEnv()
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast?.error(`Git 检测失败: ${msg}`)
  }
}

async function copyHintCmd(cmd: string) {
  try {
    await navigator.clipboard.writeText(cmd)
    toast?.success('命令已复制到剪贴板')
  } catch {
    toast?.error('复制失败')
  }
}

// ---- Gitea ----

interface GiteaStatus {
  baseUrl: string
  hasToken: boolean
  tokenSettingsUrl: string
}
interface GiteaUser {
  username: string
  email: string
  avatarUrl: string
}

const gitea = ref({ baseURL: '', tokenInput: '' })
const giteaStatus = ref<GiteaStatus | null>(null)
const giteaUser = ref<GiteaUser | null>(null)
const savingGitea = ref(false)
const verifying = ref(false)

async function loadGiteaStatus() {
  try {
    giteaStatus.value = await GetGiteaStatus()
    if (giteaStatus.value) {
      gitea.value.baseURL = giteaStatus.value.baseUrl
    }
  } catch (e) {
    console.error('load gitea status failed:', e)
  }
}

async function saveGitea() {
  savingGitea.value = true
  try {
    await SaveGiteaConfig({
      baseUrl: gitea.value.baseURL,
      token: gitea.value.tokenInput,
    })
    gitea.value.tokenInput = ''
    await loadGiteaStatus()
    toast?.success('Gitea 配置已保存')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast?.error(`保存失败: ${msg}`)
  }
  savingGitea.value = false
}

async function verifyGitea() {
  verifying.value = true
  giteaUser.value = null
  try {
    giteaUser.value = await VerifyGitea()
    toast?.success(`验证成功：${giteaUser.value?.username}`)
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast?.error(`验证失败: ${msg}`)
  }
  verifying.value = false
}

async function refreshAIKeyStatus() {
  try {
    const s = await GetAIKeyStatus()
    aiKeyConfigured.value = !!s?.configured
  } catch {
    aiKeyConfigured.value = false
  }
}

async function saveAISettings() {
  try {
    await SaveSettings({
      settings: {
        aiBaseURL: settings.value.aiBaseURL,
        aiModel: settings.value.aiModel,
      },
    })
    // API key 单独走 keyring；空字符串保留旧值（不改），全空格视为清除
    if (aiKeyInput.value !== '') {
      const trimmed = aiKeyInput.value.trim()
      await SaveAIKey({ apiKey: trimmed })
      aiKeyInput.value = ''
      await refreshAIKeyStatus()
    }
    toast?.success('AI 配置已保存')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast?.error(`保存失败: ${msg}`)
  }
}

async function loadSettings() {
  try {
    const result = await GetSettings()
    if (result.defaultRegistry) settings.value.defaultRegistry = result.defaultRegistry
    if (result.defaultWorkdir) settings.value.defaultWorkdir = result.defaultWorkdir
    if (result.dockerHost) settings.value.dockerHost = result.dockerHost
    if (result.aiBaseURL) settings.value.aiBaseURL = result.aiBaseURL
    if (result.aiModel) settings.value.aiModel = result.aiModel
    if (result.theme) {
      settings.value.theme = result.theme
      themeContext?.setTheme(result.theme)
    }
  } catch (e) {
    console.error('Failed to load settings:', e)
  }
  await refreshAIKeyStatus()
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
  checkGit()
  loadGiteaStatus()
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

.hint {
  margin-top: 6px;
  font-size: 12px;
  color: var(--text-muted, #94a3b8);
  line-height: 1.5;
}

.env-report {
  margin-top: 16px;
  padding: 12px 16px;
  background: var(--bg-primary, #f5f7fa);
  border-radius: 8px;
  border: 1px solid var(--border-color, #e0e0e0);
}

.env-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 0;
  font-size: 13px;
}

.env-row .dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.env-row.ok   .dot { background: var(--success-fg, #16a34a); }
.env-row.warn .dot { background: var(--warning-fg, #d97706); }
.env-row.fail .dot { background: var(--danger-fg,  #dc2626); }

.env-row .env-label {
  width: 110px;
  color: var(--text-secondary, #666);
}

.env-row .env-value {
  color: var(--text-primary, #1a1a2e);
  flex: 1;
}

.ai-key-status {
  margin-left: 8px;
  font-size: 12px;
  font-weight: normal;
  color: var(--success-fg, #16a34a);
}
.ai-key-status.missing {
  color: var(--text-muted, #94a3b8);
}

.env-value.mono {
  font-family: 'JetBrains Mono', 'Consolas', monospace;
  font-size: 12px;
  color: var(--text-secondary, #666);
  word-break: break-all;
}

.install-hints {
  margin-top: 16px;
  padding: 12px 16px;
  background: var(--warning-bg, #fffbeb);
  border-left: 3px solid var(--warning-fg, #d97706);
  border-radius: 6px;
}

.install-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--warning-fg, #92400e);
  margin: 0 0 10px 0;
}

.hint-row {
  margin-bottom: 10px;
}
.hint-row:last-child {
  margin-bottom: 0;
}

.hint-label {
  font-size: 12px;
  color: var(--text-primary, #1a1a2e);
  font-weight: 500;
  margin-bottom: 4px;
}

.hint-cmd {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--bg-secondary, #fff);
  padding: 6px 10px;
  border-radius: 4px;
  border: 1px solid var(--border-color, #e0e0e0);
}
.hint-cmd code {
  flex: 1;
  font-family: 'JetBrains Mono', 'Consolas', monospace;
  font-size: 12px;
  color: var(--text-primary, #1a1a2e);
  white-space: nowrap;
  overflow-x: auto;
}
.btn-copy {
  background: transparent;
  border: 1px solid var(--brand-start, #667eea);
  color: var(--brand-start, #667eea);
  padding: 2px 10px;
  border-radius: 4px;
  font-size: 11px;
  cursor: pointer;
}
.btn-copy:hover {
  background: var(--brand-soft, rgba(102, 126, 234, 0.12));
}

.hint-link {
  display: inline-block;
  margin-top: 4px;
  font-size: 12px;
  color: var(--brand-start, #667eea);
  text-decoration: none;
}
.hint-link:hover {
  text-decoration: underline;
}

.gitea-help {
  margin-top: -4px;
  margin-bottom: 8px;
  padding: 10px 14px;
  background: var(--info-bg, #eff6ff);
  border-left: 3px solid var(--info-fg, #1e40af);
  border-radius: 4px;
  font-size: 12px;
  color: var(--text-secondary, #4a5568);
  line-height: 1.6;
}
.gitea-help summary {
  cursor: pointer;
  font-weight: 500;
  color: var(--info-fg, #1e40af);
  user-select: none;
  outline: none;
}
.gitea-help ol {
  margin: 8px 0 8px 20px;
  padding: 0;
}
.gitea-help li {
  margin-bottom: 4px;
}
.gitea-help code {
  background: rgba(0, 0, 0, 0.06);
  padding: 1px 5px;
  border-radius: 3px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
}
</style>
