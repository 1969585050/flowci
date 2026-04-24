package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PipelineStep struct {
	Type    string `json:"type"`    // build, push, deploy
	Name    string `json:"name"`    // step name
	Config  map[string]interface{} `json:"config"`
	Retry   int    `json:"retry"`   // retry count
	OnFail  string `json:"on_fail"` // continue, stop
}

type PipelineConfig struct {
	Parallel   bool `json:"parallel"`
	StopOnFail bool `json:"stop_on_fail"`
}

type Pipeline struct {
	ID        string          `json:"id"`
	ProjectID string         `json:"project_id"`
	Name      string          `json:"name"`
	Steps     []PipelineStep  `json:"steps"`
	Config    PipelineConfig  `json:"config"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type CreatePipelineInput struct {
	ProjectID string         `json:"project_id"`
	Name      string         `json:"name"`
	Steps     []PipelineStep `json:"steps"`
	Config    PipelineConfig `json:"config"`
}

type UpdatePipelineInput struct {
	Name   string         `json:"name"`
	Steps  []PipelineStep `json:"steps"`
	Config PipelineConfig `json:"config"`
}

func ListPipelines(projectID string) ([]Pipeline, error) {
	rows, err := DB.Query(`SELECT id, project_id, name, steps, config, created_at, updated_at FROM pipelines WHERE project_id = ? ORDER BY created_at DESC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query pipelines: %w", err)
	}
	defer rows.Close()

	var pipelines []Pipeline
	for rows.Next() {
		var p Pipeline
		var stepsJSON, configJSON string
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.Name, &stepsJSON, &configJSON, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan pipeline: %w", err)
		}
		if err := json.Unmarshal([]byte(stepsJSON), &p.Steps); err != nil {
			p.Steps = []PipelineStep{}
		}
		if err := json.Unmarshal([]byte(configJSON), &p.Config); err != nil {
			p.Config = PipelineConfig{StopOnFail: true}
		}
		pipelines = append(pipelines, p)
	}
	return pipelines, rows.Err()
}

func GetPipeline(id string) (*Pipeline, error) {
	var p Pipeline
	var stepsJSON, configJSON string
	err := DB.QueryRow(`SELECT id, project_id, name, steps, config, created_at, updated_at FROM pipelines WHERE id = ?`, id).
		Scan(&p.ID, &p.ProjectID, &p.Name, &stepsJSON, &configJSON, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("pipeline not found")
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

func CreatePipeline(input CreatePipelineInput) (*Pipeline, error) {
	id := uuid.New().String()
	now := time.Now()

	stepsJSON, _ := json.Marshal(input.Steps)
	configJSON, _ := json.Marshal(input.Config)

	_, err := DB.Exec(`INSERT INTO pipelines (id, project_id, name, steps, config, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
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

func UpdatePipeline(id string, input UpdatePipelineInput) (*Pipeline, error) {
	now := time.Now()
	stepsJSON, _ := json.Marshal(input.Steps)
	configJSON, _ := json.Marshal(input.Config)

	result, err := DB.Exec(`UPDATE pipelines SET name = ?, steps = ?, config = ?, updated_at = ? WHERE id = ?`,
		input.Name, string(stepsJSON), string(configJSON), now, id)
	if err != nil {
		return nil, fmt.Errorf("update pipeline: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("pipeline not found")
	}

	return GetPipeline(id)
}

func DeletePipeline(id string) error {
	result, err := DB.Exec(`DELETE FROM pipelines WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete pipeline: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("pipeline not found")
	}
	return nil
}
