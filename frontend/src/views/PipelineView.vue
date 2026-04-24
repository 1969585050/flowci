<template>
  <div class="pipeline-view">
    <div class="header">
      <div class="header-left">
        <h1>流水线</h1>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="showImportDialog = true">📥 导入 YAML</button>
        <button class="btn-primary" @click="showCreateDialog = true">新建流水线</button>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else-if="pipelines.length === 0" class="empty-state">
      <div class="empty-icon">🔧</div>
      <h3>暂无流水线</h3>
      <p>创建第一个流水线来自动化您的构建和部署流程</p>
    </div>

    <div v-else class="pipeline-list">
      <div v-for="pipeline in pipelines" :key="pipeline.id" class="pipeline-card">
        <div class="pipeline-header">
          <div class="pipeline-info">
            <h3>{{ pipeline.name }}</h3>
            <span class="step-count">{{ pipeline.steps?.length || 0 }} 个步骤</span>
          </div>
          <div class="pipeline-actions">
            <button class="btn-action" @click="exportPipeline(pipeline)">📤 导出</button>
            <button class="btn-action" @click="runPipeline(pipeline)">▶ 运行</button>
            <button class="btn-action btn-danger" @click="deletePipelineConfirm(pipeline)">删除</button>
          </div>
        </div>
        <div class="pipeline-steps">
          <div v-for="(step, idx) in pipeline.steps" :key="idx" class="step-badge" :class="step.type">
            <span class="step-type">{{ stepIcon(step.type) }}</span>
            <span class="step-name">{{ step.name || step.type }}</span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showCreateDialog" class="modal-overlay" @click.self="closeCreateDialog">
      <div class="modal">
        <h2>新建流水线</h2>
        <div class="form-group">
          <label>流水线名称</label>
          <input v-model="newPipeline.name" type="text" placeholder="my-pipeline" />
        </div>
        <div class="form-group">
          <label>选择项目</label>
          <select v-model="newPipeline.projectId">
            <option value="">请选择项目</option>
            <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
        </div>

        <div class="steps-section">
          <div class="steps-header">
            <label>步骤</label>
            <button class="btn-small" @click="addStep">+ 添加步骤</button>
          </div>
          <div v-for="(step, idx) in newPipeline.steps" :key="idx" class="step-item">
            <select v-model="step.type" class="step-type-select">
              <option value="build">构建</option>
              <option value="push">推送</option>
              <option value="deploy">部署</option>
            </select>
            <input v-model="step.name" type="text" placeholder="步骤名称" class="step-name-input" />
            <input v-model.number="step.retry" type="number" min="0" max="5" placeholder="重试次数" class="step-retry-input" />
            <select v-model="step.onFail" class="step-fail-select">
              <option value="stop">失败停止</option>
              <option value="continue">失败继续</option>
            </select>
            <button class="btn-remove" @click="removeStep(idx)">✕</button>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn-cancel" @click="closeCreateDialog">取消</button>
          <button class="btn-primary" @click="createPipeline" :disabled="creating">创建</button>
        </div>
      </div>
    </div>

    <div v-if="showRunDialog" class="modal-overlay" @click.self="closeRunDialog">
      <div class="modal">
        <h2>运行流水线</h2>
        <div v-if="runningPipeline" class="pipeline-run-info">
          <p><strong>流水线:</strong> {{ runningPipeline.name }}</p>
          <p><strong>项目:</strong> {{ getProjectName(runningPipeline.project_id) }}</p>
        </div>

        <div v-if="!pipelineResult" class="running-status">
          <div class="spinner"></div>
          <p>正在执行流水线...</p>
        </div>

        <div v-else class="pipeline-result">
          <div v-for="(log, idx) in pipelineResult.logs" :key="idx" class="log-item" :class="log.status">
            <span class="log-icon">{{ log.status === 'success' ? '✅' : '❌' }}</span>
            <span class="log-step">{{ log.step }}</span>
            <span class="log-message">{{ log.message || log.error }}</span>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn-cancel" @click="closeRunDialog">关闭</button>
        </div>
      </div>
    </div>

    <div v-if="showImportDialog" class="modal-overlay" @click.self="closeImportDialog">
      <div class="modal">
        <h2>导入 YAML</h2>
        <div class="form-group">
          <label>选择项目</label>
          <select v-model="importConfig.projectId">
            <option value="">请选择项目</option>
            <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>YAML 内容</label>
          <textarea v-model="importConfig.yaml" rows="12" placeholder="name: my-pipeline
