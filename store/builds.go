package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type BuildResult struct {
	ID         string     `json:"id"`
	ProjectID  string     `json:"project_id"`
	ImageName  string     `json:"image_name"`
	ImageTag   string     `json:"image_tag"`
	Status     string     `json:"status"`
	Log        string     `json:"log"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

func CreateBuildRecord(projectID, imageName, imageTag string) (BuildResult, error) {
	now := time.Now()
	r := BuildResult{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		ImageName: imageName,
		ImageTag:  imageTag,
		Status:    "building",
		StartedAt: now,
	}

	_, err := DB.Exec(
		`INSERT INTO build_records (id, project_id, image_name, image_tag, status, started_at) VALUES (?, ?, ?, ?, ?, ?)`,
		r.ID, projectID, r.ImageName, r.ImageTag, r.Status, r.StartedAt,
	)
	if err != nil {
		return BuildResult{}, fmt.Errorf("create build record: %w", err)
	}
	return r, nil
}

func FinishBuildRecord(id, status, log string) error {
	now := time.Now()
	_, err := DB.Exec(
		`UPDATE build_records SET status=?, log=?, finished_at=? WHERE id=?`,
		status, log, now, id,
	)
	if err != nil {
		return fmt.Errorf("finish build record: %w", err)
	}
	return nil
}

func GetBuildRecord(id string) (BuildResult, error) {
	var r BuildResult
	var finishedAt sql.NullTime
	err := DB.QueryRow(
		`SELECT id, project_id, image_name, image_tag, status, log, started_at, finished_at FROM build_records WHERE id=?`, id,
	).Scan(&r.ID, &r.ProjectID, &r.ImageName, &r.ImageTag, &r.Status, &r.Log, &r.StartedAt, &finishedAt)
	if err != nil {
		return BuildResult{}, fmt.Errorf("get build record %s: %w", id, err)
	}
	if finishedAt.Valid {
		r.FinishedAt = &finishedAt.Time
	}
	return r, nil
}

func ListBuildRecords(projectID string) ([]BuildResult, error) {
	rows, err := DB.Query(
		`SELECT id, project_id, image_name, image_tag, status, log, started_at, finished_at FROM build_records WHERE project_id=? ORDER BY started_at DESC LIMIT 50`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list build records: %w", err)
	}
	defer rows.Close()

	records := make([]BuildResult, 0)
	for rows.Next() {
		var r BuildResult
		var finishedAt sql.NullTime
		if err := rows.Scan(&r.ID, &r.ProjectID, &r.ImageName, &r.ImageTag, &r.Status, &r.Log, &r.StartedAt, &finishedAt); err != nil {
			return nil, fmt.Errorf("scan build record: %w", err)
		}
		if finishedAt.Valid {
			r.FinishedAt = &finishedAt.Time
		}
		records = append(records, r)
	}
	return records, nil
}
