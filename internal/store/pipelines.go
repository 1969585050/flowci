package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PipelineStep 流水线单步定义。
// Type: build / push / deploy（见 internal/pipeline/validator.go）
// OnFail: stop / continue（空串等同 stop）
type PipelineStep struct {
	Type   string                 `json:"type"`
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
	Retry  int                    `json:"retry"`
	OnFail string                 `json:"onFail"`
}

// PipelineConfig 流水线全局配置。
// TODO(phase-2 末): Parallel 字段目前声明未实现；executor 串行执行。
type PipelineConfig struct {
	Parallel   bool `json:"parallel"`
	StopOnFail bool `json:"stopOnFail"`
}

// Pipeline 流水线实体。
type Pipeline struct {
	ID        string         `json:"id"`
	ProjectID string         `json:"projectId"`
	Name      string         `json:"name"`
	Steps     []PipelineStep `json:"steps"`
	Config    PipelineConfig `json:"config"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// CreatePipelineInput 入库参数。
type CreatePipelineInput struct {
	ProjectID string         `json:"projectId"`
	Name      string         `json:"name"`
	Steps     []PipelineStep `json:"steps"`
	Config    PipelineConfig `json:"config"`
}

// UpdatePipelineInput 更新参数（不含 ID、projectID）。
type UpdatePipelineInput struct {
	Name   string         `json:"name"`
	Steps  []PipelineStep `json:"steps"`
	Config PipelineConfig `json:"config"`
}

// ListPipelines 列出指定项目下所有流水线，按创建时间倒序。
func ListPipelines(projectID string) ([]Pipeline, error) {
	rows, err := DB.Query(
		`SELECT id, project_id, name, steps, config, created_at, updated_at
		 FROM pipelines WHERE project_id = ?
		 ORDER BY created_at DESC`,
		projectID)
	if err != nil {
		return nil, fmt.Errorf("query pipelines: %w", err)
	}
	defer rows.Close()

	pipelines := make([]Pipeline, 0)
	for rows.Next() {
		p, err := scanPipeline(rows)
		if err != nil {
			return nil, err
		}
		pipelines = append(pipelines, p)
	}
	return pipelines, rows.Err()
}

// CountPipelines 流水线总数（Dashboard 用）。
func CountPipelines() (int, error) {
	var n int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM pipelines`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count pipelines: %w", err)
	}
	return n, nil
}

// ListAllPipelines 列出所有项目下的流水线（前端一次拉取，避免 N+1）。
func ListAllPipelines() ([]Pipeline, error) {
	rows, err := DB.Query(
		`SELECT id, project_id, name, steps, config, created_at, updated_at
		 FROM pipelines ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query all pipelines: %w", err)
	}
	defer rows.Close()

	pipelines := make([]Pipeline, 0)
	for rows.Next() {
		p, err := scanPipeline(rows)
		if err != nil {
			return nil, err
		}
		pipelines = append(pipelines, p)
	}
	return pipelines, rows.Err()
}

// GetPipeline 按 ID 查询；不存在时返回 ErrNotFound。
func GetPipeline(id string) (*Pipeline, error) {
	var p Pipeline
	var stepsJSON, configJSON string
	err := DB.QueryRow(
		`SELECT id, project_id, name, steps, config, created_at, updated_at
		 FROM pipelines WHERE id = ?`, id).
		Scan(&p.ID, &p.ProjectID, &p.Name, &stepsJSON, &configJSON, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query pipeline: %w", err)
	}
	if err := json.Unmarshal([]byte(stepsJSON), &p.Steps); err != nil {
		p.Steps = []PipelineStep{}
	}
	if err := json.Unmarshal([]byte(configJSON), &p.Config); err != nil {
		p.Config = PipelineConfig{StopOnFail: true}
	}
	return &p, nil
}

// CreatePipeline 落一条新流水线。steps/config 以 JSON 存表里。
func CreatePipeline(input CreatePipelineInput) (*Pipeline, error) {
	id := uuid.New().String()
	now := time.Now().UTC()

	stepsJSON, _ := json.Marshal(input.Steps)
	configJSON, _ := json.Marshal(input.Config)

	_, err := DB.Exec(
		`INSERT INTO pipelines (id, project_id, name, steps, config, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, input.ProjectID, input.Name, string(stepsJSON), string(configJSON), now, now)
	if err != nil {
		return nil, fmt.Errorf("insert pipeline: %w", err)
	}

	return &Pipeline{
		ID:        id,
		ProjectID: input.ProjectID,
		Name:      input.Name,
		Steps:     input.Steps,
		Config:    input.Config,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdatePipeline 更新；目标不存在时返回 ErrNotFound。
func UpdatePipeline(id string, input UpdatePipelineInput) (*Pipeline, error) {
	now := time.Now().UTC()
	stepsJSON, _ := json.Marshal(input.Steps)
	configJSON, _ := json.Marshal(input.Config)

	result, err := DB.Exec(
		`UPDATE pipelines SET name = ?, steps = ?, config = ?, updated_at = ? WHERE id = ?`,
		input.Name, string(stepsJSON), string(configJSON), now, id)
	if err != nil {
		return nil, fmt.Errorf("update pipeline: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return nil, ErrNotFound
	}
	return GetPipeline(id)
}

// DeletePipeline 按 ID 删除；目标不存在时返回 ErrNotFound。
func DeletePipeline(id string) error {
	result, err := DB.Exec(`DELETE FROM pipelines WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete pipeline: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// scanPipeline 将一行结果装配成 Pipeline（含 JSON 字段反序列化容错）。
func scanPipeline(rows *sql.Rows) (Pipeline, error) {
	var p Pipeline
	var stepsJSON, configJSON string
	if err := rows.Scan(&p.ID, &p.ProjectID, &p.Name, &stepsJSON, &configJSON, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return Pipeline{}, fmt.Errorf("scan pipeline: %w", err)
	}
	if err := json.Unmarshal([]byte(stepsJSON), &p.Steps); err != nil {
		p.Steps = []PipelineStep{}
	}
	if err := json.Unmarshal([]byte(configJSON), &p.Config); err != nil {
		p.Config = PipelineConfig{StopOnFail: true}
	}
	return p, nil
}
