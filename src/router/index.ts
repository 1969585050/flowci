import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'projects',
      component: () => import('../views/ProjectsView.vue')
    },
    {
      path: '/build',
      name: 'build',
      component: () => import('../views/BuildView.vue')
    },
    {
      path: '/deploy',
      name: 'deploy',
      component: () => import('../views/DeployView.vue')
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('../views/SettingsView.vue')
    }
  ]
})

export default router
