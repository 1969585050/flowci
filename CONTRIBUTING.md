# Contributing to FlowCI

感谢您对 FlowCI 项目的贡献！请仔细阅读以下指南。

## 目录

- [行为准则](#行为准则)
- [快速开始](#快速开始)
- [开发环境](#开发环境)
- [开发流程](#开发流程)
- [代码规范](#代码规范)
- [Git 工作流](#git-工作流)
- [Pull Request 指南](#pull-request-指南)
- [测试要求](#测试要求)
- [文档更新](#文档更新)

---

## 行为准则

我们承诺为 FlowCI 社区提供一个友好、安全的环境，无论经验水平如何。所有参与者都必须遵守我们的行为准则。

### 我们承诺

- 用包容性的语言
- 尊重不同的观点和经验
- 接受建设性批评
- 关注社区最佳利益
- 对其他社区成员表现出同理心

### 不可接受的行为

- 公开或私下骚扰
- 人身攻击或贬低性评论
- 未经同意发布私人信息
- 其他不道德或不专业的行为

---

## 快速开始

### 1. Fork 仓库

点击 GitHub 页面右上角的 "Fork" 按钮。

### 2. 克隆代码

```bash
git clone https://github.com/YOUR_USERNAME/flowci.git
cd flowci
```

### 3. 添加上游仓库

```bash
git remote add upstream https://github.com/flowci/flowci.git
```

### 4. 创建开发分支

```bash
git checkout -b feature/your-feature-name
```

---

## 开发环境

### 前置要求

| 工具 | 版本要求 |
|------|----------|
| Node.js | >= 18.0.0 |
| Go | >= 1.21 |
| Rust | >= 1.70 |
| Docker | >= 24.0 |
| Git | >= 2.40 |

### 环境设置脚本

```bash
# 克隆后运行设置脚本
./scripts/setup.sh
```

### 启动开发服务器

```bash
# 终端 1: 启动 Go API 服务器
cd go && go run cmd/server/main.go

# 终端 2: 启动前端开发服务器
cd src && npm run dev

# 终端 3: 启动 Tauri 桌面应用
cd src-tauri && cargo tauri dev
```

---

## 开发流程

### Issue 创建

在开始工作之前，请先创建或选择一个 Issue：

1. 搜索现有 Issue 确保没有重复
2. 创建新 Issue 并描述：
   - 清晰的问题描述
   - 复现步骤（如果是 Bug）
   - 预期行为
   - 截图或日志（如适用）

### 分支命名规范

| 分支类型 | 命名格式 | 示例 |
|----------|----------|------|
| 功能分支 | feature/issue-id-description | feature/123-add-dark-mode |
| 修复分支 | fix/issue-id-description | fix/456-container-timeout |
| 重构分支 | refactor/description | refactor/api-error-handling |
| 文档分支 | docs/description | docs/update-api-doc |
| 实验分支 | experiment/description | experiment/new-builder |

### 提交信息规范

使用 Angular Commit Message 格式：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 类型**:

| Type | Description |
|------|-------------|
| feat | 新功能 |
| fix | Bug 修复 |
| docs | 文档变更 |
| style | 代码格式（不影响功能） |
| refactor | 重构（不是新功能或修复） |
| perf | 性能优化 |
| test | 测试相关 |
| chore | 构建/工具变更 |

**Examples**:

```
feat(build): add multi-stage Dockerfile support

fix(deploy): resolve container restart loop issue

docs(api): update deployment API specification

refactor(docker): simplify image build context creation

test(builder): add unit tests for Dockerfile generation
```

---

## 代码规范

### Go 代码规范

遵循 [Google Go 代码规范](https://golang.org/doc/effective_go) 并使用以下工具：

```bash
# 格式化代码
go fmt ./...

# 运行 linter
golangci-lint run

# 运行测试
go test ./... -cover
```

**关键规则**:
- 使用 `gofmt` 格式化代码
- 导出函数必须有注释
- 错误必须被处理
- 优先使用结构体而非接口
- 避免嵌套过深（最多 3 层）

### TypeScript 代码规范

遵循 ESLint 配置和以下规则：

```bash
# 运行 linter
npm run lint

# 运行类型检查
npm run typecheck
```

**关键规则**:
- 使用 TypeScript 严格模式
- 避免使用 `any` 类型
- 使用 `const` 优先于 `let`
- 组件使用 PascalCase
- 函数使用 camelCase
- 导入排序：内置 → 外部 → 内部

### 文件组织

```
模块/
├── module.go           # 主文件
├── module_test.go      # 测试文件
├── submodule/
│   ├── submodule.go
│   └── submodule_test.go
└── doc.go             # 包文档
```

---

## Git 工作流

### 基础流程

```
     upstream/main
           │
           │   feature/123-feature
           │        │
           │        ▼
           │    develop ────────────► PR & Merge
           │        ▲
           │        │
           └────────┘
              merge
```

### 详细步骤

1. **同步上游代码**

```bash
git fetch upstream
git checkout main
git merge upstream/main
```

2. **创建功能分支**

```bash
git checkout -b feature/123-add-xxx
```

3. **提交代码**

```bash
git add .
git commit -m "feat(scope): add xxx feature"
```

4. **同步并变基**

```bash
git fetch upstream
git rebase upstream/main
```

5. **推送并创建 PR**

```bash
git push -u origin feature/123-add-xxx
```

6. **创建 PR**

在 GitHub 上创建 Pull Request，填写 PR 模板。

### Commit 规范检查

提交前自动检查：

```bash
# 安装 commitlint
npm install -g @commitlint/cli @commitlint/config-conventional

# 验证 commit
commitlint --from HEAD~1
```

---

## Pull Request 指南

### PR 标题格式

```
<type>(<scope>): <subject> (#issue_number)
```

**Examples**:
```
feat(build): add multi-arch image support (#123)
fix(deploy): resolve rollback failure on timeout (#456)
```

### PR 描述模板

```markdown
## 概述
<!-- 简短描述这个 PR 做什么 -->

## 变更类型
- [ ] 新功能 (feat)
- [ ] Bug 修复 (fix)
- [ ] 重构 (refactor)
- [ ] 文档更新 (docs)
- [ ] 测试更新 (test)

## 测试方案
<!-- 描述如何测试这些更改 -->

## 截图
<!-- 如果有 UI 变更，添加截图 -->

## Checklist
- [ ] 代码遵循项目代码规范
- [ ] 自测通过
- [ ] 添加/更新测试用例
- [ ] 更新相关文档
```

### Review 流程

1. **自动化检查** 必须通过：
   - ✅ CI/CD 流水线
   - ✅ 代码风格检查
   - ✅ 单元测试覆盖率

2. **人工 Review**：
   - 至少 1 人 approve
   - 所有 conversation 已解决
   - 分支已同步最新 main

### Merge 策略

| 分支类型 | Merge 方式 | Squash |
|----------|------------|--------|
| feature/* | Squash and Merge | Yes |
| fix/* | Squash and Merge | Yes |
| release/* | Merge Commit | No |
| hotfix/* | Merge Commit | No |

---

## 测试要求

### 覆盖率要求

| 模块 | 最低覆盖率 |
|------|-----------|
| Go API handlers | 80% |
| Go business logic | 90% |
| TypeScript components | 70% |
| TypeScript utilities | 80% |

### 运行测试

```bash
# Go 测试
cd go
go test ./... -v -coverprofile=coverage.out

# 前端测试
cd src
npm run test:unit
npm run test:e2e

# Tauri 测试
cd src-tauri
cargo test
```

### 测试文件命名

```
module.go           -> module_test.go
internal/
├── api/
│   ├── handler.go  -> handler_test.go
│   └── client.go   -> client_test.go
```

---

## 文档更新

### 必须更新的文档

| 变更类型 | 必须更新 |
|----------|----------|
| 新功能 | README, API 文档 |
| API 变更 | docs/api/openapi.yaml |
| 配置变更 | 环境变量说明 |
| 新增依赖 | README 依赖列表 |

### 文档格式

- Markdown 文件使用中文标点符号
- 代码块必须指定语言
- 图片放在 `docs/assets/` 目录
- API 文档使用 OpenAPI 3.0 格式

---

## 问题反馈

如果您发现 Bug 或有功能请求，请创建 Issue 并使用合适的模板。

感谢您的贡献！ 🎉
