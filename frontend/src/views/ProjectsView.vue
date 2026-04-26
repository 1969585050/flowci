<template>
  <div class="projects-view">
    <div class="header">
      <h1>项目列表</h1>
      <div style="display: flex; gap: 12px;">
        <button class="btn-outline" @click="goToRepositories">🌿 从仓库源导入</button>
        <button class="btn-primary" @click="showCreateDialog = true">+ 新建项目</button>
      </div>
    </div>

    <div v-if="loading && stats.length === 0" class="project-grid">
      <div v-for="i in 4" :key="i" class="project-card skeleton-card">
        <div class="skel-row" style="width: 60%; height: 18px;"></div>
        <div class="skel-row" style="width: 40%; height: 12px; margin-top: 8px;"></div>
        <div class="skel-row" style="width: 100%; height: 36px; margin-top: 14px; border-radius: 6px;"></div>
        <div style="display: flex; gap: 6px; margin-top: 14px;">
          <div class="skel-row" style="width: 70px; height: 24px;"></div>
          <div class="skel-row" style="width: 70px; height: 24px;"></div>
        </div>
      </div>
    </div>

    <div v-else-if="projects.length === 0" class="empty-state">
      <div class="empty-icon">📦</div>
      <h3>暂无项目</h3>
      <p>点击右上角按钮创建你的第一个项目</p>
    </div>

    <div v-else class="project-grid">
      <div v-for="s in stats" :key="s.project.id" class="project-card" :class="{ pinned: s.project.pinnedAt }">
        <span v-if="s.project.pinnedAt" class="pin-mark" title="已置顶">📌</span>
        <!-- 标题行：项目名 + 语言 + 来源标签 -->
        <div class="project-header">
          <div class="project-title">
            <h3 class="project-name">{{ s.project.name }}</h3>
            <div class="title-tags">
              <span class="lang-badge">{{ getLangName(s.project.language) }}</span>
              <span v-if="s.project.repoUrl" class="src-badge git" title="Git 仓库项目">🌿 Git</span>
              <span v-else class="src-badge local" title="本地路径项目">📁 本地</span>
            </div>
          </div>
          <button class="btn-icon-menu" @click="toggleMenu(s.project.id)" title="更多操作">⋯</button>
        </div>

        <!-- 路径 / Git 信息 -->
        <div class="project-info">
          <div class="info-row" :title="s.project.path">
            <span class="info-icon">📂</span>
            <span class="info-text mono">{{ s.project.path || '(未配置路径)' }}</span>
          </div>
          <div v-if="s.project.repoBranch" class="info-row">
            <span class="info-icon">🌿</span>
            <span class="info-text">{{ s.project.repoBranch }}</span>
            <span v-if="s.headCommit" class="commit-tag" :title="s.headSubject">
              {{ s.headCommit }}
            </span>
          </div>
          <div v-if="s.headSubject" class="info-row" :title="s.headSubject">
            <span class="info-icon">💬</span>
            <span class="info-text commit-subject">{{ s.headSubject }}</span>
          </div>
        </div>

        <!-- 构建状态条 -->
        <div class="build-strip">
          <template v-if="s.lastBuild">
            <span class="status-pill" :class="s.lastBuild.status">
              {{ statusIcon(s.lastBuild.status) }} {{ statusText(s.lastBuild.status) }}
            </span>
            <span class="build-meta">
              <span class="mono">{{ s.lastBuild.imageName }}:{{ s.lastBuild.imageTag }}</span>
              <span class="dot-sep">·</span>
              <span :title="formatDate(s.lastBuild.startedAt)">{{ relativeTime(s.lastBuild.startedAt) }}</span>
              <template v-if="s.lastBuild.finishedAt">
                <span class="dot-sep">·</span>
                <span>{{ calcDuration(s.lastBuild.startedAt, s.lastBuild.finishedAt) }}</span>
              </template>
            </span>
          </template>
          <template v-else>
            <span class="status-pill never">⏳ 从未构建</span>
          </template>
          <span class="build-count" v-if="s.buildCount > 0">{{ s.buildCount }} 次构建</span>
        </div>

        <!-- 主操作 + 折叠菜单 -->
        <div class="project-actions">
          <button class="btn-primary-sm" @click="buildProject(s.project)">▶ 构建</button>
          <button class="btn-outline-sm" @click="showHistory(s.project)">📊 历史</button>
          <transition name="menu-fade">
            <div v-if="openMenu === s.project.id" class="more-menu" @click.stop>
              <button v-if="s.project.pinnedAt" @click="togglePin(s.project, false); closeMenu()">📍 取消置顶</button>
              <button v-else @click="togglePin(s.project, true); closeMenu()">📌 置顶</button>
              <button @click="deployProject(s.project); closeMenu()">🌐 部署</button>
              <button @click="editProject(s.project); closeMenu()">✏️ 编辑</button>
              <button class="danger" @click="deleteProject(s.project); closeMenu()">🗑 删除</button>
            </div>
          </transition>
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
import { ListProjectsWithStats, CreateProject, DeleteProject, UpdateProject, PinProject, UnpinProject } from '../wailsjs/go/handler/App'
import { useConfirm } from '../composables/useConfirm'

