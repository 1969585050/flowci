package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"flowci/internal/docker"
	"flowci/internal/validate"
)

// GenerateCompose 由规格产出 docker-compose.yml 文本。
// 基础字段过白名单，避免产出非法 YAML 或注入。
func (a *App) GenerateCompose(ctx context.Context, req *GenerateComposeRequest) (string, error) {
	if req == nil || strings.TrimSpace(req.Image) == "" || strings.TrimSpace(req.Name) == "" {
		return "", fmt.Errorf("%w: image and name required", ErrBadRequest)
	}
	if err := validate.ImageRef(req.Image); err != nil {
		return "", fmt.Errorf("%w: %v", ErrBadRequest, err)
	}
	if err := validate.ContainerName(req.Name); err != nil {
		return "", fmt.Errorf("%w: %v", ErrBadRequest, err)
	}
	if err := validate.Port(req.HostPort); err != nil {
		return "", fmt.Errorf("%w: hostPort %v", ErrBadRequest, err)
	}
	if err := validate.Port(req.ContainerPort); err != nil {
		return "", fmt.Errorf("%w: containerPort %v", ErrBadRequest, err)
	}
	if err := validate.EnvMultiline(req.Env); err != nil {
		return "", fmt.Errorf("%w: %v", ErrBadRequest, err)
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
// workdir 必须在 App.dataDir 下，防止覆盖任意路径的已有 compose 文件。
func (a *App) DeployWithCompose(ctx context.Context, req *DeployWithComposeRequest) (*docker.ComposeDeployResult, error) {
	if req == nil || strings.TrimSpace(req.Compose) == "" {
		return nil, fmt.Errorf("%w: compose required", ErrBadRequest)
	}

	workdir, err := a.resolveComposeWorkdir(req.Workdir)
	if err != nil {
		return nil, err
	}

	res, err := a.docker.DeployWithCompose(ctx, req.Compose, workdir)
	if err != nil {
		return &res, err
	}
	return &res, nil
}

// resolveComposeWorkdir 把前端传入的 workdir 规范化为 dataDir 下的绝对路径。
// 空字符串默认落到 dataDir/tmp/compose；其它路径必须位于 dataDir 下。
func (a *App) resolveComposeWorkdir(raw string) (string, error) {
	base, err := filepath.Abs(a.dataDir)
	if err != nil {
		return "", fmt.Errorf("resolve data dir: %w", err)
	}

	if strings.TrimSpace(raw) == "" {
		workdir := filepath.Join(base, "tmp", "compose")
		if err := os.MkdirAll(workdir, 0755); err != nil {
			return "", fmt.Errorf("create default compose workdir: %w", err)
		}
		return workdir, nil
	}

	absRaw, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("%w: workdir %q: %v", ErrBadRequest, raw, err)
	}
	// 必须落在 dataDir 之内，防止覆盖任意路径
	if !strings.HasPrefix(absRaw+string(filepath.Separator),
		base+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: workdir %q must be under data dir %q",
			ErrBadRequest, absRaw, base)
	}
	if err := os.MkdirAll(absRaw, 0755); err != nil {
		return "", fmt.Errorf("create compose workdir: %w", err)
	}
	return absRaw, nil
}
