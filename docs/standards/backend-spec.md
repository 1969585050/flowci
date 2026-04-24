# FlowCI 后端开发规范 (Go)

> 版本: v1.0 (2026-04-25)
> 适用范围: `main.go` 与 `internal/**` 下所有 Go 代码
> 强制级别: **HIGH**（违反必须修）/ **MED**（PR 前整改）/ **LOW**（长期优化）

---

## 1. 目录结构 [HIGH]

```
flowci/
├── main.go                          # 只做 Wails 启动 + Bind，≤ 80 行
├── internal/
│   ├── handler/                    # Wails Bind 层（薄封装，无业务）
│   │   ├── app.go                  # App struct + Bind 方法集合
│   │   ├── dto.go                  # 所有 Request/Response struct
│   │   └── errors.go               # 错误协议、常量
│   ├── docker/                     # docker exec 封装
│   │   ├── client.go               # 统一入口、CheckDocker
│   │   ├── image.go                # ListImages / RemoveImage / PushImage
│   │   ├── container.go            # ListContainers / Start / Stop / Remove / Logs
│   │   ├── build.go                # BuildImage (buildx)
│   │   └── compose.go              # GenerateCompose / DeployWithCompose
│   ├── pipeline/                   # 流水线
│   │   ├── pipeline.go             # 类型定义
│   │   ├── executor.go             # 串行/并行执行器
│   │   ├── yaml.go                 # YAML 导入导出（唯一定义处）
│   │   └── validator.go            # 步骤与配置校验
│   ├── store/                      # SQLite 持久化
│   │   ├── store.go                # 连接池、PRAGMA、Close
│   │   ├── migrate.go              # 版本化迁移
│   │   ├── projects.go
│   │   ├── pipelines.go
│   │   ├── builds.go
│   │   └── settings.go
│   ├── secret/                     # 凭证加密 (OS keyring)
│   │   └── keyring.go
│   ├── logger/                     # slog 初始化
│   │   └── logger.go
│   └── config/                     # 路径、默认值
│       └── path.go
├── frontend/                        # 见 frontend-spec.md
├── docs/standards/                  # 本规范所在目录
└── migrations/                      # SQL 迁移文件 (见 data-spec.md)
```

**硬规则**:
- `main.go` 仅允许：`//go:embed`、`NewApp()`、`wails.Run()`、`main()`。不允许任何业务方法
- `internal/` 包之间只能**向内依赖**：`handler → docker / pipeline / store / secret / logger`，`pipeline → docker / store`，`store / docker / secret / logger / config` 互不依赖
- 禁止包名用复数、下划线、缩写（例：`stores` ❌、`docker_util` ❌、`db` ❌）

---

## 2. 文件与函数体量 [HIGH]

| 指标 | 上限 | 违反处理 |
|---|---|---|
| 单文件行数 | **500** | 立即拆分 |
| 单函数行数 | **50** | 抽子函数 |
| 函数嵌套深度 | **3** | 提前 return / 抽函数 |
| 单个 switch-case 分支数 | **8** | 改查表或多态 |

陌生工程师读 30 秒能说出意图和边界，则合格。

---

## 3. 命名 [HIGH]

### 包名
- 全小写，单数名词，无下划线：`docker` / `pipeline` / `store` / `handler`

### 导出标识符
- 类型、函数、常量：`PascalCase`
- struct 字段 JSON tag：`camelCase`（见 [ipc-spec.md](./ipc-spec.md) § 4）
- 错误变量：`Err` 前缀，如 `ErrProjectNotFound`

### 未导出标识符
- 局部变量、私有函数：`camelCase`
- 短命周期变量可 1-2 字母（`i`、`p`、`err`），跨 20 行以上的变量必须语义化命名

### 禁用
- ❌ 匈牙利命名（`strName`、`iCount`）
- ❌ 中文拼音缩写（`yhmm` → 应为 `password`）
- ❌ `Util`、`Helper`、`Common` 这类兜底包名

---

