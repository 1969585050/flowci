# `src/api/` — IPC 封装层

## 为什么需要

按 `docs/standards/ipc-spec.md` § 7.1 和 `docs/standards/frontend-spec.md` § 5.1：

> 所有 Wails 调用必须通过 `src/api/**`。
> `views/` 只 `import * as api from '../api/xxx'`，不得直接 `import ... from '../wailsjs/...'`。

好处：
- 一次改 `src/api/xxx.ts`，所有 view 自动生效
- 类型收敛到一处（不再每个 view 手写 interface）
- 将来切换 IPC 实现（HTTP、WebSocket、mock）不需要改 view

## 现在的状态

**骨架待补**：具体 `src/api/*.ts` 文件尚未建，因为它们依赖 wails 自动生成的
`wailsjs/go/handler/App`（需本地执行 `wails generate module`）。

## 你要做的

按 `docs/standards/FRONTEND_MIGRATION.md` 第 1、2、3 节顺序走：

```bash
# 1. 重新生成 Wails 绑定
wails generate module
```

之后新建：
- `src/api/index.ts` — 聚合 re-export
- `src/api/projects.ts` — `listProjects / create / update / delete`
- `src/api/pipelines.ts` — CRUD + YAML 导入导出 + execute
- `src/api/builds.ts` — list + get
- `src/api/images.ts` — list + remove
- `src/api/containers.ts` — list + start/stop/remove + logs + deploy
- `src/api/compose.ts` — generate + deploy
- `src/api/push.ts` — push
- `src/api/settings.ts` — get + save
- `src/api/docker.ts` — checkDocker

每个文件的模板：

```ts
// src/api/projects.ts
import {
  ListProjects, CreateProject, UpdateProject, DeleteProject,
} from '../wailsjs/go/handler/App'
import type { handler, store } from '../wailsjs/go/models'

export const list = () => ListProjects()

export const create = (req: handler.CreateProjectRequest) =>
  CreateProject(req)

export const update = (req: handler.UpdateProjectRequest) =>
  UpdateProject(req)

export const remove = (id: string) =>
  DeleteProject(id)

export type Project = store.Project
```

## 配套

- Toast 提示：`import { useToast, withErrorToast } from '../composables/useToast'`
- 确认对话框：`import { useConfirm } from '../composables/useConfirm'`
- 设置：`import { useSettings } from '../composables/useSettings'`

views/*.vue 调用示例：

```ts
import * as projectsApi from '../api/projects'
import { useToast, withErrorToast } from '../composables/useToast'
import { useConfirm } from '../composables/useConfirm'

const toast = useToast()
const { ask } = useConfirm()

async function deleteProject(id: string, name: string) {
  const ok = await ask({
    title: '删除项目',
    message: `确定删除 "${name}" 吗？此操作不可撤销。`,
    variant: 'danger',
  })
  if (!ok) return
  await withErrorToast(projectsApi.remove(id), '删除失败')
  toast.success('已删除')
  await refreshList()
}
```
