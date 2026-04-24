package docker

import (
	"context"
	"fmt"
	"strings"
)

// Container 表示一条 docker ps 列表项。
type Container struct {
	ID     string   `json:"id"`
	Names  []string `json:"names"`
	Image  string   `json:"image"`
	State  string   `json:"state"`
	Status string   `json:"status"`
	Ports  string   `json:"ports"`
}

// DeployRequest 启动新容器的参数。EnvMultiline 为多行 KEY=VAL，\n 分隔。
type DeployRequest struct {
	Image         string
	Name          string
	HostPort      string
	ContainerPort string
	RestartPolicy string
	EnvMultiline  string
}

// DeployResult 是 Deploy 成功时的返回值。
type DeployResult struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// ListContainers 列出全部容器（包括停止的）。
func ListContainers(ctx context.Context) ([]Container, error) {
	out, err := run(ctx, TimeoutQuery, "ps", "-a",
		"--format", "{{.ID}}|{{.Names}}|{{.Image}}|{{.State}}|{{.Status}}|{{.Ports}}")
	if err != nil {
		return nil, err
	}
	return parseContainers(string(out)), nil
}

func parseContainers(s string) []Container {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return []Container{}
	}
	lines := strings.Split(trimmed, "\n")
	cs := make([]Container, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		// docker ps Ports 字段可能为空；补齐到 6 段避免越界
		for len(parts) < 6 {
			parts = append(parts, "")
		}
		cs = append(cs, Container{
			ID:     parts[0],
			Names:  []string{parts[1]},
			Image:  parts[2],
			State:  parts[3],
			Status: parts[4],
			Ports:  parts[5],
		})
	}
	return cs
}

// StartContainer 启动指定容器。
func StartContainer(ctx context.Context, id string) error {
	_, err := run(ctx, TimeoutLifecycle, "start", id)
	return err
}

// StopContainer 停止指定容器。
func StopContainer(ctx context.Context, id string) error {
	_, err := run(ctx, TimeoutLifecycle, "stop", id)
	return err
}

// RemoveContainer 强制删除容器（-f 等同于 kill + rm）。
func RemoveContainer(ctx context.Context, id string) error {
	_, err := run(ctx, TimeoutLifecycle, "rm", "-f", id)
	return err
}

// GetContainerLogs 拉取最近 tail 行日志；tail ≤ 0 时默认 100。
func GetContainerLogs(ctx context.Context, id string, tail int) (string, error) {
	if tail <= 0 {
		tail = 100
	}
	out, err := run(ctx, TimeoutQuery, "logs", "--tail", fmt.Sprintf("%d", tail), id)
	return string(out), err
}

// Deploy 启动新容器。EnvMultiline 的每行会成为独立 `-e KEY=VAL` 参数（不走 shell）。
func Deploy(ctx context.Context, req DeployRequest) (DeployResult, error) {
	policy := req.RestartPolicy
	if policy == "" {
		policy = "unless-stopped"
	}

	args := []string{"run", "-d", "--name", req.Name, "--restart", policy}
	if req.HostPort != "" && req.ContainerPort != "" {
		args = append(args, "-p", req.HostPort+":"+req.ContainerPort)
	}
	if req.EnvMultiline != "" {
		for _, line := range strings.Split(req.EnvMultiline, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				args = append(args, "-e", line)
			}
		}
	}
	args = append(args, req.Image)

	out, err := run(ctx, TimeoutLifecycle, args...)
	if err != nil {
		return DeployResult{}, err
	}
	return DeployResult{
		ID:      strings.TrimSpace(string(out)),
		Message: "Container deployed successfully",
	}, nil
}
