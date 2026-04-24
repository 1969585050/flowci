<template>
  <div id="app">
    <div class="sidebar">
      <div class="logo">
        <span class="logo-icon">🚀</span>
        <span class="logo-text">FlowCI</span>
      </div>
      <div class="nav">
        <router-link to="/projects" class="nav-item" active-class="active">
          <span class="nav-icon">📦</span>
          <span>项目</span>
        </router-link>
        <router-link to="/build" class="nav-item" active-class="active">
          <span class="nav-icon">🔨</span>
          <span>构建</span>
        </router-link>
        <router-link to="/deploy" class="nav-item" active-class="active">
          <span class="nav-icon">🌐</span>
          <span>部署</span>
        </router-link>
        <router-link to="/settings" class="nav-item" active-class="active">
          <span class="nav-icon">⚙️</span>
          <span>设置</span>
        </router-link>
      </div>
    </div>
    <div class="content">
      <router-view />
    </div>
  </div>
</template>

<script setup lang="ts">
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background: #f5f5f5;
  color: #333;
}

#app {
  display: flex;
  height: 100vh;
}

.sidebar {
  width: 240px;
  background: linear-gradient(180deg, #1e1e2e 0%, #2d2d44 100%);
  color: #fff;
  display: flex;
  flex-direction: column;
  padding: 20px 0;
}

.logo {
  padding: 0 20px 30px;
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 24px;
  font-weight: bold;
}

.logo-icon {
  font-size: 32px;
}

.logo-text {
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.nav {
  flex: 1;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 24px;
  color: #a0a0b0;
  text-decoration: none;
  transition: all 0.2s;
  cursor: pointer;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #fff;
}

.nav-item.active {
  background: rgba(102, 126, 234, 0.2);
  color: #667eea;
  border-left: 3px solid #667eea;
}

.nav-icon {
  font-size: 20px;
}

.content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}
</style>
