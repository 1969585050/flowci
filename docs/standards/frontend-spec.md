# FlowCI 前端开发规范 (Vue 3 + TypeScript)

> 版本: v1.0 (2026-04-25)
> 适用范围: `frontend/` 下所有代码
> 强制级别: **HIGH** / **MED** / **LOW**

---

## 1. 目录结构 [HIGH]

```
frontend/
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
└── src/
    ├── main.ts                     # createApp + router 注册
    ├── App.vue                     # 顶层壳（侧边栏 + router-view）
    ├── router/
    │   └── index.ts                # 路由独立文件，不在 main.ts 堆
    ├── views/                      # 路由页面（一路由一文件）
    │   ├── ProjectsView.vue
    │   ├── PipelineView.vue
    │   └── ...
    ├── components/                 # 可复用 UI 组件
    │   ├── Toast.vue
    │   ├── ConfirmDialog.vue
    │   └── ...
    ├── composables/                # 组合式函数 (use*)
    │   ├── useToast.ts
    │   ├── useAsync.ts
    │   └── useTheme.ts
    ├── api/                        # 对 wailsjs 的二次封装
    │   ├── index.ts                # 统一导出
    │   ├── projects.ts
    │   ├── pipelines.ts
    │   └── ...
    ├── types/                      # 手写补充类型
    │   └── index.ts
    ├── stores/                     # Pinia store（可选，跨组件状态）
    │   └── settings.ts
    ├── utils/                      # 纯函数工具
    │   └── format.ts
    ├── styles/                     # 全局样式变量
    │   ├── theme.css
    │   └── reset.css
    └── wailsjs/                    # Wails 自动生成（不手改）
        ├── go/main/App.d.ts
        └── runtime/
```

**硬规则**:
- `wailsjs/` 目录下文件**严禁手改**，改 Go struct 后跑 `wails generate module` 同步
- `views/` 只放带路由的页面，不放局部对话框（对话框归 `components/`）
- `api/` 是前端与 Go 通信的**唯一出口**，`views/` 不得直接 `import ... from '../wailsjs/...'`

---

## 2. 文件体量 [HIGH]

| 指标 | 上限 | 违反处理 |
|---|---|---|
| 单 `.vue` 文件总行数 | **400** | 拆组件或 composable |
| `<script setup>` 段行数 | **200** | 抽 composable |
| `<template>` 段行数 | **150** | 拆子组件 |
| 单 `.ts` 文件行数 | **300** | 拆模块 |

---

## 3. 命名 [HIGH]

### 文件
- `.vue` 组件文件：`PascalCase.vue`
- `.ts` 模块文件：`camelCase.ts`
- 组合式函数文件：`use + 名词`，如 `useToast.ts`
- store 文件：业务名 `.ts`，如 `settings.ts`

### 标识符
- 组件名 / 类型：`PascalCase`
- 变量 / 函数：`camelCase`
- 常量：`SCREAMING_SNAKE_CASE`（仅模块级）
- 事件 emit：`kebab-case`（如 `update:value`、`item-click`）
- Props：`camelCase` 定义，模板里用 `kebab-case`（Vue 惯例）

### 禁用
- ❌ 组件名带后缀 `Component`、`View`（View 仅 `views/` 目录内允许）
- ❌ 中文拼音、中英混拼

---

## 4. 类型系统 [HIGH]

### 4.1 类型来源优先级

1. **Wails 生成类型**（`wailsjs/go/models.ts` 等） — 最高优先级
2. **`src/types/` 手写补充** — 仅前端内部状态使用
3. **`any` / `unknown`** — 禁止（除 wailsjs runtime 调用返回）

### 4.2 禁止重复定义

```ts
// ❌ 禁止：手写和 wailsjs 重复
interface Pipeline {
  id: string
  project_id: string
  name: string
  steps: PipelineStep[]
}

// ✅ 正确：直接从 wailsjs 导入
import type { handler } from '../wailsjs/go/models'
type Pipeline = handler.Pipeline
```

### 4.3 前端内部状态类型

路由参数、UI 表单中间态、组合式函数返回值等不经过 IPC 的类型写 `src/types/`。

---

## 5. API 调用 [HIGH]

### 5.1 统一出口

所有 Wails 调用必须通过 `src/api/**`：

```ts
// src/api/pipelines.ts
import { ListPipelines, CreatePipeline } from '../wailsjs/go/handler/App'
import type { handler } from '../wailsjs/go/models'

export async function listPipelines(projectId: string): Promise<handler.Pipeline[]> {
  return await ListPipelines(projectId)
}

export async function createPipeline(
  req: handler.CreatePipelineRequest,
): Promise<handler.Pipeline> {
  return await CreatePipeline(req)
}
```

