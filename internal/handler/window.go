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
