<template>
  <div class="projects-view">
    <div class="header">
      <h1>项目列表</h1>
      <div style="display: flex; gap: 12px;">
        <button class="btn-outline" @click="openGiteaImport">🦊 从 Gitea 导入</button>
        <button class="btn-primary" @click="showCreateDialog = true">+ 新建项目</button>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <span>加载中...</span>
    </div>

    <div v-else-if="projects.length === 0" class="empty-state">
      <div class="empty-icon">📦</div>
      <h3>暂无项目</h3>
      <p>点击右上角按钮创建你的第一个项目</p>
    </div>

    <div v-else class="project-grid">
      <div v-for="project in projects" :key="project.id" class="project-card">
        <div class="project-header">
          <div class="project-name">{{ project.name }}</div>
          <span class="lang-badge">{{ getLangName(project.language) }}</span>
        </div>
        <div class="project-path">{{ project.path }}</div>
        <div class="project-meta">
          <span>创建: {{ formatDate(project.createdAt) }}</span>
        </div>
        <div class="project-actions">
          <button class="btn-small" @click="buildProject(project)">构建</button>
          <button class="btn-small btn-secondary" @click="deployProject(project)">部署</button>
          <button class="btn-small btn-outline" @click="showHistory(project)">历史</button>
          <button class="btn-small btn-outline" @click="editProject(project)">编辑</button>
          <button class="btn-small btn-danger" @click="deleteProject(project)">删除</button>
        </div>
      </div>
    </div>

    <div v-if="showCreateDialog" class="modal-overlay" @click.self="showCreateDialog = false">
      <div class="modal">
        <h2>新建项目</h2>
        <div class="form-group">
          <label>项目名称</label>
          <input v-model="newProject.name" type="text" placeholder="my-project" />
        </div>
        <div class="form-group">
          <label>项目路径</label>
          <input v-model="newProject.path" type="text" placeholder="/workspace/my-project" />
        </div>
        <div class="form-group">
          <label>语言/框架</label>
          <select v-model="newProject.language">
            <option value="nodejs">Node.js</option>
            <option value="go">Go</option>
            <option value="python">Python</option>
            <option value="java-maven">Java (Maven)</option>
            <option value="java-gradle">Java (Gradle)</option>
            <option value="php">PHP</option>
            <option value="ruby">Ruby</option>
            <option value="dotnet">.NET</option>
            <option value="rust">Rust</option>
            <option value="c">C/C++</option>
          </select>
        </div>
        <div class="modal-actions">
          <button class="btn-cancel" @click="showCreateDialog = false">取消</button>
          <button class="btn-primary" @click="createProject" :disabled="creating">创建</button>
        </div>
      </div>
    </div>

    <!-- Gitea 导入弹窗 -->
    <div v-if="giteaModal.open" class="modal-overlay" @click.self="closeGiteaImport">
      <div class="modal modal-lg">
        <h2>从 Gitea 导入仓库</h2>

        <div v-if="giteaModal.loading" class="loading" style="padding: 40px 0;">
          <div class="spinner"></div>
          <span>正在拉取你的仓库列表…</span>
        </div>

        <div v-else-if="giteaModal.error" class="env-row fail" style="padding: 16px;">
          <span class="dot"></span>
          <span class="env-value">{{ giteaModal.error }}</span>
        </div>

        <template v-else>
          <div class="repo-toolbar">
            <input
              v-model="giteaModal.search"
              type="text"
              class="repo-search"
              placeholder="🔍 搜索仓库名 / owner..."
            />
            <button class="btn-link" @click="toggleAll">
              {{ allSelected ? '取消全选' : `全选 (${filteredRepos.length})` }}
            </button>
            <span class="repo-count">已选 <strong>{{ giteaModal.selected.size }}</strong> / {{ giteaModal.repos.length }}</span>
          </div>

          <div class="repo-list">
            <label
              v-for="r in filteredRepos"
              :key="r.fullName"
              class="repo-item"
              :class="{ selected: giteaModal.selected.has(r.fullName) }"
            >
              <input
                type="checkbox"
                :checked="giteaModal.selected.has(r.fullName)"
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

          <div v-if="giteaModal.importing" class="env-row" style="padding: 12px;">
            <div class="spinner spinner-sm"></div>
            <span>导入中… 已完成 {{ giteaModal.importedCount }} / {{ giteaModal.selected.size }}</span>
          </div>

          <div v-if="giteaModal.result" class="import-result">
            <div class="env-row ok" v-if="giteaModal.result.imported.length">
              <span class="dot"></span>
              <span>成功导入 {{ giteaModal.result.imported.length }} 个：{{ giteaModal.result.imported.map(p => p.name).join(', ') }}</span>
            </div>
            <div v-if="giteaModal.result.errors?.length" class="env-row fail" style="margin-top: 6px;">
              <span class="dot"></span>
              <span>失败 {{ giteaModal.result.errors.length }} 个：</span>
            </div>
            <ul v-if="giteaModal.result.errors?.length" class="error-list">
              <li v-for="e in giteaModal.result.errors" :key="e.fullName">
                <strong>{{ e.fullName }}</strong>: {{ e.error }}
              </li>
            </ul>
          </div>
        </template>

        <div class="modal-actions">
          <button class="btn-cancel" @click="closeGiteaImport">关闭</button>
          <button
            v-if="!giteaModal.loading && !giteaModal.error"
            class="btn-primary"
            :disabled="giteaModal.importing || giteaModal.selected.size === 0"
            @click="confirmImport"
          >
            {{ giteaModal.importing ? '导入中…' : `导入 ${giteaModal.selected.size} 个` }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="showEditDialog" class="modal-overlay" @click.self="showEditDialog = false">
      <div class="modal">
        <h2>编辑项目</h2>
        <div class="form-group">
          <label>项目名称</label>
          <input v-model="editForm.name" type="text" placeholder="my-project" />
        </div>
        <div class="form-group">
          <label>项目路径</label>
          <input v-model="editForm.path" type="text" placeholder="/workspace/my-project" />
        </div>
        <div class="form-group">
          <label>语言/框架</label>
          <select v-model="editForm.language">
            <option value="nodejs">Node.js</option>
            <option value="go">Go</option>
            <option value="python">Python</option>
            <option value="java-maven">Java (Maven)</option>
            <option value="java-gradle">Java (Gradle)</option>
            <option value="php">PHP</option>
            <option value="ruby">Ruby</option>
            <option value="dotnet">.NET</option>
            <option value="rust">Rust</option>
            <option value="c">C/C++</option>
          </select>
        </div>
        <div class="modal-actions">
          <button class="btn-cancel" @click="showEditDialog = false">取消</button>
          <button class="btn-primary" @click="updateProject" :disabled="updating">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, inject, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ListProjects, CreateProject, DeleteProject, UpdateProject, ListGiteaRepos, ImportGiteaRepos } from '../wailsjs/go/handler/App'
