# FlowCI 数据持久化规范 (SQLite)

> 版本: v1.0 (2026-04-25)
> 适用范围: `internal/store/**`、`migrations/**`、以及任何直接操作 SQLite 的代码
> 强制级别: **HIGH** / **MED** / **LOW**

---

## 1. 存储分层 [HIGH]

| 类型 | 存储位置 | 说明 |
|---|---|---|
| 结构化业务数据 | SQLite `flowci.db` | 项目 / 流水线 / 构建记录 / 设置 |
| 大文本（build log） | 文件系统 `logs/<build-id>.log` | 数据库只存路径 |
| 敏感凭证 | **OS keyring**（DPAPI / Keychain） | 数据库不存原文 |
| 临时缓存 | 内存（不持久化） | Docker 进程状态、connection check |
| 前端偏好 | `settings` 表（经 IPC） | 禁用 localStorage 持久化业务数据 |

### 1.1 数据目录

- Windows: `%APPDATA%\FlowCI\`
- macOS/Linux: `~/.local/share/FlowCI/`（`XDG_DATA_HOME` 优先）

结构：
```
FlowCI/
├── flowci.db                # 主数据库
├── flowci.db-wal            # WAL 文件（自动）
├── flowci.db-shm            # 共享内存（自动）
├── logs/
│   ├── flowci-2026-04-25.log   # 应用日志
│   └── builds/
│       └── <build-id>.log       # 构建日志
└── tmp/                      # 临时 compose 文件等
```

---

## 2. 连接配置 [HIGH]

### 2.1 启动 PRAGMA

`internal/store/store.go` Init 流程：

```go
db, err := sql.Open("sqlite", dbPath)
// 必须全部执行成功才继续
pragmas := []string{
    `PRAGMA journal_mode = WAL`,
    `PRAGMA synchronous = NORMAL`,
    `PRAGMA busy_timeout = 5000`,
    `PRAGMA foreign_keys = ON`,
    `PRAGMA temp_store = MEMORY`,
}
for _, p := range pragmas {
    if _, err := db.Exec(p); err != nil {
        return fmt.Errorf("pragma %s: %w", p, err)
    }
}
```

### 2.2 连接池

```go
db.SetMaxOpenConns(4)       // 禁止 1（WAL 下并发读不阻塞）
db.SetMaxIdleConns(2)
db.SetConnMaxLifetime(30 * time.Minute)
```

### 2.3 禁止

- ❌ `SetMaxOpenConns(1)`（退化到 journal 模式阻塞）
- ❌ 不开 WAL（写期间读也被锁）
- ❌ 不开 foreign_keys（SQLite 默认关闭）

---

## 3. 迁移管理 [HIGH]

### 3.1 迁移表

每次启动自动建：

```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version     INTEGER PRIMARY KEY,
    description TEXT NOT NULL,
    applied_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 3.2 迁移文件

位置：`migrations/NNNN_description.sql`（项目根）

命名规则：4 位数字版本号 + 下划线 + 英文描述

```
migrations/
├── 0001_initial_schema.sql
├── 0002_add_pipelines_table.sql
├── 0003_add_schema_migrations.sql
├── 0004_add_build_log_path.sql
└── 0005_add_index_on_project_id.sql
```

文件内容：纯 SQL，单文件单事务（程序启动时用 `BEGIN; ... COMMIT;` 包裹）。

### 3.3 执行逻辑

启动时：
1. 确保 `schema_migrations` 表存在
2. 扫描 `migrations/` 目录（embed）获取所有版本号
3. 查询当前数据库最大 `version`
4. 按序执行大于当前版本的迁移文件
5. 每个迁移单独事务，失败即中止启动

### 3.4 规则

- ✅ 迁移文件**永远 append-only**，不删不改（即使 bug）
- ✅ 修复上一个迁移的 bug 就写新迁移
- ✅ 版本号**单调递增**，无跳号
- ❌ 不支持 rollback（向前修复哲学，回滚会制造更多问题）
- ❌ 不在代码里硬编码 `CREATE TABLE IF NOT EXISTS`（除 migrations 表自己）

---

