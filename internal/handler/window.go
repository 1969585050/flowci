package handler

import (
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"flowci/internal/store"
)

// settings key 常量
const settingWindowAlwaysOnTop = "windowAlwaysOnTop"

// SetWindowAlwaysOnTop 切换 Wails 窗口的"始终置顶"状态。
// 同时把状态写入 settings 表，下次启动时 OnStartup 自动恢复。
func (a *App) SetWindowAlwaysOnTop(on bool) error {
	runtime.WindowSetAlwaysOnTop(a.ctx, on)
	v := "false"
	if on {
		v = "true"
	}
	return store.SaveSettings(settingWindowAlwaysOnTop, v)
}

// GetWindowAlwaysOnTop 返回当前窗口置顶设置（来自 settings 表）。
func (a *App) GetWindowAlwaysOnTop() bool {
	settings, err := store.GetSettings()
	if err != nil {
		return false
	}
	return settings[settingWindowAlwaysOnTop] == "true"
}

// ---- 窗口控制（自绘标题栏调用） ----

// WindowMinimise 最小化到任务栏。
func (a *App) WindowMinimise() {
	runtime.WindowMinimise(a.ctx)
}

// WindowToggleMaximise 最大化 ⇄ 还原。
func (a *App) WindowToggleMaximise() {
	if runtime.WindowIsMaximised(a.ctx) {
		runtime.WindowUnmaximise(a.ctx)
	} else {
		runtime.WindowMaximise(a.ctx)
	}
}

// WindowIsMaximised 当前是否已最大化（前端切换图标 ⬜/❐ 用）。
func (a *App) WindowIsMaximised() bool {
	return runtime.WindowIsMaximised(a.ctx)
}

// QuitApp 关闭应用（标题栏 ✕ 按钮）。
func (a *App) QuitApp() {
	runtime.Quit(a.ctx)
}
