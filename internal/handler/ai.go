package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"flowci/internal/ai"
	"flowci/internal/secret"
	"flowci/internal/store"
)

// AI 在 keyring 与 settings 中的 key
const (
	keyringAIKey      = "ai:apiKey"
	settingAIBaseURL  = "aiBaseURL"
	settingAIModel    = "aiModel"
)

// SaveAIKey 把 API key 写入 OS keyring。空字符串视为删除。
func (a *App) SaveAIKey(req *SaveAIKeyRequest) error {
	if req == nil {
		return fmt.Errorf("%w: missing key", ErrBadRequest)
	}
	key := strings.TrimSpace(req.APIKey)
	if key == "" {
		// 删除
		_ = secret.Delete(keyringAIKey)
		return nil
	}
	return secret.Set(keyringAIKey, key)
}

// GetAIKeyStatus 仅返回 keyring 中是否已配置 key（不回传 key 本身）。
func (a *App) GetAIKeyStatus() (*AIKeyStatus, error) {
	_, err := secret.Get(keyringAIKey)
	if errors.Is(err, secret.ErrNotFound) {
		return &AIKeyStatus{Configured: false}, nil
	}
	if err != nil {
		return nil, err
	}
	return &AIKeyStatus{Configured: true}, nil
}

// DiagnoseBuild 用配置的 AI provider 分析指定构建的日志。
// 流程：读 settings (baseURL/model) + keyring (apiKey) → 调 ai.Provider.DiagnoseBuild。
func (a *App) DiagnoseBuild(req *DiagnoseBuildRequest) (*DiagnoseBuildResponse, error) {
	if req == nil || strings.TrimSpace(req.BuildID) == "" {
		return nil, fmt.Errorf("%w: buildId required", ErrBadRequest)
	}

	// 1) 拉构建记录（含 log）
	record, err := store.GetBuildRecord(req.BuildID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrBuildNotFound
		}
		return nil, err
	}
	if strings.TrimSpace(record.Log) == "" {
		return nil, fmt.Errorf("%w: build log is empty", ErrBadRequest)
	}

	// 2) 读 AI 配置（settings + keyring）
	settings, err := store.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}
	apiKey, err := secret.Get(keyringAIKey)
	if err != nil && !errors.Is(err, secret.ErrNotFound) {
		return nil, fmt.Errorf("read AI key from keyring: %w", err)
	}

	provider := ai.New(ai.Config{
		BaseURL: settings[settingAIBaseURL],
		APIKey:  apiKey,
		Model:   settings[settingAIModel],
	})
	if !provider.Configured() {
		return nil, fmt.Errorf("%w: 请先在 设置 → AI 助手 配置 baseURL 和 model", ErrBadRequest)
	}

	// 3) 调 AI（90s 超时；用 a.ctx 派生）
	ctx, cancel := context.WithTimeout(a.ctx, 90*time.Second)
	defer cancel()

	md, err := provider.DiagnoseBuild(ctx, record.Log)
	if err != nil {
		return nil, fmt.Errorf("AI 诊断失败: %w", err)
	}
	return &DiagnoseBuildResponse{
		Markdown: md,
		Model:    settings[settingAIModel],
	}, nil
}
