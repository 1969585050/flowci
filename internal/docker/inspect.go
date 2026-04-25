package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// EnvReport 一次性 Docker 环境探测的结果。
// 即使部分项失败也尽量填充其余字段，UI 端展示哪些可用、哪些不可用。
type EnvReport struct {
	Host          string `json:"host"`          // 实际生效的 DOCKER_HOST（空表示本地 daemon）
	Connected     bool   `json:"connected"`     // server 是否可达
	ClientVersion string `json:"clientVersion"` // 本地 docker CLI 版本
	ServerVersion string `json:"serverVersion"` // 远端 daemon 版本
	ServerOS      string `json:"serverOS"`      // 远端 OS（linux / windows）
	ServerArch    string `json:"serverArch"`    // 远端架构
	HasBuildx     bool   `json:"hasBuildx"`     // docker buildx 插件是否可用
	HasCompose    bool   `json:"hasCompose"`    // docker compose 插件是否可用
	Message       string `json:"message"`       // 失败原因 / 友好提示
}

// DetectEnv 一次性探测目标 Docker 环境。
//   - host 为空：使用当前 SetDockerHost 配置的值（可能是本地）
//   - host 非空：仅本次调用覆盖 DOCKER_HOST，不影响全局配置
//
// 永不返回 error；探测失败的项会留空，原因写入 Message。
func DetectEnv(ctx context.Context, host string) EnvReport {
	actualHost := strings.TrimSpace(host)
	if actualHost == "" {
		actualHost = GetDockerHost()
	}
	report := EnvReport{Host: actualHost}

	// 1) docker version --format json：客户端 + 服务端版本
	out, err := runProbe(ctx, actualHost, TimeoutQuery, "version", "--format", "{{json .}}")
	if err != nil {
		report.Message = friendlyDockerErr(err)
		// 尝试拿 client 版本（即使 server 不通）
		if cv, e := runProbe(ctx, actualHost, TimeoutQuery, "version", "--format", "{{.Client.Version}}"); e == nil {
			report.ClientVersion = strings.TrimSpace(string(cv))
		}
		return report
	}

	var v struct {
		Client struct{ Version string }
		Server struct {
			Version string
			Os      string
			Arch    string
		}
	}
	if jerr := json.Unmarshal(out, &v); jerr != nil {
		report.Message = fmt.Sprintf("parse docker version json: %v", jerr)
		return report
	}
	report.ClientVersion = v.Client.Version
	if v.Server.Version == "" {
		report.Message = "Docker CLI 已安装，但远端 daemon 不可达"
		return report
	}
	report.Connected = true
	report.ServerVersion = v.Server.Version
	report.ServerOS = v.Server.Os
	report.ServerArch = v.Server.Arch

	// 2) buildx 插件
	if _, err := runProbe(ctx, actualHost, TimeoutQuery, "buildx", "version"); err == nil {
		report.HasBuildx = true
	}

	// 3) compose 插件（v2，docker compose 而非 docker-compose）
	if _, err := runProbe(ctx, actualHost, TimeoutQuery, "compose", "version"); err == nil {
		report.HasCompose = true
	}

	return report
}

// runProbe 是 DetectEnv 专用的 exec 包装，独立于 run() 是因为它要支持显式 host 覆盖。
// host 为空时不注入 DOCKER_HOST（沿用进程 env）。
func runProbe(ctx context.Context, host string, timeout time.Duration, args ...string) ([]byte, error) {
	ctxTO, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctxTO, "docker", args...)
	if host != "" {
		cmd.Env = append(os.Environ(), "DOCKER_HOST="+host)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return out, nil
}

// friendlyDockerErr 把 docker 命令错误转成对用户更友好的中文提示。
func friendlyDockerErr(err error) string {
	if err == nil {
		return ""
	}
	s := strings.ToLower(err.Error())
	switch {
	case strings.Contains(s, "executable file not found"):
		return "未安装 docker CLI；请安装 Docker Desktop 或独立 docker.exe"
	case strings.Contains(s, "cannot connect to the docker daemon"),
		strings.Contains(s, "is the docker daemon running"):
		return "docker daemon 未运行或无法连接（检查 DOCKER_HOST、权限、防火墙）"
	case strings.Contains(s, "permission denied"):
		return "无 docker 操作权限"
	case strings.Contains(s, "context deadline exceeded"):
		return "连接超时（远端 daemon 无响应）"
	}
	return err.Error()
}
