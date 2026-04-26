<template>
  <div class="repos-view">
    <!-- 顶部：标题 + 横向 provider tab -->
    <header class="page-head">
      <div>
        <h1>仓库源</h1>
        <p class="subtitle">配置 Git 平台凭证 → 自动扫描全部仓库 → 勾选导入为 FlowCI 项目</p>
      </div>
      <div class="provider-tabs">
        <button
          v-for="p in providers"
          :key="p.id"
          class="ptab"
          :class="{ active: activeProvider === p.id, disabled: !p.enabled }"
          :disabled="!p.enabled"
          @click="activeProvider = p.id"
        >
          <span class="ptab-icon">{{ p.icon }}</span>
          <span class="ptab-name">{{ p.name }}</span>
          <span v-if="!p.enabled" class="ptab-tag soon">即将支持</span>
          <span v-else-if="p.id === 'gitea' && giteaStatus?.hasToken" class="ptab-tag ok">✓</span>
        </button>
      </div>
    </header>

    <!-- Gitea 主面板 -->
    <main v-if="activeProvider === 'gitea'" class="provider-main">
      <!-- 配置卡：未配置 → 展开表单；已配置 → 折叠为一行 summary -->
      <section class="config-card" :class="{ compact: configCollapsed }">
        <!-- 折叠态摘要 -->
        <div v-if="configCollapsed" class="config-summary">
          <div class="config-line">
            <span class="provider-emoji">🦊</span>
            <span class="config-url mono">{{ gitea.baseURL || '(未设置)' }}</span>
            <span v-if="giteaUser?.username" class="config-user">
              <span class="user-dot"></span>
              {{ giteaUser.username }}
            </span>
            <span v-else class="config-user warn">
              <span class="user-dot warn"></span>
              已配置 token (未拉取用户名)
            </span>
          </div>
          <button class="btn-ghost" @click="configCollapsed = false">⚙ 修改配置</button>
        </div>

        <!-- 展开态：完整配置表单 -->
        <div v-else class="config-form">
          <div class="config-form-head">
            <h3>🦊 Gitea 配置</h3>
            <button v-if="giteaStatus?.hasToken" class="btn-ghost" @click="configCollapsed = true">收起</button>
          </div>

          <div class="form-row">
            <label>Gitea 实例 URL</label>
            <input
              v-model="gitea.baseURL"
              type="text"
              placeholder="https://gitea.example.com 或 http://localhost:3000"
            />
          </div>

          <div class="form-row">
            <label>
              Access Token
              <span v-if="giteaStatus?.hasToken" class="ai-key-status">✓ 已配置</span>
              <span v-else class="ai-key-status missing">未配置</span>
            </label>
            <input
              v-model="gitea.tokenInput"
              type="password"
              :placeholder="giteaStatus?.hasToken ? '已保存（留空保留旧值；填新值覆盖）' : '从 Gitea 设置 → 应用 生成新令牌'"
            />
          </div>

          <details class="token-help" :open="!giteaStatus?.hasToken">
            <summary>📘 如何生成 Token？</summary>
            <ol>
              <li>右上角头像 → <strong>设置</strong> → 左侧 <strong>应用</strong></li>
              <li>"管理访问令牌" → 名称 <code>FlowCI</code></li>
              <li>勾选 scope：<code>repository</code>=Read（必选）+ <code>user</code>=Read（可选）</li>
              <li><strong>生成</strong> → 复制 token（只显示一次！）粘贴到上方</li>
            </ol>
            <a v-if="giteaStatus?.tokenSettingsUrl" :href="giteaStatus.tokenSettingsUrl" target="_blank" rel="noopener" class="hint-link">
              🔗 直接打开 Token 设置页 →
            </a>
          </details>

          <div class="form-actions">
            <button class="btn-primary" @click="saveGitea" :disabled="savingGitea">
              {{ savingGitea ? '保存中…' : '保存配置' }}
            </button>
            <button class="btn-outline" @click="verifyGitea" :disabled="verifying || !giteaStatus?.hasToken">
              {{ verifying ? '验证中…' : '验证连接' }}
            </button>
          </div>

          <!-- 验证错误 inline alert -->
          <div v-if="verifyError" class="alert error">
            <div class="alert-head">❌ 验证失败：{{ verifyError }}</div>
            <details class="alert-help">
              <summary>排查建议</summary>
              <ul>
                <li><strong>HTTP 404</strong> → URL 应填 Gitea 根，不要带 <code>/api/v1</code></li>
                <li><strong>401/403</strong> → token 错、过期，或缺 <code>repository=Read</code> scope</li>
                <li><strong>connection refused</strong> → URL 写错 / Gitea 未启动 / 防火墙阻挡</li>
                <li><strong>tls / x509</strong> → 自签名证书，改 <code>http://</code> 或装可信证书</li>
                <li>浏览器打开 <code>{{ gitea.baseURL || '<URL>' }}/api/v1/version</code> 自查</li>
              </ul>
            </details>
          </div>
        </div>
      </section>

      <!-- 仓库列表区 -->
      <section v-if="giteaStatus?.hasToken" class="repos-pane">
        <!-- 工具栏 -->
        <div class="repos-toolbar">
          <input
            v-model="repoState.search"
            type="text"
            class="search-input"
            placeholder="🔍 搜索仓库名 / owner..."
          />
          <button class="btn-ghost" @click="scanRepos" :disabled="repoState.loading" title="重新扫描">
            <span :class="{ spinning: repoState.loading }">🔄</span>
            {{ repoState.loading ? '扫描中' : '刷新' }}
          </button>
          <span class="toolbar-stat">
            <strong>{{ repoState.repos.length }}</strong> 个仓库 · 来自
            <strong>{{ groupedRepos.length }}</strong> 个组织
          </span>
        </div>

        <!-- 错误 -->
        <div v-if="repoState.error" class="alert error">
          <div class="alert-head">⚠ 扫描失败：{{ repoState.error }}</div>
        </div>

        <!-- 列表 -->
        <div v-else-if="repoState.loading && !repoState.scanned" class="repos-skeleton">
          <div v-for="i in 6" :key="i" class="skel-item">
            <div class="skel checkbox-skel"></div>
            <div class="skel name-skel"></div>
            <div class="skel meta-skel"></div>
          </div>
        </div>

        <div v-else-if="repoState.scanned && repoState.repos.length === 0" class="empty">
          <p>未找到任何仓库 — 检查 token 是否有 <code>repository=Read</code> 权限。</p>
        </div>

        <div v-else-if="repoState.scanned" class="repos-tree">
          <div v-for="group in groupedRepos" :key="group.owner" class="org-block">
            <div class="org-row" @click="toggleGroup(group.owner)">
              <span class="org-arrow" :class="{ collapsed: !isExpanded(group.owner) }">▾</span>
              <span class="org-name">{{ group.owner }}</span>
              <span class="org-count">{{ group.repos.length }}</span>
              <span v-if="group.selectedCount > 0" class="org-selected">
                选中 {{ group.selectedCount }}
              </span>
              <button class="btn-link org-action" @click.stop="toggleGroupAll(group)">
                {{ group.allSelected ? '取消全选' : '全选' }}
              </button>
            </div>

            <ul v-if="isExpanded(group.owner)" class="repo-rows">
              <li
                v-for="r in group.repos"
                :key="r.fullName"
                class="repo-row"
                :class="{ selected: repoState.selected.has(r.fullName) }"
                @click="toggleRepo(r.fullName)"
              >
                <input
                  type="checkbox"
                  :checked="repoState.selected.has(r.fullName)"
                  @click.stop
                  @change="toggleRepo(r.fullName)"
                />
                <span class="repo-name">{{ repoShortName(r.fullName) }}</span>
                <span class="repo-branch mono">{{ r.defaultBranch }}</span>
                <span v-if="r.private" class="repo-priv">🔒</span>
                <span class="repo-desc">{{ r.description || '—' }}</span>
              </li>
            </ul>
          </div>
        </div>

        <!-- 导入结果（覆盖在底部 above sticky bar） -->
        <div v-if="repoState.result" class="alert" :class="repoState.result.errors?.length ? 'mixed' : 'success'">
          <div class="alert-head">
            <template v-if="repoState.result.imported.length">
              ✓ 成功导入 {{ repoState.result.imported.length }} 个
            </template>
            <template v-if="repoState.result.errors?.length">
              · ✗ 失败 {{ repoState.result.errors.length }} 个
            </template>
          </div>
          <ul v-if="repoState.result.errors?.length" class="alert-list">
            <li v-for="e in repoState.result.errors" :key="e.fullName">
              <code>{{ e.fullName }}</code>: {{ e.error }}
            </li>
          </ul>
          <button class="btn-link" @click="repoState.result = null">关闭</button>
        </div>
      </section>

      <!-- Sticky 底部操作栏 -->
      <div v-if="giteaStatus?.hasToken && (repoState.selected.size > 0 || repoState.importing)" class="sticky-bar">
        <div class="bar-info">
          <strong>{{ repoState.selected.size }}</strong> 个仓库已选
        </div>
        <div class="bar-actions">
          <button class="btn-ghost" @click="clearSelection" :disabled="repoState.importing">清除</button>
          <button class="btn-primary" :disabled="repoState.importing || repoState.selected.size === 0" @click="confirmImport">
            {{ repoState.importing ? '导入中…' : `📥 导入 ${repoState.selected.size} 个` }}
          </button>
          <button class="btn-outline" @click="goToProjects">项目列表 →</button>
        </div>
      </div>
    </main>

    <!-- 占位 panel（disabled providers） -->
    <main v-else class="provider-main">
      <div class="empty-card">
        {{ providerPlaceholder }}
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, inject, computed, onMounted, watch } from 'vue'
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
const toast = inject('toast') as { success: (m: string) => void; error: (m: string) => void; info?: (m: string) => void; warning?: (m: string) => void }

