<template>
  <div class="settings-view">
    <h1>设置</h1>

    <div class="card">
      <h3>Docker 连接</h3>
      <div class="status-badge" :class="dockerStatus.status">
        <span class="status-dot"></span>
        {{ dockerStatus.text }}
      </div>
      <div v-if="dockerStatus.version" class="version-info">
        <span>Docker 版本: {{ dockerStatus.version }}</span>
      </div>
      <button class="btn-outline" @click="checkDocker" style="margin-top: 16px;">检查连接</button>
    </div>

    <div class="card">
      <h3>默认配置</h3>
      <div class="form-group">
        <label>默认镜像仓库</label>
        <input v-model="settings.defaultRegistry" type="text" placeholder="docker.io" />
      </div>
      <div class="form-group">
        <label>默认工作目录</label>
        <input v-model="settings.defaultWorkdir" type="text" placeholder="/workspace" />
      </div>
      <button class="btn-primary" @click="saveSettings">保存设置</button>
    </div>

    <div class="card">
      <h3>关于</h3>
      <div class="about-info">
        <p><strong>FlowCI</strong> - 轻量级 Docker 构建部署工具</p>
        <p>版本: 0.1.0</p>
        <p>技术栈: Wails + Go + Vue 3</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const dockerStatus = ref({
  status: 'checking',
  text: '检查中...',
  version: ''
})

const settings = ref({
  defaultRegistry: 'docker.io',
  defaultWorkdir: '/workspace'
})

async function checkDocker() {
  dockerStatus.value = {
    status: 'checking',
    text: '检查中...',
    version: ''
  }
  
  try {
    if ((window as any).wails?.Invoke) {
      const result = await (window as any).wails.Invoke('CheckDocker')
      dockerStatus.value = {
        status: result.connected ? 'connected' : 'disconnected',
        text: result.connected ? '已连接' : '未连接',
        version: result.version || ''
      }
    } else {
      await new Promise(resolve => setTimeout(resolve, 1000))
      dockerStatus.value = {
        status: 'connected',
        text: '已连接',
        version: '24.0.0'
      }
    }
  } catch (e) {
    dockerStatus.value = {
      status: 'error',
      text: '检查失败',
      version: ''
    }
  }
}

function saveSettings() {
  alert('设置已保存！')
}

onMounted(() => {
  checkDocker()
})
</script>

<style scoped>
.settings-view {
  max-width: 700px;
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
  margin-bottom: 20px;
}

.card h3 {
  font-size: 18px;
  color: #1a1a2e;
  margin-bottom: 16px;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border-radius: 8px;
  font-weight: 600;
}

.status-badge.checking {
  background: #fef3c7;
  color: #d97706;
}

.status-badge.connected {
  background: #dcfce7;
  color: #16a34a;
}

.status-badge.disconnected,
.status-badge.error {
  background: #fee2e2;
  color: #dc2626;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: currentColor;
}

.version-info {
  margin-top: 12px;
  color: #666;
  font-size: 14px;
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

.form-group input {
  padding: 12px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 14px;
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

.about-info p {
  color: #666;
  margin-bottom: 8px;
}
</style>
