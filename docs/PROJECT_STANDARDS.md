# FlowCI 项目规范 v1.0

## 1. 项目概述

本文档定义了 FlowCI 项目的开发规范、代码风格和工作流程，确保团队协作一致性和代码质量。

---

## 2. 技术栈规范

### 2.1 技术选型

| 层级 | 技术 | 版本要求 |
|------|------|----------|
| 桌面框架 | Tauri | 2.0+ |
| 前端框架 | Vue 3 | 3.4+ |
| 语言 | TypeScript | 5.3+ |
| UI 组件库 | Naive UI | 2.36+ |
| 构建工具 | Vite | 5.1+ |
| 后端语言 | Go | 1.21+ |
| Docker SDK | docker-sdk-go | 1.0+ |
| 数据库 | SQLite | 3.44+ |

---

## 3. 项目结构规范

### 3.1 整体目录结构

```
FlowCI/
├── docs/                    # 项目文档
│   ├── SPEC.md             # 详细设计文档
│   ├── API.md              # API 接口文档
│   └── PROJECT_STANDARDS.md # 本文档
├── go/                      # Go 后端代码
├── src/                     # Vue 前端代码
├── src-tauri/              # Tauri 原生代码
├── scripts/                 # 辅助脚本
├── configs/                # 配置文件
└── README.md               # 项目说明
```

### 3.2 Go 项目结构

```
go/
├── cmd/                    # 应用程序入口
│   └── server/            # 主服务程序
│       └── main.go
├── pkg/                    # 公共包（可被外部引用）
│   └── docker/
│       └── client.go
├── internal/               # 内部包（仅限本项目使用）
│   ├── builder/           # 镜像构建模块
│   ├── deployer/          # 部署模块
│   ├── config/            # 配置管理模块
│   ├── pipeline/          # 流水线模块
│   └── registry/          # 镜像仓库模块
├── api/                    # API 定义和协议
│   └── v1/
├── pkgutil/                # 工具函数包
└── go.mod
```

### 3.3 前端项目结构

```
src/
├── assets/                 # 静态资源
│   ├── images/
│   └── styles/
├── components/            # 公共组件
│   ├── common/           # 通用组件
│   ├── build/            # 构建相关组件
│   ├── deploy/           # 部署相关组件
│   └── settings/         # 设置相关组件
├── views/                 # 页面视图
│   ├── ProjectsView.vue
│   ├── BuildView.vue
│   ├── DeployView.vue
│   └── SettingsView.vue
├── stores/               # Pinia 状态管理
│   ├── project.ts
│   ├── build.ts
│   └── settings.ts
├── router/               # Vue Router 配置
│   └── index.ts
├── utils/                # 工具函数
│   ├── api.ts            # API 请求封装
│   └── docker.ts         # Docker 相关工具
├── types/                # TypeScript 类型定义
│   ├── project.ts
│   ├── build.ts
│   └── deploy.ts
├── App.vue
└── main.ts
```

---

## 4. 命名规范

### 4.1 文件命名

| 类型 | 规范 | 示例 |
|------|------|------|
| Go 源文件 | 小写下划线 | `build_config.go` |
| Go 测试文件 | `*_test.go` | `builder_test.go` |
| Vue 组件文件 | PascalCase | `ProjectCard.vue` |
| TypeScript 类型文件 | 小写下划线 | `project_type.ts` |
| 配置文件 | 小写下划线或 kebab-case | `vite.config.ts` |

### 4.2 Go 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 包名 | 简短、小写、无下划线 | `builder`, `config` |
| 结构体名 | PascalCase | `BuildConfig`, `DeployResult` |
| 接口名 | PascalCase，可带 I 前缀 | `Builder`, `DockerClient` |
| 函数名 | PascalCase（导出）或 camelCase（内部） | `NewBuilder()`, `buildImage()` |
| 变量名 | camelCase | `imageTag`, `projectPath` |
| 常量名 | 全大写下划线或 PascalCase | `MaxRetries`, `StatusRunning` |
| 错误变量 | `Err` 前缀 | `ErrBuildFailed`, `ErrNotFound` |