import { computed } from 'vue'
import { useConfirm } from '../composables/useConfirm'

const { ask } = useConfirm()

interface Project {
  id: string
  name: string
  path: string
  language: string
  createdAt: string
}

const router = useRouter()
const loading = ref(false)
const creating = ref(false)
const updating = ref(false)
const projects = ref<Project[]>([])
const showCreateDialog = ref(false)
const showEditDialog = ref(false)
const newProject = ref({ name: '', path: '', language: 'nodejs' })
const editForm = ref({ id: '', name: '', path: '', language: 'nodejs' })

// ---- Gitea 导入 ----
interface GiteaRepo {
  fullName: string
  cloneUrl: string
  defaultBranch: string
  description: string
  private: boolean
}
interface ImportError { fullName: string; error: string }
interface ImportResult {
  imported: Project[]
  errors: ImportError[]
}
const giteaModal = ref({
  open: false,
  loading: false,
  importing: false,
  importedCount: 0,
  error: '',
  search: '',
  repos: [] as GiteaRepo[],
  selected: new Set<string>(),
  result: null as ImportResult | null,
})

const filteredRepos = computed(() => {
  const q = giteaModal.value.search.trim().toLowerCase()
  if (!q) return giteaModal.value.repos
  return giteaModal.value.repos.filter(r => r.fullName.toLowerCase().includes(q))
})
const allSelected = computed(() => {
  const list = filteredRepos.value
  return list.length > 0 && list.every(r => giteaModal.value.selected.has(r.fullName))
})

async function openGiteaImport() {
  giteaModal.value = {
    open: true, loading: true, importing: false, importedCount: 0,
    error: '', search: '', repos: [], selected: new Set(), result: null,
  }
  try {
    giteaModal.value.repos = await ListGiteaRepos()
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    giteaModal.value.error = msg.includes('设置') ? msg : `拉取仓库失败: ${msg}`
  }
  giteaModal.value.loading = false
}

function closeGiteaImport() {
  if (giteaModal.value.importing) return
  giteaModal.value.open = false
  if (giteaModal.value.result?.imported.length) {
    void refreshProjects()
  }
}

function toggleRepo(fullName: string) {
  const s = new Set(giteaModal.value.selected)
  if (s.has(fullName)) s.delete(fullName)
  else s.add(fullName)
  giteaModal.value.selected = s
}