### 5.2 错误处理

后端返回 error → Wails 自动 reject Promise → 前端 try/catch。

```ts
// ✅ 标准模式
try {
  const pipeline = await createPipeline(req)
  toast.success('创建成功')
} catch (e) {
  toast.error(`创建失败: ${e instanceof Error ? e.message : String(e)}`)
}
```

**禁止**靠约定的 `{ success: false, error: "..." }`（见 ipc-spec.md § 5）。

### 5.3 N+1 防御

列表聚合必须在后端完成，前端禁止在 `for` 里发 N 次 IPC。

```ts
// ❌
for (const p of projects.value) {
  const pipelines = await ListPipelines(p.id)
}

// ✅
const all = await listAllPipelines()  // 后端一次返回
```

---

## 6. 状态管理 [MED]

### 6.1 选型

| 场景 | 方案 |
|---|---|
| 组件内状态 | `ref` / `reactive` |
| 父子传值 | `props` / `emit` |
| 跨层但同子树 | `provide` / `inject` |
| 全局跨页 | **Pinia**（已在 package.json） |
| 持久化（跨启动） | **后端 SQLite settings 表**，不用 localStorage |

### 6.2 Pinia 使用

```ts
// src/stores/settings.ts
import { defineStore } from 'pinia'
import * as api from '../api/settings'

export const useSettingsStore = defineStore('settings', {
  state: () => ({ theme: 'dark' as 'dark' | 'light' | 'system' }),
  actions: {
    async load() {
      const s = await api.getSettings()
      this.theme = (s.theme as any) ?? 'dark'
    },
    async setTheme(theme: 'dark' | 'light' | 'system') {
      this.theme = theme
      await api.saveSettings({ theme })
    },
  },
})
```

### 6.3 禁止

- ❌ `localStorage` / `sessionStorage` 持久化业务数据（主题、项目列表）
  - 例外：路由 scroll 位置、临时 UI 偏好（侧边栏折叠）可以用 localStorage
- ❌ 全局 `window.xxx` 挂载业务对象

---

## 7. 组件设计 [MED]

### 7.1 Props / Events

- Props 必须定义类型 + 默认值
- 用 `defineProps<Type>()` + `withDefaults`
- 事件用 `defineEmits<{ (e: 'xxx', payload: T): void }>()`

### 7.2 对话框模式

对话框组件**不**通过 `v-if + 组件内 state` 控制，统一：

```vue
<!-- 父组件 -->
<ConfirmDialog
  v-model:visible="showDelete"
  title="删除流水线"
  :message="`确定删除 ${pipeline.name} ?`"
  @confirm="doDelete"
/>
```

### 7.3 样式隔离

- `<style scoped>` 默认
- 全局变量（主题色）写 `src/styles/theme.css`，组件内只引用 `var(--xxx)`
- **禁止**在组件里硬编码颜色（`#667eea` 之类），全走 CSS 变量

---

## 8. 主题 / 设置 [HIGH]

- **单一真相源**：`SQLite.settings.theme`
- `App.vue` 启动时 `useSettingsStore().load()`
- 系统主题监听用 `matchMedia`，不读 localStorage
- 切换主题走 Pinia action → 写 SQLite → 更新 `data-theme` attr

---

## 9. 代码注释 [MED]

- 注释用**英文**（与 m1-front 仓保持一致）
- 只解释 WHY，不解释 WHAT
- 组合式函数、store action 必须有一行 JSDoc：

```ts
/** Loads settings from SQLite and applies theme to document. */
export async function useThemeBootstrap() { ... }
```

---

## 10. 禁止清单 [HIGH]

- ❌ 手写 `fetch` / `axios`（已删除 axios 依赖）— 一切走 Wails
- ❌ 在 `<script setup>` 里拼长 HTML 字符串用 `v-html`
- ❌ 在模板里写复杂逻辑（三元嵌套、长链方法）— 抽 `computed`
- ❌ `@tauri-apps/*` 任何残留（阶段 1 清零）
- ❌ `confirm()` / `alert()` / `prompt()` — 统一用 `ConfirmDialog` + Toast

---

## 11. 审查清单（PR 前自检）

- [ ] 所有 IPC 调用都经过 `src/api/`
- [ ] 没有手写和 wailsjs 重复的类型
- [ ] 没有 `localStorage` 存业务数据
- [ ] 没有硬编码颜色（全走 CSS 变量）
- [ ] 所有 async 调用都有 try/catch + toast 提示
- [ ] 单 `.vue` ≤ 400 行
- [ ] 组件 props / emits 都有类型标注
- [ ] 没有 `any`（或有注释说明原因）
- [ ] 文件 UTF-8 无 BOM 保存
