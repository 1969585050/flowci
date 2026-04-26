<template>
  <div class="projects-view">
    <div class="header">
      <h1>项目列表</h1>
      <div class="header-actions">
        <button class="btn btn-secondary" @click="goToRepositories">
          <GitBranch :size="14" :stroke-width="1.75" />
          从仓库源导入
        </button>
        <button class="btn btn-primary" @click="showCreateDialog = true">
          <Plus :size="14" :stroke-width="2" />
          新建项目
        </button>
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
      <PackageOpen class="empty-icon" :size="56" :stroke-width="1.25" />
      <h3>暂无项目</h3>
      <p>点击右上角按钮创建你的第一个项目</p>
    </div>

    <div v-else class="project-grid">
      <div v-for="s in stats" :key="s.project.id" class="card project-card" :class="{ 'card-pinned': s.project.pinnedAt }">
        <span v-if="s.project.pinnedAt" class="pin-mark" title="已置顶">
          <Pin :size="11" :stroke-width="2" />
        </span>
        <!-- 标题行：项目名 + 语言 + 来源标签 -->
        <div class="project-header">
          <div class="project-title">
            <h3 class="project-name">{{ s.project.name }}</h3>
            <div class="title-tags">
              <span class="badge badge-neutral">{{ getLangName(s.project.language) }}</span>
              <span v-if="s.project.repoUrl" class="badge badge-success" title="Git 仓库项目">
                <GitBranch :size="11" :stroke-width="2" /> Git
              </span>
              <span v-else class="badge badge-info" title="本地路径项目">
                <Folder :size="11" :stroke-width="2" /> 本地
              </span>
            </div>
          </div>
          <button class="btn btn-icon" @click="toggleMenu(s.project.id)" title="更多操作">
            <MoreHorizontal :size="16" :stroke-width="1.75" />
          </button>
        </div>

        <!-- 路径 / Git 信息 -->
        <div class="project-info">
          <div class="info-row" :title="s.project.path">
            <FolderOpen class="info-icon" :size="12" :stroke-width="1.75" />
            <span class="info-text mono">{{ s.project.path || '(未配置路径)' }}</span>
          </div>
          <div v-if="s.project.repoBranch" class="info-row">
            <GitBranch class="info-icon" :size="12" :stroke-width="1.75" />
            <span class="info-text">{{ s.project.repoBranch }}</span>
            <span v-if="s.headCommit" class="badge badge-mono badge-neutral" :title="s.headSubject">
              {{ s.headCommit }}
            </span>
          </div>
          <div v-if="s.headSubject" class="info-row" :title="s.headSubject">
            <MessageSquareText class="info-icon" :size="12" :stroke-width="1.75" />
            <span class="info-text commit-subject">{{ s.headSubject }}</span>
          </div>
        </div>

        <!-- 构建状态条 -->
        <div class="build-strip">
          <template v-if="s.lastBuild">
            <span class="badge" :class="statusBadgeClass(s.lastBuild.status)">
              <component
                :is="statusIconComp(s.lastBuild.status)"
                :size="12"
                :stroke-width="2"
                :class="{ 'icon-spin': s.lastBuild.status === 'building' }"
              />
              {{ statusText(s.lastBuild.status) }}
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
            <span class="badge badge-neutral">
              <Clock :size="11" :stroke-width="1.75" /> 从未构建
            </span>
          </template>
          <span class="build-count" v-if="s.buildCount > 0">{{ s.buildCount }} 次构建</span>
        </div>

        <!-- 主操作 + 折叠菜单 -->
        <div class="project-actions">
          <button class="btn btn-primary btn-sm" @click="buildProject(s.project)">
            <Play :size="12" :stroke-width="2" /> 构建
          </button>
          <button class="btn btn-secondary btn-sm" @click="showHistory(s.project)">
            <History :size="12" :stroke-width="1.75" /> 历史
          </button>
          <transition name="menu-fade">
            <div v-if="openMenu === s.project.id" class="more-menu" @click.stop>
              <button v-if="s.project.pinnedAt" @click="togglePin(s.project, false); closeMenu()">
                <PinOff :size="13" :stroke-width="1.75" /> 取消置顶
              </button>
              <button v-else @click="togglePin(s.project, true); closeMenu()">
                <Pin :size="13" :stroke-width="1.75" /> 置顶
              </button>
              <button @click="deployProject(s.project); closeMenu()">
                <Rocket :size="13" :stroke-width="1.75" /> 部署
              </button>
              <button @click="editProject(s.project); closeMenu()">
                <Pencil :size="13" :stroke-width="1.75" /> 编辑
              </button>
              <button class="danger" @click="deleteProject(s.project); closeMenu()">
                <Trash2 :size="13" :stroke-width="1.75" /> 删除
              </button>
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
          <input class="input" v-model="newProject.name" type="text" placeholder="my-project" />
        </div>
        <div class="form-group">
          <label>项目路径</label>
          <input class="input" v-model="newProject.path" type="text" placeholder="/workspace/my-project" />
        </div>
        <div class="form-group">
          <label>语言/框架</label>
          <select class="select" v-model="newProject.language">
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
          <button class="btn btn-secondary" @click="showCreateDialog = false">取消</button>
          <button class="btn btn-primary" @click="createProject" :disabled="creating">创建</button>
        </div>
      </div>
    </div>

    <div v-if="showEditDialog" class="modal-overlay" @click.self="showEditDialog = false">
      <div class="modal">
        <h2>编辑项目</h2>
        <div class="form-group">
          <label>项目名称</label>
          <input class="input" v-model="editForm.name" type="text" placeholder="my-project" />
        </div>
        <div class="form-group">
          <label>项目路径</label>
          <input class="input" v-model="editForm.path" type="text" placeholder="/workspace/my-project" />
        </div>
        <div class="form-group">
          <label>语言/框架</label>
          <select class="select" v-model="editForm.language">
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
          <button class="btn btn-secondary" @click="showEditDialog = false">取消</button>
          <button class="btn btn-primary" @click="updateProject" :disabled="updating">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, inject, onMounted, type Component } from 'vue'