## 4. Schema 规范 [HIGH]

### 4.1 主键

- 一律 `id TEXT PRIMARY KEY`，值为 UUID v4 字符串
- 应用层生成（`github.com/google/uuid`），不用 SQLite 自增
- 不用复合主键（查询复杂、JOIN 难），必要时加 `UNIQUE(...)` 约束

### 4.2 时间字段

- 类型：`DATETIME NOT NULL`
- 默认：`DEFAULT CURRENT_TIMESTAMP`
- 时区：**UTC**（Go 端 `time.Now().UTC()`）
- 展示时由前端转本地时区

标准时间列：
- `created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP`
- `updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP`
- 可选：`deleted_at DATETIME`（软删，见 § 9）

### 4.3 字段类型映射

| 场景 | SQLite 类型 | Go 类型 |
|---|---|---|
| ID / 名称 / 短文本 | `TEXT NOT NULL` | `string` |
| 枚举 | `TEXT NOT NULL` | `string`（应用层校验） |
| 数字（步骤数、重试次数） | `INTEGER NOT NULL DEFAULT 0` | `int` |
| 布尔 | `INTEGER NOT NULL DEFAULT 0`（0/1） | `bool` |
| 时间 | `DATETIME NOT NULL` | `time.Time` |
| 可空时间 | `DATETIME` | `*time.Time` 或 `sql.NullTime` |
| JSON 数据（pipeline steps / config） | `TEXT NOT NULL DEFAULT '[]'` | `json.RawMessage` 或具体类型 marshal |
| 大文本（不推荐） | `TEXT` 且限制 1MB | — |

### 4.4 必填 & 默认

- 所有字段必须显式 `NOT NULL` 或有 `DEFAULT`
- 字符串字段无业务默认时用 `DEFAULT ''`
- 避免字段可空，能用默认就用默认

### 4.5 外键

```sql
FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
```

- 所有外键显式声明 `ON DELETE` 策略
- `CASCADE` 用于附属数据（build_records 跟随 project 删除）
- `RESTRICT` 用于不应跟删的引用

---

## 5. 敏感数据 [HIGH]

### 5.1 绝对禁止入库的字段

- 明文密码、API token、OAuth secret
- registry 凭证（username 可入库，password 走 keyring）
- 任何私钥 / 证书原文

### 5.2 keyring 使用

`internal/secret/keyring.go` 封装 `zalando/go-keyring`：

```go
// Set 存储到 OS keychain，key 形如 "flowci:registry:docker.io"
secret.Set(ctx, "registry:docker.io", password)

// Get 读取
password, err := secret.Get(ctx, "registry:docker.io")

// Delete 删除
secret.Delete(ctx, "registry:docker.io")
```

### 5.3 存储引用关系

数据库表里只存 **用户名 + 是否已配置密码** 两个字段：

```sql
CREATE TABLE registry_credentials (
    id            TEXT PRIMARY KEY,
    registry_url  TEXT NOT NULL UNIQUE,
    username      TEXT NOT NULL,
    has_password  INTEGER NOT NULL DEFAULT 0,  -- 0/1
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

实际密码通过 `secret.Get("registry:" + registry_url)` 取。

---

## 6. 大字段 & 日志 [HIGH]

### 6.1 规则

**任何 TEXT 字段预期内容 > 1MB 的必须落文件。**

### 6.2 build log 设计

```sql
CREATE TABLE build_records (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL,
    image_name  TEXT NOT NULL,
    image_tag   TEXT NOT NULL DEFAULT 'latest',
    status      TEXT NOT NULL DEFAULT 'pending',
    log_path    TEXT NOT NULL DEFAULT '',     -- 相对 data dir 的路径
    log_size    INTEGER NOT NULL DEFAULT 0,   -- 字节数，便于展示
    started_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at DATETIME,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);
