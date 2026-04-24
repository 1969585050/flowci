package handler

import (
	"context"
	"fmt"
	"strings"

	"flowci/internal/docker"
)

// GenerateCompose 由规格产出 docker-compose.yml 文本。
func (a *App) GenerateCompose(ctx context.Context, req *GenerateComposeRequest) (string, error) {
	if req == nil || strings.TrimSpace(req.Image) == "" || strings.TrimSpace(req.Name) == "" {
		return "", fmt.Errorf("%w: image and name required", ErrBadRequest)
	}
	return docker.GenerateCompose(docker.ComposeSpec{
		Image:         req.Image,
		Name:          req.Name,
		HostPort:      req.HostPort,
		ContainerPort: req.ContainerPort,
		RestartPolicy: req.RestartPolicy,
		EnvMultiline:  req.Env,
	})
}

// DeployWithCompose 将 compose 文本写到 workdir/docker-compose.yml 并 up -d。
// TODO(phase-3)：workdir 必须限制在 data dir 下，防止覆盖任意路径。
func (a *App) DeployWithCompose(ctx context.Context, req *DeployWithComposeRequest) (*docker.ComposeDeployResult, error) {
	if req == nil || strings.TrimSpace(req.Compose) == "" {
		return nil, fmt.Errorf("%w: compose required", ErrBadRequest)
	}
	res, err := docker.DeployWithCompose(ctx, req.Compose, req.Workdir)
	if err != nil {
		return &res, err
	}
	return &res, nil
}