// ---- Gitea ----

interface GiteaStatus { baseUrl: string; hasToken: boolean; tokenSettingsUrl: string }
interface GiteaUser { username: string; email: string; avatarUrl: string }

const gitea = ref({ baseURL: '', tokenInput: '' })
const giteaStatus = ref<GiteaStatus | null>(null)
const giteaUser = ref<GiteaUser | null>(null)
const savingGitea = ref(false)
const verifying = ref(false)
const verifyError = ref('')
const configCollapsed = ref(false)  // 已配置 + 已 verify → 自动折叠

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

// 按 owner 分组
interface RepoGroup {
  owner: string
  repos: GiteaRepo[]
  selectedCount: number
  allSelected: boolean
}
const groupedRepos = computed<RepoGroup[]>(() => {
  const map = new Map<string, GiteaRepo[]>()
  for (const r of filteredRepos.value) {
    const owner = (r.fullName.split('/')[0] || 'unknown').toLowerCase()
    const arr = map.get(owner) ?? []
    arr.push(r)
    map.set(owner, arr)
  }
  return Array.from(map.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([owner, repos]) => {
      const selectedCount = repos.filter(r => repoState.value.selected.has(r.fullName)).length
      return { owner, repos, selectedCount, allSelected: selectedCount === repos.length }
    })
})

