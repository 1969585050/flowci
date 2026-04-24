package handler

import (
	"context"
	"log/slog"
	"path/filepath"

	"flowci/internal/docker"
	"flowci/internal/pipeline"
	"flowci/internal/store"
)

// App 是 Wails Bind 的主入口，持有跨请求的状态（executor 的 per-pipeline 锁）。
type App struct {
	ctx      context.Context
	dataDir  string
	executor *pipeline.Executor
}

// NewApp 构造 App 实例。dataDir 来自 config.DataDir()，用于 store.Init 和 build log 落盘。
func NewApp(dataDir string) *App {
	buildLogsDir := filepath.Join(dataDir, "logs", "builds")
	return &App{
		dataDir:  dataDir,
		executor: pipeline.NewExecutor(buildLogsDir),
	}
}

// Startup 实现 wails options.App.OnStartup，由 wails runtime 回调。
// 此时 context 已可用；store 在这里初始化。
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	slog.Info("application starting")
	if err := store.Init(a.dataDir); err != nil {
		slog.Error("initialize store failed", "err", err)
	}
	slog.Info("data directory", "path", a.dataDir)
}

// Shutdown 实现 wails options.App.OnShutdown，释放资源。
func (a *App) Shutdown(ctx context.Context) {
	store.Close()
	slog.Info("application stopped")
}

// CheckDocker 探测本机 docker daemon 连通性。
// 永不返回 error：连不上时返回 Status{Connected: false}。
func (a *App) CheckDocker(ctx context.Context) docker.Status {
	return docker.Check(ctx)
}
