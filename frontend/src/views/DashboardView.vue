<template>
  <div class="dashboard">
    <div class="header">
      <div>
        <h1>仪表盘</h1>
        <p class="subtitle">{{ greeting }}，FlowCI 在你身边。</p>
      </div>
      <button class="btn-outline" :disabled="loading" @click="refresh">
        {{ loading ? '刷新中…' : '🔄 刷新' }}
      </button>
    </div>

    <!-- KPI 大数字卡片 -->
    <div class="kpi-grid">
      <router-link to="/projects" class="kpi-card">
        <div class="kpi-icon">📦</div>
        <div class="kpi-body">
          <div class="kpi-value">{{ stats?.projects ?? '-' }}</div>
          <div class="kpi-label">项目</div>
          <div v-if="stats?.gitProjects" class="kpi-sub">含 {{ stats.gitProjects }} 个 Git</div>
        </div>
      </router-link>

      <router-link to="/pipelines" class="kpi-card">
        <div class="kpi-icon">🔧</div>
        <div class="kpi-body">
          <div class="kpi-value">{{ stats?.pipelines ?? '-' }}</div>
          <div class="kpi-label">流水线</div>
        </div>
      </router-link>

      <router-link to="/deploy" class="kpi-card">
        <div class="kpi-icon">🌐</div>
        <div class="kpi-body">
          <div class="kpi-value">{{ stats?.containers?.total ?? '-' }}</div>
          <div class="kpi-label">容器</div>
          <div v-if="stats?.containers" class="kpi-sub">
            <span class="dot-running"></span>{{ stats.containers.running }} 运行
            <span class="dot-stopped" style="margin-left: 8px;"></span>{{ stats.containers.stopped }} 停止
          </div>
        </div>
      </router-link>

      <router-link to="/images" class="kpi-card">
        <div class="kpi-icon">🗃️</div>
        <div class="kpi-body">
          <div class="kpi-value">{{ stats?.images ?? '-' }}</div>
          <div class="kpi-label">镜像</div>
        </div>
      </router-link>
    </div>

    <div class="row-grid">
      <!-- Docker 状态卡 -->
      <div class="card docker-card" :class="dockerClass">
        <div class="card-head">
          <h3>🐳 Docker 状态</h3>
        </div>
        <div v-if="loading && !stats" class="skel-line lg"></div>
        <template v-else-if="stats?.docker?.connected">
          <div class="docker-ok">
            <div class="docker-pulse"></div>
            <div>
              <div class="docker-status">已连接</div>
              <div class="docker-version">v{{ stats.docker.version }}</div>
            </div>
          </div>
        </template>
        <template v-else>
          <div class="docker-fail">
            <div class="docker-icon">⚠️</div>
            <div>
              <div class="docker-status">未连接</div>
              <router-link to="/settings" class="docker-link">前往 设置 → Docker 检查 →</router-link>
            </div>
          </div>
        </template>
      </div>

      <!-- 24h 构建摘要 -->
      <div class="card">
        <div class="card-head">
          <h3>📊 最近 24 小时构建</h3>
        </div>
        <div class="build-summary">
          <div class="bs-item ok">
            <div class="bs-num">{{ stats?.buildSummary?.success ?? 0 }}</div>
            <div class="bs-label">✅ 成功</div>
          </div>
          <div class="bs-item fail">
            <div class="bs-num">{{ stats?.buildSummary?.failed ?? 0 }}</div>
            <div class="bs-label">❌ 失败</div>
          </div>
          <div class="bs-item busy">
            <div class="bs-num">{{ stats?.buildSummary?.building ?? 0 }}</div>
            <div class="bs-label">⚙️ 进行中</div>
          </div>
        </div>
      </div>

      <!-- 快捷操作 -->
      <div class="card">
        <div class="card-head"><h3>⚡ 快捷操作</h3></div>
        <div class="quick-actions">
          <router-link to="/projects" class="qa-btn">+ 新建项目</router-link>
          <router-link to="/repositories" class="qa-btn">🌿 从仓库源导入</router-link>
          <router-link to="/build" class="qa-btn">🔨 触发构建</router-link>
          <router-link to="/pipelines" class="qa-btn">🔧 管理流水线</router-link>
        </div>
      </div>
    </div>

    <!-- 最近构建活动 -->
    <div class="card">
      <div class="card-head">
        <h3>🕒 最近构建</h3>
        <router-link to="/build-history" class="card-link">查看全部 →</router-link>
      </div>
      <div v-if="loading && !stats" class="empty-hint">加载中…</div>
      <div v-else-if="!stats?.recentBuilds?.length" class="empty-hint">
        还没有任何构建记录。去 <router-link to="/build">构建</router-link> 触发一次吧。
      </div>
      <div v-else class="recent-list">
        <div v-for="b in stats.recentBuilds" :key="b.id" class="recent-item">
          <span class="status-pill" :class="b.status">
            {{ statusIcon(b.status) }}
          </span>
          <div class="recent-meta">
            <div class="recent-image">{{ b.image_name || b.imageName }}<span class="dim">:{{ b.image_tag || b.imageTag }}</span></div>
            <div class="recent-time">{{ relativeTime(b.started_at || b.startedAt) }}</div>
          </div>
          <span class="recent-duration" v-if="b.finished_at || b.finishedAt">
            {{ calcDuration(b.started_at || b.startedAt, b.finished_at || b.finishedAt) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { GetDashboardStats } from '../wailsjs/go/handler/App'

interface BuildItem {
  id: string
  image_name?: string; imageName?: string
  image_tag?: string;  imageTag?: string
  status: string
  started_at?: string; startedAt?: string
  finished_at?: string; finishedAt?: string
}

interface Stats {
  projects: number
  gitProjects: number
  pipelines: number
  containers: { total: number; running: number; stopped: number }
  images: number
  docker: { connected: boolean; version: string }
  recentBuilds: BuildItem[]
  buildSummary: { success: number; failed: number; building: number }
}

const stats = ref<Stats | null>(null)
const loading = ref(false)

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 6)  return '凌晨好'
  if (h < 11) return '早上好'
  if (h < 14) return '中午好'
  if (h < 18) return '下午好'
  if (h < 22) return '晚上好'
  return '深夜了'
})

