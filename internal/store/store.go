package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// DB 是全局数据库句柄。
// TODO(phase-3)：将 Init 返回 *Store 替代全局，便于测试注入和多实例管理。
var DB *sql.DB

// Project 项目实体（对应 projects 表）。
//
// Path 含义因模式而异：
//   - RepoURL 为空：用户指定的本地路径
//   - RepoURL 非空：FlowCI 自动 clone 到的本地目录（通常 <dataDir>/repos/<id>）
type Project struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Path       string     `json:"path"`
	Language   string     `json:"language"`
	RepoURL    string     `json:"repoUrl"`
	RepoBranch string     `json:"repoBranch"`
	LastPullAt *time.Time `json:"lastPullAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

// BuildRecord 在 builds.go 中定义（保留 CRUD 与类型定义同文件）。

// Init 初始化数据库连接、设置 PRAGMA、跑迁移脚本。
//
// 数据库文件位于 dataDir/flowci.db；WAL / SHM 文件由 SQLite 自动管理。
// 启用 PRAGMA：journal_mode=WAL / synchronous=NORMAL / busy_timeout=5000
// / foreign_keys=ON / temp_store=MEMORY。
func Init(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "flowci.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	if err := applyPragmas(db); err != nil {
		db.Close()
		return err
	}

	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("ping db: %w", err)
	}

	if err := runMigrations(db); err != nil {
		db.Close()
		return fmt.Errorf("run migrations: %w", err)
	}

	DB = db
	return nil
}

// Close 关闭全局数据库连接；幂等，可重复调用。
func Close() {
	if DB != nil {
		DB.Close()
		DB = nil
	}
}

// applyPragmas 开启 WAL 与性能/一致性参数。
// 必须在迁移和业务 SQL 之前执行。
func applyPragmas(db *sql.DB) error {
	pragmas := []string{
		`PRAGMA journal_mode = WAL`,
		`PRAGMA synchronous = NORMAL`,
		`PRAGMA busy_timeout = 5000`,
		`PRAGMA foreign_keys = ON`,
		`PRAGMA temp_store = MEMORY`,
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("pragma %q: %w", p, err)
		}
	}
	return nil
}
