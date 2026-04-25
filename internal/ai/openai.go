// Package ai 封装对 OpenAI Chat Completions 兼容 API 的调用。
//
// 兼容范围：所有提供 /v1/chat/completions 端点的 provider:
//   - OpenAI (api.openai.com)
//   - DeepSeek (api.deepseek.com)
//   - 月之暗面 / Moonshot (api.moonshot.cn)
//   - Together / Fireworks / Groq 等
//   - 本地 ollama (http://localhost:11434/v1)
//
// 不支持原生 Anthropic Messages API（如需可后续扩 internal/ai/anthropic.go）。
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// 业务错误
var (
	ErrNotConfigured = errors.New("AI provider not configured (missing baseURL/apiKey/model)")
	ErrAPIFailed     = errors.New("AI API call failed")
)

// Config Provider 构造参数。
type Config struct {
	BaseURL string // 形如 "https://api.openai.com" 或 "http://localhost:11434"
	APIKey  string // API key；空时不发 Authorization header（本地 ollama 场景）
	Model   string // 模型 ID 例如 "gpt-4o-mini" / "deepseek-chat" / "llama3"
}

// Provider 是 OpenAI Chat Completions 兼容客户端。
type Provider struct {
	cfg    Config
	client *http.Client
}

// New 构造 Provider。HTTP client 默认 60s 超时。
func New(cfg Config) *Provider {
	return &Provider{
		cfg:    cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// Configured 是否具备最低可调用条件。
func (p *Provider) Configured() bool {
	return strings.TrimSpace(p.cfg.BaseURL) != "" && strings.TrimSpace(p.cfg.Model) != ""
}

// Chat 发起一次对话；system 可空。
// 返回 assistant 的 message.content 文本。
func (p *Provider) Chat(ctx context.Context, system, user string) (string, error) {
	if !p.Configured() {
		return "", ErrNotConfigured
	}

	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	messages := make([]msg, 0, 2)
	if strings.TrimSpace(system) != "" {
		messages = append(messages, msg{Role: "system", Content: system})
	}
	messages = append(messages, msg{Role: "user", Content: user})

	body, err := json.Marshal(map[string]any{
		"model":    p.cfg.Model,
		"messages": messages,
		"stream":   false,
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := strings.TrimRight(p.cfg.BaseURL, "/") + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if p.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPIFailed, err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("%w: HTTP %d: %s", ErrAPIFailed, resp.StatusCode, truncate(string(respBytes), 500))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return "", fmt.Errorf("parse response: %w (body: %s)", err, truncate(string(respBytes), 200))
	}
	if len(parsed.Choices) == 0 || parsed.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("%w: empty response", ErrAPIFailed)
	}
	return parsed.Choices[0].Message.Content, nil
}

// truncate 截断到 max 个字节，超长加 "...(N more bytes)" 后缀。
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + fmt.Sprintf("...(%d more bytes)", len(s)-max)
}
