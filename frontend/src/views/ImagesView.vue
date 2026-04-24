<template>
  <div class="images-view">
    <div class="page-header">
      <h2>镜像管理</h2>
      <div class="header-actions">
        <button class="btn btn-secondary" @click="refreshImages" :disabled="loading">
          {{ loading ? '加载中...' : '刷新' }}
        </button>
      </div>
    </div>

    <div v-if="loading && images.length === 0" class="loading">
      加载镜像列表...
    </div>

    <div v-else-if="images.length === 0" class="empty">
      暂无 Docker 镜像。请先在构建页面构建镜像。
    </div>

    <div v-else class="image-grid">
      <div v-for="image in images" :key="image.id" class="image-card card">
        <div class="image-card-header">
          <div class="image-name">
            <span class="img-icon">🗃️</span>
            <div class="name-detail">
              <span class="repository">{{ image.repository }}</span>
              <span class="tag">:{{ image.tag }}</span>
            </div>
          </div>
          <span class="image-size">{{ image.size }}</span>
        </div>
        <div class="image-card-body">
          <div class="detail-row">
            <span class="label">镜像 ID</span>
            <span class="value id-value">{{ shortId(image.id) }}</span>
          </div>
          <div class="detail-row">
            <span class="label">创建时间</span>
            <span class="value">{{ image.created_at }}</span>
          </div>
        </div>
        <div class="image-card-footer">
          <button class="btn btn-danger" @click="removeImage(image)">删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, inject, onMounted } from 'vue'
import { ListImages, RemoveImage } from '../wailsjs/go/main/App'

const toast = inject('toast') as { success: (msg: string) => void; error: (msg: string) => void; info: (msg: string) => void }

interface Image {
  id: string
  repository: string
  tag: string
  size: string
  created_at: string
}

const images = ref<Image[]>([])
const loading = ref(false)

function shortId(id: string) {
  return id.length > 12 ? id.substring(0, 12) : id
}

async function refreshImages() {
  loading.value = true
  try {
    images.value = await ListImages()
  } catch (e) {
    console.error('Failed to load images:', e)
    toast?.error('加载镜像列表失败')
  }
  loading.value = false
}

async function removeImage(image: Image) {
  const tag = `${image.repository}:${image.tag}`
  if (!confirm(`确定要删除镜像 ${tag} 吗？`)) return
  try {
    const result = await RemoveImage(image.id)
    if (result.success) {
      toast?.success(`镜像 ${tag} 已删除`)
      await refreshImages()
    } else {
      toast?.error(`${result.error || '删除失败'}`)
    }
  } catch (e) {
    toast?.error(`删除失败: ${e}`)
  }
}

onMounted(refreshImages)
</script>

<style scoped>
.images-view {
  max-width: 1000px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.page-header h2 {
  font-size: 24px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.loading,
.empty {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-secondary, #999);
  font-size: 16px;
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 16px;
}

.image-card {
  display: flex;
  flex-direction: column;
}

.image-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color, #f0f0f0);
}

.image-name {
  display: flex;
  align-items: center;
  gap: 10px;
}

.img-icon {
  font-size: 24px;
}

.name-detail {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.repository {
  font-weight: 600;
  font-size: 15px;
}

.tag {
  color: #667eea;
  font-weight: 500;
  font-size: 14px;
}

.image-size {
  color: var(--text-secondary, #999);
  font-size: 13px;
  white-space: nowrap;
}

.image-card-body {
  padding: 12px 20px;
  flex: 1;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  font-size: 14px;
}

.detail-row .label {
  color: var(--text-secondary, #999);
}

.detail-row .value {
  color: var(--text-primary, #555);
}

.id-value {
  font-family: monospace;
}

.image-card-footer {
  padding: 12px 20px;
  border-top: 1px solid var(--border-color, #f0f0f0);
  display: flex;
  justify-content: flex-end;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: var(--bg-primary, #f0f0f0);
  color: var(--text-primary, #333);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--border-color, #e0e0e0);
}

.btn-danger {
  background: var(--card-bg, #fff);
  color: #e74c3c;
  border: 1px solid #e74c3c;
}

.btn-danger:hover {
  background: #e74c3c;
  color: #fff;
}
</style>
