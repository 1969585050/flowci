# FlowCI IPC 规范 (Wails Bind)

> 版本: v1.0 (2026-04-25)
> 适用范围: Go `internal/handler/` 与前端 `src/api/` 之间的边界契约
> 强制级别: 全部 **HIGH**（破坏契约 = 前后端两侧同时返工）

---

## 1. Bind 表面原则 [HIGH]

1. **薄**：handler 方法**不写业务**，只做「校验参数 → 调 service → 返回」
2. **稳定**：对外方法签名一旦发布，改动必须走 deprecation 流程
3. **类型化**：参数和返回都是命名 struct，禁止裸 map
4. **幂等**：同一请求重复调用应产生相同结果（写操作用业务 ID 去重）

---

## 2. 方法命名 [HIGH]

| 动作 | 前缀 | 示例 |
|---|---|---|
| 列表查询 | `List` | `ListProjects` / `ListPipelines(projectId)` |
| 单体查询 | `Get` | `GetProject(id)` / `GetBuildRecord(id)` |
| 新建 | `Create` | `CreateProject(req)` |
| 更新 | `Update` | `UpdateProject(req)` |
| 删除 | `Delete` | `DeleteProject(id)` |
| 执行 / 动作 | 动词原形 | `BuildImage(req)` / `PushImage(req)` / `ExecutePipeline(req)` |
| 导入导出 | `Import` / `Export` | `ImportPipelineFromYaml(req)` |
| 检查 | `Check` | `CheckDocker()` |

### 禁用
- ❌ `Do*` / `Process*` / `Handle*` 等无意义动词
- ❌ 同一资源复数形式混用（统一：`ListProjects` + `GetProject`，不出现 `GetProjects`）

---

## 3. 方法签名 [HIGH]

### 3.1 统一形态

```go
func (a *App) MethodName(ctx context.Context, req *MethodRequest) (*MethodResponse, error)
```

- 第一参数固定 `ctx context.Context`（Wails 会注入）
- 只有一个入参：**`req` 指针 struct**
- 返回：**`response` 指针 struct + error**

### 3.2 允许的例外

- **无参数查询**：`func (a *App) ListProjects(ctx) ([]*Project, error)` — 省略 req
- **仅需 id 的查询/删除**：`func (a *App) DeleteProject(ctx, id string) error` — id 直传
- **批量返回 slice**：`[]*T`（而非 `*[]T` 或 `[]T`）

### 3.3 禁止形态

```go
// ❌ map 入参
func (a *App) CreateProject(ctx context.Context, data map[string]interface{}) map[string]interface{}

// ❌ 只返回 bool 表成败
func (a *App) DeleteProject(ctx context.Context, id string) bool

// ❌ 返回带错误信息的 map
func (a *App) Build(...) map[string]interface{} {
    return map[string]interface{}{"success": false, "error": "..."}
}
```

---

## 4. DTO 规范 [HIGH]

### 4.1 位置

**所有 Request / Response struct 定义在 `internal/handler/dto.go`**，handler 其他文件 import。

### 4.2 命名

- Request：`<Method>Request`，如 `CreateProjectRequest`
- Response：`<Method>Response`，如 `BuildImageResponse`
- 单体实体（复用跨 Method）：`Project` / `Pipeline` / `BuildRecord`

### 4.3 JSON tag 一律 camelCase

```go
// ✅
type CreatePipelineRequest struct {
    ProjectID string         `json:"projectId"`
    Name      string         `json:"name"`
    Steps     []PipelineStep `json:"steps"`
    Config    PipelineConfig `json:"config"`
}

type PipelineStep struct {
    Type   string                 `json:"type"`
    Name   string                 `json:"name"`
    Retry  int                    `json:"retry"`
    OnFail string                 `json:"onFail"`
    Config map[string]interface{} `json:"config"`
}
```

**已知例外**：YAML 导入导出里的字段名跟 YAML 约定走（snake_case `on_fail` / `stop_on_fail`），仅限 YAML 文件内，不影响 IPC。

### 4.4 字段规约

- 时间字段 Go 端用 `time.Time`，Wails 自动序列化为 ISO 8601 字符串
- ID 字段一律 `string`（UUID v4）
- 可选字段用指针或 `omitempty`，必填字段非指针
- 敏感字段（password 等）**不得**出现在 Response，只能在 Request
- 不传敏感字段的 middleware：写入 IPC 日志前用 `secret.Mask()` 遮蔽

---

## 5. 错误协议 [HIGH]

### 5.1 契约

**Go 返回 error → Wails 自动 reject Promise → 前端 catch。**

```go
// ✅ 唯一模式
func (a *App) GetProject(ctx context.Context, id string) (*Project, error) {
    p, err := a.store.GetProject(id)
    if err != nil {
        if errors.Is(err, store.ErrNotFound) {
            return nil, handler.ErrProjectNotFound
        }
        return nil, fmt.Errorf("get project: %w", err)
    }
    return p, nil
}
```

