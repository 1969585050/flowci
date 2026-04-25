package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"flowci/internal/docker"
	"flowci/internal/secret"
	"flowci/internal/validate"
)

// PushImage 推送镜像到 Registry。
//
// 凭证获取顺序（Password 字段）：
//  1. 请求中显式提供的 Password（一次性使用，不落盘）
//  2. OS keyring 中 key=`registry:<Registry>` 的条目（阶段 3 起推荐）
//  3. 都没有则不登录，期望 docker 已有本地缓存
//
// 日志输出 Password 以 *** 遮蔽。
func (a *App) PushImage(req *PushImageRequest) (*docker.PushResult, error) {
	if req == nil || strings.TrimSpace(req.Image) == "" {
		return nil, fmt.Errorf("%w: image required", ErrBadRequest)
	}
	if err := validate.ImageRef(req.Image); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadRequest, err)
	}
	if err := validate.RegistryHost(req.Registry); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadRequest, err)
	}

	password := req.Password
	if password == "" && req.Username != "" {
		keyringKey := registryKeyringKey(req.Registry)
		if v, err := secret.Get(keyringKey); err == nil {
			password = v
		} else if !errors.Is(err, secret.ErrNotFound) {
			slog.Warn("keyring lookup failed", "key", keyringKey, "err", err)
		}
	}

	slog.Info("push image",
		"image", req.Image,
		"registry", req.Registry,
		"username", req.Username,
		"password", secret.Mask(password))

	res, err := a.docker.PushImage(a.ctx, docker.PushRequest{
		Image:    req.Image,
		Registry: req.Registry,
		Username: req.Username,
		Password: password,
	})
	if err != nil {
		return &res, err
	}
	return &res, nil
}

// registryKeyringKey 将 registry host 规范化为 keyring key。
// 空 / "docker.io" 统一用 "registry:docker.io" 作 Hub 默认键。
func registryKeyringKey(registry string) string {
	r := strings.TrimSpace(registry)
	if r == "" {
		r = "docker.io"
	}
	return "registry:" + r
}
