package handler

import "flowci/internal/git"

// DetectGitEnv 检查本机 git 安装状态；不连远程，仅本地命令探测。
// 永不返回 error：未装时 Installed=false + Message + InstallHints。
func (a *App) DetectGitEnv() git.EnvReport {
	return git.DetectEnv(a.ctx)
}
