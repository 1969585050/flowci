<template>
  <div class="build-history-view">
    <div class="header">
      <div class="header-left">
        <button class="btn-back" @click="goBack">← 返回</button>
        <h1 v-if="project">构建历史 - {{ project.name }}</h1>
        <h1 v-else>构建历史</h1>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <span>加载中...</span>
    </div>

    <div v-else-if="records.length === 0" class="empty-state">
      <div class="empty-icon">🏗️</div>
      <h3>暂无构建记录</h3>
      <p>该项目还没有进行过构建</p>
    </div>

    <div v-else class="records-list">
      <div
        v-for="record in records"
        :key="record.id"
        class="record-card"
        @click="viewDetail(record)"
      >
        <div class="record-main">
          <div class="record-info">
            <div class="record-image">
              {{ record.imageName }}<span class="tag">:{{ record.imageTag }}</span>
            </div>
            <div class="record-time">
              <span>{{ formatDate(record.startedAt) }}</span>
              <span v-if="record.finishedAt" class="duration">
                ({{ calcDuration(record.startedAt, record.finishedAt) }})
              </span>
            </div>
          </div>
          <span :class="['status-badge', record.status]">
            {{ statusText(record.status) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ListProjects, ListBuildRecords } from '../wailsjs/go/handler/App'

interface BuildRecord {
  id: string
  projectId: string
  imageName: string
  imageTag: string
  status: string
  startedAt: string
  finishedAt?: string
}

interface Project {
  id: string
  name: string
}

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const records = ref<BuildRecord[]>([])
const project = ref<Project | null>(null)

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
  router.push('/projects')
}

function viewDetail(record: BuildRecord) {
  router.push({ path: '/build-detail', query: { recordId: record.id } })
}

async function loadProject(projectId: string) {
  try {
    const projects = await ListProjects() as Project[]
    const found = projects.find((p: Project) => p.id === projectId)
    if (found) project.value = found
  } catch (e) {
    console.error('Failed to load project:', e)
  }
}

async function loadRecords(projectId: string) {
  loading.value = true
  try {
    records.value = await ListBuildRecords(projectId) as BuildRecord[]
  } catch (e) {
    console.error('Failed to load build records:', e)
  }
  loading.value = false
}

onMounted(() => {
  const projectId = route.query.projectId as string
  if (projectId) {
    loadProject(projectId)
    loadRecords(projectId)
  }
})
</script>

<style scoped>
.build-history-view {
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
  margin: 0;
}

.empty-state p {
  color: var(--text-secondary, #999);
  margin: 0;
}

.records-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.record-card {
  background: var(--card-bg, white);
  border-radius: 12px;
  padding: 16px 20px;
  box-shadow: var(--shadow, 0 2px 12px rgba(0, 0, 0, 0.05));
  cursor: pointer;
  transition: all 0.2s;
}

.record-card:hover {
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  transform: translateX(4px);
}

.record-main {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.record-info {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.record-image {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #1a1a2e);
}

.record-image .tag {
  color: #667eea;
  font-weight: 400;
}

.record-time {
  font-size: 13px;
  color: var(--text-secondary, #999);
  display: flex;
  gap: 8px;
  align-items: center;
}

.duration {
  color: var(--text-secondary, #666);
}

.status-badge {
  padding: 6px 16px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;
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
</style>
