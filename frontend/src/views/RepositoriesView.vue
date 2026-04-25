<template>
  <div class="repos-view">
    <h1>仓库源</h1>
    <p class="subtitle">配置 Git 托管平台凭证，自动扫描你的全部仓库并一键导入为 FlowCI 项目。</p>

    <div class="layout">
      <!-- 左侧 provider 列表 -->
      <aside class="provider-tabs">
        <div
          v-for="p in providers"
          :key="p.id"
          class="provider-tab"
          :class="{ active: activeProvider === p.id, disabled: !p.enabled }"
          @click="p.enabled && (activeProvider = p.id)"
        >
          <div class="provider-icon">{{ p.icon }}</div>
          <div class="provider-meta">
            <div class="provider-name">{{ p.name }}</div>
            <div class="provider-status">
              <span v-if="!p.enabled" class="badge badge-muted">即将支持</span>
              <span v-else-if="p.id === 'gitea' && giteaStatus?.hasToken" class="badge badge-ok">已配置</span>
              <span v-else class="badge badge-warn">未配置</span>
            </div>
          </div>
        </div>
      </aside>

      <!-- 右侧 - Gitea 详细 -->
      <section v-if="activeProvider === 'gitea'" class="provider-panel">
        <!-- 配置区 -->
        <div class="card">
          <h3>🦊 Gitea 配置</h3>
          <p class="hint" style="margin-top: 0;">
            支持本地部署或线上 Gitea 实例。Token 存 OS keyring，不入库。
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
              <span v-if="giteaStatus?.hasToken" class="badge badge-ok inline">✓ 已配置</span>
              <span v-else class="badge badge-warn inline">未配置</span>
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
              <li>“管理访问令牌”区域 → 输入名称（如 <code>FlowCI</code>）→ 选 scope <code>read:repository</code></li>
              <li><strong>生成令牌</strong> → 复制显示的 token（只显示一次！）粘贴到上方</li>
              <li>保存 → 验证连接看到用户名即成功</li>
            </ol>
            <a v-if="giteaStatus?.tokenSettingsUrl" :href="giteaStatus.tokenSettingsUrl" target="_blank" rel="noopener" class="hint-link">
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
            <span class="label">当前用户</span>
            <span class="value">{{ giteaUser.username }} ({{ giteaUser.email || '无邮箱' }})</span>
          </div>

          <div v-if="verifyError" class="verify-error">
            <div class="env-row fail" style="padding: 0;">
              <span class="dot"></span>
              <span class="value"><strong>验证失败</strong>：{{ verifyError }}</span>
            </div>
            <details class="verify-help">
              <summary>排查建议</summary>
              <ul>
                <li><strong>HTTP 404</strong> → URL 应填 Gitea 根 (如 <code>http://192.168.3.128:3000</code>)，不要带 <code>/api/v1</code></li>
                <li><strong>HTTP 401/403 / token invalid</strong> → token 错、过期，或权限缺 <code>read:repository</code></li>
                <li><strong>unreachable / connection refused</strong> → URL 写错、Gitea 没启动、防火墙阻挡</li>
                <li><strong>tls / x509</strong> → 自签名证书；改 <code>http://</code> 或在 Gitea 装可信证书</li>
                <li>检查能否在浏览器里打开 <code>{{ gitea.baseURL || '<URL>' }}/api/v1/version</code></li>
              </ul>
            </details>
          </div>
        </div>

        <!-- 仓库扫描 + 导入 -->
        <div v-if="giteaStatus?.hasToken" class="card">
          <div style="display: flex; align-items: center; justify-content: space-between;">
            <h3 style="margin: 0;">📚 你的仓库</h3>
            <button class="btn-outline" @click="scanRepos" :disabled="repoState.loading">
              {{ repoState.loading ? '扫描中…' : (repoState.scanned ? '🔄 重新扫描' : '🔍 扫描仓库') }}
            </button>
          </div>

          <div v-if="repoState.error" class="env-row fail" style="margin-top: 16px;">
            <span class="dot"></span>
            <span class="value">{{ repoState.error }}</span>
          </div>

          <template v-else-if="repoState.scanned">
            <div class="repo-toolbar">
              <input
                v-model="repoState.search"
                type="text"
                class="repo-search"
                placeholder="🔍 搜索仓库名 / owner..."
              />
              <button class="btn-link" @click="toggleAll">
                {{ allSelected ? '取消全选' : `全选 (${filteredRepos.length})` }}
              </button>
              <span class="repo-count">已选 <strong>{{ repoState.selected.size }}</strong> / {{ repoState.repos.length }}</span>
            </div>

            <div class="repo-list" v-if="repoState.repos.length > 0">
              <label
                v-for="r in filteredRepos"
                :key="r.fullName"
                class="repo-item"
                :class="{ selected: repoState.selected.has(r.fullName) }"
              >
                <input
                  type="checkbox"
                  :checked="repoState.selected.has(r.fullName)"
                  @change="toggleRepo(r.fullName)"
                />
                <div class="repo-meta">
                  <div class="repo-name">
                    {{ r.fullName }}
                    <span v-if="r.private" class="repo-tag private">私有</span>
                    <span class="repo-tag branch">{{ r.defaultBranch }}</span>
                  </div>
                  <div v-if="r.description" class="repo-desc">{{ r.description }}</div>
                </div>
              </label>
            </div>

            <div v-else class="env-row" style="padding: 20px; justify-content: center;">
              <span>未找到任何仓库（token 是否有 read:repository 权限？）</span>
            </div>

            <div v-if="repoState.importing" class="env-row" style="padding: 12px;">
              <div class="spinner-sm"></div>
              <span style="margin-left: 8px;">导入中…</span>
            </div>

            <div v-if="repoState.result" class="import-result">
              <div v-if="repoState.result.imported.length" class="env-row ok">
                <span class="dot"></span>
                <span>成功导入 {{ repoState.result.imported.length }} 个：{{ repoState.result.imported.map(p => p.name).join(', ') }}</span>
              </div>
              <div v-if="repoState.result.errors?.length" class="env-row fail" style="margin-top: 6px;">
                <span class="dot"></span>
                <span>失败 {{ repoState.result.errors.length }} 个：</span>
              </div>
              <ul v-if="repoState.result.errors?.length" class="error-list">
                <li v-for="e in repoState.result.errors" :key="e.fullName">
                  <strong>{{ e.fullName }}</strong>: {{ e.error }}
                </li>
              </ul>
            </div>

            <div style="display: flex; gap: 12px; margin-top: 16px; justify-content: flex-end;">
              <button
                class="btn-primary"
                :disabled="repoState.importing || repoState.selected.size === 0"
                @click="confirmImport"
              >
                {{ repoState.importing ? '导入中…' : `导入选中 ${repoState.selected.size} 个` }}
              </button>
              <button class="btn-outline" @click="goToProjects">查看项目列表 →</button>
            </div>
          </template>

          <div v-else class="env-row" style="padding: 20px; justify-content: center;">
            <span>点击"扫描仓库"查看你能访问的所有仓库</span>
          </div>
        </div>
      </section>

      <!-- 占位面板（disabled providers） -->
      <section v-else class="provider-panel">
        <div class="card empty">
          <p>{{ providerPlaceholder }}</p>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, inject, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  GetGiteaStatus, SaveGiteaConfig, VerifyGitea,
  ListGiteaRepos, ImportGiteaRepos,
} from '../wailsjs/go/handler/App'