### 4.3 TypeScript/Vue 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 组件名 | PascalCase，3 个单词以上 | `ProjectCard.vue`, `BuildLogViewer.vue` |
| 组件目录 | kebab-case | `build-config/` |
| 变量名 | camelCase | `imageTag`, `projectList` |
| 常量名 | 全大写下划线 | `MAX_RETRY_COUNT`, `API_BASE_URL` |
| 接口名 | PascalCase，可带 I 前缀 | `Project`, `IProject` |
| 类型名 | PascalCase | `BuildStatus`, `DeployResult` |
| 枚举名 | PascalCase | `BuildStatus` |
| 事件名 | kebab-case | `on-project-created` |
| CSS 类名 | BEM 或 kebab-case | `.project-card`, `.build-log__item` |

### 4.4 数据库命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 表名 | 全小写下划线 | `projects`, `build_records` |
| 列名 | 全小写下划线 | `project_name`, `created_at` |
| 索引名 | `idx_` 前缀 | `idx_project_id` |
| 外键名 | `fk_` 前缀 | `fk_project_build` |

---

## 5. 代码风格规范

### 5.1 Go 代码风格

**格式化**：
- 使用 `gofmt` 或 `goimports` 自动格式化
- 编辑器配置：VS Code Go 插件 / GoLand

**注释规范**：
```go
// Package builder provides Docker image building functionality.
//
// Usage:
//   b := builder.NewBuilder(dockerClient)
//   result, err := b.Build(ctx, config)
package builder

// BuildConfig holds the configuration for building an image.
type BuildConfig struct {
    ImageTag string
    Language Language
}

// NewBuilder creates a new Builder instance.
func NewBuilder(docker *docker.Client) *Builder {
    return &Builder{docker: docker}
}
```

**错误处理**：
```go
// 错误定义
var (
    ErrBuildFailed = errors.New("build failed")
    ErrImageNotFound = errors.New("image not found")
)

// 错误返回
if err != nil {
    return nil, fmt.Errorf("failed to build image: %w", err)
}
```

**Context 使用**：
```go
func (b *Builder) Build(ctx context.Context, cfg *BuildConfig) error {
    // 确保所有操作都支持 context
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // 继续执行
    }
}
```

### 5.2 TypeScript/Vue 代码风格

**类型定义**：
```typescript
// 使用 interface 定义对象类型
interface Project {
  id: string
  name: string
  language: Language
  createdAt: Date
}

// 使用 type 定义联合类型或别名
type BuildStatus = 'pending' | 'running' | 'success' | 'failed'

// 枚举
enum DeployStatus {
  Pending = 'pending',
  Running = 'running',
  Success = 'success',
  Failed = 'failed'
}
```

**组件规范**：
```vue
<template>
  <div class="project-card">
    <h3>{{ project.name }}</h3>
  </div>
</template>

<script setup lang="ts">
// 定义 props
interface Props {
  project: Project
  readonly?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  readonly: false
})

// 定义 emits
const emit = defineEmits<{
  (e: 'select', project: Project): void
  (e: 'delete', id: string): void
}>()

// Composition API
const handleSelect = () => {
  emit('select', props.project)
}
</script>

<style scoped>
.project-card {
  padding: 1rem;
}
</style>
```

**API 请求封装**：
```typescript
// utils/api.ts
import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000
})

api.interceptors.response.use(
  response => response.data,
  error => {
    console.error('API Error:', error)
    return Promise.reject(error)
  }
)

export default api

// 使用
import api from '@/utils/api'

export const getProjects = () => api.get('/projects')
export const createProject = (data: CreateProjectDTO) => api.post('/projects', data)
```

---

## 6. Git 规范

### 6.1 分支命名规范

| 分支类型 | 命名格式 | 示例 |
|----------|----------|------|
| 主分支 | `main` | `main` |
| 开发分支 | `develop` | `develop` |
| 功能分支 | `feature/<issue-id>-<描述>` | `feature/123-add-user-auth` |
| 修复分支 | `fix/<issue-id>-<描述>` | `fix/456-fix-docker-connect` |
| 热修复分支 | `hotfix/<issue-id>-<描述>` | `hotfix/789-critical-security` |
| 发布分支 | `release/<version>` | `release/v0.1.0` |

