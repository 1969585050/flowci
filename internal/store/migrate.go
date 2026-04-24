package store

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// migration 表示 migrations/ 目录下的一条 SQL 迁移脚本。
// 文件命名约定：NNNN_description.sql（NNNN 为 4 位数字版本号，单调递增）。
type migration struct {
	version     int
	description string
	content     []byte
}

// runMigrations 按版本号升序执行未应用的迁移脚本。
// 每个迁移单独事务，失败立即中止并返回错误（向前修复，不支持 rollback）。
func runMigrations(db *sql.DB) error {
	if err := ensureMigrationsTable(db); err != nil {
		return err
	}

	applied, err := loadAppliedVersions(db)
	if err != nil {
		return err
	}

	migs, err := loadMigrationFiles()
	if err != nil {
		return err
	}

	for _, m := range migs {
		if applied[m.version] {
			continue
		}
		if err := applyMigration(db, m); err != nil {
			return err
		}
	}
	return nil
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version     INTEGER PRIMARY KEY,
		description TEXT NOT NULL,
		applied_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	return nil
}

func loadAppliedVersions(db *sql.DB) (map[int]bool, error) {
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("query schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := map[int]bool{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("scan schema_migrations: %w", err)
		}
		applied[v] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate schema_migrations: %w", err)
	}
	return applied, nil
}

func loadMigrationFiles() ([]migration, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	migs := make([]migration, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		trimmed := strings.TrimSuffix(name, ".sql")
		parts := strings.SplitN(trimmed, "_", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid migration filename %q: expected NNNN_description.sql", name)
		}
		v, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid version in %q: %w", name, err)
		}
		content, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", name, err)
		}
		migs = append(migs, migration{version: v, description: parts[1], content: content})
	}

	sort.Slice(migs, func(i, j int) bool { return migs[i].version < migs[j].version })
	return migs, nil
}

func applyMigration(db *sql.DB, m migration) error {
	slog.Info("applying migration", "version", m.version, "description", m.description)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx for migration %d: %w", m.version, err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(string(m.content)); err != nil {
		return fmt.Errorf("apply migration %d (%s): %w", m.version, m.description, err)
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations (version, description) VALUES (?, ?)`,
		m.version, m.description); err != nil {
		return fmt.Errorf("record migration %d: %w", m.version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %d: %w", m.version, err)
	}
	return nil
}
