package docker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// PushRequest 推送镜像参数。
// Registry 为空或为 "docker.io" 时表示 Docker Hub。
type PushRequest struct {
	Image    string
	Registry string
	Username string
	Password string
}

// PushResult 推送结果。
type PushResult struct {
	Log string `json:"log"`
}

// PushImage 推送镜像到 Registry。
// 流程：
//  1. 如 Registry 非 Hub 且带凭证 → docker login Registry
//  2. 如 Registry 非 Hub → docker tag <Image> <Registry>/<Image>
//  3. 如 Registry 为 Hub 且带凭证 → docker login
//  4. docker push <target>
//
// 密码通过 stdin 传给 docker login，避免进入进程参数表。
func PushImage(ctx context.Context, req PushRequest) (PushResult, error) {
	target := req.Image

	if req.Registry != "" && req.Registry != "docker.io" {
		if req.Username != "" && req.Password != "" {
			if err := dockerLogin(ctx, req.Registry, req.Username, req.Password); err != nil {
				return PushResult{}, fmt.Errorf("login failed: %w", err)
			}
		}
		target = req.Registry + "/" + req.Image
		if _, err := run(ctx, TimeoutLifecycle, "tag", req.Image, target); err != nil {
			return PushResult{}, fmt.Errorf("tag failed: %w", err)
		}
	} else if req.Registry == "docker.io" && req.Username != "" && req.Password != "" {
		if err := dockerLogin(ctx, "", req.Username, req.Password); err != nil {
			return PushResult{}, fmt.Errorf("docker hub login failed: %w", err)
		}
	}

	out, err := run(ctx, TimeoutPush, "push", target)
	return PushResult{Log: string(out)}, err
}

// dockerLogin 封装 docker login；registry 为空则登录 Docker Hub。
// 密码通过 stdin 传入，进程参数表中不出现密码。
func dockerLogin(ctx context.Context, registry, username, password string) error {
	ctxTO, cancel := context.WithTimeout(ctx, TimeoutQuery)
	defer cancel()
	args := []string{"login"}
	if registry != "" {
		args = append(args, registry)
	}
	args = append(args, "-u", username, "--password-stdin")
	cmd := exec.CommandContext(ctxTO, "docker", args...)
	cmd.Stdin = strings.NewReader(password)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
