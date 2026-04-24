package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCreateBuildRecord(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()
	pid := createTestProject(t)

	r, err := CreateBuildRecord(pid, "myapp", "latest")
	if err != nil {
		t.Fatalf("CreateBuildRecord: %v", err)
	}
	if r.ID == "" {
		t.Fatal("ID empty")
	}
	if r.Status != "building" {
		t.Errorf("initial status = %q, want building", r.Status)
	}
	if r.StartedAt.IsZero() {
		t.Error("StartedAt zero")
	}
	if r.FinishedAt != nil {
		t.Errorf("FinishedAt should be nil initially, got %v", r.FinishedAt)
	}
}

func TestFinishBuildRecord_WritesLogToDisk(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()
	pid := createTestProject(t)

	r, _ := CreateBuildRecord(pid, "myapp", "v1")

	logsDir, err := os.MkdirTemp("", "flowci-buildlogs-*")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	defer os.RemoveAll(logsDir)

	logContent := "Step 1/3: FROM alpine\nSuccessfully built abc123\n"
	if err := FinishBuildRecord(r.ID, "success", logContent, logsDir); err != nil {
		t.Fatalf("FinishBuildRecord: %v", err)
	}

	// 文件应落到 logsDir/<id>.log
	path := filepath.Join(logsDir, r.ID+".log")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if string(data) != logContent {
		t.Errorf("log file content mismatch:\n got %q\nwant %q", string(data), logContent)
	}

	// GetBuildRecord 应该从文件读回
	got, err := GetBuildRecord(r.ID)
	if err != nil {
		t.Fatalf("GetBuildRecord: %v", err)
	}
	if got.Status != "success" {
		t.Errorf("status = %q, want success", got.Status)
	}
	if got.Log != logContent {
		t.Errorf("Log from GetBuildRecord mismatch")
	}
	if got.LogSize != int64(len(logContent)) {
		t.Errorf("LogSize = %d, want %d", got.LogSize, len(logContent))
	}
	if got.FinishedAt == nil {
		t.Error("FinishedAt should be set after Finish")
	}
}

// TestFinishBuildRecord_EmptyLog 空 log 不应触发文件写入。
func TestFinishBuildRecord_EmptyLog(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()
	pid := createTestProject(t)

	r, _ := CreateBuildRecord(pid, "app", "v1")

	logsDir, _ := os.MkdirTemp("", "flowci-empty-*")
	defer os.RemoveAll(logsDir)

	if err := FinishBuildRecord(r.ID, "failed", "", logsDir); err != nil {
		t.Fatalf("FinishBuildRecord: %v", err)
	}

	// 不该有文件生成
	entries, _ := os.ReadDir(logsDir)
	if len(entries) != 0 {
		t.Errorf("empty log should not write file, got %d entries", len(entries))
	}

	got, _ := GetBuildRecord(r.ID)
	if got.Status != "failed" {
		t.Errorf("status = %q, want failed", got.Status)
	}
	if got.Log != "" {
		t.Errorf("Log should be empty, got %q", got.Log)
	}
}

// TestGetBuildRecord_FallbackToLegacyLogField 验证 log_path 为空 + log 字段有值时
// 走老字段回退（兼容阶段 3 之前落库的数据）。
func TestGetBuildRecord_FallbackToLegacyLogField(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()
	pid := createTestProject(t)

	r, _ := CreateBuildRecord(pid, "legacy", "v1")
	// 直接写入老字段（模拟阶段 3 前落库的数据）
	_, err := DB.Exec(`UPDATE build_records SET status='success', log=? WHERE id=?`,
		"legacy log content", r.ID)
	if err != nil {
		t.Fatalf("manual UPDATE: %v", err)
	}

	got, err := GetBuildRecord(r.ID)
	if err != nil {
		t.Fatalf("GetBuildRecord: %v", err)
	}
	if got.Log != "legacy log content" {
		t.Errorf("legacy fallback failed: got %q", got.Log)
	}
}

