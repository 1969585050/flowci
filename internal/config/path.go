// Package config 统一管理运行时路径与常量。
//
// 所有访问本地文件系统的代码（store、logger、docker compose 临时文件等）
// 都应通过本包取路径，避免各处硬编码 %APPDATA% 或 ~/.local/share。
package config

import (
	"os"
	"path/filepath"
)

// AppName 应用名；作为数据目录最后一级子目录。
const AppName = "FlowCI"

// DataDir 返回应用数据根目录：
//   - Windows：%APPDATA%\FlowCI
//   - 其它平台：$XDG_DATA_HOME/FlowCI 优先，回退到 ~/.local/share/FlowCI
//
// 目录不保证存在，调用方按需 os.MkdirAll。
func DataDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, AppName)
	}
	if appData := os.Getenv("APPDATA"); appData != "" {
		return filepath.Join(appData, AppName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", AppName)
}

// LogDir 应用日志目录（slog 文件落地）。
func LogDir() string {
	return filepath.Join(DataDir(), "logs")
}

// BuildLogsDir 构建日志子目录（阶段 3 build log 从 DB 迁移到文件时使用）。
func BuildLogsDir() string {
	return filepath.Join(LogDir(), "builds")
}

// ComposeTmpDir docker-compose 临时文件目录。
func ComposeTmpDir() string {
	return filepath.Join(DataDir(), "tmp")
}
