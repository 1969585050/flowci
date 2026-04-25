// Package git 封装对本机 git CLI 的调用。
//
// 设计原则跟 internal/docker 对齐：
//  1. 所有对外函数接 ctx + 超时
//  2. 返回强类型 struct + error
//  3. 不做业务级写入，由 handler 层负责持久化元数据
//
// MVP 范围只有环境探测 (DetectEnv)；clone / pull 等操作待需求确认后扩。
package git

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// EnvReport 一次性 git 环境探测结果。
type EnvReport struct {
	Installed    bool          `json:"installed"`
	Version      string        `json:"version"`        // "git version 2.43.0.windows.1"
	Path         string        `json:"path"`           // exec.LookPath 结果
	Message      string        `json:"message"`        // 失败原因 / 友好提示
	InstallHints []InstallHint `json:"installHints"`   // 当前平台的安装建议（不重叠展示）
}

// InstallHint 给前端提示如何安装 git；前端可直接展示一个"复制命令"按钮 + 一个"打开链接"按钮。
type InstallHint struct {
	Method  string `json:"method"`  // 简短标签：winget / choco / scoop / brew / apt / yum / official
	Label   string `json:"label"`   // 中文显示标签
	Command string `json:"command"` // 可复制的命令；为空表示只给链接
	URL     string `json:"url"`     // 官方下载页
}

// DetectEnv 检查 git 是否可用，给出版本、路径与安装提示。
// 永不返回 error；探测失败信息写在 Message。
func DetectEnv(ctx context.Context) EnvReport {
	report := EnvReport{InstallHints: hintsForCurrentOS()}

	path, err := exec.LookPath("git")
	if err != nil {
		report.Message = "未在 PATH 中找到 git，需要安装"
		return report
	}
	report.Path = path

	ctxTO, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctxTO, "git", "--version").Output()
	if err != nil {
		report.Message = friendlyErr(err)
		return report
	}
	report.Installed = true
	report.Version = strings.TrimSpace(string(out))
	return report
}

// hintsForCurrentOS 按当前操作系统给出 1-3 个推荐安装方式。
func hintsForCurrentOS() []InstallHint {
	switch runtime.GOOS {
	case "windows":
		return []InstallHint{
			{Method: "winget", Label: "winget（推荐 Win 10/11）", Command: "winget install --id Git.Git -e --source winget"},
			{Method: "scoop", Label: "Scoop", Command: "scoop install git"},
			{Method: "official", Label: "官方安装包", URL: "https://git-scm.com/download/win"},
		}
	case "darwin":
		return []InstallHint{
			{Method: "brew", Label: "Homebrew", Command: "brew install git"},
			{Method: "official", Label: "官方安装包", URL: "https://git-scm.com/download/mac"},
		}
	case "linux":
		return []InstallHint{
			{Method: "apt", Label: "Debian/Ubuntu", Command: "sudo apt-get update && sudo apt-get install -y git"},
			{Method: "yum", Label: "RHEL/CentOS", Command: "sudo yum install -y git"},
			{Method: "official", Label: "官方文档", URL: "https://git-scm.com/download/linux"},
		}
	default:
		return []InstallHint{
			{Method: "official", Label: "官方下载", URL: "https://git-scm.com/downloads"},
		}
	}
}

// friendlyErr 将常见 git 错误转中文提示。
func friendlyErr(err error) string {
	if err == nil {
		return ""
	}
	s := strings.ToLower(err.Error())
	switch {
	case strings.Contains(s, "executable file not found"):
		return "未安装 git CLI"
	case strings.Contains(s, "permission denied"):
		return "git 无执行权限"
	case strings.Contains(s, "context deadline exceeded"):
		return "git 命令超时（5s）"
	case errors.Is(err, exec.ErrNotFound):
		return "未在 PATH 中找到 git"
	}
	return err.Error()
}