```ts
// ✅ 前端
try {
  const p = await api.getProject(id)
} catch (e) {
  toast.error(e instanceof Error ? e.message : String(e))
}
```

### 5.2 禁止

```go
// ❌ 不得用 Response 字段表达错误
return &Response{Success: false, Error: "xxx"}, nil

// ❌ 不得既 return error 又在 struct 里带 error
return &Response{Error: "xxx"}, errors.New("xxx")
```

### 5.3 错误码（可选）

需要前端按错误类型做分支处理时，用哨兵错误 + 错误包装：

```go
var ErrDockerNotReady = errors.New("docker_not_ready")
```

前端：`e.message.includes('docker_not_ready')` 判断。不引入独立错误码表以保持简单。

---

## 6. 参数校验 [HIGH]

### 6.1 位置

- **handler 层做结构校验**：必填字段非空、枚举值合法、字符串格式（正则）、数值范围
- **service 层做业务校验**：资源存在、权限、状态机允许
- 禁止 handler 直接对 DB 操作

### 6.2 结构校验标准

```go
func (a *App) CreateProject(ctx context.Context, req *CreateProjectRequest) (*Project, error) {
    if req == nil {
        return nil, ErrBadRequest
    }
    if strings.TrimSpace(req.Name) == "" {
        return nil, fmt.Errorf("%w: name required", ErrBadRequest)
    }
    if !isValidLanguage(req.Language) {
        return nil, fmt.Errorf("%w: invalid language %q", ErrBadRequest, req.Language)
    }
    return a.projectSvc.Create(ctx, req)
}
```

### 6.3 白名单校验（安全硬要求）

以下字段必须过正则白名单（见 `internal/docker/validator.go`）：

| 字段 | 正则 |
|---|---|
| container name | `^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,62}$` |
| image name | `^[a-z0-9]+(?:[._-][a-z0-9]+)*(?:/[a-z0-9]+(?:[._-][a-z0-9]+)*)*(?::[a-zA-Z0-9_.-]+)?$` |
| port | `^\d{1,5}$` 且 1 ≤ n ≤ 65535 |
| registry host | `^[a-zA-Z0-9.-]+(?::\d{1,5})?$` |
| env key | `^[A-Z_][A-Z0-9_]*$` |

---

## 7. 前端绑定 [HIGH]

### 7.1 类型同步流程

```
Go handler/dto.go 改 struct
        ↓
wails generate module
        ↓
frontend/src/wailsjs/ 更新
        ↓
前端 src/api/ 直接 import 新类型
```

**禁止**前端手动维护与 Wails 生成的类型重复的 interface。

### 7.2 封装层

```ts
// src/api/projects.ts
import { ListProjects, CreateProject } from '../wailsjs/go/handler/App'
import type { handler } from '../wailsjs/go/models'

export async function listProjects(): Promise<handler.Project[]> {
  return await ListProjects()
}

export async function createProject(
  req: handler.CreateProjectRequest,
): Promise<handler.Project> {
  return await CreateProject(req)
}
```

`views/` 只 `import * as api from '../api/projects'`。

---

## 8. 日志 & 脱敏 [HIGH]

handler 入口统一日志中间件（或每方法首行写）：

```go
slog.Info("ipc",
    "method", "CreateProject",
    "req", secret.MaskStruct(req),  // 自动遮蔽敏感字段
)
```

敏感字段 tag：在 struct 上标注

```go
type PushImageRequest struct {
    Image    string `json:"image"`
    Registry string `json:"registry"`
    Username string `json:"username"`
    Password string `json:"password" mask:"true"`
}
```

`secret.MaskStruct` 识别 `mask:"true"` tag 输出 `***`。

---

## 9. 版本与兼容 [MED]

个人小团队工具，**不做多版本兼容**。但以下变更需提前沟通：

- 方法重命名
- 方法删除
- Request 字段删除 / 重命名
- Response 字段改变语义

变更流程：
1. 写新字段 / 新方法，保留老的
2. 前端逐步迁移
3. 老接口删除，打一条 commit：`refactor(ipc): drop deprecated Xxx`

---

## 10. 审查清单

- [ ] 方法名匹配 § 2 表格
- [ ] 签名是 `(ctx, *Request) (*Response, error)` 或其允许的例外
- [ ] DTO 定义在 `internal/handler/dto.go`
- [ ] JSON tag 全部 camelCase
- [ ] 错误通过 return error，不走 Response 字段
- [ ] 敏感字段打了 `mask:"true"` tag
- [ ] 所有用户输入都过白名单校验
- [ ] 前端通过 `src/api/` 调用，无直接 wailsjs import
- [ ] Go 改 struct 后跑过 `wails generate module`