import { useRouter } from 'vue-router'
import {
  Plus, GitBranch, Folder, MoreHorizontal, FolderOpen, MessageSquareText,
  Play, History, Pin, PinOff, Rocket, Pencil, Trash2, PackageOpen,
  CheckCircle2, XCircle, Loader2, Clock, Circle,
} from 'lucide-vue-next'
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

function statusIconComp(s: string): Component {
  switch (s) {
    case 'success':  return CheckCircle2
    case 'failed':   return XCircle
    case 'building': return Loader2
    case 'pending':  return Clock
    default:         return Circle
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

function statusBadgeClass(s: string): string {
  switch (s) {
    case 'success':  return 'badge-success'
    case 'failed':   return 'badge-danger'
    case 'building': return 'badge-warning'
    case 'pending':  return 'badge-info'
    default:         return 'badge-neutral'
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
/* ---- Page header ---- */
.projects-view {
  max-width: 1200px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-6);
}
.header h1 {
  font-size: var(--text-3xl);
  font-weight: var(--weight-semibold);
  color: var(--text-primary);
  letter-spacing: -0.01em;
}
.header-actions {
  display: flex;
  gap: var(--space-3);
}

/* ---- Empty state ---- */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px var(--space-5);
  gap: var(--space-4);
}
.empty-icon { color: var(--text-muted); }
.empty-state h3 {
  font-size: var(--text-xl);
  font-weight: var(--weight-medium);
  color: var(--text-secondary);
}
.empty-state p { color: var(--text-muted); }

/* ---- Icon spin animation (used by Loader2 for "building" status) ---- */
.icon-spin {
  animation: icon-spin 1s linear infinite;
}
@keyframes icon-spin {
  to { transform: rotate(360deg); }
}

/* ---- Project grid ---- */
.project-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: var(--space-4);
}

/* ---- Project card (extends .card from components.css) ---- */
.project-card {
  padding: var(--space-4) var(--space-5);
}

.pin-mark {
  position: absolute;
  top: -8px;
  left: var(--space-4);
  background: var(--brand-500);
  color: var(--text-on-brand);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  z-index: 1;
  box-shadow: var(--shadow-sm);
}

.project-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--space-2);
  margin-bottom: var(--space-3);
}
.project-title { flex: 1; min-width: 0; }
.project-name {
  font-size: var(--text-md);
  font-weight: var(--weight-semibold);
  color: var(--text-primary);
  margin: 0 0 var(--space-1) 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.title-tags {
  display: flex; gap: var(--space-1); flex-wrap: wrap;
}

.project-info {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  margin-bottom: var(--space-3);
  font-size: var(--text-sm);
  color: var(--text-secondary);
}
.info-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
}
.info-icon { font-size: var(--text-xs); flex-shrink: 0; }
.info-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.info-text.mono {
  font-family: var(--font-mono);
  font-size: var(--text-xs);
  color: var(--text-muted);
}
.info-text.commit-subject {
  color: var(--text-primary);
  font-style: italic;
}

