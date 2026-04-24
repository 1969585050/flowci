package handler

import (
	"context"
	"fmt"
	"strings"

	"flowci/internal/docker"
)

// ListContainers 列出全部容器（含已停止的）。
func (a *App) ListContainers(ctx context.Context) ([]docker.Container, error) {
	return docker.ListContainers(ctx)
}

// StartContainer 启动容器。
func (a *App) StartContainer(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: id required", ErrBadRequest)
	}
	return docker.StartContainer(ctx, id)
}

// StopContainer 停止容器。
func (a *App) StopContainer(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: id required", ErrBadRequest)
	}
	return docker.StopContainer(ctx, id)
}

// RemoveContainer 强制删除容器（等效 kill + rm）。
func (a *App) RemoveContainer(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: id required", ErrBadRequest)
	}
	return docker.RemoveContainer(ctx, id)
}

// GetContainerLogs 拉取最近 tail 行日志；tail ≤ 0 时默认 100。
func (a *App) GetContainerLogs(ctx context.Context, id string, tail int) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("%w: id required", ErrBadRequest)
	}
	return docker.GetContainerLogs(ctx, id, tail)
}

// DeployContainer 直接 docker run 启动容器（非 compose 路径）。
func (a *App) DeployContainer(ctx context.Context, req *DeployContainerRequest) (*docker.DeployResult, error) {
	if req == nil || strings.TrimSpace(req.Image) == "" || strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("%w: image and name required", ErrBadRequest)
	}
	res, err := docker.Deploy(ctx, docker.DeployRequest{
		Image:         req.Image,
		Name:          req.Name,
		HostPort:      req.HostPort,
		ContainerPort: req.ContainerPort,
		RestartPolicy: req.RestartPolicy,
		EnvMultiline:  req.Env,
	})
	if err != nil {
		return nil, err
	}
	return &res, nil
}
