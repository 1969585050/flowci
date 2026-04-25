package gitprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Gitea 错误
var (
	ErrGiteaNotConfigured = errors.New("gitea not configured (missing baseURL/token)")
	ErrGiteaUnauthorized  = errors.New("gitea token invalid or expired")
	ErrGiteaUnreachable   = errors.New("gitea instance unreachable")
)

// GiteaClient Gitea REST API v1 客户端。
type GiteaClient struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewGitea 构造 Gitea 客户端。
// baseURL 是 Gitea 实例根（不含 /api/v1），如 "https://gitea.example.com"。
func NewGitea(baseURL, token string) *GiteaClient {
	return &GiteaClient{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		token:   strings.TrimSpace(token),
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

// Configured 是否具备最低调用条件。
func (g *GiteaClient) Configured() bool {
	return g.baseURL != "" && g.token != ""
}

// TokenSettingsURL 返回 Gitea 用户生成 token 的页面 URL，前端可一键打开。
// 形如 "{baseURL}/user/settings/applications"
func (g *GiteaClient) TokenSettingsURL() string {
	if g.baseURL == "" {
		return ""
	}
	return g.baseURL + "/user/settings/applications"
}

// Verify ping /api/v1/user 验证 token 是否有效，返回当前用户信息。
func (g *GiteaClient) Verify(ctx context.Context) (*UserInfo, error) {
	if !g.Configured() {
		return nil, ErrGiteaNotConfigured
	}
	body, err := g.get(ctx, "/api/v1/user")
	if err != nil {
		return nil, err
	}
	var raw struct {
		Login     string `json:"login"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse user response: %w", err)
	}
	return &UserInfo{
		Username:  raw.Login,
		Email:     raw.Email,
		AvatarURL: raw.AvatarURL,
	}, nil
}

// ListRepos 拉取当前 token 用户能访问的所有仓库（含个人 + 组织 + 协作）。
// 自动分页，最多 1000 条（20 页 × 50/页）。
func (g *GiteaClient) ListRepos(ctx context.Context) ([]Repo, error) {
	if !g.Configured() {
		return nil, ErrGiteaNotConfigured
	}
	const pageSize = 50
	const maxPages = 20

	all := make([]Repo, 0, pageSize)
	for page := 1; page <= maxPages; page++ {
		body, err := g.get(ctx, fmt.Sprintf("/api/v1/user/repos?limit=%d&page=%d", pageSize, page))
		if err != nil {
			return nil, err
		}
		var raw []struct {
			Name          string `json:"name"`
			FullName      string `json:"full_name"`
			CloneURL      string `json:"clone_url"`
			HTMLURL       string `json:"html_url"`
			DefaultBranch string `json:"default_branch"`
			Description   string `json:"description"`
			Private       bool   `json:"private"`
			UpdatedAt     string `json:"updated_at"`
		}
		if err := json.Unmarshal(body, &raw); err != nil {
			return nil, fmt.Errorf("parse repos response: %w", err)
		}
		for _, r := range raw {
			all = append(all, Repo{
				Name:          r.Name,
				FullName:      r.FullName,
				CloneURL:      r.CloneURL,
				HTMLURL:       r.HTMLURL,
				DefaultBranch: r.DefaultBranch,
				Description:   r.Description,
				Private:       r.Private,
				UpdatedAt:     r.UpdatedAt,
			})
		}
		if len(raw) < pageSize {
			break
		}
	}
	return all, nil
}

// CloneURLWithToken 返回带 token 的 HTTPS clone URL。
// Gitea 接受 https://oauth2:<token>@host/owner/repo.git
func (g *GiteaClient) CloneURLWithToken(cloneURL string) string {
	if g.token == "" {
		return cloneURL
	}
	u, err := url.Parse(cloneURL)
	if err != nil {
		return cloneURL
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		// SSH 不支持 token 注入，原样返回
		return cloneURL
	}
	u.User = url.UserPassword("oauth2", g.token)
	return u.String()
}

// get GET + Authorization: token <T>，统一错误转译为业务错误。
func (g *GiteaClient) get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "token "+g.token)
	req.Header.Set("Accept", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGiteaUnreachable, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, fmt.Errorf("%w: HTTP %d", ErrGiteaUnauthorized, resp.StatusCode)
	default:
		return nil, fmt.Errorf("gitea HTTP %d: %s", resp.StatusCode, truncate(string(body), 300))
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