### 6.2 提交信息规范

**格式**：
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 类型**：

| Type | 说明 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `docs` | 文档更新 |
| `style` | 代码格式（不影响功能） |
| `refactor` | 重构（不是新功能或修复） |
| `perf` | 性能优化 |
| `test` | 测试相关 |
| `chore` | 构建或辅助工具更新 |
| `ci` | CI 配置更新 |

**示例**：
```
feat(builder): add multi-stage build support

- Add Java multi-stage Dockerfile template
- Add Node.js multi-stage Dockerfile template
- Support custom build args

Closes #123
```

```
fix(docker): resolve connection timeout issue

- Add retry mechanism for Docker ping
- Increase connection timeout to 10s
- Add better error messages

Closes #456
```

### 6.3 Pull Request 规范

**PR 标题**：
```
[Type] Short description (#issue)

示例：
[Feature] Add project management module (#123)
[Bugfix] Fix Docker connection timeout (#456)
```

**PR 描述模板**：
```markdown
## 概述
<!-- 简要描述这个 PR 的目的 -->

## 改动内容
<!-- 列出具体的改动 -->

## 测试情况
<!-- 说明你做了哪些测试 -->

## 截图
<!-- 如果有 UI 改动，附上截图 -->

## Checklist
- [ ] 代码符合代码规范
- [ ] 单元测试通过
- [ ] 文档已更新
- [ ] 没有警告或错误
```

---

## 7. API 设计规范

### 7.1 RESTful API 规范

**Base URL**: `/api/v1`

**资源命名**: 使用复数名词

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/projects` | 获取项目列表 |
| POST | `/projects` | 创建项目 |
| GET | `/projects/:id` | 获取单个项目 |
| PUT | `/projects/:id` | 更新项目 |
| DELETE | `/projects/:id` | 删除项目 |
| POST | `/projects/:id/build` | 触发构建 |
| POST | `/projects/:id/deploy` | 触发部署 |

### 7.2 请求响应格式

**成功响应**：
```json
{
  "success": true,
  "data": {
    "id": "123",
    "name": "my-project"
  }
}
```

**错误响应**：
```json
{
  "success": false,
  "error": {
    "code": "BUILD_FAILED",
    "message": "镜像构建失败",
    "details": "Docker daemon is not running"
  }
}
```

---

## 8. 文档规范

### 8.1 代码注释要求

| 文件类型 | 要求 |
|----------|------|
| Go 公共函数 | 必须有注释，说明功能、参数、返回值 |
| Go 私有函数 | 建议注释复杂逻辑 |
| TypeScript 接口 | 必须有注释 |
| Vue 组件 | 必须有组件说明注释 |
| 业务逻辑 | 必须有注释说明业务规则 |

### 8.2 文档更新要求

| 改动类型 | 必须更新的文档 |
|----------|---------------|
| 新增 API | API.md |
| 新增功能模块 | SPEC.md |
| 数据库表变更 | 数据库设计文档 |
| 部署方式变更 | README.md |
| 规范变更 | PROJECT_STANDARDS.md |

---

## 9. 测试规范

### 9.1 测试覆盖率要求

| 模块 | 覆盖率要求 |
|------|------------|
| 核心业务逻辑 | ≥ 80% |
| 工具函数 | ≥ 90% |
| API Handler | ≥ 70% |
| 前端组件 | ≥ 60% |

### 9.2 测试命名规范

**Go 测试**：
```go
func TestBuild_Success(t *testing.T) { ... }
func TestBuild_FailedWithInvalidDockerfile(t *testing.T) { ... }
func TestDeploy_Rollback(t *testing.T) { ... }
```

**Vue 组件测试**：
```typescript
describe('ProjectCard.vue', () => {
  it('should render project name', () => { ... })
  it('should emit select event when clicked', () => { ... })
})
```

---

## 10. 版本规范

### 10.1 版本号格式

使用 **语义化版本 (SemVer)**：
```
主版本号.次版本号.修订号
MAJOR.MINOR.PATCH

