<template>
  <div class="projects-view">
    <div class="header">
      <h1>项目列表</h1>
      <button class="btn-primary" @click="showCreateDialog = true">+ 新建项目</button>
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
          <span>创建: {{ formatDate(project.created_at) }}</span>
        </div>
        <div class="project-actions">
          <button class="btn-small" @click="buildProject(project)">构建</button>
          <button class="btn-small btn-secondary" @click="deployProject(project)">部署</button>
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
  </div>
</template>

<script setup lang="ts">
import { ref, inject, onMounted } from 'vue'
import { useRouter } from 'vue-router'

interface Project {
  id: string
  name: string
  path: string
  language: string
  created_at: string
}

const router = useRouter()
const loading = ref(false)
const creating = ref(false)
const projects = ref<Project[]>([])
const showCreateDialog = ref(false)
const newProject = ref({ name: '', path: '', language: 'nodejs' })

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
    if ((window as any).wails?.Invoke) {
      projects.value = await (window as any).wails.Invoke('ListProjects')
    }
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
    if ((window as any).wails?.Invoke) {
      await (window as any).wails.Invoke('CreateProject', {
        name: newProject.value.name,
        path: newProject.value.path,
        language: newProject.value.language
      })
      toast?.success('项目创建成功')
      showCreateDialog.value = false
      newProject.value = { name: '', path: '', language: 'nodejs' }
      await refreshProjects()
    }
  } catch (e) {
    toast?.error(`创建失败: ${e}`)
  }
  creating.value = false
}

async function deleteProject(project: Project) {
  if (!confirm(`确定要删除项目 "${project.name}" 吗？`)) return
  try {
    if ((window as any).wails?.Invoke) {
      await (window as any).wails.Invoke('DeleteProject', project.id)
      toast?.success('项目已删除')
      await refreshProjects()
    }
  } catch (e) {
    toast?.error(`删除失败: ${e}`)
  }
}

function buildProject(project: Project) {
  router.push('/build')
}

function deployProject(project: Project) {
  router.push('/deploy')
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
  color: #1a1a2e;
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
  border: 4px solid #e0e0e0;
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
  color: #666;
}

.empty-state p {
  color: #999;
}

.project-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.project-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
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
  color: #1a1a2e;
}

.lang-badge {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
}

.project-path {
  color: #666;
  font-size: 13px;
  margin-bottom: 12px;
}

.project-meta {
  color: #999;
  font-size: 12px;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid #eee;
}

.project-actions {
  display: flex;
  gap: 12px;
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
  background: #f0f0f0;
  color: #333;
}

.btn-danger {
  background: #fee2e2;
  color: #dc2626;
}

.btn-danger:hover {
  background: #fecaca;
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
  background: white;
  border-radius: 16px;
  padding: 32px;
  width: 480px;
  max-width: 90vw;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}

.modal h2 {
  font-size: 22px;
  color: #1a1a2e;
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
  color: #333;
}

.form-group input,
.form-group select {
  padding: 12px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 14px;
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
  background: #f0f0f0;
  color: #333;
  border: none;
  padding: 12px 24px;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
}
</style>
