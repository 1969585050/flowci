<template>
  <div class="dashboard">
    <div class="header">
      <div>
        <h1>仪表盘</h1>
        <p class="subtitle">{{ greeting }}，FlowCI 在你身边。</p>
      </div>
      <button class="btn btn-secondary btn-sm" :disabled="loading" @click="refresh">
        <RefreshCw :size="13" :stroke-width="1.75" :class="{ 'icon-spin': loading }" />
        {{ loading ? '刷新中…' : '刷新' }}
      </button>
    </div>

    <!-- KPI 大数字卡片 -->
    <div class="kpi-grid">
      <router-link to="/projects" class="card card-interactive kpi-card">
        <div class="kpi-icon"><Package :size="20" :stroke-width="1.75" /></div>
        <div class="kpi-body">
          <div class="kpi-value">{{ stats?.projects ?? '-' }}</div>
          <div class="kpi-label">项目</div>
          <div v-if="stats?.gitProjects" class="kpi-sub">含 {{ stats.gitProjects }} 个 Git</div>
        </div>
      </router-link>

      <router-link to="/pipelines" class="card card-interactive kpi-card">
        <div class="kpi-icon"><Workflow :size="20" :stroke-width="1.75" /></div>
        <div class="kpi-body">
          <div class="kpi-value">{{ stats?.pipelines ?? '-' }}</div>
          <div class="kpi-label">流水线</div>
        </div>
      </router-link>

      <router-link to="/deploy" class="card card-interactive kpi-card">
        <div class="kpi-icon"><Boxes :size="20" :stroke-width="1.75" /></div>
        <div class="kpi-body">
          <div class="kpi-value">{{ stats?.containers?.total ?? '-' }}</div>
          <div class="kpi-label">容器</div>
          <div v-if="stats?.containers" class="kpi-sub">
            <span class="dot dot-running"></span>{{ stats.containers.running }} 运行
            <span class="dot dot-stopped"></span>{{ stats.containers.stopped }} 停止
          </div>
        </div>
      </router-link>

      <router-link to="/images" class="card card-interactive kpi-card">
        <div class="kpi-icon"><Layers :size="20" :stroke-width="1.75" /></div>
        <div class="kpi-body">
          <div class="kpi-value">{{ stats?.images ?? '-' }}</div>
          <div class="kpi-label">镜像</div>
        </div>
      </router-link>
    </div>

    <div class="row-grid">
      <!-- Docker 状态卡 -->
      <div class="card">
        <div class="card-head">
          <h3><Container :size="14" :stroke-width="1.75" /> Docker 状态</h3>
        </div>
        <div v-if="loading && !stats" class="skel-line lg"></div>
        <template v-else-if="stats?.docker?.connected">
          <div class="docker-row">
            <div class="docker-pulse"></div>
            <div>
              <div class="docker-status">
                <span class="badge badge-success badge-dot">已连接</span>
              </div>
              <div class="docker-version mono">v{{ stats.docker.version }}</div>
            </div>
          </div>
        </template>
        <template v-else>
          <div class="docker-row">
            <AlertTriangle class="docker-warn-icon" :size="22" :stroke-width="1.75" />
            <div>
              <div class="docker-status">
                <span class="badge badge-danger">未连接</span>
              </div>
              <router-link to="/settings" class="docker-link">前往 设置 → Docker 检查 →</router-link>
            </div>
          </div>
        </template>
      </div>

      <!-- 24h 构建摘要 -->
      <div class="card">
        <div class="card-head">
          <h3><BarChart3 :size="14" :stroke-width="1.75" /> 最近 24 小时构建</h3>
        </div>
        <div class="build-summary">
          <div class="bs-item bs-success">
            <div class="bs-num">{{ stats?.buildSummary?.success ?? 0 }}</div>
            <div class="bs-label"><CheckCircle2 :size="11" :stroke-width="2" /> 成功</div>
          </div>
          <div class="bs-item bs-failed">
            <div class="bs-num">{{ stats?.buildSummary?.failed ?? 0 }}</div>
            <div class="bs-label"><XCircle :size="11" :stroke-width="2" /> 失败</div>
          </div>
          <div class="bs-item bs-building">
            <div class="bs-num">{{ stats?.buildSummary?.building ?? 0 }}</div>
            <div class="bs-label"><Loader2 :size="11" :stroke-width="2" /> 进行中</div>
          </div>
        </div>
      </div>

      <!-- 快捷操作 -->
      <div class="card">
        <div class="card-head"><h3><Zap :size="14" :stroke-width="1.75" /> 快捷操作</h3></div>
        <div class="quick-actions">
          <router-link to="/projects" class="qa-btn">
            <Plus :size="14" :stroke-width="1.75" /> 新建项目
          </router-link>
          <router-link to="/repositories" class="qa-btn">
            <GitBranch :size="14" :stroke-width="1.75" /> 从仓库源导入
          </router-link>
          <router-link to="/build" class="qa-btn">
            <Hammer :size="14" :stroke-width="1.75" /> 触发构建
          </router-link>
          <router-link to="/pipelines" class="qa-btn">
            <Workflow :size="14" :stroke-width="1.75" /> 管理流水线
          </router-link>
        </div>
      </div>
    </div>

    <!-- 最近构建活动 -->
    <div class="card">
      <div class="card-head">
        <h3><Clock :size="14" :stroke-width="1.75" /> 最近构建</h3>
        <router-link to="/build-history" class="card-link">查看全部 →</router-link>
      </div>
      <div v-if="loading && !stats" class="empty-hint">加载中…</div>
      <div v-else-if="!stats?.recentBuilds?.length" class="empty-hint">
        还没有任何构建记录。去 <router-link to="/build">构建</router-link> 触发一次吧。
      </div>
      <div v-else class="recent-list">
        <router-link
          v-for="b in stats.recentBuilds"
          :key="b.id"
          :to="{ path: '/build-detail', query: { recordId: b.id } }"
          class="recent-item"
        >
          <span class="recent-status" :class="statusColorClass(b.status)">
            <component
              :is="statusIconComp(b.status)"
              :size="14"
              :stroke-width="2"
              :class="{ 'icon-spin': b.status === 'building' }"
            />
          </span>
          <div class="recent-meta">
            <div class="recent-image mono">{{ b.image_name || b.imageName }}<span class="dim">:{{ b.image_tag || b.imageTag }}</span></div>
            <div class="recent-time">{{ relativeTime(b.started_at || b.startedAt) }}</div>
          </div>
          <span class="recent-duration mono" v-if="b.finished_at || b.finishedAt">
            {{ calcDuration(b.started_at || b.startedAt, b.finished_at || b.finishedAt) }}
          </span>
          <ChevronRight class="recent-arrow" :size="16" :stroke-width="1.75" />
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, type Component } from 'vue'
import {
  Package, Workflow, Boxes, Layers, Container, AlertTriangle,
  BarChart3, Zap, Clock, RefreshCw, Plus, GitBranch, Hammer,
  CheckCircle2, XCircle, Loader2, Circle, ChevronRight,
} from 'lucide-vue-next'
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

