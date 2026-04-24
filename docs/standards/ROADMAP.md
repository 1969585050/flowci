# FlowCI 整改路线图

> 版本: v1.0 (2026-04-25)
> 用途: 从当前 AI 生成的"能跑但不专业"状态迁移到"符合 backend/frontend/ipc/data-spec 的工程化状态"

---

## 整体概览

| 阶段 | 目标 | 预计工时 | 状态 |
|---|---|---|---|
| **阶段 0** | 立规范、定基线 | 0.5 天 | ✅ 完成 (commit `ea8ae00`) |
| **阶段 1** | 清理 Tauri 残留 + 死代码 + slog 日志 | 1 天 | ✅ 完成 (commit `565fb8d`) |
| **阶段 2** | 拆 main.go + 强类型化 + SQLite 改造 + exec 超时 | 3-5 天 | 🟢 后端完成；前端 2.G 需手动（见 FRONTEND_MIGRATION.md） |
| **阶段 3** | 凭证 keyring + build log 瘦身 + 参数校验 + 并发锁 | 2-3 天 | ⏳ 待开始 |

**总计**: 约 7-10 个工作日。每个阶段结束都是一次稳定基线，可以暂停。

---

## 阶段 0 — 立规范

**目标**: 把工程规则写清楚，后续改造有据可依。

### 交付物

- [x] `docs/standards/backend-spec.md`
- [x] `docs/standards/frontend-spec.md`
- [x] `docs/standards/ipc-spec.md`
- [x] `docs/standards/data-spec.md`
- [x] `CLAUDE.md`（仓根）
- [x] `docs/standards/ROADMAP.md`（本文件）
- [ ] commit 规范基线（单次 commit 提交所有规范）

### 验收标准

- 6 份文档全部可读
- `CLAUDE.md` 正确索引四份 spec
- 未来改动前能查得到对应规则

---

## 阶段 1 — 清理 ✅

**已完成** (commit `565fb8d`)

### 任务清单

#### 1.1 删 Tauri 残留 ✅
- [x] 删根目录 `package.json` / `vite.config.ts` / `index.html`
- [x] `wails.json` 改 `frontend:install` / `frontend:build` / `frontend:dev` / `frontend:dir` 指向 `frontend/`
- [x] `main.go` 的 `//go:embed` 改成 `all:frontend/dist`
- [x] `.gitignore` 加 `frontend/dist/.gitkeep` 占位（裸 `go build` 可过）
- [ ] 实际 `wails build -clean` 验证（阶段末或 v0.2.0 前手动验证）

#### 1.2 删死代码 ✅
- [x] 删 `pushImage`（无凭证版，原 main.go:618-620）
- [x] `PipelineConfig.Parallel` 加 TODO 注释（阶段 2 实现）
- [x] 删 `main_test.go` 里自写的 `contains()`，改 `strings.Contains`

#### 1.3 统一 YAML 类型 ✅
- [x] 新建 `internal/pipeline/yaml.go`，统一 `YamlPipeline`/`YamlStep`/`YamlConfig`
- [x] `main.go` + `main_test.go` + `main_integration_test.go` 全部使用包级定义
- [x] `main_integration_test.go` 遮蔽包名的 `pipeline` 局部变量重命名为 `created`

#### 1.4 日志改 slog ✅
- [x] 新建 `internal/logger/logger.go`：TextHandler + stderr + 日期切割文件
- [x] 19 处 `fmt.Printf`/`fmt.Println` 全部替换为 `slog.Info/Error/Debug`
- [x] `main()` 内提前初始化 logger

#### 1.5 统一命名 ⏸
- [ ] JSON tag 统一 camelCase → **推迟到阶段 2.2 与强类型化一起做**（先改命名再重写签名是返工）
- [ ] 前端手写 interface 与 wailsjs 对齐 → **推迟到阶段 2.2**

### 验收标准

- 根目录无 Tauri 相关文件
- `grep -r "fmt.Printf\|fmt.Println" main.go internal/` 无命中（除测试）
- YAML 类型在全仓唯一定义
- `go test ./...` 通过
- `wails build -clean` 生成 EXE 能启动

---

## 阶段 2 — 重构（重头戏）🟢 后端完成

**后端**: 完成 (commit `1f4cac7`)
**前端 2.G**: 需要本地 `wails generate module` 后手动迁移；详见 [FRONTEND_MIGRATION.md](./FRONTEND_MIGRATION.md)

### 完成项

