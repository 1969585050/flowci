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

    <div class="card">
      <div class="card-header-compact">
        <h3>Docker Compose</h3>
        <label class="toggle">
          <input v-model="useCompose" type="checkbox" />
          <span>使用 docker-compose 部署</span>
        </label>
      </div>
      <div v-if="useCompose" class="compose-section">
        <div class="form-group">
          <label>工作目录 (docker-compose.yml 保存位置)</label>
          <input v-model="composeWorkdir" type="text" placeholder="." />
        </div>
        <div class="form-group">
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <label>docker-compose.yml 预览</label>
            <button class="btn-small" @click="generateCompose">重新生成</button>
          </div>
          <div class="compose-preview">
            <pre>{{ composePreview || '点击"重新生成"生成 Compose 文件' }}</pre>
          </div>
        </div>
      </div>
    </div>

    <button class="btn-primary btn-large" @click="deployContainer" :disabled="deploying">
      {{ deploying ? '部署中...' : (useCompose ? '使用 Compose 部署' : '开始部署') }}
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
            <button class="btn-action" @click="viewLogs(container)">日志</button>
            <button v-if="container.status === 'running'" class="btn-action" @click="stopContainer(container.id)">停止</button>
            <button v-else class="btn-action" @click="startContainer(container.id)">启动</button>
            <button class="btn-action btn-danger" @click="removeContainer(container.id)">删除</button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showLogDialog" class="dialog-overlay" @click.self="showLogDialog = false">
      <div class="dialog log-dialog">
        <div class="dialog-header">
          <h3>容器日志 - {{ logContainer.name }}</h3>
          <button class="close-btn" @click="showLogDialog = false">&times;</button>
        </div>
        <div class="dialog-body">
          <div v-if="logLoading" class="loading-inline">
            <div class="spinner-small"></div>
            <span>加载日志中...</span>
          </div>
          <pre v-else class="log-content">{{ logContent || '(日志为空)' }}</pre>
        </div>
        <div class="dialog-footer">
          <button class="btn-outline" @click="showLogDialog = false">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, inject, onMounted } from 'vue'
import { DeployContainer, ListContainers, StartContainer, StopContainer, RemoveContainer, GetContainerLogs, GenerateCompose, DeployWithCompose } from '../wailsjs/go/handler/App'
import { useConfirm } from '../composables/useConfirm'

const { ask } = useConfirm()

interface Container {
  id: string
  name: string
  image: string
  status: 'running' | 'stopped' | 'exited'
  statusText: string
}

const toast = inject('toast') as { success: (msg: string) => void; error: (msg: string) => void; info: (msg: string) => void }

const deploying = ref(false)
const containersLoading = ref(false)
const containers = ref<Container[]>([])

const deployConfig = ref({
  image: 'my-app:latest',
  name: 'my-app-container',
  // 端口以字符串存储，后端 DTO 字段为 string；模板表单用 type="number" 仍可正常输入
  hostPort: '8080',
  containerPort: '3000',
  restartPolicy: 'unless-stopped',
  env: 'NODE_ENV=production\nPORT=3000'
})

const useCompose = ref(false)
const composePreview = ref('')
const composeWorkdir = ref('')

const showLogDialog = ref(false)
const logContainer = ref({ name: '', id: '' })
const logContent = ref('')
const logLoading = ref(false)

async function deployContainer() {
  if (deploying.value) return
  deploying.value = true

  try {
    if (useCompose.value) {
      await DeployWithCompose({
        compose: composePreview.value,
        workdir: composeWorkdir.value || '.'
      })
      toast?.success('Docker Compose 部署成功！')
    } else {
      await DeployContainer(deployConfig.value)
      toast?.success('部署成功！')
    }
    await refreshContainers()
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast?.error(`部署失败: ${msg}`)
  }

  deploying.value = false
}

async function generateCompose() {
  try {
    composePreview.value = await GenerateCompose(deployConfig.value)
  } catch (e) {
    toast?.error(`生成 Compose 文件失败: ${e}`)
  }
}

async function refreshContainers() {
  containersLoading.value = true
  try {
    containers.value = await ListContainers()
  } catch (e) {
    console.error('Failed to load containers:', e)
  }
  containersLoading.value = false
}

