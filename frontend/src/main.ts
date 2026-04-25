import { createApp } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
import App from './App.vue'

// 全局主题变量（必须在任何组件样式之前加载）
import './styles/theme.css'

import ProjectsView from './views/ProjectsView.vue'
import BuildView from './views/BuildView.vue'
import DeployView from './views/DeployView.vue'
import PushView from './views/PushView.vue'
import SettingsView from './views/SettingsView.vue'
import BuildHistoryView from './views/BuildHistoryView.vue'
import BuildDetailView from './views/BuildDetailView.vue'
import ImagesView from './views/ImagesView.vue'
import PipelineView from './views/PipelineView.vue'
import RepositoriesView from './views/RepositoriesView.vue'
import DashboardView from './views/DashboardView.vue'

const routes = [
  { path: '/', redirect: '/dashboard' },
  { path: '/dashboard', component: DashboardView },
  { path: '/projects', component: ProjectsView },
  { path: '/repositories', component: RepositoriesView },
  { path: '/build', component: BuildView },
  { path: '/deploy', component: DeployView },
  { path: '/push', component: PushView },
  { path: '/settings', component: SettingsView },
  { path: '/build-history', component: BuildHistoryView },
  { path: '/build-detail', component: BuildDetailView },
  { path: '/images', component: ImagesView },
  { path: '/pipelines', component: PipelineView },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

createApp(App).use(router).mount('#app')