/* ---- Build strip (sunken sub-panel) ---- */
.build-strip {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
  padding: var(--space-2) var(--space-3);
  background: var(--bg-sunken);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-3);
  font-size: var(--text-sm);
}
.build-meta {
  color: var(--text-secondary);
  font-size: var(--text-xs);
  display: flex; align-items: center; gap: var(--space-1); flex-wrap: wrap;
  flex: 1; min-width: 0;
}
.build-meta .mono {
  font-family: var(--font-mono);
  color: var(--text-primary);
}
.dot-sep { color: var(--text-muted); }
.build-count {
  margin-left: auto;
  font-size: var(--text-xs);
  color: var(--text-muted);
  font-family: var(--font-mono);
}

/* ---- Actions row ---- */
.project-actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  position: relative;
}

/* ---- More menu ---- */
.more-menu {
  position: absolute;
  top: -8px;
  right: 0;
  transform: translateY(-100%);
  background: var(--bg-elevated);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-md);
  padding: var(--space-1);
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 140px;
  z-index: var(--z-dropdown);
}
.more-menu button {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  background: transparent;
  border: none;
  text-align: left;
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  color: var(--text-primary);
  cursor: pointer;
  transition: background var(--transition-fast);
}
.more-menu button:hover { background: var(--bg-hover); }
.more-menu button.danger { color: var(--danger-fg); }
.more-menu button.danger:hover { background: var(--danger-bg); }

.menu-fade-enter-active, .menu-fade-leave-active {
  transition: opacity var(--duration-fast), transform var(--duration-fast);
}
.menu-fade-enter-from, .menu-fade-leave-to {
  opacity: 0;
  transform: translateY(-95%);
}

/* ---- Modal ---- */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: var(--z-modal);
}
.modal {
  background: var(--bg-elevated);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-xl);
  padding: var(--space-8);
  width: 480px;
  max-width: 90vw;
  box-shadow: var(--shadow-lg);
}
.modal h2 {
  font-size: var(--text-xl);
  font-weight: var(--weight-semibold);
  color: var(--text-primary);
  margin-bottom: var(--space-6);
}
.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  margin-bottom: var(--space-4);
}
.form-group label {
  font-size: var(--text-sm);
  font-weight: var(--weight-medium);
  color: var(--text-secondary);
}
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  margin-top: var(--space-6);
}

/* ---- Skeleton ---- */
.skeleton-card { pointer-events: none; }
.skel-row {
  background: linear-gradient(90deg,
    var(--bg-sunken) 0%,
    var(--border-subtle) 50%,
    var(--bg-sunken) 100%);
  background-size: 200% 100%;
  border-radius: var(--radius-sm);
  animation: shimmer 1.4s linear infinite;
}
@keyframes shimmer {
  0%   { background-position: 100% 0; }
  100% { background-position: -100% 0; }
}

</style>
