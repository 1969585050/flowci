package store

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound 通用的"记录不存在"哨兵错误；handler 层可用 errors.Is 分类响应。
var ErrNotFound = errors.New("record not found")

// BuildRecord 构建记录实体（对应 build_records 表）。
//
// 日志存储策略（阶段 3）：
//   - 新建记录：log 字段保持空；构建完成后 FinishBuildRecord 将 docker 输出写到
//     <logsDir>/<id>.log，并把路径 + 字节数写入 log_path / log_size
//   - 读取：GetBuildRecord 优先读 log_path 文件填入 Log 字段；log_path 为空时
//     回退到 log 字段（兼容阶段 3 之前落库的历史数据）
type BuildRecord struct {
	ID         string     `json:"id"`
	ProjectID  string     `json:"projectId"`
	ImageName  string     `json:"imageName"`
	ImageTag   string     `json:"imageTag"`
	Status     string     `json:"status"` // pending, building, success, failed
	Log        string     `json:"log,omitempty"`
	LogSize    int64      `json:"logSize"`
	StartedAt  time.Time  `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
}

// CreateBuildRecord 在构建开始前落一条 building 状态记录，后续由 FinishBuildRecord 更新终态。
func CreateBuildRecord(projectID, imageName, imageTag string) (BuildRecord, error) {
	now := time.Now().UTC()
	r := BuildRecord{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		ImageName: imageName,
		ImageTag:  imageTag,
		Status:    "building",
		StartedAt: now,
	}

	_, err := DB.Exec(
		`INSERT INTO build_records (id, project_id, image_name, image_tag, status, started_at) VALUES (?, ?, ?, ?, ?, ?)`,
		r.ID, r.ProjectID, r.ImageName, r.ImageTag, r.Status, r.StartedAt,
	)
	if err != nil {
		return BuildRecord{}, fmt.Errorf("create build record: %w", err)
	}
	return r, nil
}

// FinishBuildRecord 更新记录终态（success / failed），并把 docker 输出写到磁盘。
//
// logsDir 为构建日志的根目录（调用方通常传 config.BuildLogsDir()）。
// 如果写文件失败，只打 warning 并把状态写回 DB，不影响构建结果。
func FinishBuildRecord(id, status, logContent, logsDir string) error {
	logPath, logSize := persistBuildLog(id, logContent, logsDir)

	_, err := DB.Exec(
		`UPDATE build_records SET status=?, log_path=?, log_size=?, finished_at=? WHERE id=?`,
		status, logPath, logSize, time.Now().UTC(), id,
	)
	if err != nil {
		return fmt.Errorf("finish build record: %w", err)
	}
	return nil
}

// persistBuildLog 尽最大努力把 logContent 写到 logsDir/<id>.log；
// 失败时返回 ("", 0) 并打 warning，DB 端 log_path 保持空串。
func persistBuildLog(id, logContent, logsDir string) (string, int64) {
	if logsDir == "" || logContent == "" {
		return "", 0
	}
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		slog.Warn("create build logs dir failed", "dir", logsDir, "err", err)
		return "", 0
	}
	path := filepath.Join(logsDir, id+".log")
	if err := os.WriteFile(path, []byte(logContent), 0644); err != nil {
		slog.Warn("write build log failed", "id", id, "path", path, "err", err)
		return "", 0
	}
	return path, int64(len(logContent))
}

// GetBuildRecord 按 ID 取单条记录（含完整 log）。不存在时返回 ErrNotFound。
// log 优先从 log_path 文件读；log_path 为空时回退老 log 字段（历史数据）。
func GetBuildRecord(id string) (BuildRecord, error) {
	var r BuildRecord
	var finishedAt sql.NullTime
	var logLegacy string
	var logPath string
	err := DB.QueryRow(
		`SELECT id, project_id, image_name, image_tag, status, log, log_path, log_size, started_at, finished_at
		 FROM build_records WHERE id=?`, id,
	).Scan(&r.ID, &r.ProjectID, &r.ImageName, &r.ImageTag, &r.Status, &logLegacy, &logPath, &r.LogSize, &r.StartedAt, &finishedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return BuildRecord{}, ErrNotFound
	}
	if err != nil {
		return BuildRecord{}, fmt.Errorf("get build record %s: %w", id, err)
	}
	if finishedAt.Valid {
		r.FinishedAt = &finishedAt.Time
	}

	r.Log = loadBuildLog(logPath, logLegacy)
	if r.LogSize == 0 && r.Log != "" {
		r.LogSize = int64(len(r.Log))
	}
	return r, nil
}

// loadBuildLog 按策略加载 log 内容：文件优先 → 历史 log 字段回退。
func loadBuildLog(logPath, legacy string) string {
	if logPath == "" {
		return legacy
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		slog.Warn("read build log file failed", "path", logPath, "err", err)
		return legacy
	}
	return string(data)
}

// LatestBuildRecord 取某项目最近一条构建记录（不含 log），用于项目卡片摘要。
// 不存在返回 (BuildRecord{}, ErrNotFound)。
func LatestBuildRecord(projectID string) (BuildRecord, error) {
	var r BuildRecord
	var finishedAt sql.NullTime
	err := DB.QueryRow(
		`SELECT id, project_id, image_name, image_tag, status, log_size, started_at, finished_at
		 FROM build_records WHERE project_id=?
		 ORDER BY started_at DESC LIMIT 1`, projectID,
	).Scan(&r.ID, &r.ProjectID, &r.ImageName, &r.ImageTag, &r.Status, &r.LogSize, &r.StartedAt, &finishedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return BuildRecord{}, ErrNotFound
	}
	if err != nil {
		return BuildRecord{}, fmt.Errorf("latest build record %s: %w", projectID, err)
	}
	if finishedAt.Valid {
		r.FinishedAt = &finishedAt.Time
	}
	return r, nil
}

// CountBuildRecords 该项目历史构建数（含进行中）。
func CountBuildRecords(projectID string) (int, error) {
	var n int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM build_records WHERE project_id=?`, projectID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count build records %s: %w", projectID, err)
	}
	return n, nil
}

// ListBuildRecords 列出某项目最近的构建记录；不含 log 内容，仅带 LogSize 方便前端提示。
func ListBuildRecords(projectID string) ([]BuildRecord, error) {
	rows, err := DB.Query(
		`SELECT id, project_id, image_name, image_tag, status, log_size, started_at, finished_at
		 FROM build_records WHERE project_id=?
		 ORDER BY started_at DESC LIMIT 50`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list build records: %w", err)
	}
	defer rows.Close()

	records := make([]BuildRecord, 0)
	for rows.Next() {
		var r BuildRecord
		var finishedAt sql.NullTime
		if err := rows.Scan(&r.ID, &r.ProjectID, &r.ImageName, &r.ImageTag, &r.Status, &r.LogSize, &r.StartedAt, &finishedAt); err != nil {
			return nil, fmt.Errorf("scan build record: %w", err)
		}
		if finishedAt.Valid {
			r.FinishedAt = &finishedAt.Time
		}
		records = append(records, r)
	}
	return records, rows.Err()
}