const collapsedOwners = ref(new Set<string>())
function isExpanded(owner: string): boolean {
  if (repoState.value.search.trim() !== '') return true
  return !collapsedOwners.value.has(owner)
}
function toggleGroup(owner: string) {
  if (repoState.value.search.trim() !== '') return
  const s = new Set(collapsedOwners.value)
  if (s.has(owner)) s.delete(owner)
  else s.add(owner)
  collapsedOwners.value = s
}
function toggleGroupAll(group: RepoGroup) {
  const s = new Set(repoState.value.selected)
  if (group.allSelected) group.repos.forEach(r => s.delete(r.fullName))
  else group.repos.forEach(r => s.add(r.fullName))
  repoState.value.selected = s
}
function toggleRepo(fullName: string) {
  const s = new Set(repoState.value.selected)
  if (s.has(fullName)) s.delete(fullName)
  else s.add(fullName)
  repoState.value.selected = s
}
function clearSelection() {
  repoState.value.selected = new Set()
}

function repoShortName(fullName: string): string {
  const i = fullName.indexOf('/')
  return i >= 0 ? fullName.slice(i + 1) : fullName
}

// 加载 + 自动扫描 + 自动折叠配置
async function loadGiteaStatus() {
  try {
    giteaStatus.value = await GetGiteaStatus()
    if (giteaStatus.value) gitea.value.baseURL = giteaStatus.value.baseUrl
    // 已配置 → 默认折叠配置区
    if (giteaStatus.value?.hasToken) configCollapsed.value = true
  } catch (e) {
    console.error(e)
  }
}

