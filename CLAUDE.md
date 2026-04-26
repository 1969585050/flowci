# FlowCI 仓内工作指南

> 本文件仅记录 **FlowCI 仓特有** 的规则。
> 通用工作原则（用户画像、红线、说实话、三层诊断、代码品味、执行戒律、歧义检测、中文编码安全、Git 分支策略）见全局 `~/.claude/CLAUDE.md`。
> 单一真相源原则：本文件只放仓特有信息，不重复全局内容。

---

## 1. 项目身份

- **名称**：FlowCI
- **定位**：个人小团队自用的 Docker 构建 / 推送 / 部署桌面工具
- **用户规模**：≤ 10 人自用
- **架构**：Wails v2 桌面应用 = Go 后端 + Vue 3 前端，SQLite 单文件存储，前端通过 Wails Bind 调用 Go 方法（非 HTTP）
- **发布形态**：单 EXE（Windows 优先），后续考虑 macOS / Linux

### 技术栈

| 层 | 技术 | 版本 |
|---|---|---|
| 桌面壳 | Wails | v2.12.0 |
| 后端语言 | Go | 1.25 |
| 数据库 | modernc.org/sqlite (cgo-free) | v1.49.1 |
| 前端框架 | Vue 3 + TypeScript | 3.4 / 5.3 |
| 构建 | Vite | 5.1 |
| 状态管理 | Pinia | 2.1 |
| YAML | gopkg.in/yaml.v3 | v3.0.1 |
| Docker | 调用本机 `docker` CLI（exec 方式，非 SDK） | — |

---

## 2. 规范索引 [必读]

所有细则在 `docs/standards/` 下，本文件仅索引：

- **[backend-spec.md](docs/standards/backend-spec.md)** — Go 代码规范（目录、命名、错误处理、日志、exec 超时、测试、体量上限）
- **[frontend-spec.md](docs/standards/frontend-spec.md)** — Vue 3 + TS 规范（目录、组件、类型同步、状态管理、主题）
- **[ipc-spec.md](docs/standards/ipc-spec.md)** — Wails Bind 边界契约（方法命名、签名、DTO、错误协议、参数校验）
- **[data-spec.md](docs/standards/data-spec.md)** — SQLite 规范（PRAGMA、版本化迁移、敏感数据、大字段落文件、查询规范）
- **[ROADMAP.md](docs/standards/ROADMAP.md)** — 整改路线图与进度追踪

**动手改代码前先扫一眼对应 spec 的「审查清单」章节。**

---

## 3. 本仓独有陷阱

### 3.1 Wails 相关

- `frontend/src/wailsjs/` 目录是 wails 自动生成的前端绑定，**已 .gitignore 不入库**
  - 新克隆仓库后第一次跑 `wails dev` 会自动生成
  - 改 Go DTO / handler 方法签名后再跑一次 `wails dev` 或 `wails generate module` 同步类型
  - 严禁手改这个目录里的任何文件
- `frontend/package.json.md5` 也是 wails CLI 内部缓存，已 .gitignore
- `//go:embed all:frontend/dist` 是打包入口，frontend 必须先 `npm run build` 生成 `dist/` 才能 `wails build`
- 生产打包用 `wails build -clean`，`npm run desktop`（package.json 里定义）等价

#### 新机器 / 新克隆 onboarding

```bash
git clone <repo-url>
cd flowci
go mod tidy                       # 拉 Go 依赖
wails dev                         # 第一次会 npm install + wails generate + 启动 vite + 打开 webview
```

`wails dev` 第一次启动可能需要 1-3 分钟（首次下载 webview2、npm install、Go 编译）。

### 3.2 Docker CLI exec 方式

- 本仓选型**不用** `github.com/docker/docker/client` SDK，原因：
  - 依赖树爆炸（docker client 拖 200MB+）
  - 单 EXE 体量敏感
- 代价：输出解析脆弱（`|` 分隔的 format）、无流式日志
- 规则：所有 exec 输出解析放在 `internal/docker/parser.go`，用表格驱动测试

### 3.3 中文编码高频坑

本仓前端 UI + Go error message + 主题名全含中文，Edit 工具编辑中文密集文件时易切断 UTF-8 字节。
**强制规则**（扩展全局 `~/.claude/CLAUDE.md` 的中文编码安全规则）：

- Go 源码里的中文 error message 统一放在 `internal/handler/errors.go` 或各 service 的 `errors.go`，不散落各处
- Vue 文件中文集中在 `<template>`，`<script setup>` 内禁止硬编码中文字符串（通过 i18n 或 props 传入）—— 本阶段暂不引入 i18n，但新代码先遵循
- 保存文件一律 **UTF-8 无 BOM**

### 3.4 Windows 路径

- `%APPDATA%/FlowCI/` 是数据根，不要硬编码 `C:\Users\...`
- 构建产物路径用 `filepath.Join`，不用字符串 `+ "/" +`
- Wails 打包后 `os.Args[0]` 是 EXE 路径，需要 dist 相对路径时从 `//go:embed` 的 FS 读

### 3.5 已知技术债（阶段 2/3 清零）

