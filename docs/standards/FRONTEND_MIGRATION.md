# 前端迁移指南（阶段 2.G + 3.5）

> 目的: 后端阶段 2 重构把 App 从 `main` 包迁到了 `internal/handler/`，并改了所有 Bind 方法签名。
> 前端 `wailsjs/` 绑定是 wails CLI 自动生成的，本次迁移必须**在本地环境**由你手动跑 `wails generate module` 再改前端代码。
> （AI 会话在 sandbox 下跑不了 wails CLI 的 go 子进程，所以这部分无法代做。）

---

## 为什么要迁移

| 项 | 旧 | 新 |
|---|---|---|
| Go 包 | `package main` 持有 App | `package handler` 持有 App |
| 方法签名 | `func (App) XxxYyy(ctx, data map[string]interface{}) map[string]interface{}` | `func (App) XxxYyy(ctx, *XxxRequest) (*XxxResponse, error)` |
| 错误协议 | `return map{success: false, error: "..."}, nil` | `return nil, fmt.Errorf(...)` |
| JSON tag | snake_case (`project_id`, `created_at`) | camelCase (`projectId`, `createdAt`) |
| Wails 绑定位置 | `wailsjs/go/main/App` | `wailsjs/go/handler/App` |
| Pipeline.Parallel | 有字段但未实现 | 同（阶段 3 后实现） |

所有改动符合 `docs/standards/ipc-spec.md`。

---

## 迁移步骤

### 1. 重新生成 Wails 绑定

```bash
cd /path/to/flowci
wails generate module
```

这会：
- 扫描 `main.go` 的 `Bind: []interface{}{app}`（现为 `handler.App`）
- 重新生成 `frontend/src/wailsjs/go/handler/App.{d.ts,js}`
- 重新生成 `frontend/src/wailsjs/go/models.ts`（含所有 handler 下 Request/Response struct 类型）
- 删除旧的 `wailsjs/go/main/App.*`（如果存在）

确认生成后，`frontend/src/wailsjs/go/handler/App.d.ts` 应该包含类似：

```ts
export function ListProjects(arg1:context.Context):Promise<Array<store.Project>>;
export function CreateProject(arg1:context.Context, arg2:handler.CreateProjectRequest):Promise<store.Project>;
export function ExportPipelineToYaml(arg1:context.Context, arg2:string):Promise<string>;
// ...
```

### 2. 建立 API 封装层

在 `frontend/src/api/` 下为每个领域建一个文件。骨架已经在仓里（见下面的文件），你只需要解开注释并按模板补充其余方法。

所有 `views/*.vue` 的 Wails import 必须改为 import from `src/api/`，**不得直接 import `wailsjs/go/handler/App`**（见 ipc-spec.md § 7.1 / frontend-spec.md § 1）。

### 3. 修改 views/*.vue 的 IPC 调用

各文件需要做的改动类型：

#### 3.1 import 路径

```diff
- import { ListPipelines, CreatePipeline, ... } from '../wailsjs/go/main/App'
+ import * as pipelinesApi from '../api/pipelines'
```

#### 3.2 方法调用

```diff
- const list = await ListPipelines(projectId)
+ const list = await pipelinesApi.list(projectId)
```

#### 3.3 入参结构化

原本前端传 object 字面量，wails 会当 `map[string]interface{}` 收；现在是强类型 struct。字段名不变（都是 camelCase）。

```diff
- await CreatePipeline({
-   projectId: '...',
-   name: '...',
-   steps: [...],
-   config: { stopOnFail: true }
- })
+ await pipelinesApi.create({
+   projectId: '...',
+   name: '...',
+   steps: [...],
+   config: { parallel: false, stopOnFail: true }
+ })
```

**关键差异**：后端现在强制 `config` 结构，`parallel` 是必须字段（即使为 false 也要传）。

#### 3.4 错误处理

旧协议：返回 `{success: false, error: "..."}`，前端读 `result.error` 判断。
新协议：失败直接 Promise reject，前端 try/catch。