## 4. 错误处理 [HIGH]

### 4.1 返回值协议

**所有对外（Bind 到 Wails）与包间方法必须返回 `(T, error)`。**

```go
// ✅ 正确
func ListProjects(ctx context.Context) ([]Project, error) { ... }

// ❌ 禁止：map[string]interface{} 作为对外返回
func ListProjects(...) map[string]interface{} { ... }

// ❌ 禁止：靠 {success: false, error: "..."} 协议
func Build(...) map[string]interface{} {
    return map[string]interface{}{"success": false, "error": "xxx"}
}
```

包内辅助函数返回 `map` 或基本类型均可。

### 4.2 错误包装

```go
// ✅ 用 %w 保留错误链
if err := db.Exec(...); err != nil {
    return fmt.Errorf("create project: %w", err)
}

// ❌ 吞上下文
if err != nil { return err }                // 没有业务上下文
if err != nil { return errors.New("failed") } // 丢失 %w 链
```

### 4.3 哨兵错误

业务错误集中定义在 `internal/handler/errors.go`：

```go
var (
    ErrProjectNotFound  = errors.New("project not found")
    ErrPipelineNotFound = errors.New("pipeline not found")
    ErrDockerNotReady   = errors.New("docker daemon not ready")
)
```

调用方用 `errors.Is` 判断，不靠字符串 match。

### 4.4 禁止

- ❌ `panic` / `log.Fatal`（除 main 启动阶段）
- ❌ 吞错：`_ = err` 或 `if err != nil { /* ignore */ }`
- ❌ 用 `fmt.Printf` 记录错误（见 § 5）

---

## 5. 日志 [HIGH]

### 5.1 统一使用 `log/slog`

`internal/logger/logger.go` 初始化：
- 开发模式：`slog.NewTextHandler(os.Stdout, ...)`
- 生产模式：JSON handler + 文件 rotation，路径 `%APPDATA%/FlowCI/logs/flowci-YYYY-MM-DD.log`
- 日志级别通过 settings 表可调（默认 `INFO`）

### 5.2 调用规范

```go
// ✅
slog.Info("project created", "id", p.ID, "name", p.Name)
slog.Error("list projects failed", "err", err)

// ❌
fmt.Printf("ListProjects error: %v\n", err)     // 全文件禁用
fmt.Println("Application starting...")          // 启动信息也用 slog
log.Printf("xxx")                                // 标准 log 包禁用
```

### 5.3 敏感字段

password / token / authorization header 写日志前必须遮蔽为 `***`。
`internal/secret/mask.go` 提供 `Mask(s string) string`。

---

## 6. 并发与上下文 [HIGH]

### 6.1 context 贯通

- 所有 Bind 方法第一参数 `ctx context.Context`（Wails 会注入）
- 所有可能阻塞的调用（exec / DB / HTTP）必须传 `ctx`

### 6.2 exec.Command 规范

```go
// ✅ 必须带 timeout
ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
defer cancel()
cmd := exec.CommandContext(ctx, "docker", "version")

// ❌ 裸 exec.Command 一律禁用
cmd := exec.Command("docker", ...) // 不允许
```

默认超时（可在 `internal/docker/client.go` 集中定义）：

| 操作 | 超时 |
|---|---|
| 查询类（version / ps / images） | 10s |
| 启停容器 | 30s |
| 拉镜像 | 10min |
| 构建镜像 | 30min |
| push | 15min |
| compose up | 10min |

### 6.3 参数拼接

- **永远用 arg 数组**，禁止字符串拼接
- 用户输入的 image name / container name / tag / port 必须先过 `internal/docker/validator.go` 白名单正则

```go
// ✅
cmd := exec.CommandContext(ctx, "docker", "run", "-d", "--name", name, image)

// ❌
cmd := exec.CommandContext(ctx, "sh", "-c", "docker run -d --name "+name+" "+image)
```

### 6.4 并发锁