function toggleAll() {
  if (allSelected.value) {
    const s = new Set(giteaModal.value.selected)
    filteredRepos.value.forEach(r => s.delete(r.fullName))
    giteaModal.value.selected = s
  } else {
    const s = new Set(giteaModal.value.selected)
    filteredRepos.value.forEach(r => s.add(r.fullName))
    giteaModal.value.selected = s
  }
}

async function confirmImport() {
  giteaModal.value.importing = true
  giteaModal.value.importedCount = 0
  giteaModal.value.result = null

  const picked = giteaModal.value.repos.filter(r => giteaModal.value.selected.has(r.fullName))
  const payload = picked.map(r => ({
    fullName: r.fullName,
    cloneUrl: r.cloneUrl,
    branch: r.defaultBranch,
  }))
  try {
    const result = await ImportGiteaRepos({ repos: payload })
    giteaModal.value.result = result
    giteaModal.value.importedCount = result.imported?.length ?? 0
    if (result.errors?.length) {
      toast?.error(`导入完成：成功 ${result.imported.length}，失败 ${result.errors.length}`)
    } else {
      toast?.success(`成功导入 ${result.imported.length} 个项目`)
    }
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast?.error(`导入失败: ${msg}`)
  }
  giteaModal.value.importing = false
}

const toast = inject('toast') as { success: (msg: string) => void; error: (msg: string) => void; info: (msg: string) => void }

const langMap: Record<string, string> = {
  nodejs: 'Node.js',
  go: 'Go',
  python: 'Python',
  'java-maven': 'Java (Maven)',
  'java-gradle': 'Java (Gradle)',
  php: 'PHP',
  ruby: 'Ruby',
  dotnet: '.NET',
  rust: 'Rust',
  c: 'C/C++'
}

function getLangName(lang: string) {
  return langMap[lang] || lang
}

function formatDate(date: string) {
  return new Date(date).toLocaleDateString('zh-CN')
}

async function refreshProjects() {
  loading.value = true
  try {
    projects.value = await ListProjects()
  } catch (e) {
    console.error('Failed to load projects:', e)
    toast?.error('加载项目列表失败')
  }
  loading.value = false
}

async function createProject() {
  if (!newProject.value.name || !newProject.value.path) {
    toast?.error('请填写项目名称和路径')
    return
  }
  creating.value = true
  try {
    await CreateProject({
      name: newProject.value.name,
      path: newProject.value.path,
      language: newProject.value.language
    })
    toast?.success('项目创建成功')
    showCreateDialog.value = false
    newProject.value = { name: '', path: '', language: 'nodejs' }
    await refreshProjects()
  } catch (e) {
    toast?.error(`创建失败: ${e}`)
  }
  creating.value = false
}

async function deleteProject(project: Project) {
  const ok = await ask({
    title: '删除项目',
    message: `确定要删除项目 "${project.name}" 吗？此操作不可撤销，相关流水线与构建记录将一并级联删除。`,
    variant: 'danger',
    confirmText: '删除',
  })
  if (!ok) return
  try {
    await DeleteProject(project.id)
    toast?.success('项目已删除')
    await refreshProjects()
  } catch (e) {
    toast?.error(`删除失败: ${e}`)
  }
}

function showHistory(project: Project) {
  router.push({ path: '/build-history', query: { projectId: project.id } })
}

function editProject(project: Project) {
  editForm.value = {
    id: project.id,
    name: project.name,
    path: project.path,
    language: project.language
  }
  showEditDialog.value = true
}

async function updateProject() {
  if (!editForm.value.name || !editForm.value.path) {
    toast?.error('请填写项目名称和路径')
    return
  }
  updating.value = true
  try {
    await UpdateProject({
      id: editForm.value.id,
      name: editForm.value.name,
      path: editForm.value.path,
      language: editForm.value.language
    })
    toast?.success('项目已更新')
    showEditDialog.value = false
    await refreshProjects()
  } catch (e) {
    toast?.error(`更新失败: ${e}`)
  }
  updating.value = false
}

function buildProject(project: Project) {
  router.push({ path: '/build', query: { projectId: project.id } })
}

function deployProject(project: Project) {
  router.push({ path: '/deploy', query: { projectId: project.id } })
}

onMounted(() => {
  refreshProjects()
})
</script>

