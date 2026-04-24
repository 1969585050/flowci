<template>
  <div class="projects-view">
    <div class="header">
      <h1>项目列表</h1>
      <button class="btn-primary" @click="refreshProjects">刷新</button>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <span>加载中...</span>
    </div>

    <div v-else-if="projects.length === 0" class="empty-state">
      <div class="empty-icon">📦</div>
      <h3>暂无项目</h3>
      <p>开始添加你的第一个项目吧</p>
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
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
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
const projects = ref<Project[]>([])

const langMap: Record<string, string> = {
  nodejs: 'Node.js',
  go: 'Go',
  python: 'Python',
  java: 'Java',
  rust: 'Rust',
  dotnet: '.NET',
  php: 'PHP',
  ruby: 'Ruby',
  elixir: 'Elixir'
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
      const result = await (window as any).wails.Invoke('ListProjects')
      projects.value = result
    } else {
      projects.value = [
        { id: '1', name: '示例项目 (Node.js)', path: '/workspace/my-app', language: 'nodejs', created_at: new Date().toISOString() },
        { id: '2', name: 'API 服务 (Go)', path: '/workspace/api-service', language: 'go', created_at: new Date().toISOString() }
      ]
    }
  } catch (e) {
    console.error('Failed to load projects:', e)
  }
  loading.value = false
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
</style>
