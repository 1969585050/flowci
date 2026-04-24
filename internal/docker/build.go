package docker

import (
	"context"
	"strings"
)

// BuildRequest 构建镜像的参数。
// ContextPath 为空时默认 "."（当前工作目录）。
type BuildRequest struct {
	Tag         string
	ContextPath string
	NoCache     bool
	PullLatest  bool
}

// BuildResult 构建结果。
// 即使 BuildImage 返回 error，Log 仍会填充 docker 输出供上层落 build_record.log。
type BuildResult struct {
	ImageName string `json:"imageName"`
	ImageTag  string `json:"imageTag"`
	Log       string `json:"log"`
}

// BuildImage 执行 `docker buildx build -t <tag> --progress=plain <context>`。
// 失败时返回的 BuildResult.Log 带完整输出，上层可根据 err 决定是否落盘。
func BuildImage(ctx context.Context, req BuildRequest) (BuildResult, error) {
	args := []string{"buildx", "build", "-t", req.Tag, "--progress=plain"}
	if req.NoCache {
		args = append(args, "--no-cache")
	}
	if req.PullLatest {
		args = append(args, "--pull")
	}
	contextPath := req.ContextPath
	if contextPath == "" {
		contextPath = "."
	}
	args = append(args, contextPath)

	out, err := run(ctx, TimeoutBuild, args...)
	name, tag := splitTag(req.Tag)
	return BuildResult{ImageName: name, ImageTag: tag, Log: string(out)}, err
}

// splitTag 把 "name:tag" 拆成 (name, tag)；无 tag 部分默认为 "latest"。
func splitTag(s string) (string, string) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], "latest"
}