const { ask } = useConfirm()

interface Project {
  id: string
  name: string
  path: string
  language: string
  createdAt: string
  repoUrl?: string
  repoBranch?: string
  pinnedAt?: string | null
}

interface BuildSummary {
  id: string
  projectId: string
  imageName: string
  imageTag: string
  status: string
  startedAt: string
  finishedAt?: string
}

interface ProjectStats {
  project: Project
  lastBuild?: BuildSummary
  buildCount: number
  headCommit?: string
  headSubject?: string
}

const router = useRouter()
const loading = ref(false)
const creating = ref(false)
const updating = ref(false)
const stats = ref<ProjectStats[]>([])
const openMenu = ref<string | null>(null)

function toggleMenu(id: string) {
  openMenu.value = openMenu.value === id ? null : id
}
function closeMenu() {
  openMenu.value = null
}

async function togglePin(project: Project, pin: boolean) {
  try {
    if (pin) await PinProject(project.id)
    else     await UnpinProject(project.id)
    toast?.success(pin ? `${project.name} 已置顶` : `${project.name} 已取消置顶`)
    await refreshProjects()
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast?.error(`${pin ? '置顶' : '取消置顶'}失败: ${msg}`)
  }
}

// 派生 projects 列表给原有 deleteProject 等函数复用
const projects = ref<Project[]>([])
const showCreateDialog = ref(false)
const showEditDialog = ref(false)
const newProject = ref({ name: '', path: '', language: 'nodejs' })
const editForm = ref({ id: '', name: '', path: '', language: 'nodejs' })

function goToRepositories() {
  router.push('/repositories')
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
    const result = await ListProjectsWithStats()
    stats.value = result || []
    projects.value = stats.value.map(s => s.project)
  } catch (e) {
    console.error('Failed to load projects:', e)
    toast?.error('加载项目列表失败')
    stats.value = []
    projects.value = []
  }
  loading.value = false
}

// ---- 卡片显示辅助 ----

function statusIcon(s: string): string {
  switch (s) {
    case 'success': return '✅'
    case 'failed':  return '❌'
    case 'building': return '⚙️'
    default: return '·'
  }
}

function statusText(s: string): string {
  switch (s) {
    case 'success':  return '成功'
    case 'failed':   return '失败'
    case 'building': return '构建中'
    case 'pending':  return '排队'
    default: return s
  }
}

