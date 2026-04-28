package handler

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"flowci/internal/docker"
	"flowci/internal/pipeline"
	"flowci/internal/store"
)

// App 是 Wails Bind 的主入口，持有跨请求的状态（executor 的 per-pipeline 锁 + docker 客户端）。
// docker 字段是接口，测试可通过 NewAppWithClient 注入 fake。
type App struct {
	ctx      context.Context
	dataDir  string
	executor *pipeline.Executor
	docker   docker.Client
}

// NewApp 构造 App 实例，内部用默认 docker.NewClient()。
// dataDir 来自 config.DataDir()，用于 store.Init 和 build log 落盘。
func NewApp(dataDir string) *App {
	return NewAppWithClient(dataDir, docker.NewClient())
}

// NewAppWithClient 同 NewApp，但允许注入自定义 docker.Client（测试用）。
func NewAppWithClient(dataDir string, client docker.Client) *App {
	buildLogsDir := filepath.Join(dataDir, "logs", "builds")
	return &App{
		dataDir:  dataDir,
		executor: pipeline.NewExecutorWithClient(buildLogsDir, client),
		docker:   client,
	}
}

// Startup 实现 wails options.App.OnStartup，由 wails runtime 回调。
// 此时 context 已可用；store 在这里初始化；运行时配置（如 dockerHost）也在此加载。
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	slog.Info("application starting")
	if err := store.Init(a.dataDir); err != nil {
		slog.Error("initialize store failed", "err", err)
		return
	}
	slog.Info("data directory", "path", a.dataDir)

	// 从 settings 表加载运行时配置
	if settings, err := store.GetSettings(); err == nil {
		if host := settings["dockerHost"]; host != "" {
			docker.SetDockerHost(host)
			slog.Info("docker host from settings", "host", host)
		}
		// 恢复"窗口始终置顶"状态
		if settings[settingWindowAlwaysOnTop] == "true" {
			wruntime.WindowSetAlwaysOnTop(ctx, true)
			slog.Info("window always on top restored")
		}
		// 恢复窗口最大化状态
		if settings[settingWindowMaximised] == "true" {
			wruntime.WindowMaximise(ctx)
			slog.Info("window maximised restored")
		} else {
			// 恢复窗口位置和大小
			if x, errX := strconv.Atoi(settings[settingWindowX]); errX == nil {
				if y, errY := strconv.Atoi(settings[settingWindowY]); errY == nil {
					wruntime.WindowSetPosition(ctx, x, y)
					slog.Info("window position restored", "x", x, "y", y)
				}
			}
			if w, errW := strconv.Atoi(settings[settingWindowWidth]); errW == nil && w > 0 {
				if h, errH := strconv.Atoi(settings[settingWindowHeight]); errH == nil && h > 0 {
					wruntime.WindowSetSize(ctx, w, h)
					slog.Info("window size restored", "w", w, "h", h)
				}
			}
		}
	}
}

// Shutdown 实现 wails options.App.OnShutdown，保存窗口状态后释放资源。
func (a *App) Shutdown(ctx context.Context) {
	_ = ctx

	if a.ctx != nil {
		x, y := wruntime.WindowGetPosition(a.ctx)
		w, h := wruntime.WindowGetSize(a.ctx)
		isMax := wruntime.WindowIsMaximised(a.ctx)

		pairs := map[string]string{
			settingWindowX:         strconv.Itoa(x),
			settingWindowY:         strconv.Itoa(y),
			settingWindowWidth:     strconv.Itoa(w),
			settingWindowHeight:    strconv.Itoa(h),
			settingWindowMaximised: fmt.Sprintf("%t", isMax),
		}
		for k, v := range pairs {
			if err := store.SaveSettings(k, v); err != nil {
				slog.Warn("failed to save window state", "key", k, "err", err)
			}
		}
	}

	store.Close()
	slog.Info("application stopped")
}

// CheckDocker 探测本机 docker daemon 连通性。
// 永不返回 error：连不上时返回 Status{Connected: false}。
func (a *App) CheckDocker() docker.Status {
	return a.docker.Check(a.ctx)
}