<style scoped>
.projects-view {
  max-width: 1200px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header h1 {
  font-size: 28px;
  color: var(--text-primary, #1a1a2e);
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
  transition: transform 0.2s;
}

.btn-primary:hover {
  transform: translateY(-2px);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

.loading, .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  gap: 16px;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid var(--border-color, #e0e0e0);
  border-top-color: #667eea;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-icon {
  font-size: 64px;
}

.empty-state h3 {
  font-size: 20px;
  color: var(--text-secondary, #666);
}

.empty-state p {
  color: var(--text-secondary, #999);
}

.project-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.project-card {
  background: var(--card-bg, white);
  border-radius: 12px;
  padding: 20px;
  box-shadow: var(--shadow, 0 2px 12px rgba(0, 0, 0, 0.05));
  transition: all 0.2s;
}

.project-card:hover {
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}

.project-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.project-name {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary, #1a1a2e);
}

.lang-badge {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
}

.project-path {
  color: var(--text-secondary, #666);
  font-size: 13px;
  margin-bottom: 12px;
}

.project-meta {
  color: var(--text-secondary, #999);
  font-size: 12px;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-color, #eee);
}

.project-actions {
  display: flex;
  gap: 8px;
}

.btn-small {
  flex: 1;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
}

.btn-secondary {
  background: var(--bg-primary, #f0f0f0);
  color: var(--text-primary, #333);
}

.btn-danger {
  background: #fee2e2;
  color: #dc2626;
}

.btn-danger:hover {
  background: #fecaca;
}

.btn-outline {
  background: var(--card-bg, white);
  color: #667eea;
  border: 1.5px solid #667eea;
}

.btn-outline:hover {
  background: rgba(102, 126, 234, 0.05);
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: var(--card-bg, white);
  border-radius: 16px;
  padding: 32px;
  width: 480px;
  max-width: 90vw;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}

.modal h2 {
  font-size: 22px;
  color: var(--text-primary, #1a1a2e);
  margin-bottom: 24px;
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

.form-group input,
.form-group select {
  padding: 12px;
  border: 2px solid var(--border-color, #e0e0e0);
  border-radius: 8px;
  font-size: 14px;
  background: var(--card-bg, white);
  color: var(--text-primary, #333);
  transition: border-color 0.2s;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #667eea;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.btn-cancel {
  background: var(--bg-primary, #f0f0f0);
  color: var(--text-primary, #333);
  border: none;
  padding: 12px 24px;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
}

/* ---- Gitea 导入弹窗 ---- */

.modal-lg {
  width: 720px !important;
  max-width: 92vw;
}

.repo-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 12px 0;
}

.repo-search {
  flex: 1;
  padding: 10px 12px;
  border: 2px solid var(--border-color, #e0e0e0);
  border-radius: 8px;
  font-size: 13px;
}

.repo-count {
  font-size: 12px;
  color: var(--text-secondary, #666);
  margin-left: auto;
}

.btn-link {
  background: none;
  border: none;
  color: var(--brand-start, #667eea);
  cursor: pointer;
  font-size: 13px;
}
.btn-link:hover {
  text-decoration: underline;
}

.repo-list {
  max-height: 50vh;
  overflow-y: auto;
  border: 1px solid var(--border-color, #e0e0e0);
  border-radius: 8px;
}

.repo-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 10px 14px;
  border-bottom: 1px solid var(--border-color, #f0f0f0);
  cursor: pointer;
  transition: background 0.12s;
}
.repo-item:hover {
  background: var(--bg-primary, #f9fafb);
}
.repo-item.selected {
  background: var(--brand-soft, rgba(102, 126, 234, 0.08));
}
.repo-item input[type="checkbox"] {
  margin-top: 4px;
  cursor: pointer;
}

.repo-meta {
  flex: 1;
  min-width: 0;
}

.repo-name {
  font-weight: 600;
  color: var(--text-primary, #1a1a2e);
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.repo-tag {
  font-size: 10px;
  font-weight: 500;
  padding: 1px 6px;
  border-radius: 10px;
  font-family: 'JetBrains Mono', monospace;
}
.repo-tag.private {
  background: var(--warning-bg, #fffbeb);
  color: var(--warning-fg, #92400e);
}
.repo-tag.branch {
  background: var(--info-bg, #eff6ff);
  color: var(--info-fg, #1e40af);
}

.repo-desc {
  margin-top: 2px;
  font-size: 12px;
  color: var(--text-muted, #94a3b8);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.spinner-sm {
  width: 16px;
  height: 16px;
  border: 2px solid var(--border-color);
  border-top-color: var(--brand-start);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  display: inline-block;
}

.import-result {
  margin-top: 12px;
  padding: 10px 14px;
  background: var(--bg-primary, #f5f7fa);
  border-radius: 6px;
}
.error-list {
  margin: 6px 0 0 24px;
  padding: 0;
  font-size: 12px;
  color: var(--text-secondary, #666);
}
.error-list li {
  margin-bottom: 3px;
}
.env-row {
  display: flex; align-items: center; gap: 8px; padding: 4px 0; font-size: 13px;
}
.env-row .dot {
  width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0;
}
.env-row.ok   .dot { background: #16a34a; }
.env-row.fail .dot { background: #dc2626; }
.env-row.warn .dot { background: #d97706; }
.env-value { flex: 1; word-break: break-word; }
</style>
