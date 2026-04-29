package handler

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"flowci/internal/store"
)

// settings key 常量
const (
	settingWindowAlwaysOnTop = "windowAlwaysOnTop"
	settingWindowX           = "windowX"
	settingWindowY           = "windowY"
	settingWindowWidth       = "windowWidth"
	settingWindowHeight      = "windowHeight"
	settingWindowMaximised   = "windowMaximised"
)

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

// SetWindowSize 设置窗口大小（宽 × 高），立即生效并持久化到 settings。
func (a *App) SetWindowSize(w, h int) error {
	if w <= 0 || h <= 0 {
		return fmt.Errorf("invalid window size: %d×%d", w, h)
	}
	runtime.WindowSetSize(a.ctx, w, h)
	if err := store.SaveSettings(settingWindowWidth, fmt.Sprintf("%d", w)); err != nil {
		return err
	}
	return store.SaveSettings(settingWindowHeight, fmt.Sprintf("%d", h))
}

// WindowSize 是 GetWindowSize 的返回值。
type WindowSize struct {
	W int `json:"w"`
	H int `json:"h"`
}

// GetWindowSize 返回当前窗口实时宽高（来自 runtime，非 settings）。
func (a *App) GetWindowSize() WindowSize {
	w, h := runtime.WindowGetSize(a.ctx)
	return WindowSize{W: w, H: h}
}

// RestartApp 重启应用：dev 模式下直接退出，production 模式下启动新进程后退出。
func (a *App) RestartApp() error {
	if runtime.Environment(a.ctx).BuildType == "dev" {
		runtime.Quit(a.ctx)
		return nil
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	runtime.Quit(a.ctx)
	return nil
}