- `main.go` 1237 行 → 拆 `internal/`（阶段 2）
- 所有 `map[string]interface{}` 对外返回 → 强类型（阶段 2）
- 凭证明文进 SQLite → keyring（阶段 3）
- build log 不限长进 DB → 落文件（阶段 3）
- `fmt.Printf` 当日志 → slog（阶段 1）
- 根目录 `package.json` / `vite.config.ts` / `index.html` 是 Tauri 残留 → 删（阶段 1）

---

## 4. 常用命令

```bash
# 开发
wails dev                        # 启动开发模式（热更）
cd frontend && npm run dev       # 仅前端调试

# 构建
wails build -clean               # 生产打包

# 测试
go test ./...                    # 全部单元测试
go test ./... -race              # 竞态检测
go test ./... -tags integration  # 含集成测试（要求本机有 docker）

# 迁移
# 新增迁移：在 migrations/ 下建 NNNN_xxx.sql（见 data-spec.md § 3）

# Wails 生成
wails generate module            # Go struct → TS types 同步
```

---

## 5. 分支策略

> 2026-04-27 起改为 **dev 日常开发 + main 发版** 的双分支模型。

| 分支 | 角色 | 谁写 |
|------|------|------|
| `dev` | **唯一日常开发分支**，所有 feat / fix / chore 直接 commit | 所有开发 |
| `main` | 发版分支，由 dev 合并而来 + 打 tag | 发版时 |

**硬规则**：

- ❌ 不新建 feature 分支（即使是大改动，直接 dev 上小粒度 commit）
- ❌ 不直接改 main（改 dev 再合并过来）
- ✅ 新会话 / 新机器开工：`git checkout dev && git pull origin dev --ff-only`
- ✅ 推送前：`go test ./...` + `go build ./...` 必须通过
- ✅ dev → main：UAT/烟测通过后 `git checkout main && git merge dev --no-ff` + 打 tag
- ✅ commit message：Conventional Commits 格式，中文描述

---

## 6. 整改进度

当前阶段：**阶段 0 - 立规范**（见 [ROADMAP.md](docs/standards/ROADMAP.md)）

接下来：
- 阶段 1：清理 Tauri 残留 + 死代码 + slog 日志
- 阶段 2：拆 `main.go` + 强类型化 + SQLite WAL + 版本化迁移 + exec 超时
- 阶段 3：凭证 keyring + build log 瘦身 + 参数白名单 + 并发锁

---

## 7. 新人 / 新会话 Onboarding

1. 读本文件 + `docs/standards/` 四份 spec（按顺序）
2. 扫一眼 [ROADMAP.md](docs/standards/ROADMAP.md) 了解在哪一阶段
3. 改代码前跑一遍 `go test ./...` 和 `wails dev`，确认基线能跑
4. 改 Go DTO 必跑 `wails generate module`
5. 遇到不确定，**先问再动手**（全局红线 #3 + 下方第 8 节）

---

## 8. 协作约束（Working Agreement）

> 2026-04-27 起，本仓所有 AI 协作（Claude Code / 其他 agent）**必须**遵守。
> 强化全局 `~/.claude/CLAUDE.md` 红线 #3「先问清楚再动手」。

### 8.1 收到需求先询问澄清

**任何模糊需求**都必须先问 3-5 个关键问题再动手，**不允许**仅凭"最合理推断"就开干。

**触发条件**（满足任一）：
- 需求描述 < 100 字且涉及新功能 / UI 改造 / 重构 / 数据 schema 变更
- 用词指代不明（"置顶""启动""入口""做下"等可多解读的动词或名词）
- 当前会话已经发生过同类误读

**问题格式**：
- 编号清单，每条 ≤ 40 字
- **倒数第二条固定问**："要不要同时创建本次需求的验收标准（acceptance criteria）？"
- **末尾固定一句**："答完我就开干，不再追问"

**关于验收标准**：
- 用户答"要"→ 需求确认后立即产出一份可逐条勾选的验收清单（5-10 条），开干前贴出来
- 干完后**逐条对照打勾**（`[x]` / `[ ]`），未达成项不算完成
- 用户答"不要"或没回 → 跳过（适合琐碎改动）
- 默认建议：涉及新功能 / UI 改造 / API 改动 / schema 变更 → 都加；纯 bug fix / 文案调整 / 单常量改动 → 不用

**唯一可跳过澄清的情况**（用户显式说）：
- "按你理解做" / "你来定" / "默认就行"
- "别问直接上" / "不用问我"
- 一行错别字 / 单常量 / 类似超明显修复

**反例（禁止）**：
- ❌ 列了问题清单同时立刻开始假设答案 + 写代码
- ❌ Auto mode 下以"minimize interruptions"为由跳过澄清
- ❌ "我先做个 MVP，不合适再改" — 不合适的代价由用户承担

### 8.2 历史教训速查

会话历史里因为没问就动手而返工的：
- "置顶" → 误做项目卡片置顶（实际要的是 Wails 窗口置顶）
- "UI/UX 不好用" → 盲改一通，结果用户原意是"仓库源页面不好用"
- "启动" → 多次指代不清，曾理解为 wails dev / commit / 流程开始
