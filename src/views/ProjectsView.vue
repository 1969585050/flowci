<template>
  <div class="projects-view">
    <div class="view-header">
      <h2>📁 项目列表</h2>
      <button class="btn-primary" @click="showCreateDialog = true">+ 新建项目</button>
    </div>

    <div class="docker-status" :class="{ connected: dockerStatus.connected }">
      <span class="status-icon">{{ dockerStatus.connected ? '✅' : '❌' }}</span>
      <span>Docker: {{ dockerStatus.connected ? '已连接' : '未连接' }}</span>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="projects.length === 0" class="empty-state">
      <p>暂无项目</p>
      <p>点击上方按钮创建第一个项目</p>
    </div>

    <div v-else class="projects-grid">
      <div v-for="project in projects" :key="project.id" class="project-card" @click="selectProject(project)">
        <div class="project-icon">📦</div>
        <div class="project-info">
          <h3>{{ project.name }}</h3>
          <p class="project-path">{{ project.path }}</p>
          <p class="project-language">语言: {{ project.language }}</p>
        </div>
        <div class="project-actions">
          <button @click.stop="buildProject(project)">构建</button>
          <button @click.stop="deployProject(project)">部署</button>
          <button @click.stop="deleteProject(project.id)" class="btn-danger">删除</button>
        </div>
      </div>
    </div>

    <div v-if="showCreateDialog" class="dialog-overlay" @click.self="showCreateDialog = false">
      <div class="dialog">
        <h3>创建新项目</h3>
        <form @submit.prevent="createProject">
          <div class="form-group">
            <label>项目名称</label>
            <input v-model="newProject.name" type="text" required placeholder="例如: my-app" />
          </div>
          <div class="form-group">
            <label>项目路径</label>
            <input v-model="newProject.path" type="text" required placeholder="例如: /path/to/project" />
          </div>
          <div class="form-group">
            <label>编程语言</label>
            <select v-model="newProject.language" required>
              <option value="java-maven">Java (Maven)</option>
              <option value="java-gradle">Java (Gradle)</option>
              <option value="nodejs">Node.js</option>
              <option value="python">Python</option>
              <option value="go">Go</option>
              <option value="php">PHP</option>
              <option value="ruby">Ruby</option>
              <option value="dotnet">.NET</option>
            </select>
          </div>
          <div class="dialog-actions">
            <button type="button" @click="showCreateDialog = false">取消</button>
            <button type="submit" class="btn-primary">创建</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { invoke } from '@tauri-apps/api/core'

interface Project {
  id: string
  name: string
  path: string
  language: string
  created_at: string
  updated_at: string
}

interface DockerStatus {
  connected: boolean
  version?: string
  error?: string
}

const projects = ref<Project[]>([])
const loading = ref(true)
const showCreateDialog = ref(false)
const dockerStatus = ref<DockerStatus>({ connected: false })

const newProject = ref({
  name: '',
  path: '',
  language: 'nodejs'
})

onMounted(async () => {
  await checkDocker()
  await loadProjects()
})

async function checkDocker() {
  try {
    dockerStatus.value = await invoke('check_docker')
  } catch (e) {
    dockerStatus.value = { connected: false, error: String(e) }
  }
}

async function loadProjects() {
  try {
    loading.value = true
    projects.value = await invoke('list_projects')
  } catch (e) {
    console.error('Failed to load projects:', e)
  } finally {
    loading.value = false
  }
}

async function createProject() {
  try {
    await invoke('create_project', {
      request: {
        name: newProject.value.name,
        path: newProject.value.path,
        language: newProject.value.language
      }
    })
    showCreateDialog.value = false
    newProject.value = { name: '', path: '', language: 'nodejs' }
    await loadProjects()
  } catch (e) {
    console.error('Failed to create project:', e)
  }
}

async function deleteProject(id: string) {
  if (confirm('确定要删除这个项目吗？')) {
    try {
      await invoke('delete_project', { projectId: id })
      await loadProjects()
    } catch (e) {
      console.error('Failed to delete project:', e)
    }
  }
}

function selectProject(project: Project) {
  console.log('Selected project:', project)
}

function buildProject(project: Project) {
  console.log('Build project:', project)
}

function deployProject(project: Project) {
  console.log('Deploy project:', project)
}
</script>

<style scoped>
.projects-view {
  max-width: 1200px;
  margin: 0 auto;
}

.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.view-header h2 {
  margin: 0;
}

.docker-status {
  background: #fee;
  border: 1px solid #fcc;
  padding: 0.75rem 1rem;
  border-radius: 8px;
  margin-bottom: 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.docker-status.connected {
  background: #efe;
  border-color: #cfc;
}

.loading, .empty-state {
  text-align: center;
  padding: 3rem;
  color: #666;
}

.projects-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1rem;
}

.project-card {
  background: white;
  border-radius: 8px;
  padding: 1.5rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.project-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.project-icon {
  font-size: 2.5rem;
  margin-bottom: 1rem;
}

.project-info h3 {
  margin: 0 0 0.5rem 0;
}

.project-path {
  color: #666;
  font-size: 0.875rem;
  word-break: break-all;
}

.project-language {
  color: #888;
  font-size: 0.875rem;
  margin-top: 0.5rem;
}

.project-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 1rem;
}

.project-actions button {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  background: #667eea;
  color: white;
  transition: background 0.2s;
}

.project-actions button:hover {
  background: #5568d3;
}

.project-actions .btn-danger {
  background: #e53e3e;
}

.project-actions .btn-danger:hover {
  background: #c53030;
}

.btn-primary {
  background: #667eea !important;
  color: white !important;
  padding: 0.75rem 1.5rem !important;
  border: none !important;
  border-radius: 6px !important;
  cursor: pointer !important;
  font-weight: 500 !important;
}

.btn-primary:hover {
  background: #5568d3 !important;
}

.dialog-overlay {
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

.dialog {
  background: white;
  border-radius: 12px;
  padding: 2rem;
  width: 100%;
  max-width: 450px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}

.dialog h3 {
  margin-top: 0;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 1rem;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1.5rem;
}
</style>
