package ai

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProvider_Configured(t *testing.T) {
	p := New(Config{})
	if p.Configured() {
		t.Error("empty config should not be configured")
	}
	p2 := New(Config{BaseURL: "http://x", Model: "m"})
	if !p2.Configured() {
		t.Error("expected configured")
	}
	// APIKey 不是必填（本地 ollama）
}

func TestChat_NotConfigured(t *testing.T) {
	p := New(Config{})
	_, err := p.Chat(context.Background(), "", "hi")
	if !errors.Is(err, ErrNotConfigured) {
		t.Errorf("expected ErrNotConfigured, got %v", err)
	}
}

func TestChat_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 校验 URL/method/auth header/请求体
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %q, want /v1/chat/completions", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("auth = %q, want 'Bearer test-key'", got)
		}

		body, _ := io.ReadAll(r.Body)
		var parsed struct {
			Model    string                   `json:"model"`
			Messages []map[string]string      `json:"messages"`
			Stream   bool                     `json:"stream"`
		}
		_ = json.Unmarshal(body, &parsed)
		if parsed.Model != "test-model" {
			t.Errorf("model = %q", parsed.Model)
		}
		if len(parsed.Messages) != 2 {
			t.Errorf("messages len = %d, want 2 (system+user)", len(parsed.Messages))
		}
		if parsed.Messages[0]["role"] != "system" || parsed.Messages[1]["role"] != "user" {
			t.Errorf("messages role mismatch: %+v", parsed.Messages)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"诊断结果"}}]}`))
	}))
	defer srv.Close()

	p := New(Config{BaseURL: srv.URL, APIKey: "test-key", Model: "test-model"})
	out, err := p.Chat(context.Background(), "system prompt", "user msg")
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if out != "诊断结果" {
		t.Errorf("got %q, want 诊断结果", out)
	}
}

func TestChat_NoSystemMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var parsed struct {
			Messages []map[string]string `json:"messages"`
		}
		_ = json.Unmarshal(body, &parsed)
		if len(parsed.Messages) != 1 || parsed.Messages[0]["role"] != "user" {
			t.Errorf("expected only user message, got %+v", parsed.Messages)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer srv.Close()

	p := New(Config{BaseURL: srv.URL, Model: "m"})
	_, err := p.Chat(context.Background(), "", "u")
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
}

func TestChat_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer srv.Close()

	p := New(Config{BaseURL: srv.URL, Model: "m"})
	_, err := p.Chat(context.Background(), "", "u")
	if !errors.Is(err, ErrAPIFailed) {
		t.Errorf("expected ErrAPIFailed, got %v", err)
	}
	if !strings.Contains(err.Error(), "HTTP 400") {
		t.Errorf("error should mention HTTP 400, got %v", err)
	}
}

func TestChat_EmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer srv.Close()

	p := New(Config{BaseURL: srv.URL, Model: "m"})
	_, err := p.Chat(context.Background(), "", "u")
	if !errors.Is(err, ErrAPIFailed) {
		t.Errorf("expected ErrAPIFailed, got %v", err)
	}
}

func TestDiagnoseBuild_TruncatesLongLog(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// 校验 prompt 中含有截断标记
		if !strings.Contains(string(body), "已截断前") {
			t.Error("expected '已截断前' marker in user message")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer srv.Close()

	p := New(Config{BaseURL: srv.URL, Model: "m"})
	bigLog := strings.Repeat("a", 50*1024) // 50KB > 30KB limit
	_, err := p.DiagnoseBuild(context.Background(), bigLog)
	if err != nil {
		t.Fatalf("DiagnoseBuild: %v", err)
	}
}
