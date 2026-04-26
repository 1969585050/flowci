package handler

import (
	"errors"
	"fmt"
	"strings"

	"flowci/internal/git"
	"flowci/internal/store"
)

// ListProjects 列出全部项目，按 updated_at DESC 排序。
func (a *App) ListProjects() ([]store.Project, error) {
	return store.ListProjects()
}

// ListProjectsWithStats 一次返回所有项目 + 摘要（最近构建、构建数、Git HEAD）。
// 比前端逐项目调 ListBuildRecords + git log 高效，避免 N+1。
// Git HEAD 信息只对有 RepoURL 且本地 path 存在的项目尝试取，失败静默。
func (a *App) ListProjectsWithStats() ([]ProjectStats, error) {
	projects, err := store.ListProjects()
	if err != nil {
		return nil, err
	}
	stats := make([]ProjectStats, 0, len(projects))
	for _, p := range projects {
		s := ProjectStats{Project: p}

		if last, err := store.LatestBuildRecord(p.ID); err == nil {
			s.LastBuild = &last
		}
		if n, err := store.CountBuildRecords(p.ID); err == nil {
			s.BuildCount = n
		}
		if p.RepoURL != "" && p.Path != "" {
			if sha, subj, err := git.HeadCommit(a.ctx, p.Path); err == nil {
				s.HeadCommit = sha
				s.HeadSubject = subj
			}
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// CreateProject 新建项目。必填：name、path。
func (a *App) CreateProject(req *CreateProjectRequest) (*store.Project, error) {
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
func (a *App) UpdateProject(req *UpdateProjectRequest) (*store.Project, error) {
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
func (a *App) DeleteProject(id string) error {
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

// PinProject 把项目置顶到列表最前。
func (a *App) PinProject(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: id required", ErrBadRequest)
	}
	if err := store.PinProject(id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrProjectNotFound
		}
		return err
	}
	return nil
}

// UnpinProject 取消置顶。
func (a *App) UnpinProject(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: id required", ErrBadRequest)
	}
	if err := store.UnpinProject(id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrProjectNotFound
		}
		return err
	}
	return nil
}
