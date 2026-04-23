---
name: "sdlc-manager"
description: "Software Development Life Cycle Manager - manages projects from PRD to deployment with quality gates. Invoke when user wants to start a new project, review PRD, or manage project milestones."
---

# SDLC Manager

Professional software development lifecycle management skill with quality gates and milestone tracking.

## Core Philosophy

**"专业的事情需要专业的流程"** - Professional work requires professional processes.

Every stage must pass quality review (≥95 score after 3 rounds) before proceeding.

---

## Phase 1: Project Initiation

### 1.1 Create PRD (Product Requirements Document)

**Trigger**: User wants to start a new project

**PRD Template** (`PROJECT_NAME/PRD.md`):

```markdown
# [Project Name] PRD

## 1. 项目概述
- 项目名称:
- 项目类型: (桌面应用/Web服务/移动应用/微服务/CLI工具)
- 核心价值: 一句话说清楚项目解决什么问题
- 目标用户: 谁会使用这个产品

## 2. 功能需求

### 2.1 用户故事 (User Stories)
| ID | 标题 | 作为... | 我想... | 以便... | 优先级 |
|----|------|---------|---------|---------|--------|
| US-001 | xxx | 用户 | 做xxx | 达到xxx | P0 |

### 2.2 功能列表
| 功能 | 描述 | 验收标准 |
|------|------|----------|
| F001 | xxx | 可衡量、可测试的标准 |

### 2.3 非功能需求
| 需求类型 | 描述 | 指标 |
|----------|------|------|
| 性能 | 响应时间 | <200ms |
| 可用性 | 系统 uptime | 99.9% |
| 安全 | 数据加密 | AES-256 |

## 3. 技术选型

| 层级 | 技术 | 版本 | 选择理由 |
|------|------|------|----------|
| 前端框架 | | | |
| 后端框架 | | | |
| 数据库 | | | |
| 部署方式 | | | |

## 4. 项目里程碑

| 阶段 | 交付物 | 时间线 | DoD |
|------|--------|--------|-----|
| M1 | 原型/POC | 2周 | 可演示的原型 |
| M2 | MVP | 4周 | 核心功能可用 |
| M3 | Beta | 6周 | 稳定可测试 |
| M4 | GA | 8周 | 生产可用 |

## 5. 团队角色

| 角色 | 职责 | 人数 |
|------|------|------|
| PM | 产品管理 | 1 |
| 开发 | 编码实现 | 2-4 |
| 测试 | 质量保障 | 1 |

## 6. 风险评估

| 风险 | 影响 | 概率 | 应对策略 |
|------|------|------|----------|
| 技术难点 | 高 | 中 | 提前POC验证 |
```

### 1.2 PRD Quality Review

**Trigger**: After PRD draft is created

**Process**:
1. **Round 1**: Functional completeness check
2. **Round 2**: Technical feasibility check
3. **Round 3**: Risk and constraint check

**Scoring Criteria** (100 points):
- 功能完整性: 25分
- 技术可行性: 25分
- 需求清晰度: 20分
- 非功能需求: 15分
- 风险可控性: 15分

**Pass Threshold**: ≥95/100 after any 3 rounds

**Output Format**:
```json
{
  "round": 1,
  "score": 82,
  "issues": [
    {"severity": "high", "issue": "缺少性能指标定义", "suggestion": "补充响应时间要求"}
  ],
  "next_action": "修复高优先级问题后进入Round 2"
}
```

---

## Phase 2: Project Setup

### 2.1 Project Structure

**Trigger**: After PRD approved

**Standard Structure**:
```
project/
├── .github/
│   └── workflows/          # CI/CD
├── docs/                   # 项目文档
│   ├── api/               # API 规范
│   ├── guides/            # 开发指南
│   └── standards/          # 规范文档
├── src/                    # 前端代码
├── backend/               # 后端代码
├── scripts/               # 构建脚本
├── tests/                 # 测试
├── Makefile
├── README.md
└── CONTRIBUTING.md
```

### 2.2 Tech Stack Selection

**Decision Framework**:

| 维度 | 权重 | 评估项 |
|------|------|--------|
| 团队熟悉度 | 30% | 团队是否已有经验 |
| 社区活跃度 | 20% | GitHub stars, 更新频率 |
| 生态系统 | 20% | 周边工具、库的支持 |
| 性能表现 | 15% | 基准测试数据 |
| 学习曲线 | 15% | 新人上手时间 |

**Current Best Practices** (2026):

| 类别 | 推荐技术 | 最低版本 |
|------|----------|-----------|
| 桌面应用 | Tauri | 2.10+ |
| 前端框架 | Vue 3 | 3.5+ |
| 类型系统 | TypeScript | 5.7+ |
| 构建工具 | Vite | 6.2+ |
| 后端语言 | Go | 1.26+ |
| HTTP框架(Go) | Chi | 5.1+ |
| 容器 | Docker SDK | 24.0+ |