config:
  stop_on_fail: true
steps:
  - type: build
    name: build-image
  - type: push
    name: push-image"></textarea>
        </div>
        <div class="modal-actions">
          <button class="btn-cancel" @click="closeImportDialog">取消</button>
          <button class="btn-primary" @click="importPipeline" :disabled="importing">导入</button>
        </div>
      </div>
    </div>

    <div v-if="showExportDialog" class="modal-overlay" @click.self="closeExportDialog">
      <div class="modal">
        <h2>导出 YAML</h2>
        <div class="yaml-preview">
          <pre>{{ exportYaml }}</pre>
        </div>
        <div class="modal-actions">
          <button class="btn-cancel" @click="closeExportDialog">关闭</button>
          <button class="btn-primary" @click="copyYaml">复制</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, inject, onMounted } from 'vue'
import { ListPipelines, CreatePipeline, DeletePipeline, ExecutePipeline, ListProjects, ExportPipelineToYaml, ImportPipelineFromYaml } from '../wailsjs/go/main/App'

interface Pipeline {
  id: string
  project_id: string
  name: string
  steps: PipelineStep[]
  config: PipelineConfig
}

interface PipelineStep {
  type: string
  name: string
  config: Record<string, any>
  retry: number
  onFail: string
}

interface PipelineConfig {
  parallel: boolean
  stopOnFail: boolean
}

interface Project {
  id: string
  name: string
}

const toast = inject('toast') as { success: (msg: string) => void; error: (msg: string) => void }

const loading = ref(false)
const pipelines = ref<Pipeline[]>([])
const projects = ref<Project[]>([])
const showCreateDialog = ref(false)
const showRunDialog = ref(false)
const showImportDialog = ref(false)
const showExportDialog = ref(false)
const creating = ref(false)
const importing = ref(false)
const runningPipeline = ref<Pipeline | null>(null)
const pipelineResult = ref<any>(null)
const exportYaml = ref('')
const importConfig = ref({
  projectId: '',
  yaml: ''
})

const newPipeline = ref({
  name: '',
  projectId: '',
  steps: [] as PipelineStep[]
})

function stepIcon(type_: string): string {
  const icons: Record<string, string> = {
    build: '🔨',
    push: '📤',
    deploy: '🚀'
  }
  return icons[type_] || '📦'
}

function getProjectName(projectId: string): string {
  const p = projects.value.find(p => p.id === projectId)
  return p?.name || projectId
}

async function loadPipelines() {
  loading.value = true
  try {
    const result = await ListProjects()
    projects.value = result.map((p: any) => ({ id: p.id, name: p.name }))

    const allPipelines: Pipeline[] = []
    for (const p of projects.value) {
      const pls = await ListPipelines(p.id)
      allPipelines.push(...pls)
    }
    pipelines.value = allPipelines
  } catch (e) {
    console.error('Failed to load pipelines:', e)
  }
  loading.value = false
}

function addStep() {
  newPipeline.value.steps.push({
    type: 'build',
    name: '',
    config: {},
    retry: 0,
    onFail: 'stop'
  })
}

function removeStep(idx: number) {
  newPipeline.value.steps.splice(idx, 1)
}

function closeCreateDialog() {
  showCreateDialog.value = false
  newPipeline.value = { name: '', projectId: '', steps: [] }
}