示例：v0.1.0, v1.2.3
```

| 版本类型 | 触发条件 |
|----------|----------|
| MAJOR | 破坏性 API 变更 |
| MINOR | 新功能（向后兼容） |
| PATCH | Bug 修复（向后兼容） |

### 10.2 版本状态

| 状态 | 说明 |
|------|------|
| `dev` | 开发中，不稳定 |
| `alpha` | 内部测试 |
| `beta` | 公开测试 |
| `stable` | 正式发布 |

---

## 11. 环境规范

### 11.1 环境类型

| 环境 | 用途 | 配置来源 |
|------|------|----------|
| 开发环境 | 本地开发 | `config/dev.yaml` |
| 测试环境 | CI/CD 测试 | `config/test.yaml` |
| 预生产环境 | 上线前验证 | `config/staging.yaml` |
| 生产环境 | 正式环境 | 环境变量 |

### 11.2 环境变量配置

所有配置必须通过环境变量管理，禁止硬编码。

#### Go 后端环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `FLOWCI_API_ADDR` | `localhost:3847` | API 服务器监听地址 |
| `FLOWCI_CONFIG_PATH` | `~/.config/flowci/config.yaml` | 配置文件路径 |
| `FLOWCI_DOCKER_HOST` | `unix:///var/run/docker.sock` | Docker daemon 地址 |
| `FLOWCI_LOG_LEVEL` | `info` | 日志级别 (debug/info/warn/error) |
| `FLOWCI_DB_PATH` | `~/.config/flowci/flowci.db` | SQLite 数据库路径 |

#### Tauri 前端环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `FLOWCI_API_URL` | `http://localhost:3847` | Go API 服务器地址 |
| `VITE_API_BASE_URL` | `/api/v1` | 前端 API 基础路径 |

#### 前端 .env 示例

```bash
VITE_API_BASE_URL=/api/v1
```

#### 开发环境启动

```bash
# 终端 1: 启动 Go API
export FLOWCI_API_ADDR=localhost:3847
cd go && go run cmd/server/main.go

# 终端 2: 启动前端
cd src && npm run dev

# 终端 3: 启动 Tauri
cd src-tauri && cargo tauri dev
```

### 11.3 配置优先级

1. 命令行参数（最高）
2. 环境变量
3. 配置文件 (`~/.config/flowci/config.yaml`)
4. 默认值（最低）

---

## 12. 持续集成规范

### 12.1 CI 流程

```
代码提交 → 静态检查 → 单元测试 → 构建 → 集成测试 → 构建产物
```

| 步骤 | 工具 | 超时 |
|------|------|------|
| 静态检查 (Go) | golangci-lint | 5min |
| 静态检查 (TS) | eslint | 5min |
| 单元测试 | go test / vitest | 10min |
| 构建 | go build / npm run build | 10min |
| 集成测试 | Docker 容器 | 15min |

### 12.2 合入要求

所有合入 `main` 分支的代码必须：
- [ ] 通过所有 CI 检查
- [ ] 至少一个 Review 通过
- [ ] 无未解决的警告
- [ ] 测试覆盖率达标

---

## 13. 附录

### 13.1 常用命令

**Go**：
```bash
# 格式化代码
go fmt ./...

# 运行测试
go test -v ./...

# 构建
go build -o flowci ./cmd/server

# 依赖检查
go mod tidy
```

**前端**：
```bash
# 安装依赖
npm install

# 开发模式
npm run dev

# 构建
npm run build

# 类型检查
npm run typecheck
```

### 13.2 规范维护

| 角色 | 职责 |
|------|------|
| 所有开发者 | 遵守本规范，发现问题及时反馈 |
| 项目负责人 | 审核规范变更，解释规范细节 |
| Tech Lead | 最终决策规范相关争议 |

---

**最后更新**: 2026-04-24
**版本**: v1.0
