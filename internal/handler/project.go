package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"flowci/internal/store"
)

// ListProjects 列出全部项目，按 updated_at DESC 排序。
func (a *App) ListProjects(ctx context.Context) ([]store.Project, error) {
	return store.ListProjects()
}

// CreateProject 新建项目。必填：name、path。
func (a *App) CreateProject(ctx context.Context, req *CreateProjectRequest) (*store.Project, error) {
	if req == nil {
		return nil, ErrBadRequest
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Path) == "" {
		return nil, fmt.Errorf("%w: name and path are required", ErrBadRequest)
	}
	p, err := store.CreateProject(store.CreateProjectInput{
		Name:     req.Name,
		Path:     req.Path,
		Language: req.Language,
	})
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// UpdateProject 更新指定 ID 的项目；ID 不存在时返回 ErrProjectNotFound。
func (a *App) UpdateProject(ctx context.Context, req *UpdateProjectRequest) (*store.Project, error) {
	if req == nil || strings.TrimSpace(req.ID) == "" {
		return nil, fmt.Errorf("%w: id required", ErrBadRequest)
	}
	p, err := store.UpdateProject(req.ID, store.UpdateProjectInput{
		Name:     req.Name,
		Path:     req.Path,
		Language: req.Language,
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return &p, nil
}

// DeleteProject 按 ID 删除项目；不存在时返回 ErrProjectNotFound。
func (a *App) DeleteProject(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: id required", ErrBadRequest)
	}
	if err := store.DeleteProject(id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrProjectNotFound
		}
		return err
	}
	return nil
}
