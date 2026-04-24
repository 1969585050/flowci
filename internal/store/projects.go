package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateProjectInput 新建项目入参。
type CreateProjectInput struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Language string `json:"language"`
}

// UpdateProjectInput 更新项目入参（不含 ID，ID 通过单独参数传入）。
type UpdateProjectInput struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Language string `json:"language"`
}

// ListProjects 按 updated_at DESC 返回全部项目。
func ListProjects() ([]Project, error) {
	if DB == nil {
		return []Project{}, fmt.Errorf("database not initialized")
	}
	rows, err := DB.Query(
		`SELECT id, name, path, language, created_at, updated_at FROM projects ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	projects := make([]Project, 0)
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Path, &p.Language, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// CreateProject 新建项目。
func CreateProject(input CreateProjectInput) (Project, error) {
	now := time.Now().UTC()
	p := Project{
		ID:        uuid.NewString(),
		Name:      input.Name,
		Path:      input.Path,
		Language:  input.Language,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := DB.Exec(
		`INSERT INTO projects (id, name, path, language, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.Path, p.Language, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}
	return p, nil
}

// GetProject 按 ID 查询；不存在返回 ErrNotFound。
func GetProject(id string) (Project, error) {
	var p Project
	err := DB.QueryRow(
		`SELECT id, name, path, language, created_at, updated_at FROM projects WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.Path, &p.Language, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	if err != nil {
		return Project{}, fmt.Errorf("get project %s: %w", id, err)
	}
	return p, nil
}

// UpdateProject 更新项目；不存在返回 ErrNotFound。
func UpdateProject(id string, input UpdateProjectInput) (Project, error) {
	now := time.Now().UTC()
	result, err := DB.Exec(
		`UPDATE projects SET name=?, path=?, language=?, updated_at=? WHERE id=?`,
		input.Name, input.Path, input.Language, now, id,
	)
	if err != nil {
		return Project{}, fmt.Errorf("update project %s: %w", id, err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return Project{}, ErrNotFound
	}
	return GetProject(id)
}

// DeleteProject 按 ID 删除；不存在返回 ErrNotFound。
func DeleteProject(id string) error {
	result, err := DB.Exec(`DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete project %s: %w", id, err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