async function autoScanIfConfigured() {
  if (giteaStatus.value?.hasToken && !repoState.value.scanned && !repoState.value.loading) {
    await scanRepos()
  }
}

async function saveGitea() {
  savingGitea.value = true
  const tokenWasUpdated = gitea.value.tokenInput !== ''
  try {
    await SaveGiteaConfig({ baseUrl: gitea.value.baseURL, token: gitea.value.tokenInput })
    gitea.value.tokenInput = ''
    await loadGiteaStatus()
    toast?.success('Gitea 配置已保存')
    if (tokenWasUpdated || gitea.value.baseURL) {
      repoState.value.scanned = false
      repoState.value.repos = []
      repoState.value.selected = new Set()
      repoState.value.result = null
      if (giteaStatus.value?.hasToken) {
        configCollapsed.value = true
        await scanRepos()
      }
    }
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
    toast?.success(`验证成功${giteaUser.value?.username ? `：${giteaUser.value.username}` : ''}`)
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
    // 顺便补 username（成功扫描说明 token OK）
    if (!giteaUser.value) {
      try { giteaUser.value = await VerifyGitea() } catch { /* ignore */ }
    }
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    repoState.value.error = msg
  }
  repoState.value.loading = false
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
      toast?.warning?.(`导入完成：成功 ${repoState.value.result.imported.length}，失败 ${repoState.value.result.errors.length}`)
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

watch(verifyError, (v) => { if (v) configCollapsed.value = false })

onMounted(async () => {
  await loadGiteaStatus()
  await autoScanIfConfigured()
})
</script>

<style scoped>
.repos-view {
  max-width: 1400px;
  margin: 0 auto;
  padding-bottom: 80px; /* 给 sticky bar 留位 */
}

/* === 顶部 head === */
.page-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 20px;
}
.page-head h1 {
  font-size: 26px;
  margin: 0 0 4px;
  color: var(--text-primary);
}
.subtitle {
  color: var(--text-secondary);
  margin: 0;
  font-size: 13px;
}

.provider-tabs {
  display: flex;
  gap: 6px;
  background: var(--bg-primary);
  padding: 4px;
  border-radius: var(--radius-md);
}
.ptab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}
.ptab:hover:not(.disabled):not(.active) {
  background: var(--card-bg);
  color: var(--text-primary);
}
.ptab.active {
  background: var(--card-bg);
  color: var(--text-primary);
  font-weight: 600;
  box-shadow: var(--shadow-sm);
}
.ptab.disabled { opacity: 0.5; cursor: not-allowed; }
.ptab-icon { font-size: 16px; }
.ptab-tag {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 8px;
  font-weight: 500;
}
.ptab-tag.soon { background: var(--bg-secondary); color: var(--text-muted); }
.ptab-tag.ok   { background: var(--success-bg); color: var(--success-fg); }

/* === 配置卡 === */
.config-card {
  background: var(--card-bg);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
  margin-bottom: 16px;
  transition: padding 0.2s;
}
.config-card.compact { padding: 0; }

/* 折叠态 summary */
.config-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  gap: 16px;
}
.config-line {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}
.provider-emoji { font-size: 20px; }
.config-url {
  color: var(--text-primary);
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.config-user {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--success-fg);
  background: var(--success-bg);
  padding: 4px 10px;
  border-radius: 12px;
}
.config-user.warn {
  color: var(--warning-fg);
  background: var(--warning-bg);
}
.user-dot {
  width: 6px; height: 6px;
  border-radius: 50%;
  background: var(--success-fg);
}
.user-dot.warn { background: var(--warning-fg); }

