// Package logger 全局 slog 初始化。
//
// 输出: logDir/flowci-YYYY-MM-DD.log 文件 + stderr (多路)
// 级别: 默认 INFO，可通过环境变量 FLOWCI_LOG_LEVEL=DEBUG|INFO|WARN|ERROR 调整
//
// 阶段 1 不做按大小切割，仅按日期命名文件；阶段 3 视情况引入 lumberjack。
package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Init 初始化全局 slog logger，将输出挂到 slog.Default()。
// logDir 不存在则尝试创建；创建失败时退化为仅 stderr 输出并返回错误。
func Init(logDir string) (io.Closer, error) {
	level := parseLevel(os.Getenv("FLOWCI_LOG_LEVEL"))

	var writers []io.Writer
	writers = append(writers, os.Stderr)

	var file *os.File
	if logDir != "" {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
			return nopCloser{}, fmt.Errorf("create log dir: %w", err)
		}
		path := filepath.Join(logDir, fmt.Sprintf("flowci-%s.log", time.Now().Format("2006-01-02")))
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
			return nopCloser{}, fmt.Errorf("open log file: %w", err)
		}
		file = f
		writers = append(writers, f)
	}

	handler := slog.NewTextHandler(io.MultiWriter(writers...), &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	})
	slog.SetDefault(slog.New(handler))

	if file != nil {
		return file, nil
	}
	return nopCloser{}, nil
}

func parseLevel(s string) slog.Level {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type nopCloser struct{}

func (nopCloser) Close() error { return nil }
