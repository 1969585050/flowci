<template>
  <div class="build-detail-view">
    <div class="header">
      <div class="header-left">
        <button class="btn-back" @click="goBack">← 返回</button>
        <h1>构建详情</h1>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <span>加载中...</span>
    </div>

    <template v-else-if="record">
      <div class="info-grid">
        <div class="info-card">
          <div class="info-label">镜像名称</div>
          <div class="info-value">{{ record.imageName }}:{{ record.imageTag }}</div>
        </div>
        <div class="info-card">
          <div class="info-label">构建状态</div>
          <div class="info-value">
            <span :class="['status-badge', record.status]">
              {{ statusText(record.status) }}
            </span>
          </div>
        </div>
        <div class="info-card">
          <div class="info-label">开始时间</div>
          <div class="info-value">{{ formatDate(record.startedAt) }}</div>
        </div>
        <div class="info-card" v-if="record.finishedAt">
          <div class="info-label">完成时间</div>
          <div class="info-value">{{ formatDate(record.finishedAt) }}</div>
        </div>
        <div class="info-card" v-if="record.startedAt && record.finishedAt">
          <div class="info-label">耗时</div>
          <div class="info-value">{{ calcDuration(record.startedAt, record.finishedAt) }}</div>
        </div>
      </div>

      <div class="log-section">
        <h3>构建日志</h3>
        <div class="log-container">
          <pre>{{ record.log || '(无日志输出)' }}</pre>
        </div>
      </div>
    </template>

    <div v-else-if="!loading" class="empty-state">
      <div class="empty-icon">🔍</div>
      <h3>未找到构建记录</h3>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { GetBuildRecord } from '../wailsjs/go/handler/App'

interface BuildRecord {
  id: string
  projectId: string
  imageName: string
  imageTag: string
  status: string
  log: string
  startedAt: string
  finishedAt?: string
}

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const record = ref<BuildRecord | null>(null)

function statusText(status: string) {
  const map: Record<string, string> = {
    building: '构建中',
    success: '成功',
    failed: '失败'
  }
  return map[status] || status
}

function formatDate(date: string) {
  return new Date(date).toLocaleString('zh-CN')
}

function calcDuration(start: string, end: string) {
  const ms = new Date(end).getTime() - new Date(start).getTime()
  const secs = Math.floor(ms / 1000)
  if (secs < 60) return `${secs}秒`
  const mins = Math.floor(secs / 60)
  const remainSecs = secs % 60
  return `${mins}分${remainSecs}秒`
}

function goBack() {
  router.back()
}

async function loadRecord(recordId: string) {
  loading.value = true
  try {
    record.value = await GetBuildRecord(recordId)
  } catch (e) {
    console.error('Failed to load build record:', e)
  }
  loading.value = false
}

onMounted(() => {
  const recordId = route.query.recordId as string
  if (recordId) {
    loadRecord(recordId)
  }
})
</script>

<style scoped>
.build-detail-view {
  max-width: 960px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-left h1 {
  font-size: 24px;
  color: var(--text-primary, #1a1a2e);
  margin: 0;
}

.btn-back {
  background: var(--bg-primary, #f0f0f0);
  color: var(--text-primary, #333);
  border: none;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-back:hover {
  background: var(--border-color, #e0e0e0);
}

.loading, .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  gap: 16px;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid var(--border-color, #e0e0e0);
  border-top-color: #667eea;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-icon {
  font-size: 64px;
}

.empty-state h3 {
  font-size: 20px;
  color: var(--text-secondary, #666);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.info-card {
  background: var(--card-bg, white);
  border-radius: 12px;
  padding: 16px;
  box-shadow: var(--shadow, 0 2px 12px rgba(0, 0, 0, 0.05));
}

.info-label {
  font-size: 12px;
  color: var(--text-secondary, #999);
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.info-value {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary, #1a1a2e);
}

.status-badge {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 600;
  display: inline-block;
}

.status-badge.building {
  background: #fef3c7;
  color: #d97706;
}

.status-badge.success {
  background: #d1fae5;
  color: #059669;
}

.status-badge.failed {
  background: #fee2e2;
  color: #dc2626;
}

.log-section {
  background: var(--card-bg, white);
  border-radius: 12px;
  padding: 20px;
  box-shadow: var(--shadow, 0 2px 12px rgba(0, 0, 0, 0.05));
}

.log-section h3 {
  font-size: 16px;
  color: var(--text-primary, #1a1a2e);
  margin: 0 0 12px 0;
}

.log-container {
  background: #1e1e2e;
  color: #cdd6f4;
  border-radius: 8px;
  padding: 16px;
  font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace;
  font-size: 13px;
  line-height: 1.6;
  max-height: 500px;
  overflow: auto;
}

.log-container pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