func TestGetBuildRecord_NotFound(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	_, err := GetBuildRecord("nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// TestListBuildRecords_ExcludesLogField 验证列表查询不带 log 内容（避免 N 条 × MB 的查询）。
func TestListBuildRecords_ExcludesLogField(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()
	pid := createTestProject(t)

	logsDir, _ := os.MkdirTemp("", "flowci-list-*")
	defer os.RemoveAll(logsDir)

	// 建 3 条记录，都落了磁盘 log
	bigContent := strings.Repeat("x", 10_000) // 10KB 模拟大 log
	for i := 0; i < 3; i++ {
		r, _ := CreateBuildRecord(pid, "app", "v1")
		_ = FinishBuildRecord(r.ID, "success", bigContent, logsDir)
	}

	list, err := ListBuildRecords(pid)
	if err != nil {
		t.Fatalf("ListBuildRecords: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 records, got %d", len(list))
	}
	for i, r := range list {
		if r.Log != "" {
			t.Errorf("list[%d].Log should be empty in list query, got %d bytes", i, len(r.Log))
		}
		if r.LogSize != int64(len(bigContent)) {
			t.Errorf("list[%d].LogSize = %d, want %d", i, r.LogSize, len(bigContent))
		}
	}
}

func TestListBuildRecords_OrderByStartedDesc(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()
	pid := createTestProject(t)

	// 顺序建 3 条；sleep 保证 started_at 严格递增（SQLite DATETIME 分辨率最小到微秒）
	r1, _ := CreateBuildRecord(pid, "a", "v1")
	time.Sleep(5 * time.Millisecond)
	r2, _ := CreateBuildRecord(pid, "b", "v2")
	time.Sleep(5 * time.Millisecond)
	r3, _ := CreateBuildRecord(pid, "c", "v3")

	list, err := ListBuildRecords(pid)
	if err != nil {
		t.Fatalf("ListBuildRecords: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 records, got %d", len(list))
	}
	// started_at DESC：最后建的最先
	if list[0].ID != r3.ID || list[1].ID != r2.ID || list[2].ID != r1.ID {
		t.Errorf("order wrong: %q, %q, %q", list[0].ID, list[1].ID, list[2].ID)
	}
}

func TestListBuildRecords_ScopedByProject(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()

	p1, _ := CreateProject(CreateProjectInput{Name: "p1", Path: "/1", Language: "go"})
	p2, _ := CreateProject(CreateProjectInput{Name: "p2", Path: "/2", Language: "go"})

	_, _ = CreateBuildRecord(p1.ID, "a", "v1")
	_, _ = CreateBuildRecord(p1.ID, "b", "v1")
	_, _ = CreateBuildRecord(p2.ID, "c", "v1")

	l1, _ := ListBuildRecords(p1.ID)
	l2, _ := ListBuildRecords(p2.ID)
	if len(l1) != 2 {
		t.Errorf("p1 expected 2, got %d", len(l1))
	}
	if len(l2) != 1 {
		t.Errorf("p2 expected 1, got %d", len(l2))
	}
}

// TestPersistBuildLog_MissingDirIsRecovered 验证 logsDir 不存在时 FinishBuildRecord
// 会自动 MkdirAll，而不是 panic 或返回错误。
func TestPersistBuildLog_MissingDirIsRecovered(t *testing.T) {
	cleanup := setupTestStore(t)
	defer cleanup()
	pid := createTestProject(t)

	r, _ := CreateBuildRecord(pid, "app", "v1")

	root, _ := os.MkdirTemp("", "flowci-nested-*")
	defer os.RemoveAll(root)
	nested := filepath.Join(root, "deeply", "nested", "not-exist")

	if err := FinishBuildRecord(r.ID, "success", "some log", nested); err != nil {
		t.Fatalf("FinishBuildRecord: %v", err)
	}

	if _, err := os.Stat(filepath.Join(nested, r.ID+".log")); err != nil {
		t.Errorf("expected log file created in nested dir: %v", err)
	}
}
