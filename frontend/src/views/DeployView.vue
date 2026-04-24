<template>
  <div class="deploy-view">
    <h1>容器部署</h1>

    <div class="card">
      <h3>部署配置</h3>
      <div class="form-row">
        <div class="form-group">
          <label>镜像名称</label>
          <input v-model="deployConfig.image" type="text" placeholder="my-app:latest" />
        </div>
        <div class="form-group">
          <label>容器名称</label>
          <input v-model="deployConfig.name" type="text" placeholder="my-app-container" />
        </div>
      </div>
      <div class="form-row">
        <div class="form-group">
          <label>主机端口</label>
          <input v-model="deployConfig.hostPort" type="number" placeholder="8080" />
        </div>
        <div class="form-group">
          <label>容器端口</label>
          <input v-model="deployConfig.containerPort" type="number" placeholder="3000" />
        </div>
      </div>
      <div class="form-group">
        <label>重启策略</label>
        <select v-model="deployConfig.restartPolicy">
          <option value="no">不重启</option>
          <option value="always">总是重启</option>
          <option value="on-failure">失败时重启</option>
          <option value="unless-stopped">除非停止否则重启</option>
        </select>
      </div>
      <div class="form-group">
        <label>环境变量 (每行一个，格式: KEY=VALUE)</label>
        <textarea v-model="deployConfig.env" rows="4" placeholder="NODE_ENV=production&#10;PORT=3000"></textarea>
      </div>
    </div>

    <button class="btn-primary btn-large" @click="deployContainer" :disabled="deploying">
      {{ deploying ? '部署中...' : '开始部署' }}
    </button>

    <div class="card" style="margin-top: 24px;">
      <div style="display: flex; justify-content: space-between; align-items: center;">
        <h3>运行中的容器</h3>
        <button class="btn-outline" @click="refreshContainers">刷新</button>
      </div>
      
      <div v-if="containersLoading" class="loading-inline">
        <div class="spinner-small"></div>
        <span>加载中...</span>
      </div>
      
      <div v-else-if="containers.length === 0" class="empty-state-inline">
        暂无运行中的容器
      </div>

      <div v-else class="container-list">
        <div v-for="container in containers" :key="container.id" class="container-item">
          <div class="container-info">
            <div class="container-name">{{ container.name }}</div>
            <div class="container-status" :class="container.status">
              <span class="status-dot"></span>
              {{ container.statusText }}
            </div>
            <div class="container-image">{{ container.image }}</div>
          </div>
          <div class="container-actions">
            <button v-if="container.status === 'running'" class="btn-action" @click="stopContainer(container.id)">停止</button>
            <button v-else class="btn-action" @click="startContainer(container.id)">启动</button>
            <button class="btn-action btn-danger" @click="removeContainer(container.id)">删除</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface Container {
  id: string
  name: string
  image: string
  status: 'running' | 'stopped' | 'exited'
  statusText: string
}

const deploying = ref(false)
const containersLoading = ref(false)
const containers = ref<Container[]>([])

const deployConfig = ref({
  image: 'my-app:latest',
  name: 'my-app-container',
  hostPort: 8080,
  containerPort: 3000,
  restartPolicy: 'unless-stopped',
  env: 'NODE_ENV=production\nPORT=3000'
})

async function deployContainer() {
  if (deploying.value) return
  deploying.value = true
  
  try {
    if ((window as any).wails?.Invoke) {
      await (window as any).wails.Invoke('DeployContainer', deployConfig.value)
    } else {
      await new Promise(resolve => setTimeout(resolve, 2000))
    }
    alert('部署成功！')
    await refreshContainers()
  } catch (e) {
    alert(`部署失败: ${e}`)
  }
  
  deploying.value = false
}

async function refreshContainers() {
  containersLoading.value = true
  try {
    if ((window as any).wails?.Invoke) {
      containers.value = await (window as any).wails.Invoke('ListContainers')
    } else {
      containers.value = [
        { id: '1', name: 'my-app-1', image: 'my-app:latest', status: 'running', statusText: '运行中' },
        { id: '2', name: 'my-db', image: 'postgres:15', status: 'running', statusText: '运行中' }
      ]
    }
  } catch (e) {
    console.error('Failed to load containers:', e)
  }
  containersLoading.value = false
}

async function startContainer(id: string) {
  try {
    if ((window as any).wails?.Invoke) {
      await (window as any).wails.Invoke('StartContainer', id)
    }
    await refreshContainers()
  } catch (e) {
    alert(`启动失败: ${e}`)
  }
}

async function stopContainer(id: string) {
  try {
    if ((window as any).wails?.Invoke) {
      await (window as any).wails.Invoke('StopContainer', id)
    }
    await refreshContainers()
  } catch (e) {
    alert(`停止失败: ${e}`)
  }
}

async function removeContainer(id: string) {
  if (!confirm('确定要删除此容器吗？')) return
  try {
    if ((window as any).wails?.Invoke) {
      await (window as any).wails.Invoke('RemoveContainer', id)
    }
    await refreshContainers()
  } catch (e) {
    alert(`删除失败: ${e}`)
  }
}

onMounted(() => {
  refreshContainers()
})
</script>

<style scoped>
.deploy-view {
  max-width: 900px;
}

h1 {
  font-size: 28px;
  color: #1a1a2e;
  margin-bottom: 24px;
}

.card {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
}

.card h3 {
  font-size: 18px;
  color: #1a1a2e;
  margin-bottom: 16px;
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
  margin-bottom: 16px;
}

.form-group label {
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.form-group input,
.form-group select,
.form-group textarea {
  padding: 12px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 14px;
  transition: border-color 0.2s;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #667eea;
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

.btn-large {
  width: 100%;
  padding: 16px;
  font-size: 16px;
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

.loading-inline,
.empty-state-inline {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #999;
}

.spinner-small {
  width: 24px;
  height: 24px;
  border: 3px solid #e0e0e0;
  border-top-color: #667eea;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-right: 12px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.container-list {
  margin-top: 16px;
}

.container-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: #f8f9ff;
  border-radius: 8px;
  margin-bottom: 12px;
}

.container-info {
  flex: 1;
}

.container-name {
  font-size: 16px;
  font-weight: 600;
  color: #1a1a2e;
  margin-bottom: 4px;
}

.container-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 4px;
}

.container-status.running {
  color: #22c55e;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #22c55e;
}

.container-image {
  font-size: 13px;
  color: #666;
}

.container-actions {
  display: flex;
  gap: 8px;
}

.btn-action {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  background: #e0e0e0;
  color: #333;
}

.btn-action:hover {
  background: #d0d0d0;
}

.btn-danger {
  background: #fee2e2;
  color: #dc2626;
}

.btn-danger:hover {
  background: #fecaca;
}
</style>