- [x] 2.A SQLite 改造: WAL + schema_migrations 表 + migrations/*.sql embed runner + MaxOpenConns 放宽
- [x] 2.B 新建 internal/config: DataDir/LogDir 路径出口
- [x] 2.C 新建 internal/docker: 6 文件，超时常量表，exec.CommandContext，所有返回值强类型
- [x] 2.D 扩展 internal/pipeline: executor.go per-pipeline 锁 + validator.go
- [x] 2.E 新建 internal/handler: 11 文件，DTO/errors/各领域 Bind，全 camelCase JSON tag，error 替代 success-map 协议
- [x] 2.F main.go 瘦身到 54 行
- [x] 2.H 后端测试全绿 (go test ./... 通过；handler 4 个集成测试 + pipeline 6 个 YAML 测试)

### 未完成项（留给本地手动）

- [ ] 2.G 前端 `src/api/` 封装层 + 删手写 interface + 跟新 import 路径 + 错误协议从 map 改 try/catch
  - 步骤见 `docs/standards/FRONTEND_MIGRATION.md`
  - 依赖本地 `wails generate module`，AI 会话环境下 wails CLI 无法调用 go

**目标**: 代码结构符合 backend-spec，类型系统用起来，SQLite 用对。

### 2.1 拆 main.go

**目标结构**（见 [backend-spec.md § 1](./backend-spec.md)）:

- [ ] 建 `internal/` 目录树
- [ ] `internal/docker/` — client / image / container / build / compose
- [ ] `internal/pipeline/` — pipeline / executor / yaml / validator
- [ ] `internal/store/` — 保留现有，完善迁移
- [ ] `internal/handler/` — App struct + dto + errors
- [ ] `internal/logger/`（阶段 1 已建）
- [ ] `internal/config/` — 路径、默认值
- [ ] `main.go` 瘦到 ≤ 80 行（只启动）

### 2.2 强类型化

- [ ] `internal/handler/dto.go` 定义所有 Request / Response struct
- [ ] 所有 Bind 方法改签名为 `(ctx, *Req) (*Resp, error)`
- [ ] 删除所有对外的 `map[string]interface{}`
- [ ] 前端 `src/api/` 封装层建立
- [ ] 前端删除手写重复 interface，改 import wailsjs 类型

### 2.3 SQLite 改造

- [ ] `internal/store/store.go` 开 WAL + 放开 MaxOpenConns
- [ ] 建 `schema_migrations` 表
- [ ] 将现有 `CREATE TABLE IF NOT EXISTS` 拆成 `migrations/0001_initial.sql` ~ `0004_xxx.sql`
- [ ] 实现 migration runner（`internal/store/migrate.go`）
- [ ] 外键字段加索引

### 2.4 exec 超时

- [ ] `internal/docker/client.go` 定义超时常量表
- [ ] 所有 `exec.Command` 改 `exec.CommandContext`
- [ ] 所有调用点传入带 timeout 的 ctx

### 2.5 测试

- [ ] `internal/store/` 单元测试覆盖 ≥ 70%
- [ ] `internal/pipeline/` 单元测试覆盖 ≥ 60%
- [ ] 集成测试迁到 `main_integration_test.go` 或 `internal/handler/`

### 验收标准

- `main.go` ≤ 80 行
- 全仓 `grep "map\[string\]interface{}"` 命中 ≤ 前端传入参数（pipeline step config 这种动态字段）
- `go test ./... -race` 通过
- 前端类型来自 wailsjs，无手写重复
- SQLite WAL 开启（`.db-wal` 文件出现）
- `migrations/` 目录有有效迁移脚本

---

## 阶段 3 — 安全 & 体验

**目标**: 修补安全隐患、提升使用体验。

### 3.1 凭证 keyring

- [ ] 引入 `github.com/zalando/go-keyring`
- [ ] 新建 `internal/secret/keyring.go`（Set / Get / Delete / Mask）
- [ ] 修改 `PushImage`：password 不进 IPC Request 或仅短生命周期
- [ ] 新建 `registry_credentials` 表（存 registry_url + username + has_password）
- [ ] IPC 日志中间件遮蔽敏感字段（`mask:"true"` tag）

### 3.2 build log 瘦身

- [ ] `build_records` 表改 `log` TEXT → `log_path` TEXT + `log_size` INTEGER
- [ ] 构建时 docker output 写 `logs/builds/<build-id>.log`
- [ ] `GetBuildRecord` 按需读文件；列表查询不读
- [ ] `CleanupOldBuilds()` 清理超过 200 条的记录 + 对应 log 文件

### 3.3 参数白名单

- [ ] `internal/docker/validator.go` 实现正则校验（见 [ipc-spec.md § 6.3](./ipc-spec.md)）
- [ ] Container name / image ref / port / registry host / env key 全部过一遍
- [ ] DeployWithCompose 的 `workDir` 限制在 data dir 下

### 3.4 并发锁

- [ ] `internal/pipeline/executor.go` 加 per-pipeline `sync.Mutex`
- [ ] 重复提交同一 pipeline 返回 `ErrPipelineBusy`
- [ ] 前端 "运行" 按钮执行中禁用

### 3.5 前端体验

- [ ] 主题移除 localStorage，单走 SQLite + Pinia store
- [ ] `ListAllPipelines()` 后端聚合，消除前端 N+1
- [ ] 错误提示统一 toast，不用 alert / confirm 原生对话框
- [ ] 对话框组件抽出 `ConfirmDialog.vue`

### 验收标准

- 凭证在 OS keyring 可见（Windows Credentials Manager）
- DB 文件 < 50MB 在 200 条 build 记录下（log 都落文件）
- 非法 container name / port 提示清晰
- 同一 pipeline 双击运行只跑一次
- 主题切换后重启保留
- `go test ./... -race` 通过

---

## 跨阶段纪律

1. **每完成一阶段**：跑完验收标准，打一个 git tag（`v0.1.0-phase1` 等）
2. **每次 commit**：必须能 build，`go test` 必须过
3. **遇到范围外问题**：记到本文档"已知问题清单"，不在当前阶段处理
4. **规范修订**：spec 文档是活的，发现规则不合理当场改 spec 再改代码

---

## 已知问题清单（暂不处理）

- [ ] Docker client SDK 替换（体量 vs 能力 trade-off，暂保持 exec 方式）
- [ ] i18n（目前仅中文，若后续要出海再引入 vue-i18n）
- [ ] 多租户 / 用户账号系统（个人/小团队工具不做）
- [ ] CI/CD 流水线（GitHub Actions 自动出 release）— 放到 v1.0 后

---

## 变更记录

| 日期 | 变更 | 作者 |
|---|---|---|
| 2026-04-25 | 初版：规范 + 4 阶段路线图 | Claude |