async function createPipeline() {
  if (!newPipeline.value.name || !newPipeline.value.projectId) {
    toast?.error('请填写流水线名称和选择项目')
    return
  }
  if (newPipeline.value.steps.length === 0) {
    toast?.error('请添加至少一个步骤')
    return
  }

  creating.value = true
  try {
    await CreatePipeline({
      projectId: newPipeline.value.projectId,
      name: newPipeline.value.name,
      steps: newPipeline.value.steps,
      config: { stopOnFail: true }
    })
    toast?.success('流水线创建成功')
    closeCreateDialog()
    await loadPipelines()
  } catch (e) {
    toast?.error(`创建失败: ${e}`)
  }
  creating.value = false
}

async function deletePipelineConfirm(pipeline: Pipeline) {
  if (!confirm(`确定要删除流水线 "${pipeline.name}" 吗？`)) return
  try {
    await DeletePipeline(pipeline.id)
    toast?.success('流水线已删除')
    await loadPipelines()
  } catch (e) {
    toast?.error(`删除失败: ${e}`)
  }
}

async function runPipeline(pipeline: Pipeline) {
  runningPipeline.value = pipeline
  pipelineResult.value = null
  showRunDialog.value = true

  try {
    pipelineResult.value = await ExecutePipeline({
      pipelineId: pipeline.id,
      projectId: pipeline.project_id
    })
  } catch (e) {
    pipelineResult.value = { success: false, logs: [], error: String(e) }
  }
}

function closeRunDialog() {
  showRunDialog.value = false
  runningPipeline.value = null
  pipelineResult.value = null
}

function closeImportDialog() {
  showImportDialog.value = false
  importConfig.value = { projectId: '', yaml: '' }
}

async function importPipeline() {
  if (!importConfig.value.projectId) {
    toast?.error('请选择项目')
    return
  }
  if (!importConfig.value.yaml.trim()) {
    toast?.error('请输入 YAML 内容')
    return
  }

  importing.value = true
  try {
    const result = await ImportPipelineFromYaml({
      projectId: importConfig.value.projectId,
      yaml: importConfig.value.yaml
    })
    if (result.error) {
      toast?.error(`导入失败: ${result.error}`)
    } else {
      toast?.success('流水线导入成功')
      closeImportDialog()
      await loadPipelines()
    }
  } catch (e) {
    toast?.error(`导入失败: ${e}`)
  }
  importing.value = false
}

async function exportPipeline(pipeline: Pipeline) {
  try {
    exportYaml.value = await ExportPipelineToYaml(pipeline.id)
    showExportDialog.value = true
  } catch (e) {
    toast?.error(`导出失败: ${e}`)
  }
}

function closeExportDialog() {
  showExportDialog.value = false
  exportYaml.value = ''
}

async function copyYaml() {
  try {
    await navigator.clipboard.writeText(exportYaml.value)
    toast?.success('已复制到剪贴板')
  } catch (e) {
    toast?.error('复制失败')
  }
}

onMounted(loadPipelines)
</script>