/* 展开态表单 */
.config-form { padding: 20px; }
.config-form-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}
.config-form-head h3 {
  font-size: 16px;
  margin: 0;
  color: var(--text-primary);
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 12px;
}
.form-row label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 8px;
}
.form-row input {
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: 13px;
  background: var(--card-bg);
  color: var(--text-primary);
  transition: border-color 0.15s;
}
.form-row input:focus {
  outline: none;
  border-color: var(--brand-start);
}
.ai-key-status {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 8px;
  background: var(--success-bg);
  color: var(--success-fg);
  font-weight: 500;
}
.ai-key-status.missing {
  background: var(--bg-secondary);
  color: var(--text-muted);
}

.token-help {
  margin: 4px 0 14px;
  padding: 10px 14px;
  background: var(--info-bg);
  border-left: 3px solid var(--info-fg);
  border-radius: 4px;
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.6;
}
.token-help summary {
  cursor: pointer;
  font-weight: 500;
  color: var(--info-fg);
  user-select: none;
}
.token-help ol { margin: 8px 0 4px 20px; padding: 0; }
.token-help li { margin-bottom: 3px; }
.token-help code {
  background: rgba(0, 0, 0, 0.06);
  padding: 1px 5px;
  border-radius: 3px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
}
.hint-link {
  color: var(--brand-start);
  text-decoration: none;
  font-size: 12px;
}
.hint-link:hover { text-decoration: underline; }

.form-actions {
  display: flex;
  gap: 10px;
}

/* === 按钮统一 === */
.btn-primary {
  background: linear-gradient(135deg, var(--brand-start), var(--brand-end));
  color: #fff;
  border: none;
  padding: 8px 18px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.1s, box-shadow 0.15s;
}
.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 10px var(--brand-soft);
}
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-outline {
  background: transparent;
  color: var(--brand-start);
  border: 1px solid var(--brand-start);
  padding: 7px 16px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}
.btn-outline:hover:not(:disabled) { background: var(--brand-soft); }
.btn-outline:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-ghost {
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 5px;
}
.btn-ghost:hover:not(:disabled) {
  border-color: var(--brand-start);
  color: var(--brand-start);
}
.btn-ghost:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-link {
  background: none; border: none;
  color: var(--brand-start);
  font-size: 12px;
  cursor: pointer;
  padding: 0 4px;
}
.btn-link:hover { text-decoration: underline; }

.spinning { animation: rot 1s linear infinite; display: inline-block; }
@keyframes rot { to { transform: rotate(360deg); } }

/* === 仓库列表 === */
.repos-pane {
  background: var(--card-bg);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
  overflow: hidden;
}
.repos-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 20px;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-secondary);
}
.search-input {
  flex: 1;
  max-width: 360px;
  padding: 7px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: 13px;
  background: var(--card-bg);
  color: var(--text-primary);
}
.search-input:focus { outline: none; border-color: var(--brand-start); }
.toolbar-stat {
  margin-left: auto;
  font-size: 12px;
  color: var(--text-secondary);
}

.repos-tree {
  max-height: 60vh;
  overflow-y: auto;
}

.org-block {
  border-bottom: 1px solid var(--border-color);
}
.org-block:last-child { border-bottom: none; }

.org-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 20px;
  background: var(--bg-secondary);
  cursor: pointer;
  user-select: none;
  font-size: 13px;
  position: sticky;
  top: 0;
  z-index: 1;
}
.org-row:hover { background: var(--brand-soft); }
.org-arrow {
  font-size: 11px;
  color: var(--text-secondary);
  transition: transform 0.15s;
  width: 12px;
  display: inline-block;
}
.org-arrow.collapsed { transform: rotate(-90deg); }
.org-name {
  font-weight: 600;
  color: var(--text-primary);
}
.org-count {
  font-size: 11px;
  color: var(--text-muted);
  font-family: 'JetBrains Mono', monospace;
  background: var(--card-bg);
  padding: 1px 7px;
  border-radius: 8px;
}
.org-selected {
  font-size: 11px;
  color: var(--brand-start);
  background: var(--brand-soft);
  padding: 2px 8px;
  border-radius: 8px;
  font-weight: 500;
}
.org-action { margin-left: auto; }

