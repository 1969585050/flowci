package handler

import (
	"fmt"
	"strings"

	"flowci/internal/docker"
	"flowci/internal/validate"
)

// ListContainers 列出全部容器（含已停止的）。
func (a *App) ListContainers() ([]docker.Container, error) {
	return a.docker.ListContainers(a.ctx)
}

// StartContainer 启动容器。
func (a *App) StartContainer(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: id required", ErrBadRequest)
	}
	return a.docker.StartContainer(a.ctx, id)
}

// StopContainer 停止容器。
func (a *App) StopContainer(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: id required", ErrBadRequest)
	}
	return a.docker.StopContainer(a.ctx, id)
}

// RemoveContainer 强制删除容器（等效 kill + rm）。
func (a *App) RemoveContainer(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: id required", ErrBadRequest)
	}
	return a.docker.RemoveContainer(a.ctx, id)
}

// GetContainerLogs 拉取最近 tail 行日志；tail ≤ 0 时默认 100。
func (a *App) GetContainerLogs(id string, tail int) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("%w: id required", ErrBadRequest)
	}
	return a.docker.GetContainerLogs(a.ctx, id, tail)
}

// DeployContainer 直接 docker run 启动容器（非 compose 路径）。
// 所有用户输入过白名单校验（见 internal/validate），拒绝后不会触发 docker 调用。
func (a *App) DeployContainer(req *DeployContainerRequest) (*docker.DeployResult, error) {
	if req == nil || strings.TrimSpace(req.Image) == "" || strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("%w: image and name required", ErrBadRequest)
	}
	if err := validate.ImageRef(req.Image); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadRequest, err)
	}
	if err := validate.ContainerName(req.Name); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadRequest, err)
	}
	if err := validate.Port(req.HostPort); err != nil {
		return nil, fmt.Errorf("%w: hostPort %v", ErrBadRequest, err)
	}
	if err := validate.Port(req.ContainerPort); err != nil {
		return nil, fmt.Errorf("%w: containerPort %v", ErrBadRequest, err)
	}
	if err := validate.EnvMultiline(req.Env); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadRequest, err)
	}
	res, err := a.docker.Deploy(a.ctx, docker.DeployRequest{
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
