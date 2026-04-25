package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateProjectInput 新建项目入参。
// RepoURL 非空表示 Git 仓库项目（FlowCI 自动 clone 到 Path）。
type CreateProjectInput struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Language   string `json:"language"`
	RepoURL    string `json:"repoUrl"`
	RepoBranch string `json:"repoBranch"`
}

// UpdateProjectInput 更新项目入参（不含 ID，ID 通过单独参数传入）。
type UpdateProjectInput struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Language   string `json:"language"`
	RepoURL    string `json:"repoUrl"`
	RepoBranch string `json:"repoBranch"`
}

// ListProjects 按 updated_at DESC 返回全部项目。
func ListProjects() ([]Project, error) {
	if DB == nil {
		return []Project{}, fmt.Errorf("database not initialized")
	}
	rows, err := DB.Query(
		`SELECT id, name, path, language, repo_url, repo_branch, last_pull_at, created_at, updated_at
		 FROM projects ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	projects := make([]Project, 0)
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// CreateProject 新建项目。
func CreateProject(input CreateProjectInput) (Project, error) {
	now := time.Now().UTC()
	p := Project{
		ID:         uuid.NewString(),
		Name:       input.Name,
		Path:       input.Path,
		Language:   input.Language,
		RepoURL:    input.RepoURL,
		RepoBranch: input.RepoBranch,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	_, err := DB.Exec(
		`INSERT INTO projects (id, name, path, language, repo_url, repo_branch, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.Path, p.Language, p.RepoURL, p.RepoBranch, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}
	return p, nil
}

// GetProject 按 ID 查询；不存在返回 ErrNotFound。
func GetProject(id string) (Project, error) {
	row := DB.QueryRow(
		`SELECT id, name, path, language, repo_url, repo_branch, last_pull_at, created_at, updated_at
		 FROM projects WHERE id = ?`, id)
	p, err := scanProject(row)
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
		`UPDATE projects
		 SET name=?, path=?, language=?, repo_url=?, repo_branch=?, updated_at=?
		 WHERE id=?`,
		input.Name, input.Path, input.Language, input.RepoURL, input.RepoBranch, now, id,
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

// MarkProjectPulled 写入最近一次 pull 成功的时间戳。
func MarkProjectPulled(id string) error {
	now := time.Now().UTC()
	result, err := DB.Exec(
		`UPDATE projects SET last_pull_at=?, updated_at=? WHERE id=?`,
		now, now, id,
	)
	if err != nil {
		return fmt.Errorf("mark pulled %s: %w", id, err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
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

// scanner 抽象 sql.Row / sql.Rows 的 Scan，便于 GetProject 与 ListProjects 复用。
type scanner interface {
	Scan(dest ...interface{}) error
}

func scanProject(s scanner) (Project, error) {
	var p Project
	var lastPull sql.NullTime
	if err := s.Scan(&p.ID, &p.Name, &p.Path, &p.Language, &p.RepoURL, &p.RepoBranch, &lastPull, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return Project{}, err
	}
	if lastPull.Valid {
		p.LastPullAt = &lastPull.Time
	}
	return p, nil
}
