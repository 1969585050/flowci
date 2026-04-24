package store

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CreateProjectInput struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Language string `json:"language"`
}

type UpdateProjectInput struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Language string `json:"language"`
}

func ListProjects() ([]Project, error) {
	if DB == nil {
		return []Project{}, fmt.Errorf("database not initialized")
	}
	rows, err := DB.Query(`SELECT id, name, path, language, created_at, updated_at FROM projects ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Path, &p.Language, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	if projects == nil {
		projects = []Project{}
	}
	return projects, nil
}

func CreateProject(input CreateProjectInput) (Project, error) {
	now := time.Now()
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

func GetProject(id string) (Project, error) {
	var p Project
	err := DB.QueryRow(
		`SELECT id, name, path, language, created_at, updated_at FROM projects WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.Path, &p.Language, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return Project{}, fmt.Errorf("get project %s: %w", id, err)
	}
	return p, nil
}

func UpdateProject(id string, input UpdateProjectInput) (Project, error) {
	now := time.Now()
	_, err := DB.Exec(
		`UPDATE projects SET name=?, path=?, language=?, updated_at=? WHERE id=?`,
		input.Name, input.Path, input.Language, now, id,
	)
	if err != nil {
		return Project{}, fmt.Errorf("update project %s: %w", id, err)
	}
	return GetProject(id)
}

func DeleteProject(id string) error {
	_, err := DB.Exec(`DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete project %s: %w", id, err)
	}
	return nil
}