- 同一 pipeline 的并发执行必须拒绝：`internal/pipeline/executor.go` 维护 `map[pipelineID]*sync.Mutex`
- 同一 build record 正在写入时拒绝重复提交

---

## 7. 测试 [MED]

### 7.1 覆盖率目标

| 层 | 最低覆盖率 | 备注 |
|---|---|---|
| `internal/store/**` | **70%** | CRUD + 迁移必覆盖 |
| `internal/pipeline/**` | **60%** | YAML 导入导出 + 校验 + 执行 |
| `internal/docker/**` | 不强求 | 依赖外部 docker，可用 mock |
| `internal/handler/**` | **40%** | 集成测试覆盖主路径 |

### 7.2 测试文件位置

- 单元测试：与被测文件同目录，`xxx_test.go`
- 集成测试：`main_integration_test.go` 或 `internal/<pkg>/integration_test.go`，用 `//go:build integration` tag

### 7.3 临时目录

```go
tmpDir, err := os.MkdirTemp("", "flowci-test-*")
require.NoError(t, err)
t.Cleanup(func() { os.RemoveAll(tmpDir) })  // 用 t.Cleanup 代替 defer
```

### 7.4 禁止

- ❌ 测试里 `fmt.Println` 调试（用 `t.Logf`）
- ❌ 自己手写 `contains()` 之类标准库函数（直接 `strings.Contains`）

---

## 8. 依赖管理 [MED]

- 新增 go module 必须在 commit message 中说明引入理由
- 优先选：标准库 > 已引入的依赖 > 新依赖
- 禁止依赖未打 tag 的 master 分支（除非有明确 commit hash pin）
- Go 版本在 `go.mod` 写 `go 1.25`（minor 粒度，不写 patch）

---

## 9. Commit & 文档 [MED]

- Commit message 遵循 Conventional Commits：`<type>(<scope>): <中文描述>`
- 类型：`feat` / `fix` / `docs` / `style` / `refactor` / `test` / `chore`
- scope 对应 `internal/` 子包或 `frontend`、`docs`、`build`
- Go 导出类型/函数必须有中文 Javadoc 风格注释，说明**为什么**而非**是什么**

```go
// CreateBuildRecord 在构建开始前落一条 building 状态的记录，
// 用于前端实时刷新进度；失败时由调用方 FinishBuildRecord("failed", ...) 收尾。
func CreateBuildRecord(...) (BuildRecord, error) { ... }
```

---

## 10. 代码坏味道红线 [HIGH]

发现以下任一，立即整改或提 issue：

1. **`map[string]interface{}` 作为公共 API 返回值** → 改 struct
2. **重复定义的 struct 类型**（如多处 YamlStep） → 抽唯一定义
3. **死代码**（未被任何地方调用的函数/字段） → 删
4. **配置字段无对应行为**（如 `Parallel=true` 但执行器串行） → 要么实现要么删字段
5. **`if err != nil { fmt.Printf(...); return nil }`** → 改 slog + 返回 error
6. **硬编码魔法数字**（超时、上限、重试次数） → 提常量到 `internal/config/`
7. **重复的 ok 模式**（连续 5 次 `if v, ok := m["x"].(string); ok`） → 抽成 dto 解析函数

---

## 11. 审查清单（PR 前自检）

- [ ] 新增/修改文件都 ≤ 500 行，新增/修改函数都 ≤ 50 行
- [ ] 没有引入 `map[string]interface{}` 作为对外返回
- [ ] 所有 `exec.Command*` 都是 `CommandContext` 且有超时
- [ ] 所有错误用 `%w` 包装，都有 slog 记录（error 级别）
- [ ] 没有新的 `fmt.Printf` / `log.Printf`
- [ ] 没有新的重复 struct / 死代码
- [ ] 敏感字段不进 SQLite（见 data-spec.md § 5）
- [ ] 迁移脚本编号递增，启动时幂等
- [ ] 中文注释解释 WHY，不解释 WHAT
- [ ] Go 文件保存为 UTF-8 无 BOM