```diff
- const result = await ImportPipelineFromYaml({projectId, yaml})
- if (result.error) {
-   toast.error(`导入失败: ${result.error}`)
- } else {
-   toast.success('导入成功')
- }
+ try {
+   await pipelinesApi.importFromYaml({projectId, yaml})
+   toast.success('导入成功')
+ } catch (e) {
+   toast.error(`导入失败: ${e instanceof Error ? e.message : String(e)}`)
+ }
```

同理 ExecutePipeline 现在失败走 reject：

```diff
- const res = await ExecutePipeline({pipelineId, projectId})
- if (!res.success) { ... }
+ try {
+   const res = await pipelinesApi.execute({pipelineId, projectId})
+   // res.success 字段仍有（表示步骤全部成功），但 handler 层报错才会 reject
+ } catch (e) {
+   // 流水线执行不成立（找不到、繁忙等）
+ }
```

#### 3.5 JSON 字段名

后端 tag 从 snake 改 camel：

| 旧 | 新 |
|---|---|
| `project_id` | `projectId` |
| `created_at` | `createdAt` |
| `updated_at` | `updatedAt` |
| `image_name` | `imageName` |
| `image_tag` | `imageTag` |
| `started_at` | `startedAt` |
| `finished_at` | `finishedAt` |
| `display_name` | `displayName` |
| `stop_on_fail` | `stopOnFail` |
| `on_fail` | `onFail` |

前端所有 `p.project_id` / `build.started_at` 等访问都要改。

### 4. 删除前端手写重复 interface

`PipelineView.vue` 等文件手写了 `interface Pipeline {...}` 等，现在 `models.ts` 自动生成了：

```diff
- interface Pipeline {
-   id: string
-   project_id: string   // 字段名还是错的
-   name: string
-   steps: PipelineStep[]
- }
+ import type { store } from '../wailsjs/go/models'
+ type Pipeline = store.Pipeline
```

### 5. 消除 N+1

原 `PipelineView.vue:228` 是：

```ts
for (const p of projects.value) {
  const pls = await ListPipelines(p.id)
  // ...
}
```

改用后端新加的 `ListAllPipelines`：

```ts
pipelines.value = await pipelinesApi.listAll()
```

### 6. 主题去 localStorage

原 `App.vue:100` 还有 `localStorage.getItem('flowci_theme')`。按 frontend-spec.md § 6.3，前端不持久化业务数据：

```diff
- const settings_val = localStorage.getItem('flowci_theme')
- if (settings_val === 'system') { applyTheme('system') }
+ const settings = await settingsApi.get()
+ if (settings.theme === 'system') { applyTheme('system') }
```

---

## 验证

```bash
wails dev       # 启动开发模式；前端会热加载，TS 报错会立即暴露
```

UI 验收清单：
- [ ] 项目列表能拉出（CheckDocker / ListProjects）
- [ ] 创建/编辑/删除项目正常
- [ ] 流水线列表、创建、导出 YAML、导入 YAML 正常
- [ ] 执行流水线（需要 docker 可用）
- [ ] 构建历史列表不再拉 log 字段（看 Network tab 数据大小）
- [ ] 镜像 / 容器列表正常
- [ ] Push / Deploy / Compose 正常
- [ ] 主题切换（dark/light/system）持久化到 SQLite，重启保留

---

## 常见问题

### wails generate module 报 "executable file not found"

wails 依赖 `go` 在 PATH。如 `go env` 正常而 wails 报找不到 go，可尝试：

```bash
# Windows PowerShell 中显式导出
$env:PATH = "C:\Program Files\Go\bin;$env:PATH"
wails generate module
```

### 生成后 `wailsjs/go/models.ts` 里没有 Request/Response 类型

这是 wails 已知限制：它只生成 Bind 方法签名中出现的 struct，嵌套 struct 要显式被签名引用。
如果某 Request struct 没出现，检查该 handler 方法签名参数是否是 `*Request` 而非 map。

### 类型报错 `handler.XxxRequest is not exported`

确保 struct 字段都大写开头（Go 导出规则），且在 `internal/handler/dto.go` 里。

---

## 完成后

- 更新 `docs/standards/ROADMAP.md` 把阶段 2.G 从 ⏸ 改 ✅
- 在 `CLAUDE.md` 删掉"前端未对接"的 TODO 提示
- Commit: `refactor(phase-2-frontend): 前端对接 handler 包新签名`