async function refresh() {
  loading.value = true
  try {
    stats.value = await GetDashboardStats()
  } catch (e) {
    console.error('dashboard load failed:', e)
  }
  loading.value = false
}

function statusIconComp(s: string): Component {
  switch (s) {
    case 'success':  return CheckCircle2
    case 'failed':   return XCircle
    case 'building': return Loader2
    case 'pending':  return Clock
    default:         return Circle
  }
}

function statusColorClass(s: string): string {
  switch (s) {
    case 'success':  return 'status-success'
    case 'failed':   return 'status-failed'
    case 'building': return 'status-building'
    case 'pending':  return 'status-info'
    default:         return 'status-neutral'
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
  margin-bottom: var(--space-6);
}
h1 {
  font-size: var(--text-3xl);
  font-weight: var(--weight-semibold);
  color: var(--text-primary);
  letter-spacing: -0.01em;
  margin: 0;
}
.subtitle {
  margin: var(--space-1) 0 0 0;
  color: var(--text-secondary);
  font-size: var(--text-base);
}

/* ---- KPI cards ---- */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: var(--space-4);
  margin-bottom: var(--space-5);
}
.kpi-card {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-4) var(--space-5);
  text-decoration: none;
  color: inherit;
}
.kpi-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-sunken);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  color: var(--brand-500);
  flex-shrink: 0;
}
.kpi-value {
  font-size: var(--text-3xl);
  font-weight: var(--weight-semibold);
  color: var(--text-primary);
  line-height: 1;
  letter-spacing: -0.02em;
}
.kpi-label {
  font-size: var(--text-base);
  color: var(--text-secondary);
  margin-top: var(--space-1);
}
.kpi-sub {
  font-size: var(--text-xs);
  color: var(--text-muted);
  margin-top: var(--space-1);
  display: flex;
  align-items: center;
  gap: var(--space-1);
}
.dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  margin-right: 2px;
}
.dot + .dot { margin-left: var(--space-2); }
.dot-running { background: var(--success-fg); }
.dot-stopped { background: var(--text-muted); }

/* ---- Row grid ---- */
.row-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: var(--space-4);
  margin-bottom: var(--space-5);
}

