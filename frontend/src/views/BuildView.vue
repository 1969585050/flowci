<template>
  <div class="build-view">
    <h1>镜像构建</h1>

    <div class="card">
      <h3>关联项目</h3>
      <select v-model="selectedProjectId" class="project-select">
        <option value="">-- 请选择项目 --</option>
        <option v-for="p in projects" :key="p.id" :value="p.id">
          {{ p.name }} ({{ p.path }})
        </option>
      </select>
    </div>

    <div class="card">
      <h3>选择语言/框架</h3>
      <div class="lang-grid">
        <div
          v-for="lang in languages"
          :key="lang.id"
          class="lang-card"
          :class="{ active: selectedLang === lang.id }"
          @click="selectedLang = lang.id"
        >
          <span class="lang-emoji">{{ lang.emoji }}</span>
          <span class="lang-name">{{ lang.name }}</span>
        </div>
      </div>
    </div>

    <div class="card">
      <h3>配置</h3>
      <div class="form-row">
        <div class="form-group">
          <label>项目路径</label>
          <input v-model="buildConfig.contextPath" type="text" placeholder="/workspace/my-app" />
        </div>
        <div class="form-group">
          <label>镜像标签</label>
          <input v-model="buildConfig.tag" type="text" placeholder="my-app:latest" />
        </div>
      </div>
      <div class="form-options">
        <label>
          <input v-model="buildConfig.noCache" type="checkbox" />
          <span>不使用缓存</span>
        </label>
        <label>
          <input v-model="buildConfig.pullLatest" type="checkbox" />
          <span>拉取最新基础镜像</span>
        </label>
      </div>
    </div>

    <div class="card">
      <div style="display: flex; justify-content: space-between; align-items: center;">
        <h3>Dockerfile 预览</h3>
        <button class="btn-outline" @click="generateDockerfile">重新生成</button>
      </div>
      <div class="dockerfile-preview">
        <pre>{{ dockerfileContent }}</pre>
      </div>
    </div>

    <button class="btn-primary btn-large" @click="startBuild" :disabled="building">
      {{ building ? '构建中...' : '开始构建' }}
    </button>

    <div v-if="buildLogs.length > 0" class="card">
      <h3>构建日志</h3>
      <div class="logs-container">
        <div v-for="(log, index) in buildLogs" :key="index" class="log-line" :class="log.type">
          <span class="log-time">[{{ log.time }}]</span>
          <span class="log-text">{{ log.text }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { GenerateDockerfile, ListProjects, BuildImage } from '../wailsjs/go/main/App'

const route = useRoute()

interface Project {
  id: string
  name: string
  path: string
  language: string
}

const projects = ref<Project[]>([])
const selectedProjectId = ref('')

const languages = [
  { id: 'nodejs', name: 'Node.js', emoji: '🟢' },
  { id: 'go', name: 'Go', emoji: '🔵' },
  { id: 'python', name: 'Python', emoji: '🐍' },
  { id: 'java-maven', name: 'Java (Maven)', emoji: '☕' },
  { id: 'java-gradle', name: 'Java (Gradle)', emoji: '🅰️' },
  { id: 'php', name: 'PHP', emoji: '🐘' },
  { id: 'ruby', name: 'Ruby', emoji: '💎' },
  { id: 'dotnet', name: '.NET', emoji: '💜' },
  { id: 'rust', name: 'Rust', emoji: '🦀' },
  { id: 'c', name: 'C/C++', emoji: '⚙️' }
]

const selectedLang = ref('nodejs')
const building = ref(false)
const buildLogs = ref<{ time: string; text: string; type: string }[]>([])

const buildConfig = ref({
  contextPath: '/workspace/my-app',
  tag: 'my-app:latest',
  noCache: false,
  pullLatest: false
})

const dockerfileContent = ref(`FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
EXPOSE 3000
CMD ["node", "index.js"]`)

async function generateDockerfile() {
  if (!selectedProjectId.value) {
    addLog('请先选择项目', 'error')
    return
  }
  const lang = selectedLang.value
  try {
    const content = await GenerateDockerfile(lang)
    dockerfileContent.value = content
  } catch (error) {
    addLog(`Dockerfile生成失败: ${error}`, 'error')
    dockerfileContent.value = `FROM alpine:latest
WORKDIR /app
COPY . .
CMD ["./app"]`
  }
}

function addLog(text: string, type = 'info') {
  const time = new Date().toLocaleTimeString('zh-CN')
  buildLogs.value.push({ time, text, type })
}

async function loadProjects() {
  try {
    projects.value = (await ListProjects()) as Project[]
  } catch {
    console.error('Failed to load projects')
  }
}

async function startBuild() {
  if (building.value) return
  if (!selectedProjectId.value) {
    addLog('请先选择关联项目', 'error')
    return
  }
  building.value = true
  buildLogs.value = []
  
  const project = projects.value.find(p => p.id === selectedProjectId.value)
  addLog('开始构建...')
  addLog(`项目: ${project?.name || selectedProjectId.value}`)
  addLog(`语言: ${selectedLang.value}`)
  addLog(`镜像: ${buildConfig.value.tag}`)
  
  try {
    const result = await BuildImage({
      projectId: selectedProjectId.value,
      contextPath: buildConfig.value.contextPath,
      tag: buildConfig.value.tag,
      noCache: buildConfig.value.noCache,
      pullLatest: buildConfig.value.pullLatest
    })
    if (result.log) {
      result.log.split('\n').forEach((line: string) => {
        if (line.trim()) addLog(line)
      })
    }
    if (result.success) {
      addLog(`构建成功！镜像: ${result.image_name}:${result.image_tag}`, 'success')
    } else {
      addLog(`构建失败: ${result.error}`, 'error')
    }
  } catch (e) {
    addLog(`构建失败: ${e}`, 'error')
  }
  
  building.value = false
}

watch(selectedProjectId, (newId) => {
  if (!newId) return
  const project = projects.value.find(p => p.id === newId)
  if (project) {
    buildConfig.value.contextPath = project.path
    selectedLang.value = project.language
    generateDockerfile()
  }
})

onMounted(() => {
  loadProjects().then(() => {
    const projectId = route.query.projectId as string
    if (projectId && projects.value.some(p => p.id === projectId)) {
      selectedProjectId.value = projectId
    } else {
      generateDockerfile()
    }
  })
})
</script>

<style scoped>
.build-view {
  max-width: 900px;
}

h1 {
  font-size: 28px;
  color: var(--text-primary, #1a1a2e);
  margin-bottom: 24px;
}

.card {
  background: var(--card-bg, white);
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

.lang-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 12px;
}

.lang-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px;
  border: 2px solid var(--border-color, #e0e0e0);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
}

.lang-card:hover {
  border-color: #667eea;
  background: #f8f9ff;
}

.lang-card.active {
  border-color: #667eea;
  background: #f0f3ff;
}

.lang-emoji {
  font-size: 36px;
  margin-bottom: 8px;
}

.lang-name {
  font-size: 14px;
  font-weight: 600;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
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
  background: var(--card-bg, white);
  color: var(--text-primary, #333);
  transition: border-color 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: #667eea;
}

.project-select {
  width: 100%;
  padding: 12px;
  border: 2px solid var(--border-color, #e0e0e0);
  border-radius: 8px;
  font-size: 14px;
  background: var(--card-bg, white);
  cursor: pointer;
  transition: border-color 0.2s;
}

.project-select:focus {
  outline: none;
  border-color: #667eea;
}

.form-options {
  margin-top: 16px;
  display: flex;
  gap: 24px;
}

.form-options label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  cursor: pointer;
}

.form-options input {
  width: 18px;
  height: 18px;
}

.dockerfile-preview {
  background: #1e1e2e;
  border-radius: 8px;
  padding: 16px;
  margin-top: 12px;
}

.dockerfile-preview pre {
  color: #a5b6cf;
  font-family: 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  margin: 0;
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

.btn-large {
  width: 100%;
  padding: 16px;
  font-size: 16px;
}

.logs-container {
  background: #1e1e2e;
  border-radius: 8px;
  padding: 16px;
  max-height: 300px;
  overflow-y: auto;
}

.log-line {
  font-family: 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
  padding: 4px 0;
}

.log-time {
  color: #6b7089;
  margin-right: 8px;
}

.log-text {
  color: #a5b6cf;
}

.log-line.success .log-text {
  color: #a6e3a1;
}

.log-line.error .log-text {
  color: #f38ba8;
}
</style>
