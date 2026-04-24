-- 0001_initial_schema.sql
-- FlowCI 初始 Schema：对应原 store.go 的 migrate() 硬编码表 + 补充索引。
-- 所有表使用 IF NOT EXISTS 保证对老库（无 schema_migrations）的向后兼容。

-- projects：用户创建的本地项目（name + 路径 + 语言）
CREATE TABLE IF NOT EXISTS projects (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    path       TEXT NOT NULL,
    language   TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- build_records：构建记录（log 字段阶段 3 会改 log_path，当前保留）
CREATE TABLE IF NOT EXISTS build_records (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL,
    image_name  TEXT NOT NULL,
    image_tag   TEXT NOT NULL DEFAULT 'latest',
    status      TEXT NOT NULL DEFAULT 'pending',
    log         TEXT NOT NULL DEFAULT '',
    started_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at DATETIME,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- settings：key-value 全局设置（主题、日志级别等）
CREATE TABLE IF NOT EXISTS settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT ''
);

-- pipelines：流水线定义，steps/config 为 JSON 字符串
CREATE TABLE IF NOT EXISTS pipelines (
    id         TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    name       TEXT NOT NULL,
    steps      TEXT NOT NULL DEFAULT '[]',
    config     TEXT NOT NULL DEFAULT '{}',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- 索引：外键字段 + 列表查询常用排序字段
CREATE INDEX IF NOT EXISTS idx_build_records_project_started
    ON build_records(project_id, started_at DESC);

CREATE INDEX IF NOT EXISTS idx_pipelines_project_created
    ON pipelines(project_id, created_at DESC);
