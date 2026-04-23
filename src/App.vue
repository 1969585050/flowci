<template>
  <div class="app-container">
    <header class="app-header">
      <h1>🚀 FlowCI</h1>
      <nav>
        <router-link to="/">项目</router-link>
        <router-link to="/build">构建</router-link>
        <router-link to="/deploy">部署</router-link>
        <router-link to="/settings">设置</router-link>
      </nav>
    </header>
    <main class="app-main">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.app-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 1rem 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.app-header h1 {
  margin: 0;
  font-size: 1.5rem;
}

.app-header nav {
  display: flex;
  gap: 1.5rem;
}

.app-header nav a {
  color: white;
  text-decoration: none;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  transition: background 0.2s;
}

.app-header nav a:hover {
  background: rgba(255, 255, 255, 0.2);
}

.app-header nav a.router-link-active {
  background: rgba(255, 255, 255, 0.3);
}

.app-main {
  flex: 1;
  padding: 2rem;
  background: #f5f7fa;
}
</style>
