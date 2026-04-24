import { createApp } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
import App from './App.vue'
import ProjectsView from './views/ProjectsView.vue'
import BuildView from './views/BuildView.vue'
import DeployView from './views/DeployView.vue'
import PushView from './views/PushView.vue'
import SettingsView from './views/SettingsView.vue'
import BuildHistoryView from './views/BuildHistoryView.vue'
import BuildDetailView from './views/BuildDetailView.vue'
import ImagesView from './views/ImagesView.vue'
import PipelineView from './views/PipelineView.vue'

const routes = [
  { path: '/', redirect: '/projects' },
  { path: '/projects', component: ProjectsView },
  { path: '/build', component: BuildView },
  { path: '/deploy', component: DeployView },
  { path: '/push', component: PushView },
  { path: '/settings', component: SettingsView },
  { path: '/build-history', component: BuildHistoryView },
  { path: '/build-detail', component: BuildDetailView },
  { path: '/images', component: ImagesView },
  { path: '/pipelines', component: PipelineView }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

const app = createApp(App)
app.use(router)
app.mount('#app')