<style scoped>
.pipeline-view {
  max-width: 900px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header h1 {
  font-size: 28px;
  color: var(--text-primary, #1a1a2e);
}

.header-actions {
  display: flex;
  gap: 8px;
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

.pipeline-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.pipeline-card {
  background: var(--card-bg, white);
  border-radius: 12px;
  padding: 20px;
  box-shadow: var(--shadow, 0 2px 12px rgba(0, 0, 0, 0.05));
}

.pipeline-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.pipeline-info h3 {
  font-size: 18px;
  color: var(--text-primary, #1a1a2e);
  margin: 0 0 4px 0;
}

.step-count {
  font-size: 13px;
  color: var(--text-secondary, #999);
}

.pipeline-actions {
  display: flex;
  gap: 8px;
}

.pipeline-steps {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.step-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 13px;
  background: var(--bg-primary, #f5f5f5);
  color: var(--text-primary, #333);
}

.step-badge.build {
  background: #dbeafe;
  color: #1d4ed8;
}

.step-badge.push {
  background: #f3e8ff;
  color: #7c3aed;
}

.step-badge.deploy {
  background: #d1fae5;
  color: #059669;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: var(--card-bg, white);
  border-radius: 16px;
  padding: 32px;
  width: 560px;
  max-width: 90vw;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}

.modal h2 {
  font-size: 22px;
  color: var(--text-primary, #1a1a2e);
  margin: 0 0 24px 0;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}

.form-group label {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary, #333);
}

.form-group input,
.form-group select {
  padding: 12px;
  border: 2px solid var(--border-color, #e0e0e0);
  border-radius: 8px;
  font-size: 14px;
  background: var(--card-bg, white);
  color: var(--text-primary, #333);
  transition: border-color 0.2s;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #667eea;
}

.steps-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid var(--border-color, #e0e0e0);
}

.steps-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.steps-header label {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary, #333);
}

.step-item {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  align-items: center;
}

.step-type-select {
  width: 100px;
  padding: 8px;
  border: 2px solid var(--border-color, #e0e0e0);
  border-radius: 6px;
  font-size: 13px;
  background: var(--card-bg, white);
  color: var(--text-primary, #333);
}

.step-name-input {
  flex: 1;
  padding: 8px;
  border: 2px solid var(--border-color, #e0e0e0);
  border-radius: 6px;
  font-size: 13px;
  background: var(--card-bg, white);
  color: var(--text-primary, #333);
}

.step-retry-input {
  width: 70px;
  padding: 8px;
  border: 2px solid var(--border-color, #e0e0e0);
  border-radius: 6px;
  font-size: 13px;
  background: var(--card-bg, white);
  color: var(--text-primary, #333);
  text-align: center;
}

.step-fail-select {
  width: 100px;
  padding: 8px;
  border: 2px solid var(--border-color, #e0e0e0);
  border-radius: 6px;
  font-size: 13px;
  background: var(--card-bg, white);
  color: var(--text-primary, #333);
}

.btn-remove {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 6px;
  background: #fee2e2;
  color: #dc2626;
  cursor: pointer;
  font-size: 14px;
}

.btn-small {
  padding: 6px 12px;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.btn-cancel {
  background: var(--bg-primary, #f0f0f0);
  color: var(--text-primary, #333);
  border: none;
  padding: 12px 24px;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: var(--bg-primary, #f0f0f0);
  color: var(--text-primary, #333);
  border: 1px solid var(--border-color, #e0e0e0);
}

.btn-secondary:hover {
  background: var(--border-color, #e0e0e0);
}

.yaml-preview {
  background: #1e1e2e;
  border-radius: 8px;
  padding: 16px;
  margin: 16px 0;
  max-height: 400px;
  overflow: auto;
}

.yaml-preview pre {
  color: #a5b6cf;
  font-family: 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.5;
  white-space: pre-wrap;
  margin: 0;
}

.btn-action {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  background: var(--border-color, #e0e0e0);
  color: var(--text-primary, #333);
}

.btn-action:hover {
  background: #d0d0d0;
}

.btn-danger {
  background: #fee2e2;
  color: #dc2626;
}

.btn-danger:hover {
  background: #fecaca;
}

.pipeline-run-info {
  background: var(--bg-primary, #f5f5f5);
  padding: 16px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.pipeline-run-info p {
  margin: 4px 0;
  color: var(--text-primary, #333);
}

.running-status {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px;
  gap: 16px;
}

.running-status p {
  color: var(--text-secondary, #666);
}

.pipeline-result {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 400px;
  overflow-y: auto;
}

.log-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  border-radius: 8px;
  background: var(--bg-primary, #f5f5f5);
}

.log-item.success {
  background: #d1fae5;
}

.log-item.failed {
  background: #fee2e2;
}

.log-icon {
  font-size: 16px;
}

.log-step {
  font-weight: 600;
  color: var(--text-primary, #333);
}

.log-message {
  flex: 1;
  color: var(--text-secondary, #666);
  font-size: 13px;
}
</style>