/* ---- Card head ---- */
.card-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-4);
}
.card-head h3 {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-md);
  font-weight: var(--weight-semibold);
  color: var(--text-primary);
  margin: 0;
}
.card-head h3 :deep(svg) { color: var(--text-muted); }
.card-link {
  font-size: var(--text-sm);
  color: var(--brand-500);
  text-decoration: none;
}
.card-link:hover { color: var(--brand-600); text-decoration: underline; }

/* ---- Docker card ---- */
.docker-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}
.docker-pulse {
  width: 12px;
  height: 12px;
  background: var(--success-fg);
  border-radius: 50%;
  position: relative;
  flex-shrink: 0;
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
  to { transform: scale(2.4); opacity: 0; }
}
.docker-warn-icon {
  color: var(--warning-fg);
  flex-shrink: 0;
}
.docker-status {
  font-size: var(--text-base);
  font-weight: var(--weight-medium);
}
.docker-version {
  font-size: var(--text-xs);
  color: var(--text-muted);
  margin-top: var(--space-1);
}
.docker-link {
  font-size: var(--text-sm);
  color: var(--brand-500);
  text-decoration: none;
  margin-top: var(--space-1);
  display: inline-block;
}
.docker-link:hover { color: var(--brand-600); text-decoration: underline; }

/* ---- Build summary ---- */
.build-summary {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-2);
  text-align: center;
}
.bs-item {
  padding: var(--space-3) var(--space-2);
  border-radius: var(--radius-md);
  background: var(--bg-sunken);
  border: 1px solid var(--border-subtle);
}
.bs-num {
  font-size: var(--text-2xl);
  font-weight: var(--weight-semibold);
  line-height: 1;
  color: var(--text-primary);
}
.bs-item.bs-success .bs-num { color: var(--success-fg); }
.bs-item.bs-failed  .bs-num { color: var(--danger-fg); }
.bs-item.bs-building .bs-num { color: var(--warning-fg); }
.bs-label {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--text-xs);
  margin-top: var(--space-1);
  color: var(--text-secondary);
}

/* ---- Quick actions ---- */
.quick-actions {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}
.qa-btn {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  background: transparent;
  color: var(--text-primary);
  text-decoration: none;
  border-radius: var(--radius-sm);
  font-size: var(--text-base);
  transition: background var(--transition-fast), color var(--transition-fast);
}
.qa-btn:hover {
  background: var(--bg-hover);
  color: var(--brand-500);
}

/* ---- Recent builds list ---- */
.recent-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.recent-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  transition: background var(--transition-fast);
  text-decoration: none;
  color: inherit;
  cursor: pointer;
}
.recent-item:hover { background: var(--bg-hover); }
.recent-status {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  flex-shrink: 0;
}
.recent-status.status-success  { color: var(--success-fg); }
.recent-status.status-failed   { color: var(--danger-fg); }
.recent-status.status-building { color: var(--warning-fg); }
.recent-status.status-info     { color: var(--info-fg); }
.recent-status.status-neutral  { color: var(--text-muted); }

.recent-meta { flex: 1; min-width: 0; }
.recent-image {
  font-size: var(--text-base);
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.recent-image .dim { color: var(--text-muted); }
.recent-time {
  font-size: var(--text-xs);
  color: var(--text-muted);
  margin-top: 2px;
}
.recent-duration {
  font-size: var(--text-xs);
  color: var(--text-muted);
  flex-shrink: 0;
}
.recent-arrow {
  color: var(--text-muted);
  flex-shrink: 0;
  transition: transform var(--transition-fast), color var(--transition-fast);
}
.recent-item:hover .recent-arrow {
  color: var(--brand-500);
  transform: translateX(2px);
}

/* ---- Skeleton ---- */
.skel-line {
  height: 18px;
  background: linear-gradient(90deg,
    var(--bg-sunken),
    var(--border-subtle),
    var(--bg-sunken));
  background-size: 200% 100%;
  border-radius: var(--radius-sm);
  animation: shimmer 1.4s linear infinite;
}
.skel-line.lg { height: 24px; width: 60%; }
@keyframes shimmer {
  0% { background-position: 100% 0; }
  100% { background-position: -100% 0; }
}

.empty-hint {
  text-align: center;
  padding: var(--space-5);
  color: var(--text-muted);
  font-size: var(--text-base);
}
.empty-hint a {
  color: var(--brand-500);
  text-decoration: none;
}
.empty-hint a:hover { text-decoration: underline; }

/* ---- Icon spin animation ---- */
.icon-spin {
  animation: icon-spin 1s linear infinite;
}
@keyframes icon-spin {
  to { transform: rotate(360deg); }
}
</style>