interface ProviderTab {
  id: 'gitea' | 'github' | 'gitlab'
  name: string
  icon: string
  enabled: boolean
}

const providers: ProviderTab[] = [
  { id: 'gitea',  name: 'Gitea',  icon: '🦊', enabled: true },
  { id: 'github', name: 'GitHub', icon: '🐙', enabled: false },
  { id: 'gitlab', name: 'GitLab', icon: '🦝', enabled: false },
]
const activeProvider = ref<'gitea' | 'github' | 'gitlab'>('gitea')

const providerPlaceholder = computed(() => {
  const p = providers.find(x => x.id === activeProvider.value)
  return p ? `${p.name} 集成正在路上，敬请期待。` : ''
})

const router = useRouter()
const toast = inject('toast') as { success: (m: string) => void; error: (m: string) => void; info?: (m: string) => void }

// ---- Gitea ----

interface GiteaStatus { baseUrl: string; hasToken: boolean; tokenSettingsUrl: string }
interface GiteaUser { username: string; email: string; avatarUrl: string }

const gitea = ref({ baseURL: '', tokenInput: '' })
const giteaStatus = ref<GiteaStatus | null>(null)
const giteaUser = ref<GiteaUser | null>(null)
const savingGitea = ref(false)
const verifying = ref(false)
const verifyError = ref('')

