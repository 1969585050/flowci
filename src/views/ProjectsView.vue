<template>
  <div class="projects-view">
    <div class="view-header">
      <h2>📁 项目列表</h2>
      <button class="btn-primary" @click="showCreateDialog = true">+ 新建项目</button>
    </div>

    <div class="docker-status" :class="{ connected: dockerStatus.connected }">
      <span class="status-icon">{{ dockerStatus.connected ? '✅' : '❌' }}</span>
      <span>Docker: {{ dockerStatus.connected ? '已连接' : '未连接' }}</span>
      <span v-if="dockerStatus.version" class="version">v{{ dockerStatus.version }}</span>
    </div>

    <div v-if="loading" class="loading">
      <span class="spinner"></span>
      加载中...
    </div>

    <div v-else-if="projects.length === 0" class="empty-state">
      <div class="empty-icon">📦</div>
      <p>暂无项目</p>
      <p>点击上方按钮创建第一个项目</p>
    </div>

    <div v-else class="projects-grid">
      <div
        v-for="project in projects"
        :key="project.id"
        class="project-card"
        @click="selectProject(project)"
      >
        <div class="project-icon">{{ getLanguageIcon(project.language) }}</div>
        <div class="project-info">
          <h3>{{ project.name }}</h3>
          <p class="project-path">{{ project.path }}</p>
          <p class="project-language">语言: {{ getLanguageLabel(project.language) }}</p>
          <p class="project-updated">更新: {{ formatDate(project.updated_at) }}</p>
        </div>
        <div class="project-actions">
          <button @click.stop="buildProject(project)" class="btn-build">构建</button>
          <button @click.stop="deployProject(project)" class="btn-deploy">部署</button>
          <button @click.stop="deleteProject(project.id)" class="btn-danger">删除</button>
        </div>
      </div>
    </div>

    <div v-if="showCreateDialog" class="dialog-overlay" @click.self="closeDialog">
      <div class="dialog">
        <h3>创建新项目</h3>
        <form @submit.prevent="handleCreateProject">
          <div class="form-group">
            <label>项目名称</label>
            <input
              v-model="newProject.name"
              type="text"
              required
              placeholder="例如: my-app"
              class="form-input"
            />
          </div>
          <div class="form-group">
            <label>项目路径</label>
            <input
              v-model="newProject.path"
              type="text"
              required
              placeholder="例如: /path/to/project"
              class="form-input"
            />
          </div>
          <div class="form-group">
            <label>编程语言</label>
            <select v-model="newProject.language" required class="form-input">
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
            <button type="button" @click="closeDialog" class="btn-cancel">取消</button>
            <button type="submit" class="btn-primary" :disabled="creating">
              {{ creating ? '创建中...' : '创建' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { invoke } from '@tauri-apps/api/core'
import type { Project, DockerCheckResponse } from '../types'
import { FlowCIError } from '../types/api'

const projects = ref<Project[]>([])
const loading = ref(true)
const creating = ref(false)
const showCreateDialog = ref(false)
const dockerStatus = ref<DockerCheckResponse>({
  connected: false,
  version: '',
  api_version: '',
  os: '',
  arch: ''
})

const newProject = ref({
  name: '',
  path: '',
  language: 'nodejs'
})

const languageMap: Record<string, string> = {
  'java-maven': '☕ Java (Maven)',
  'java-gradle': '☕ Java (Gradle)',
  'nodejs': '🟢 Node.js',
  'python': '🐍 Python',
  'go': '🔵 Go',
  'php': '🐘 PHP',
  'ruby': '💎 Ruby',
  'dotnet': '🔷 .NET'
}

const languageIconMap: Record<string, string> = {
  'java-maven': '☕',
  'java-gradle': '☕',
  'nodejs': '🟢',
  'python': '🐍',
  'go': '🔵',
  'php': '🐘',
  'ruby': '💎',
  'dotnet': '🔷'
}

function getLanguageLabel(lang: string): string {
  return languageMap[lang] || lang
}

function getLanguageIcon(lang: string): string {
  return languageIconMap[lang] || '📦'
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '未知'
  try {
    const date = new Date(dateStr)
    return date.toLocaleDateString('zh-CN')
  } catch {
    return dateStr
  }
}

onMounted(async () => {
  await Promise.all([checkDocker(), loadProjects()])
})

async function checkDocker(): Promise<void> {
  try {
    dockerStatus.value = await invoke<DockerCheckResponse>('check_docker')
  } catch (e) {
    console.error('Docker check failed:', e)
    dockerStatus.value = { connected: false, version: '', api_version: '', os: '', arch: '' }
  }
}

async function loadProjects(): Promise<void> {
  try {
    loading.value = true
    projects.value = await invoke<Project[]>('list_projects')
  } catch (e) {
    console.error('Failed to load projects:', e)
    projects.value = []
  } finally {
    loading.value = false
  }
}

function closeDialog(): void {
  showCreateDialog.value = false
  newProject.value = { name: '', path: '', language: 'nodejs' }
}

async function handleCreateProject(): Promise<void> {
  try {
    creating.value = true
    await invoke<Project>('create_project', {
      request: {
        name: newProject.value.name,
        path: newProject.value.path,
        language: newProject.value.language
      }
    })
    closeDialog()
    await loadProjects()
  } catch (e) {
    const error = e as FlowCIError
    console.error('Failed to create project:', error.message || e)
    alert(`创建失败: ${error.message || '未知错误'}`)
  } finally {
    creating.value = false
  }
}

async function deleteProject(id: string): Promise<void> {
  if (!confirm('确定要删除这个项目吗？此操作不可撤销。')) {
    return
  }
  try {
    await invoke<void>('delete_project', { projectId: id })
    await loadProjects()
  } catch (e) {
    const error = e as FlowCIError
    console.error('Failed to delete project:', error.message || e)
    alert(`删除失败: ${error.message || '未知错误'}`)
  }
}

function selectProject(project: Project): void {
  console.log('Selected project:', project)
}

function buildProject(project: Project): void {
  console.log('Build project:', project)
}

function deployProject(project: Project): void {
  console.log('Deploy project:', project)
}
</script>

<style scoped>
.projects-view {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1.5rem;
}

.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.view-header h2 {
  margin: 0;
  font-size: 1.75rem;
}

.docker-status {
  background: linear-gradient(135deg, #fff5f5 0%, #fed7d7 100%);
  border: 1px solid #feb2b2;
  padding: 1rem 1.25rem;
  border-radius: 12px;
  margin-bottom: 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.95rem;
}

.docker-status.connected {
  background: linear-gradient(135deg, #f0fff4 0%, #c6f6d5 100%);
  border-color: #9ae6b4;
}

.status-icon {
  font-size: 1.25rem;
}

.version {
  margin-left: auto;
  color: #666;
  font-size: 0.85rem;
}

.loading, .empty-state {
  text-align: center;
  padding: 4rem 2rem;
  color: #666;
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}

.spinner {
  display: inline-block;
  width: 20px;
  height: 20px;
  border: 2px solid #667eea;
  border-radius: 50%;
  border-top-color: transparent;
  animation: spin 0.8s linear infinite;
  margin-right: 0.5rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.projects-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.25rem;
}

.project-card {
  background: white;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid #e2e8f0;
}

.project-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  border-color: #667eea;
}

.project-icon {
  font-size: 2.5rem;
  margin-bottom: 1rem;
}

.project-info h3 {
  margin: 0 0 0.5rem 0;
  font-size: 1.15rem;
  color: #1a202c;
}

.project-path {
  color: #4a5568;
  font-size: 0.85rem;
  word-break: break-all;
  margin: 0.25rem 0;
}

.project-language {
  color: #718096;
  font-size: 0.85rem;
  margin: 0.25rem 0;
}

.project-updated {
  color: #a0aec0;
  font-size: 0.8rem;
  margin: 0.25rem 0;
}

.project-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #e2e8f0;
}

.project-actions button {
  flex: 1;
  padding: 0.5rem 0.75rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 500;
  transition: all 0.2s;
  background: #edf2f7;
  color: #4a5568;
}

.project-actions button:hover {
  background: #e2e8f0;
}

.btn-build {
  background: #667eea !important;
  color: white !important;
}

.btn-build:hover {
  background: #5568d3 !important;
}

.btn-deploy {
  background: #48bb78 !important;
  color: white !important;
}

.btn-deploy:hover {
  background: #38a169 !important;
}

.btn-danger {
  background: #e53e3e !important;
  color: white !important;
}

.btn-danger:hover {
  background: #c53030 !important;
}

.btn-primary {
  background: #667eea;
  color: white;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  font-size: 1rem;
  transition: background 0.2s;
}

.btn-primary:hover:not(:disabled) {
  background: #5568d3;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-cancel {
  background: #e2e8f0;
  color: #4a5568;
  padding: 0.75rem 1.25rem;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.dialog {
  background: white;
  border-radius: 16px;
  padding: 2rem;
  width: 100%;
  max-width: 480px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.dialog h3 {
  margin: 0 0 1.5rem 0;
  font-size: 1.35rem;
  color: #1a202c;
}

.form-group {
  margin-bottom: 1.25rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: #4a5568;
}

.form-input {
  width: 100%;
  padding: 0.875rem;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.15);
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1.75rem;
}
</style>
