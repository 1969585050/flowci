package handler

import (
	"context"
	"fmt"
	"strings"

	"flowci/internal/docker"
)

// PushImage 推送镜像到 Registry。
// Password 目前仍走 IPC 明文（阶段 3 改 keyring）。
func (a *App) PushImage(ctx context.Context, req *PushImageRequest) (*docker.PushResult, error) {
	if req == nil || strings.TrimSpace(req.Image) == "" {
		return nil, fmt.Errorf("%w: image required", ErrBadRequest)
	}
	res, err := docker.PushImage(ctx, docker.PushRequest{
		Image:    req.Image,
		Registry: req.Registry,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return &res, err
	}
	return &res, nil
}
