package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"flowci/internal/docker"
	"flowci/internal/store"
	"flowci/internal/validate"
)

// BuildImage 构建单个镜像（非 pipeline 路径，直接 UI 触发）。
// 同时落一条 build_records 记录（building → success/failed），日志写入记录。
func (a *App) BuildImage(ctx context.Context, req *BuildImageRequest) (*docker.BuildResult, error) {
	if req == nil || strings.TrimSpace(req.ProjectID) == "" {
		return nil, fmt.Errorf("%w: projectId required", ErrBadRequest)
	}
	if strings.TrimSpace(req.Tag) == "" {
		return nil, fmt.Errorf("%w: tag required", ErrBadRequest)
	}
	if err := validate.ImageRef(req.Tag); err != nil {
		return nil, fmt.Errorf("%w: tag %v", ErrBadRequest, err)
	}
	if _, err := store.GetProject(req.ProjectID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	record, err := store.CreateBuildRecord(req.ProjectID, req.Tag, "latest")
	if err != nil {
		return nil, fmt.Errorf("create build record: %w", err)
	}

	res, buildErr := a.docker.BuildImage(ctx, docker.BuildRequest{
		Tag:         req.Tag,
		ContextPath: req.ContextPath,
		NoCache:     req.NoCache,
		PullLatest:  req.PullLatest,
	})

	status := "success"
	if buildErr != nil {
		status = "failed"
	}
	logsDir := filepath.Join(a.dataDir, "logs", "builds")
	if err := store.FinishBuildRecord(record.ID, status, res.Log, logsDir); err != nil {
		slog.Error("finish build record failed", "id", record.ID, "err", err)
	}

	if buildErr != nil {
		return &res, buildErr
	}
	return &res, nil
}

// ListBuildRecords 列出某项目最近的构建记录（不含日志，日志走 GetBuildRecord）。
func (a *App) ListBuildRecords(ctx context.Context, projectID string) ([]store.BuildRecord, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, fmt.Errorf("%w: projectId required", ErrBadRequest)
	}
	return store.ListBuildRecords(projectID)
}

// GetBuildRecord 按 ID 查一条构建记录（含完整日志）。
func (a *App) GetBuildRecord(ctx context.Context, id string) (*store.BuildRecord, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("%w: id required", ErrBadRequest)
	}
	r, err := store.GetBuildRecord(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrBuildNotFound
		}
		return nil, err
	}
	return &r, nil
}