interface GiteaRepo {
  name: string
  fullName: string
  cloneUrl: string
  defaultBranch: string
  description: string
  private: boolean
}
interface ImportError { fullName: string; error: string }
interface Project { id: string; name: string }
interface ImportResult { imported: Project[]; errors: ImportError[] }

const repoState = ref({
  loading: false,
  scanned: false,
  importing: false,
  error: '',
  search: '',
  repos: [] as GiteaRepo[],
  selected: new Set<string>(),
  result: null as ImportResult | null,
})

const filteredRepos = computed(() => {
  const q = repoState.value.search.trim().toLowerCase()
  if (!q) return repoState.value.repos
  return repoState.value.repos.filter(r => r.fullName.toLowerCase().includes(q))
})
const allSelected = computed(() => {
  const list = filteredRepos.value
  return list.length > 0 && list.every(r => repoState.value.selected.has(r.fullName))
})

async function loadGiteaStatus() {
  try {
    giteaStatus.value = await GetGiteaStatus()
    if (giteaStatus.value) gitea.value.baseURL = giteaStatus.value.baseUrl
  } catch (e) {
    console.error(e)
  }
}

async function saveGitea() {
  savingGitea.value = true
  try {
    await SaveGiteaConfig({ baseUrl: gitea.value.baseURL, token: gitea.value.tokenInput })
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
  verifyError.value = ''
  try {
    giteaUser.value = await VerifyGitea()
    toast?.success(`验证成功：${giteaUser.value?.username}`)
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    verifyError.value = msg
    toast?.error(`验证失败: ${msg}`)
  }
  verifying.value = false
}

async function scanRepos() {
  repoState.value.loading = true
  repoState.value.error = ''
  repoState.value.result = null
  repoState.value.selected = new Set()
  try {
    repoState.value.repos = await ListGiteaRepos()
    repoState.value.scanned = true
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    repoState.value.error = msg
  }
  repoState.value.loading = false
}

function toggleRepo(fullName: string) {
  const s = new Set(repoState.value.selected)
  if (s.has(fullName)) s.delete(fullName)
  else s.add(fullName)
  repoState.value.selected = s
}

function toggleAll() {
  const s = new Set(repoState.value.selected)
  if (allSelected.value) {
    filteredRepos.value.forEach(r => s.delete(r.fullName))
  } else {
    filteredRepos.value.forEach(r => s.add(r.fullName))
  }
  repoState.value.selected = s
}

async function confirmImport() {
  repoState.value.importing = true
  repoState.value.result = null
  const picked = repoState.value.repos.filter(r => repoState.value.selected.has(r.fullName))
  const payload = picked.map(r => ({
    fullName: r.fullName,
    cloneUrl: r.cloneUrl,
    branch: r.defaultBranch,
  }))
  try {
    repoState.value.result = await ImportGiteaRepos({ repos: payload })
    if (repoState.value.result?.errors?.length) {
      toast?.error(`导入完成：成功 ${repoState.value.result.imported.length}，失败 ${repoState.value.result.errors.length}`)
    } else {
      toast?.success(`成功导入 ${repoState.value.result?.imported.length} 个项目`)
      repoState.value.selected = new Set()
    }
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast?.error(`导入失败: ${msg}`)
  }
  repoState.value.importing = false
}

function goToProjects() {
  router.push('/projects')
}

onMounted(() => {
  loadGiteaStatus()
})
</script>

<style scoped>
.repos-view { max-width: 1100px; }

h1 { font-size: 28px; margin-bottom: 8px; color: var(--text-primary); }
.subtitle {
  color: var(--text-secondary);
  margin-bottom: 24px;
  font-size: 14px;
}

.layout {
  display: grid;
  grid-template-columns: 240px 1fr;
  gap: 20px;
  align-items: start;
}

/* 左侧 provider tabs */
.provider-tabs {
  display: flex;
  flex-direction: column;
  gap: 8px;
  position: sticky;
  top: 0;
}
.provider-tab {
  display: flex;
  gap: 12px;
  align-items: center;
  padding: 12px 14px;
  background: var(--card-bg);
  border-radius: var(--radius-md);
  border: 2px solid transparent;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}
.provider-tab:hover:not(.disabled) { background: var(--brand-soft); }
.provider-tab.active {
  border-color: var(--brand-start);
  background: var(--brand-soft);
}
.provider-tab.disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
.provider-icon { font-size: 22px; }
.provider-meta { flex: 1; min-width: 0; }
.provider-name {
  font-weight: 600;
  font-size: 14px;
  color: var(--text-primary);
}
.provider-status { margin-top: 2px; }

.badge {
  display: inline-block;
  font-size: 10px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 10px;
}
.badge.inline { margin-left: 6px; font-weight: normal; }
.badge-ok    { background: var(--success-bg); color: var(--success-fg); }
.badge-warn  { background: var(--warning-bg); color: var(--warning-fg); }
.badge-muted { background: var(--bg-primary); color: var(--text-muted); }

/* 右侧 panel */
.provider-panel { display: flex; flex-direction: column; gap: 16px; }
.card {
  background: var(--card-bg);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  box-shadow: var(--shadow-sm);
}
.card h3 {
  font-size: var(--text-lg);
  color: var(--text-primary);
  margin-bottom: 12px;
}
.card.empty {
  text-align: center;
  padding: var(--space-12);
  color: var(--text-muted);
}

.hint {
  color: var(--text-secondary);
  font-size: 12px;
  margin-bottom: 12px;
  line-height: 1.5;
}
.hint-link {
  color: var(--brand-start);
  text-decoration: none;
  font-size: 12px;
}
.hint-link:hover { text-decoration: underline; }

.form-group {
  display: flex; flex-direction: column;
  gap: 6px; margin-bottom: 14px;
}
.form-group label {
  font-size: 13px; font-weight: 500; color: var(--text-primary);
}
.form-group input {
  padding: 10px 12px;
  border: 2px solid var(--border-color);
  border-radius: var(--radius-md);
  font-size: 13px;
  background: var(--card-bg);
  color: var(--text-primary);
  transition: border-color 0.15s;
}
.form-group input:focus { outline: none; border-color: var(--brand-start); }

.gitea-help {
  margin-top: -4px; margin-bottom: 8px;
  padding: 10px 14px;
  background: var(--info-bg);
  border-left: 3px solid var(--info-fg);
  border-radius: 4px;
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.6;
}
.gitea-help summary { cursor: pointer; font-weight: 500; color: var(--info-fg); user-select: none; }
.gitea-help ol { margin: 8px 0 8px 20px; padding: 0; }
.gitea-help li { margin-bottom: 4px; }
.gitea-help code {
  background: rgba(0, 0, 0, 0.06);
  padding: 1px 5px;
  border-radius: 3px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
}

.btn-primary {
  background: linear-gradient(135deg, var(--brand-start), var(--brand-end));
  color: white; border: none;
  padding: 10px 20px;
  border-radius: var(--radius-md);
  font-size: 13px; font-weight: 600;
  cursor: pointer;
  transition: transform 0.15s;
}
.btn-primary:hover:not(:disabled) { transform: translateY(-1px); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-outline {
  background: transparent;
  color: var(--brand-start);
  border: 2px solid var(--brand-start);
  padding: 8px 16px;
  border-radius: var(--radius-md);
  font-size: 13px; font-weight: 600;
  cursor: pointer;
}
.btn-outline:hover:not(:disabled) { background: var(--brand-soft); }
.btn-outline:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-link {
  background: none; border: none;
  color: var(--brand-start);
  font-size: 13px; cursor: pointer;
}
.btn-link:hover { text-decoration: underline; }

/* repo list */
.repo-toolbar {
  display: flex; gap: 12px; align-items: center;
  margin: 14px 0;
}
.repo-search {
  flex: 1;
  padding: 8px 12px;
  border: 2px solid var(--border-color);
  border-radius: var(--radius-md);
  font-size: 13px;
}
.repo-count { font-size: 12px; color: var(--text-secondary); margin-left: auto; }

.repo-list {
  max-height: 460px; overflow-y: auto;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
}
.repo-item {
  display: flex; align-items: flex-start;
  gap: 12px; padding: 10px 14px;
  border-bottom: 1px solid var(--border-color);
  cursor: pointer;
  transition: background 0.12s;
}
.repo-item:hover { background: var(--bg-primary); }
.repo-item.selected { background: var(--brand-soft); }
.repo-item input[type="checkbox"] { margin-top: 4px; cursor: pointer; }
.repo-meta { flex: 1; min-width: 0; }
.repo-name {
  font-weight: 600; color: var(--text-primary);
  font-size: 13px; display: flex; align-items: center; gap: 6px;
}
.repo-tag {
  font-size: 10px; font-weight: 500;
  padding: 1px 6px; border-radius: 10px;
  font-family: 'JetBrains Mono', monospace;
}
.repo-tag.private { background: var(--warning-bg); color: var(--warning-fg); }
.repo-tag.branch  { background: var(--info-bg);    color: var(--info-fg); }
.repo-desc {
  margin-top: 2px; font-size: 12px;
  color: var(--text-muted);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}

.spinner-sm {
  width: 14px; height: 14px;
  border: 2px solid var(--border-color);
  border-top-color: var(--brand-start);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  display: inline-block;
}
@keyframes spin { to { transform: rotate(360deg); } }

.import-result {
  margin-top: 12px;
  padding: 10px 14px;
  background: var(--bg-primary);
  border-radius: var(--radius-sm);
}
.error-list {
  margin: 6px 0 0 24px;
  padding: 0;
  font-size: 12px;
  color: var(--text-secondary);
}

.verify-error {
  margin-top: 12px;
  padding: 12px 16px;
  background: var(--danger-bg);
  border-left: 3px solid var(--danger-fg);
  border-radius: var(--radius-sm);
}
.verify-help {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-secondary);
}
.verify-help summary {
  cursor: pointer;
  color: var(--info-fg);
  font-weight: 500;
}
.verify-help ul {
  margin: 8px 0 0 18px;
  padding: 0;
  line-height: 1.7;
}
.verify-help code {
  background: rgba(0, 0, 0, 0.06);
  padding: 1px 5px;
  border-radius: 3px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
}

/* env-row 共用 */
.env-row {
  display: flex; align-items: center; gap: 10px;
  padding: 6px 0; font-size: 13px;
}
.env-row .dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.env-row.ok   .dot { background: var(--success-fg); }
.env-row.fail .dot { background: var(--danger-fg); }
.env-row.warn .dot { background: var(--warning-fg); }
.env-row .label { min-width: 70px; color: var(--text-secondary); }
.env-row .value { color: var(--text-primary); flex: 1; word-break: break-word; }
</style>
