package store

import (
	"os"
	"testing"
)

// setupTestStore 在临时目录初始化一套干净的 store；返回 cleanup 函数。
// 每个测试都调一次以获得隔离的 DB。
func setupTestStore(t *testing.T) func() {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "flowci-store-test-*")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	if err := Init(tmpDir); err != nil {
		t.Fatalf("store.Init: %v", err)
	}
	return func() {
		Close()
		_ = os.RemoveAll(tmpDir)
	}
}

// TestInit_SetsPragmas 验证 Init 后 WAL / foreign_keys 等 PRAGMA 确实生效。
func TestInit_SetsPragmas(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	var mode string
	if err := DB.QueryRow(`PRAGMA journal_mode`).Scan(&mode); err != nil {
		t.Fatalf("query journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Errorf("journal_mode = %q, want wal", mode)
	}

	var fk int
	if err := DB.QueryRow(`PRAGMA foreign_keys`).Scan(&fk); err != nil {
		t.Fatalf("query foreign_keys: %v", err)
	}
	if fk != 1 {
		t.Errorf("foreign_keys = %d, want 1", fk)
	}
}

// TestInit_TwiceIsIdempotent 验证在同一目录连续 Init 两次不会出错（迁移只应用一次）。
func TestInit_TwiceIsIdempotent(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "flowci-store-idem-*")
	defer os.RemoveAll(tmpDir)

	if err := Init(tmpDir); err != nil {
		t.Fatalf("first init: %v", err)
	}
	Close()

	if err := Init(tmpDir); err != nil {
		t.Fatalf("second init: %v", err)
	}
	defer Close()

	var count int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	if count < 1 {
		t.Errorf("expected at least 1 applied migration, got %d", count)
	}
}