async function startContainer(id: string) {
  try {
    await StartContainer(id)
    toast?.success('容器已启动')
    await refreshContainers()
  } catch (e) {
    toast?.error(`启动失败: ${e}`)
  }
}

async function stopContainer(id: string) {
  try {
    await StopContainer(id)
    toast?.success('容器已停止')
    await refreshContainers()
  } catch (e) {
    toast?.error(`停止失败: ${e}`)
  }
}

async function removeContainer(id: string) {
  const ok = await ask({
    title: '删除容器',
    message: '确定要删除此容器吗？正在运行的容器会被强制停止并删除。',
    variant: 'danger',
    confirmText: '删除',
  })
  if (!ok) return
  try {
    await RemoveContainer(id)
    toast?.success('容器已删除')
    await refreshContainers()
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    toast?.error(`删除失败: ${msg}`)
  }
}

async function viewLogs(container: Container) {
  logContainer.value = { name: container.name, id: container.id }
  logContent.value = ''
  showLogDialog.value = true
  logLoading.value = true
  try {
    logContent.value = await GetContainerLogs(container.id, 200)
  } catch (e) {
    logContent.value = `获取日志失败: ${e}`
  }
  logLoading.value = false
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
  color: var(--text-primary, #1a1a2e);
  margin-bottom: 24px;
}

.card {
  background: var(--card-bg, white);
  border-radius: 12px;
  padding: 24px;
  box-shadow: var(--shadow, 0 2px 12px rgba(0, 0, 0, 0.05));
}

.card h3 {
  font-size: 18px;
  color: var(--text-primary, #1a1a2e);
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
  color: var(--text-primary, #333);
}

.form-group input,
.form-group select,
.form-group textarea {
  padding: 12px;
  border: 2px solid var(--border-color, #e0e0e0);
  border-radius: 8px;
  font-size: 14px;
  background: var(--card-bg, white);
  color: var(--text-primary, #333);
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
  color: var(--text-secondary, #999);
}

.spinner-small {
  width: 24px;
  height: 24px;
  border: 3px solid var(--border-color, #e0e0e0);
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
  background: var(--bg-primary, #f8f9ff);
  border-radius: 8px;
  margin-bottom: 12px;
}

.container-info {
  flex: 1;
}

.container-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #1a1a2e);
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
  color: var(--text-secondary, #666);
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
  background: var(--border-color, #e0e0e0);
  color: var(--text-primary, #333);
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

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: var(--card-bg, white);
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
  display: flex;
  flex-direction: column;
}

.log-dialog {
  width: 700px;
  max-width: 90vw;
  max-height: 80vh;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color, #f0f0f0);
}

.dialog-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #1a1a2e);
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  color: var(--text-secondary, #999);
  cursor: pointer;
  padding: 0;
  line-height: 1;
}

.close-btn:hover {
  color: var(--text-primary, #333);
}

.dialog-body {
  padding: 16px 24px;
  flex: 1;
  overflow: auto;
  min-height: 200px;
  max-height: 50vh;
}

.log-content {
  background: #1e1e2e;
  color: #cdd6f4;
  padding: 16px;
  border-radius: 8px;
  font-size: 13px;
  line-height: 1.5;
  font-family: 'Cascadia Code', 'Fira Code', 'Consolas', monospace;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  padding: 12px 24px;
  border-top: 1px solid var(--border-color, #f0f0f0);
}

.card-header-compact {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.card-header-compact h3 {
  margin-bottom: 0;
}

.toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: var(--text-secondary, #666);
  cursor: pointer;
}

.toggle input {
  width: 18px;
  height: 18px;
}

.compose-section {
  margin-top: 16px;
}

.compose-preview {
  background: #1e1e2e;
  border-radius: 8px;
  padding: 16px;
  margin-top: 8px;
}

.compose-preview pre {
  color: #a5b6cf;
  font-family: 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.5;
  white-space: pre-wrap;
  margin: 0;
}

.btn-small {
  padding: 6px 12px;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  transition: all 0.2s;
}

.btn-small:hover {
  transform: translateY(-1px);
}
</style>