.repo-rows {
  list-style: none;
  margin: 0;
  padding: 0;
}
.repo-row {
  display: grid;
  grid-template-columns: 24px minmax(120px, 1.4fr) 80px 22px 1fr;
  align-items: center;
  gap: 12px;
  padding: 8px 20px 8px 32px;
  border-top: 1px solid var(--border-color);
  cursor: pointer;
  font-size: 13px;
  transition: background 0.1s;
}
.repo-row:hover { background: var(--bg-primary); }
.repo-row.selected { background: var(--brand-soft); }

.repo-row input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}
.repo-name {
  color: var(--text-primary);
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.repo-branch {
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
  color: var(--info-fg);
  background: var(--info-bg);
  padding: 1px 7px;
  border-radius: 8px;
  text-align: center;
}
.repo-priv { font-size: 12px; }
.repo-desc {
  color: var(--text-muted);
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mono { font-family: 'JetBrains Mono', 'Consolas', monospace; }

/* === skeleton === */
.repos-skeleton { padding: 8px 20px; }
.skel-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px solid var(--border-color);
}
.skel {
  background: linear-gradient(90deg, var(--bg-primary), var(--border-color), var(--bg-primary));
  background-size: 200% 100%;
  animation: shimmer 1.4s linear infinite;
  border-radius: 4px;
}
.checkbox-skel { width: 16px; height: 16px; }
.name-skel { width: 200px; height: 14px; }
.meta-skel { flex: 1; height: 12px; max-width: 300px; }
@keyframes shimmer {
  0%   { background-position: 100% 0; }
  100% { background-position: -100% 0; }
}

/* === alert === */
.alert {
  margin: 14px 20px;
  padding: 12px 16px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  border-left: 3px solid;
}
.alert.error   { background: var(--danger-bg); border-color: var(--danger-fg);  color: var(--danger-fg); }
.alert.success { background: var(--success-bg); border-color: var(--success-fg); color: var(--success-fg); }
.alert.mixed   { background: var(--warning-bg); border-color: var(--warning-fg); color: var(--warning-fg); }
.alert-head {
  font-weight: 500;
  margin-bottom: 4px;
}
.alert-help summary {
  cursor: pointer;
  font-size: 12px;
  margin-top: 6px;
  color: var(--text-secondary);
}
.alert-help ul {
  margin: 6px 0 0 20px;
  padding: 0;
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: normal;
  line-height: 1.7;
}
.alert-help code {
  background: rgba(0, 0, 0, 0.06);
  padding: 1px 5px;
  border-radius: 3px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
}
.alert-list {
  list-style: none;
  margin: 4px 0;
  padding: 0;
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: normal;
}
.alert-list li { margin-top: 4px; }

.empty {
  padding: 40px 20px;
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
}
.empty code {
  background: rgba(0, 0, 0, 0.06);
  padding: 1px 5px;
  border-radius: 3px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
}

.empty-card {
  padding: 60px;
  background: var(--card-bg);
  border-radius: var(--radius-lg);
  text-align: center;
  color: var(--text-muted);
  border: 1px dashed var(--border-color);
}

/* === sticky 底部操作栏 === */
.sticky-bar {
  position: fixed;
  bottom: 0;
  left: 220px; /* 让位侧边栏宽度 */
  right: 0;
  background: var(--card-bg);
  border-top: 1px solid var(--border-color);
  box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.08);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 32px;
  gap: 16px;
  z-index: 100;
  animation: slide-up 0.2s ease-out;
}
@keyframes slide-up {
  from { transform: translateY(100%); }
  to   { transform: translateY(0); }
}
.bar-info {
  font-size: 14px;
  color: var(--text-primary);
}
.bar-info strong {
  color: var(--brand-start);
  font-size: 16px;
}
.bar-actions {
  display: flex;
  gap: 10px;
}
</style>
