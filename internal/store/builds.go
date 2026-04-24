package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound 通用的"记录不存在"哨兵错误；handler 层可用 errors.Is 分类响应。
var ErrNotFound = errors.New("record not found")

// BuildRecord 构建记录实体（对应 build_records 表）。
// TODO(phase-3)：Log 改为 LogPath + LogSize，日志落磁盘（见 data-spec.md §6）。
type BuildRecord struct {
	ID         string     `json:"id"`
	ProjectID  string     `json:"projectId"`
	ImageName  string     `json:"imageName"`
	ImageTag   string     `json:"imageTag"`
	Status     string     `json:"status"` // pending, building, success, failed
	Log        string     `json:"log,omitempty"`
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

// FinishBuildRecord 更新记录的终态（success / failed）和完整日志。
// TODO(phase-3)：log 参数改为写文件，表里只存 log_path。
func FinishBuildRecord(id, status, log string) error {
	_, err := DB.Exec(
		`UPDATE build_records SET status=?, log=?, finished_at=? WHERE id=?`,
		status, log, time.Now().UTC(), id,
	)
	if err != nil {
		return fmt.Errorf("finish build record: %w", err)
	}
	return nil
}

// GetBuildRecord 按 ID 取单条记录（含完整 log）。不存在时返回 ErrNotFound。
func GetBuildRecord(id string) (BuildRecord, error) {
	var r BuildRecord
	var finishedAt sql.NullTime
	err := DB.QueryRow(
		`SELECT id, project_id, image_name, image_tag, status, log, started_at, finished_at FROM build_records WHERE id=?`, id,
	).Scan(&r.ID, &r.ProjectID, &r.ImageName, &r.ImageTag, &r.Status, &r.Log, &r.StartedAt, &finishedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return BuildRecord{}, ErrNotFound
	}
	if err != nil {
		return BuildRecord{}, fmt.Errorf("get build record %s: %w", id, err)
	}
	if finishedAt.Valid {
		r.FinishedAt = &finishedAt.Time
	}
	return r, nil
}

// ListBuildRecords 列出某项目最近的构建记录（不含 log 字段，避免列表过大）。
// 如需完整日志调用 GetBuildRecord。
func ListBuildRecords(projectID string) ([]BuildRecord, error) {
	rows, err := DB.Query(
		`SELECT id, project_id, image_name, image_tag, status, started_at, finished_at
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
		if err := rows.Scan(&r.ID, &r.ProjectID, &r.ImageName, &r.ImageTag, &r.Status, &r.StartedAt, &finishedAt); err != nil {
			return nil, fmt.Errorf("scan build record: %w", err)
		}
		if finishedAt.Valid {
			r.FinishedAt = &finishedAt.Time
		}
		records = append(records, r)
	}
	return records, rows.Err()
}
