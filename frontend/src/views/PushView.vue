<template>
  <div class="push-view">
    <h1>推送镜像</h1>

    <div class="card">
      <h3>镜像信息</h3>
      <div class="form-group">
        <label>镜像名称</label>
        <input v-model="pushConfig.image" type="text" placeholder="my-app:latest" />
      </div>
    </div>

    <div class="card">
      <div class="card-header">
        <h3>仓库认证 (可选)</h3>
        <label class="toggle">
          <input v-model="enableAuth" type="checkbox" />
          <span>需要认证</span>
        </label>
      </div>
      <template v-if="enableAuth">
        <div class="form-row">
          <div class="form-group">
            <label>仓库地址</label>
            <input v-model="pushConfig.registry" type="text" placeholder="https://index.docker.io/v1/" />
          </div>
          <div class="form-group">
            <label>用户名</label>
            <input v-model="pushConfig.username" type="text" placeholder="username" />
          </div>
        </div>
        <div class="form-group">
          <label>密码 / Token</label>
          <input v-model="pushConfig.password" type="password" placeholder="password or access token" />
        </div>
      </template>
    </div>

    <button class="btn-primary btn-large" @click="pushImage" :disabled="pushing">
      {{ pushing ? '推送中...' : '开始推送' }}
    </button>

    <div v-if="logs.length > 0" class="card">
      <h3>推送日志</h3>
      <div class="logs-container">
        <div v-for="(log, index) in logs" :key="index" class="log-line" :class="log.type">
          <span class="log-time">[{{ log.time }}]</span>
          <span class="log-text">{{ log.text }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { PushImage } from '../wailsjs/go/main/App'

const pushing = ref(false)
const enableAuth = ref(false)
const logs = ref<{ time: string; text: string; type: string }[]>([])

const pushConfig = ref({
  image: '',
  registry: '',
  username: '',
  password: ''
})

function addLog(text: string, type = 'info') {
  const time = new Date().toLocaleTimeString('zh-CN')
  logs.value.push({ time, text, type })
}

async function pushImage() {
  if (pushing.value) return
  if (!pushConfig.value.image) {
    addLog('请输入镜像名称', 'error')
    return
  }

  pushing.value = true
  logs.value = []

  addLog('开始推送...')
  addLog(`镜像: ${pushConfig.value.image}`)

  try {
    const data: Record<string, any> = {
      image: pushConfig.value.image
    }
    if (enableAuth.value) {
      data.registry = pushConfig.value.registry
      data.username = pushConfig.value.username
      data.password = pushConfig.value.password
    }

    const result = await PushImage(data)
    if (result.success) {
      addLog('推送成功！', 'success')
      if (result.log) {
        result.log.split('\n').forEach((line: string) => {
          if (line.trim()) addLog(line)
        })
      }
    } else {
      addLog(`推送失败: ${result.error}`, 'error')
    }
  } catch (e) {
    addLog(`推送失败: ${e}`, 'error')
  }

  pushing.value = false
}
</script>

<style scoped>
.push-view {
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

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.card-header h3 {
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

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
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
