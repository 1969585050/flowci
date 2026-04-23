<template>
  <div class="deploy-view">
    <h2>🚀 部署管理</h2>
    <p class="subtitle">部署应用到 Docker 环境</p>

    <div class="deploy-form">
      <div class="form-group">
        <label>选择项目</label>
        <select v-model="selectedProjectId">
          <option value="">请选择项目</option>
          <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
      </div>

      <div class="form-group">
        <label>部署类型</label>
        <select v-model="deploymentType">
          <option value="single">单个容器</option>
          <option value="compose">Docker Compose</option>
          <option value="rolling">滚动更新</option>
        </select>
      </div>

      <div class="form-group">
        <label>镜像标签</label>
        <input v-model="imageTag" type="text" placeholder="myapp:latest" />
      </div>

      <div class="form-group">
        <label>副本数</label>
        <input v-model.number="replicas" type="number" min="1" max="10" />
      </div>

      <div class="form-actions">
        <button @click="deploy" class="btn-primary">部署</button>
        <button @click="rollback" class="btn-warning">回滚</button>
      </div>
    </div>

    <div v-if="containers.length > 0" class="containers-list">
      <h3>运行中的容器</h3>
      <div v-for="c in containers" :key="c.id" class="container-card">
        <div class="container-header">
          <span class="container-name">{{ c.name }}</span>
          <span class="container-status" :class="c.status">{{ c.status }}</span>
        </div>
        <div class="container-details">
          <p>镜像: {{ c.image }}</p>
          <p>端口: {{ formatPorts(c.ports) }}</p>
        </div>
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
}

interface Container {
  id: string
  name: string
  image: string
  status: string
  ports: Array<{ host_port: number; container_port: number }>
}

const projects = ref<Project[]>([])
const containers = ref<Container[]>([])
const selectedProjectId = ref('')
const deploymentType = ref('compose')
const imageTag = ref('')
const replicas = ref(1)

async function loadProjects() {
  try {
    projects.value = await invoke('list_projects')
  } catch (e) {
    console.error('Failed to load projects:', e)
  }
}

async function loadContainers() {
  try {
    containers.value = await invoke('list_containers')
  } catch (e) {
    console.error('Failed to load containers:', e)
  }
}

async function deploy() {
  try {
    await invoke('deploy_compose', {
      request: {
        project_id: selectedProjectId.value,
        image_tag: imageTag.value,
        deployment_type: deploymentType.value,
        environment_vars: {},
        replicas: replicas.value
      }
    })
    alert('部署成功')
    await loadContainers()
  } catch (e) {
    alert('部署失败: ' + e)
  }
}

async function rollback() {
  if (!selectedProjectId.value) return
  try {
    await invoke('rollback_deploy', { projectId: selectedProjectId.value })
    alert('回滚成功')
    await loadContainers()
  } catch (e) {
    alert('回滚失败: ' + e)
  }
}

function formatPorts(ports: Array<{ host_port: number; container_port: number }>) {
  return ports.map(p => `${p.host_port}:${p.container_port}`).join(', ') || '无'
}

onMounted(() => {
  loadProjects()
  loadContainers()
})
</script>

<style scoped>
.deploy-view {
  max-width: 800px;
  margin: 0 auto;
}

.subtitle {
  color: #666;
  margin-bottom: 2rem;
}

.deploy-form {
  background: white;
  border-radius: 12px;
  padding: 2rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin-bottom: 2rem;
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

.form-actions {
  display: flex;
  gap: 1rem;
  margin-top: 1.5rem;
}

.btn-primary {
  background: #667eea;
  color: white;
  padding: 0.75rem 2rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.btn-warning {
  background: #f6993f;
  color: white;
  padding: 0.75rem 2rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.containers-list {
  background: white;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.container-card {
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 1rem;
}

.container-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.container-name {
  font-weight: 600;
}

.container-status {
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.875rem;
}

.container-status.running {
  background: #c6f6d5;
  color: #276749;
}

.container-status.exited {
  background: #fed7d7;
  color: #c53030;
}

.container-details p {
  margin: 0.25rem 0;
  color: #666;
  font-size: 0.875rem;
}
</style>