### 2.3 Standards Setup

**Trigger**: Project structure created

**Required Standards**:

1. **错误码规范** (`docs/standards/error-codes.md`)
```markdown
| 范围 | 类别 |
|------|------|
| 0 | 成功 |
| 1001-1999 | 请求错误 |
| 2001-2999 | 业务错误 |
| 3001-3999 | 系统错误 |
| 4001-4999 | 外部错误 |
```

2. **命名规范**:
   - 文件: `snake_case` (Go) / `PascalCase` (Vue组件)
   - 函数: `PascalCase` (导出) / `camelCase` (私有)
   - 常量: `PascalCase`
   - 数据库: `snake_case` 复数表名

3. **代码风格**:
   - Go: `gofmt` + golangci-lint
   - TypeScript: ESLint + strict mode
   - Vue: `<script setup lang="ts">`

4. **Git 规范**:
   - 分支: `feature/xxx`, `fix/xxx`, `release/vx.x.x`
   - Commit: `<type>(<scope>): <subject>`

---

## Phase 3: Development

### 3.1 Quality Gates

**Before Each Milestone**:
- [ ] 代码规范检查通过
- [ ] 单元测试覆盖率 ≥80%
- [ ] 集成测试通过
- [ ] 文档更新完成
- [ ] PR Review 通过

### 3.2 Task Breakdown

**Template**:
```markdown
## Phase X: [名称] (X人天)

| 任务 | 负责人 | 预计工时 | 依赖 |
|------|--------|----------|------|
| T001 | @张三 | 2d | - |
| T002 | @李四 | 3d | T001 |
```

**DoD (Definition of Done)**:
- 代码完成并通过 lint
- 单元测试 ≥80% 覆盖
- 集成测试通过
- 文档已更新
- Tech Lead 批准

### 3.3 Code Review Checklist

**Review Points**:
1. 逻辑正确性
2. 边界情况处理
3. 错误处理完整性
4. 性能考虑
5. 安全漏洞检查
6. 可维护性
7. 测试覆盖

**Review Output**:
```json
{
  "review_id": "PR-xxx",
  "result": "approved" | "changes_requested",
  "score": 85,
  "issues": []
}
```

---

## Phase 4: Deployment

### 4.1 CI/CD Pipeline

**Required Stages**:
1. `lint` - 代码规范检查
2. `test` - 单元测试
3. `build` - 构建验证
4. `e2e` - 端到端测试
5. `deploy` - 部署

### 4.2 Environment Management

| 环境 | 用途 | 部署方式 |
|------|------|----------|
| dev | 开发调试 | 本地 |
| test | QA测试 | CI自动 |
| staging | 预生产 | 手动 |
| prod | 生产 | 发布流程 |

---

## Quality Review Template

### Round Template

```markdown
## Quality Review Round N

**Date**: YYYY-MM-DD
**Reviewer**: [Name]
**Stage**: [Phase Name]

### Score: X/100

| 维度 | 得分 | 说明 |
|------|------|------|
| 功能完整性 | /25 | |
| 技术质量 | /25 | |
| 需求清晰度 | /20 | |
| 非功能需求 | /15 | |
| 风险可控性 | /15 | |

### Issues Found

| 严重度 | 问题 | 建议 | 状态 |
|--------|------|------|------|
| High | | | Open |
| Medium | | | Open |
| Low | | | Open |

### Verdict
- [ ] Pass (≥95)
- [ ] Need Fix (send back to team)
```

---

## Usage

### Quick Start

```
/sdlc create-project --name "MyApp" --type "desktop" --template "tauri-vue-go"
```

### Review PRD
```
/sdlc review-prd --file "PRD.md" --rounds 3 --threshold 95
```

### Create Milestone
```
/sdlc create-milestone --name "M1-POC" --duration "2w" --deliverables "原型演示"
```

### Check Quality Gate
```
/sdlc quality-gate --stage "development" --checklist "full"
```

---

## Best Practices Summary

1. **PRD first, code later** - 需求不清不动工
2. **3 rounds of criticism** - 质量问题早发现
3. **≥95 to pass** - 低分必须整改
4. **Standards are mandatory** - 规范必须遵守
5. **Milestones with DoD** - 每个里程碑有明确完成标准
6. **No hardcoding** - 配置外置，环境变量优先
7. **Tests are mandatory** - 测试覆盖率是质量底线
8. **Documentation update** - 代码改动必须更新文档

---

## Integration with OpenSpace

This skill can use OpenSpace for:
- PRD generation and optimization
- Code review assistance
- Documentation generation
- Risk assessment

Invoke OpenSpace when:
- User asks to "optimize" or "improve" documents
- Need AI assistance for complex decisions
- Want to auto-generate boilerplate code
