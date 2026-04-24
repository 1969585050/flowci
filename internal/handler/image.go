package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"flowci/internal/docker"
)

// ListImages 列出本机 docker 镜像。
func (a *App) ListImages(ctx context.Context) ([]docker.Image, error) {
	return docker.ListImages(ctx)
}

// RemoveImage 按 ID 删除镜像。
// 业务错误以中文消息返回给前端（docker.Err* 在此处做语义化翻译）。
func (a *App) RemoveImage(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: id required", ErrBadRequest)
	}
	if err := docker.RemoveImage(ctx, id); err != nil {
		switch {
		case errors.Is(err, docker.ErrImageNotFound):
			return fmt.Errorf("镜像不存在: %w", err)
		case errors.Is(err, docker.ErrImageInUse):
			return fmt.Errorf("镜像正在使用中，请先停止使用该镜像的容器: %w", err)
		default:
			return err
		}
	}
	return nil
}
