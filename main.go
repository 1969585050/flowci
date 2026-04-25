// Package main 是 FlowCI 的入口；仅负责启动 Wails、初始化日志、构造 handler.App 并注册 Bind。
//
// 业务方法全部迁到 internal/handler 包下；main.go 不应再承担任何 CRUD / exec / 持久化逻辑。
// 受 backend-spec.md § 1 约束，本文件行数必须 ≤ 80（不计空行与注释）。
package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"flowci/internal/config"
	"flowci/internal/handler"
	"flowci/internal/logger"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	dataDir := config.DataDir()
	if _, err := logger.Init(config.LogDir()); err != nil {
		fmt.Fprintln(os.Stderr, "logger init warning:", err)
	}

	app := handler.NewApp(dataDir)

	err := wails.Run(&options.App{
		Title:  "FlowCI",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		// dev 模式频繁 rebuild 时窗口会反复抢焦盖住 IDE；
		// 启动时最小化到任务栏，需要时手动展开。
		WindowStartState: options.Minimised,
		OnStartup:        func(ctx context.Context) { app.Startup(ctx) },
		OnShutdown:       func(ctx context.Context) { app.Shutdown(ctx) },
		Bind:             []interface{}{app},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: false,
		},
	})
	if err != nil {
		slog.Error("wails run failed", "err", err)
	}
}