```

- 构建时 docker output 写 `logs/builds/<build-id>.log`
- DB 只存 `log_path` / `log_size`
- `GetBuildRecord` 按需读文件；列表查询**绝不读 log 文件**

### 6.3 日志清理

- `build_records` 默认保留最近 200 条，老的由 `CleanupOldBuilds()` 定期清
- 清 DB 记录时同步删对应 log 文件
- 用户手动触发"清理" action 可立刻清

---

## 7. 查询规范 [HIGH]

### 7.1 列表查询

- **必须 LIMIT**：默认 50，最大 200
- **必须 ORDER BY**：明确排序字段（通常 `created_at DESC` 或 `updated_at DESC`）
- **不 SELECT 大字段**：log / steps_json / config_json 在列表时按需

```go
// ✅
rows, err := db.Query(`
    SELECT id, project_id, image_name, image_tag, status, started_at, finished_at
    FROM build_records
    WHERE project_id = ?
    ORDER BY started_at DESC
    LIMIT ?
`, projectID, limit)

// ❌
rows, err := db.Query(`SELECT * FROM build_records`)  // 无 LIMIT + 大字段
```

### 7.2 索引

| 场景 | 索引 |
|---|---|
| 外键字段（`project_id`） | 建索引 |
| 常用过滤字段（`status`） | 视查询频率 |
| ORDER BY 字段（`created_at`、`started_at`） | 高频时建复合索引 |

```sql
CREATE INDEX idx_build_records_project_started ON build_records(project_id, started_at DESC);
```

### 7.3 单体查询

```go
err := db.QueryRow(`SELECT ... FROM projects WHERE id = ?`, id).Scan(...)
if err == sql.ErrNoRows {
    return nil, store.ErrNotFound
}
if err != nil {
    return nil, fmt.Errorf("get project: %w", err)
}
```

---

## 8. 事务 [HIGH]

### 8.1 使用场景

**多语句写操作必须包在事务里**，典型：

- 创建 pipeline 并写入初始 step 记录
- 删除 project 并级联删 pipelines / builds（除非靠 FK CASCADE）
- 批量 SaveSettings

### 8.2 标准模式

```go
func (s *Store) CreatePipelineWithSteps(ctx context.Context, ...) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback()  // Commit 成功后 Rollback 是 no-op

    // ... 多条 Exec

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit: %w", err)
    }
    return nil
}
```

### 8.3 禁止

- ❌ 在事务内调 docker exec / HTTP / 长耗时操作（锁持有时间过长）
- ❌ 嵌套事务（SQLite 不真正支持）

---

## 9. 软删除 [LOW]

个人小团队工具**默认物理删除**，不引入 `deleted_at` 架构复杂度。

例外：用户数据误删高发的表（如 projects）可考虑软删 + 30 天清理，需专门评估后加字段。

---

## 10. 备份 & 导出 [MED]

### 10.1 自动备份

启动时若 `flowci.db` 大小超 100MB 给用户提示建议备份。

### 10.2 手动导出

提供 `ExportDatabase(path)` handler：
- 用 `VACUUM INTO` 导出到用户指定路径
- 导出期间前端禁止操作

```go
_, err := db.ExecContext(ctx, "VACUUM INTO ?", exportPath)
```

---

## 11. 禁止清单 [HIGH]

- ❌ 代码里出现新的 `CREATE TABLE IF NOT EXISTS`（迁移统一管）
- ❌ 在 Go 里拼 SQL 字符串（`"WHERE id = " + id`）— 一律 placeholder
- ❌ 忘 `rows.Close()` / `rows.Err()`
- ❌ `Scan` 后不处理 `sql.ErrNoRows`
- ❌ 外键字段不建索引
- ❌ `SELECT *` 上生产代码
- ❌ 把 struct 直接 `json.Marshal` 存一列，然后其他地方按字段查询（要查就拆列）

---

## 12. 审查清单

- [ ] 新增表的字段都 NOT NULL 或有 DEFAULT
- [ ] 外键都有 `ON DELETE` 策略
- [ ] 新增迁移版本号递增、文件名英文
- [ ] 列表查询有 LIMIT + ORDER BY，不选大字段
- [ ] 写操作的多语句场景包了事务
- [ ] 新增敏感字段走 keyring，DB 只存引用
- [ ] 无新增 `CREATE TABLE IF NOT EXISTS`（除 migrations 表）
- [ ] UTC 时间，time.Time 入库前 `.UTC()`