// 把 ISO 时间转人类相对时间："3 分钟前 / 2 小时前 / 5 天前"
function relativeTime(iso: string): string {
  if (!iso) return ''
  const t = new Date(iso).getTime()
  if (isNaN(t)) return iso
  const diff = Date.now() - t
  const sec = Math.floor(diff / 1000)
  if (sec < 60) return '刚刚'
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min} 分钟前`
  const hr = Math.floor(min / 60)
  if (hr < 24) return `${hr} 小时前`
  const day = Math.floor(hr / 24)
  if (day < 30) return `${day} 天前`
  const mo = Math.floor(day / 30)
  if (mo < 12) return `${mo} 个月前`
  return `${Math.floor(mo / 12)} 年前`
}

function calcDuration(startISO: string, endISO: string): string {
  const ms = new Date(endISO).getTime() - new Date(startISO).getTime()
  if (isNaN(ms) || ms < 0) return ''
  const sec = Math.floor(ms / 1000)
  if (sec < 60) return `用时 ${sec}s`
  const min = Math.floor(sec / 60)
  return `用时 ${min}m${sec % 60 ? `${sec % 60}s` : ''}`
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

/* ---- 项目卡片 v2 ---- */

.project-card.pinned {
  border-color: var(--brand-start);
  box-shadow: 0 0 0 1px var(--brand-start), var(--shadow-sm);
}
.pin-mark {
  position: absolute;
  top: -8px;
  left: 14px;
  background: linear-gradient(135deg, var(--brand-start), var(--brand-end));
  color: #fff;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  box-shadow: var(--shadow-sm);
  z-index: 1;
  letter-spacing: 0.5px;
}

.project-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 16px;
}

.project-card {
  background: var(--card-bg);
  border-radius: var(--radius-lg);
  padding: 18px 20px;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
  transition: box-shadow 0.15s, transform 0.15s;
  position: relative;
}
.project-card:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-1px);
}

.project-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 12px;
}
.project-title { flex: 1; min-width: 0; }
.project-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 6px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.title-tags {
  display: flex; gap: 6px; flex-wrap: wrap;
}
.lang-badge {
  font-size: 10px;
  padding: 2px 8px;
  border-radius: 10px;
  background: var(--bg-primary);
  color: var(--text-secondary);
  font-weight: 500;
}
.src-badge {
  font-size: 10px;
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 500;
}
.src-badge.git   { background: var(--success-bg); color: var(--success-fg); }
.src-badge.local { background: var(--info-bg);    color: var(--info-fg); }

.btn-icon-menu {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 18px;
  cursor: pointer;
  width: 28px; height: 28px;
  border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
  transition: background 0.12s;
}
.btn-icon-menu:hover { background: var(--bg-primary); color: var(--text-primary); }

.project-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 12px;
  font-size: 12px;
  color: var(--text-secondary);
}
.info-row {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}
.info-icon { font-size: 11px; flex-shrink: 0; }
.info-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.info-text.mono {
  font-family: 'JetBrains Mono', 'Consolas', monospace;
  font-size: 11px;
  color: var(--text-muted);
}
.info-text.commit-subject {
  color: var(--text-primary);
  font-style: italic;
}
.commit-tag {
  font-family: 'JetBrains Mono', monospace;
  font-size: 10px;
  background: var(--bg-primary);
  padding: 1px 6px;
  border-radius: 3px;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.build-strip {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding: 8px 10px;
  background: var(--bg-primary);
  border-radius: var(--radius-sm);
  margin-bottom: 12px;
  font-size: 12px;
}
.status-pill {
  font-weight: 600;
  padding: 2px 10px;
  border-radius: 10px;
  font-size: 11px;
  white-space: nowrap;
}
.status-pill.success  { background: var(--success-bg); color: var(--success-fg); }
.status-pill.failed   { background: var(--danger-bg);  color: var(--danger-fg); }
.status-pill.building { background: var(--warning-bg); color: var(--warning-fg); }
.status-pill.never    { background: var(--bg-secondary); color: var(--text-muted); }
.build-meta {
  color: var(--text-secondary);
  font-size: 11px;
  display: flex; align-items: center; gap: 4px; flex-wrap: wrap;
  flex: 1; min-width: 0;
}
.build-meta .mono {
  font-family: 'JetBrains Mono', monospace;
  color: var(--text-primary);
}
.dot-sep { color: var(--text-muted); }
.build-count {
  margin-left: auto;
  font-size: 10px;
  color: var(--text-muted);
  font-family: 'JetBrains Mono', monospace;
}

.project-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  position: relative;
}
.btn-primary-sm {
  background: linear-gradient(135deg, var(--brand-start), var(--brand-end));
  color: white;
  border: none;
  padding: 6px 14px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.1s;
}
.btn-primary-sm:hover { transform: translateY(-1px); }
.btn-outline-sm {
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  padding: 5px 12px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  cursor: pointer;
}
.btn-outline-sm:hover {
  border-color: var(--brand-start);
  color: var(--brand-start);
}

.more-menu {
  position: absolute;
  top: -8px;
  right: 0;
  transform: translateY(-100%);
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-md);
  padding: 4px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 140px;
  z-index: 10;
}
.more-menu button {
  background: transparent;
  border: none;
  text-align: left;
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  color: var(--text-primary);
  cursor: pointer;
}
.more-menu button:hover { background: var(--bg-primary); }
.more-menu button.danger { color: var(--danger-fg); }
.more-menu button.danger:hover { background: var(--danger-bg); }

.menu-fade-enter-active, .menu-fade-leave-active {
  transition: opacity 0.1s, transform 0.12s;
}
.menu-fade-enter-from, .menu-fade-leave-to {
  opacity: 0;
  transform: translateY(-95%);
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

/* skeleton 占位 */
.skeleton-card { pointer-events: none; }
.skel-row {
  background: linear-gradient(90deg,
    var(--bg-primary) 0%,
    var(--border-color) 50%,
    var(--bg-primary) 100%);
  background-size: 200% 100%;
  border-radius: 4px;
  animation: shimmer 1.4s linear infinite;
}
@keyframes shimmer {
  0%   { background-position: 100% 0; }
  100% { background-position: -100% 0; }
}
</style>
