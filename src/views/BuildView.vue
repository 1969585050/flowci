<template>
  <div class="build-view">
    <h2>🔨 镜像构建</h2>
    <p class="subtitle">选择项目并构建 Docker 镜像</p>

    <div class="build-form">
      <div class="form-section">
        <h3>项目配置</h3>
        <div class="form-group">
          <label>选择项目</label>
          <select v-model="selectedProjectId">
            <option value="">请选择项目</option>
            <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
        </div>

        <div class="form-group">
          <label>镜像标签</label>
          <input v-model="imageTag" type="text" placeholder="例如: myapp:latest" />
        </div>
      </div>

      <div class="form-section">
        <h3>镜像仓库</h3>
        <div class="form-group">
          <label>仓库类型</label>
          <select v-model="registryType">
            <option value="aliyun">阿里云 ACR</option>
            <option value="dockerhub">Docker Hub</option>
            <option value="harbor">Harbor</option>
          </select>
        </div>

        <div class="form-group">
          <label>仓库地址</label>
          <input v-model="registryAddress" type="text" placeholder="registry.example.com" />
        </div>

        <div class="form-group">
          <label>命名空间</label>
          <input v-model="registryNamespace" type="text" placeholder="namespace" />
        </div>
      </div>

      <div class="form-section">
        <h3>构建选项</h3>
        <div class="form-group">
          <label>
            <input v-model="pullBaseImage" type="checkbox" />
            始终拉取基础镜像
          </label>
        </div>
      </div>

      <div class="form-actions">
        <button @click="startBuild" class="btn-primary" :disabled="!canBuild">
          开始构建
        </button>
      </div>
    </div>

    <div v-if="buildLogs.length > 0" class="build-logs">
      <h3>构建日志</h3>
      <div class="logs-container">
        <div v-for="(log, i) in buildLogs" :key="i" class="log-line">{{ log }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { invoke } from '@tauri-apps/api/core'

interface Project {
  id: string
  name: string
  language: string
}

const projects = ref<Project[]>([])
const selectedProjectId = ref('')
const imageTag = ref('')
const registryType = ref('aliyun')
const registryAddress = ref('')
const registryNamespace = ref('')
const pullBaseImage = ref(true)
const buildLogs = ref<string[]>([])

const canBuild = computed(() => {
  return selectedProjectId.value && imageTag.value && registryNamespace.value
})

async function loadProjects() {
  try {
    projects.value = await invoke('list_projects')
  } catch (e) {
    console.error('Failed to load projects:', e)
  }
}

async function startBuild() {
  try {
    buildLogs.value = ['开始构建...']
    const result = await invoke('build_image', {
      request: {
        project_id: selectedProjectId.value,
        language: projects.value.find(p => p.id === selectedProjectId.value)?.language || 'custom',
        image_tag: imageTag.value,
        registry_address: registryAddress.value,
        registry_namespace: registryNamespace.value
      }
    })
    buildLogs.value.push(`构建完成: ${JSON.stringify(result)}`)
  } catch (e) {
    buildLogs.value.push(`构建失败: ${e}`)
  }
}

loadProjects()
</script>

<style scoped>
.build-view {
  max-width: 800px;
  margin: 0 auto;
}

h2 {
  margin-bottom: 0.25rem;
}

.subtitle {
  color: #666;
  margin-bottom: 2rem;
}

.build-form {
  background: white;
  border-radius: 12px;
  padding: 2rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.form-section {
  margin-bottom: 2rem;
}

.form-section h3 {
  margin-top: 0;
  color: #667eea;
  border-bottom: 1px solid #eee;
  padding-bottom: 0.5rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.form-group input[type="text"],
.form-group select {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 1rem;
}

.form-group input[type="checkbox"] {
  margin-right: 0.5rem;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
}

.btn-primary {
  background: #667eea;
  color: white;
  padding: 0.75rem 2rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
}

.btn-primary:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.build-logs {
  margin-top: 2rem;
  background: #1e1e1e;
  border-radius: 8px;
  padding: 1rem;
}

.build-logs h3 {
  color: white;
  margin-top: 0;
}

.logs-container {
  max-height: 300px;
  overflow-y: auto;
  font-family: monospace;
  font-size: 0.875rem;
}

.log-line {
  color: #0f0;
  padding: 0.25rem 0;
}
</style>
