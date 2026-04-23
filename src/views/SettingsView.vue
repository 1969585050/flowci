<template>
  <div class="settings-view">
    <h2>⚙️ 设置</h2>

    <div class="settings-section">
      <h3>🐳 Docker 配置</h3>
      <div class="setting-item">
        <span>Docker 状态:</span>
        <span :class="{ connected: dockerStatus.connected }">
          {{ dockerStatus.connected ? '✅ 已连接' : '❌ 未连接' }}
        </span>
      </div>
    </div>

    <div class="settings-section">
      <h3>📦 镜像仓库</h3>
      <button @click="showAddRegistry = true" class="btn-primary">+ 添加仓库</button>

      <div v-if="registries.length > 0" class="registries-list">
        <div v-for="r in registries" :key="r.id" class="registry-item">
          <span class="registry-name">{{ r.name }}</span>
          <span class="registry-type">{{ r.registry_type }}</span>
          <span class="registry-address">{{ r.address }}</span>
        </div>
      </div>
    </div>

    <div class="settings-section">
      <h3>🔑 凭证管理</h3>
      <button @click="showAddCredentials = true" class="btn-primary">+ 添加凭证</button>
    </div>

    <div v-if="showAddRegistry" class="dialog-overlay" @click.self="showAddRegistry = false">
      <div class="dialog">
        <h3>添加镜像仓库</h3>
        <form @submit.prevent="addRegistry">
          <div class="form-group">
            <label>仓库名称</label>
            <input v-model="newRegistry.name" type="text" required />
          </div>
          <div class="form-group">
            <label>仓库类型</label>
            <select v-model="newRegistry.type">
              <option value="aliyun">阿里云 ACR</option>
              <option value="dockerhub">Docker Hub</option>
              <option value="harbor">Harbor</option>
            </select>
          </div>
          <div class="form-group">
            <label>地址</label>
            <input v-model="newRegistry.address" type="text" required />
          </div>
          <div class="form-group">
            <label>命名空间</label>
            <input v-model="newRegistry.namespace" type="text" required />
          </div>
          <div class="dialog-actions">
            <button type="button" @click="showAddRegistry = false">取消</button>
            <button type="submit" class="btn-primary">添加</button>
          </div>
        </form>
      </div>
    </div>

    <div v-if="showAddCredentials" class="dialog-overlay" @click.self="showAddCredentials = false">
      <div class="dialog">
        <h3>添加凭证</h3>
        <form @submit.prevent="addCredentials">
          <div class="form-group">
            <label>凭证名称</label>
            <input v-model="newCredentials.name" type="text" required />
          </div>
          <div class="form-group">
            <label>用户名</label>
            <input v-model="newCredentials.username" type="text" required />
          </div>
          <div class="form-group">
            <label>密码/Token</label>
            <input v-model="newCredentials.password" type="password" required />
          </div>
          <div class="dialog-actions">
            <button type="button" @click="showAddCredentials = false">取消</button>
            <button type="submit" class="btn-primary">添加</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { invoke } from '@tauri-apps/api/core'

interface Registry {
  id: string
  name: string
  registry_type: string
  address: string
}

interface DockerStatus {
  connected: boolean
}

const dockerStatus = ref<DockerStatus>({ connected: false })
const registries = ref<Registry[]>([])
const showAddRegistry = ref(false)
const showAddCredentials = ref(false)

const newRegistry = ref({
  name: '',
  type: 'aliyun',
  address: '',
  namespace: ''
})

const newCredentials = ref({
  name: '',
  username: '',
  password: ''
})

onMounted(async () => {
  try {
    dockerStatus.value = await invoke('check_docker')
  } catch (e) {
    console.error('Failed to check docker:', e)
  }
})

async function addRegistry() {
  console.log('Add registry:', newRegistry.value)
  showAddRegistry.value = false
}

async function addCredentials() {
  console.log('Add credentials:', newCredentials.value)
  showAddCredentials.value = false
}
</script>

<style scoped>
.settings-view {
  max-width: 800px;
  margin: 0 auto;
}

.settings-section {
  background: white;
  border-radius: 12px;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.settings-section h3 {
  margin-top: 0;
  color: #667eea;
}

.setting-item {
  display: flex;
  justify-content: space-between;
  padding: 0.75rem 0;
  border-bottom: 1px solid #eee;
}

.setting-item:last-child {
  border-bottom: none;
}

.registries-list {
  margin-top: 1rem;
}

.registry-item {
  display: flex;
  gap: 1rem;
  padding: 0.75rem;
  border: 1px solid #eee;
  border-radius: 6px;
  margin-bottom: 0.5rem;
}

.registry-name {
  font-weight: 600;
}

.registry-type {
  color: #667eea;
}

.registry-address {
  color: #666;
}

.btn-primary {
  background: #667eea;
  color: white;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
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
  max-width: 400px;
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
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1.5rem;
}
</style>