const dockerClass = computed(() => stats.value?.docker?.connected ? 'docker-up' : 'docker-down')

async function refresh() {
  loading.value = true
  try {
    stats.value = await GetDashboardStats()
  } catch (e) {
    console.error('dashboard load failed:', e)
  }
  loading.value = false
}

function statusIcon(s: string): string {
  switch (s) {
    case 'success':  return '✅'
    case 'failed':   return '❌'
    case 'building': return '⚙️'
    default: return '·'
  }
}

function relativeTime(iso?: string): string {
  if (!iso) return ''
  const t = new Date(iso).getTime()
  if (isNaN(t)) return iso
  const diff = Date.now() - t
  const sec = Math.floor(diff / 1000)
  if (sec < 60) return '刚刚'
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min} 分钟前`
  const hr = Math.floor(min / 60)
  if (hr < 24) return `${hr} 小时前`
  const day = Math.floor(hr / 24)
  if (day < 30) return `${day} 天前`
  return `${Math.floor(day / 30)} 个月前`
}

function calcDuration(s?: string, e?: string): string {
  if (!s || !e) return ''
  const ms = new Date(e).getTime() - new Date(s).getTime()
  if (isNaN(ms) || ms < 0) return ''
  const sec = Math.floor(ms / 1000)
  if (sec < 60) return `${sec}s`
  const min = Math.floor(sec / 60)
  return `${min}m${sec % 60 ? `${sec % 60}s` : ''}`
}

onMounted(refresh)
</script>

<style scoped>
.dashboard { max-width: 1280px; }

.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 24px;
}
h1 { font-size: 28px; margin: 0; color: var(--text-primary); }
.subtitle {
  margin: 6px 0 0 0;
  color: var(--text-secondary);
  font-size: 13px;
}

.btn-outline {
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  padding: 8px 14px;
  border-radius: var(--radius-md);
  font-size: 13px;
  cursor: pointer;
}
.btn-outline:hover:not(:disabled) {
  border-color: var(--brand-start);
  color: var(--brand-start);
}
.btn-outline:disabled { opacity: 0.6; cursor: not-allowed; }

/* KPI cards */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}
.kpi-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: var(--card-bg);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
  text-decoration: none;
  color: inherit;
  transition: box-shadow 0.15s, transform 0.15s, border-color 0.15s;
}
.kpi-card:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
  border-color: var(--brand-start);
}
.kpi-icon {
  font-size: 36px;
  width: 56px; height: 56px;
  display: flex; align-items: center; justify-content: center;
  background: var(--brand-soft);
  border-radius: var(--radius-md);
  flex-shrink: 0;
}
.kpi-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1;
}
.kpi-label {
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 4px;
}
.kpi-sub {
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 4px;
  display: flex; align-items: center;
}
.dot-running, .dot-stopped {
  display: inline-block;
  width: 6px; height: 6px;
  border-radius: 50%;
  margin-right: 4px;
}
.dot-running { background: var(--success-fg); }
.dot-stopped { background: var(--text-muted); }

/* row grid */
.row-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.card {
  background: var(--card-bg);
  border-radius: var(--radius-lg);
  padding: 18px 20px;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
}

.card-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}
.card-head h3 {
  font-size: 14px;
  margin: 0;
  color: var(--text-primary);
}
.card-link {
  font-size: 12px;
  color: var(--brand-start);
  text-decoration: none;
}
.card-link:hover { text-decoration: underline; }

/* docker card */
.docker-card.docker-up {
  border-left: 4px solid var(--success-fg);
}
.docker-card.docker-down {
  border-left: 4px solid var(--danger-fg);
}
.docker-ok, .docker-fail {
  display: flex;
  align-items: center;
  gap: 14px;
}
.docker-pulse {
  width: 14px; height: 14px;
  background: var(--success-fg);
  border-radius: 50%;
  position: relative;
}
.docker-pulse::after {
  content: '';
  position: absolute;
  inset: -3px;
  border-radius: 50%;
  background: var(--success-fg);
  opacity: 0.3;
  animation: pulse 1.6s ease-out infinite;
}
@keyframes pulse {
  to { transform: scale(2.2); opacity: 0; }
}
.docker-icon { font-size: 24px; }
.docker-status {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}
.docker-version {
  font-size: 11px;
  color: var(--text-muted);
  font-family: 'JetBrains Mono', monospace;
  margin-top: 2px;
}
.docker-link {
  font-size: 12px;
  color: var(--brand-start);
  text-decoration: none;
  margin-top: 4px;
  display: inline-block;
}
.docker-link:hover { text-decoration: underline; }

/* build summary */
.build-summary {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  text-align: center;
}
.bs-item {
  padding: 12px 6px;
  border-radius: var(--radius-md);
}
.bs-item.ok   { background: var(--success-bg); }
.bs-item.fail { background: var(--danger-bg); }
.bs-item.busy { background: var(--warning-bg); }
.bs-num {
  font-size: 22px;
  font-weight: 700;
  line-height: 1;
}
.bs-item.ok   .bs-num { color: var(--success-fg); }
.bs-item.fail .bs-num { color: var(--danger-fg); }
.bs-item.busy .bs-num { color: var(--warning-fg); }
.bs-label {
  font-size: 11px;
  margin-top: 4px;
  color: var(--text-secondary);
}

/* quick actions */
.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.qa-btn {
  display: block;
  padding: 8px 12px;
  background: var(--bg-primary);
  color: var(--text-primary);
  text-decoration: none;
  border-radius: var(--radius-sm);
  font-size: 13px;
  transition: background 0.12s;
}
.qa-btn:hover {
  background: var(--brand-soft);
  color: var(--brand-start);
}

/* recent builds */
.recent-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.recent-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  transition: background 0.12s;
}
.recent-item:hover { background: var(--bg-primary); }
.status-pill {
  font-size: 12px;
  width: 28px;
  text-align: center;
}
.recent-meta { flex: 1; min-width: 0; }
.recent-image {
  font-size: 13px;
  color: var(--text-primary);
  font-family: 'JetBrains Mono', monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.recent-image .dim { color: var(--text-muted); }
.recent-time {
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 2px;
}
.recent-duration {
  font-size: 11px;
  color: var(--text-muted);
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
}

/* skeleton */
.skel-line {
  height: 18px;
  background: linear-gradient(90deg, var(--bg-primary), var(--border-color), var(--bg-primary));
  background-size: 200% 100%;
  border-radius: 4px;
  animation: shimmer 1.4s linear infinite;
}
.skel-line.lg { height: 24px; width: 60%; }
@keyframes shimmer {
  0% { background-position: 100% 0; }
  100% { background-position: -100% 0; }
}

.empty-hint {
  text-align: center;
  padding: 20px;
  color: var(--text-muted);
  font-size: 13px;
}
.empty-hint a {
  color: var(--brand-start);
  text-decoration: none;
}
</style>
