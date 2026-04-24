package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ComposeSpec 生成 docker-compose.yml 需要的字段。
type ComposeSpec struct {
	Image         string
	Name          string
	HostPort      string
	ContainerPort string
	RestartPolicy string
	EnvMultiline  string
}

// ComposeDeployResult docker compose up 的输出。
type ComposeDeployResult struct {
	Output string `json:"output"`
}

// GenerateCompose 根据 spec 产出 docker-compose.yml 文本。
// 用 yaml.Marshal 避免值里含引号、换行等字符破坏 YAML 结构。
func GenerateCompose(spec ComposeSpec) (string, error) {
	policy := spec.RestartPolicy
	if policy == "" {
		policy = "unless-stopped"
	}

	service := map[string]interface{}{
		"image":          spec.Image,
		"container_name": spec.Name,
		"restart":        policy,
	}
	if spec.HostPort != "" && spec.ContainerPort != "" {
		service["ports"] = []string{fmt.Sprintf("%s:%s", spec.HostPort, spec.ContainerPort)}
	}
	if spec.EnvMultiline != "" {
		var env []string
		for _, line := range strings.Split(spec.EnvMultiline, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				env = append(env, line)
			}
		}
		if len(env) > 0 {
			service["environment"] = env
		}
	}

	doc := map[string]interface{}{
		"version": "3.8",
		"services": map[string]interface{}{
			spec.Name: service,
		},
	}
	bs, err := yaml.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("marshal compose: %w", err)
	}
	return string(bs), nil
}

// DeployWithCompose 将 content 写到 workDir/docker-compose.yml 并执行 up -d。
// workDir 为空时用 "."；调用方负责保证 workDir 可写。
// TODO(phase-3)：workDir 需加白名单校验（限制在 data dir 下），防止覆盖任意路径。
func DeployWithCompose(ctx context.Context, content, workDir string) (ComposeDeployResult, error) {
	if workDir == "" {
		workDir = "."
	}
	path := filepath.Join(workDir, "docker-compose.yml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return ComposeDeployResult{}, fmt.Errorf("write compose file: %w", err)
	}

	ctxTO, cancel := context.WithTimeout(ctx, TimeoutCompose)
	defer cancel()
	cmd := exec.CommandContext(ctxTO, "docker", "compose", "up", "-d")
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	res := ComposeDeployResult{Output: string(out)}
	if err != nil {
		return res, fmt.Errorf("compose up: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return res, nil
}
