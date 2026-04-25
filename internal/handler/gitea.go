package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"flowci/internal/git"
	"flowci/internal/gitprovider"
	"flowci/internal/secret"
	"flowci/internal/store"
)

// settings / keyring 中的固定 key
const (
	settingGiteaBaseURL = "giteaBaseURL"
	keyringGiteaToken   = "gitprovider:gitea:token"
)

// SaveGiteaConfig 保存 Gitea baseURL（settings 表）和 token（keyring）。
// req.Token 为空且 baseURL 改了：保留旧 token；req.Token 全空格视为清除。
func (a *App) SaveGiteaConfig(req *SaveGiteaConfigRequest) error {
	if req == nil {
		return fmt.Errorf("%w: missing config", ErrBadRequest)
	}
	baseURL := strings.TrimSpace(req.BaseURL)
	if err := store.SaveSettings(settingGiteaBaseURL, baseURL); err != nil {
		return err
	}
	// token 处理：
	//   - req.Token == ""  → 不修改（前端"留空保留旧值"语义）
	//   - 其他              → 写 keyring（trim 空格视为删除）
	if req.Token != "" {
		token := strings.TrimSpace(req.Token)
		if token == "" {
			_ = secret.Delete(keyringGiteaToken)
		} else if err := secret.Set(keyringGiteaToken, token); err != nil {
			return fmt.Errorf("save token: %w", err)
		}
	}
	return nil
}

// GetGiteaStatus 返回当前 Gitea 配置概况。永不暴露 token 本身。
func (a *App) GetGiteaStatus() (*GiteaStatusResponse, error) {
	settings, err := store.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}
	baseURL := settings[settingGiteaBaseURL]

	hasToken := false
	if _, err := secret.Get(keyringGiteaToken); err == nil {
		hasToken = true
	} else if !errors.Is(err, secret.ErrNotFound) {
		slog.Warn("read gitea token failed", "err", err)
	}

	tokenURL := ""
	if baseURL != "" {
		tokenURL = strings.TrimRight(baseURL, "/") + "/user/settings/applications"
	}

	return &GiteaStatusResponse{
		BaseURL:          baseURL,
		HasToken:         hasToken,
		TokenSettingsURL: tokenURL,
	}, nil
}

// VerifyGitea 调 Gitea /api/v1/user 验证当前配置。
func (a *App) VerifyGitea() (*gitprovider.UserInfo, error) {
	client, err := a.giteaClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(a.ctx, 15*time.Second)
	defer cancel()
	return client.Verify(ctx)
}

// ListGiteaRepos 拉取当前 token 用户能访问的所有仓库。
func (a *App) ListGiteaRepos() ([]gitprovider.Repo, error) {
	client, err := a.giteaClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(a.ctx, 60*time.Second)
	defer cancel()
	return client.ListRepos(ctx)
}

// ImportGiteaRepos 把用户勾选的仓库创建为 FlowCI projects（自动 clone 到 dataDir/repos/<id>）。
// 串行执行；某条失败不影响其他；返回详细成功/失败列表。
func (a *App) ImportGiteaRepos(req *ImportGiteaReposRequest) (*ImportGiteaReposResponse, error) {
	if req == nil || len(req.Repos) == 0 {
		return nil, fmt.Errorf("%w: at least one repo required", ErrBadRequest)
	}
	client, err := a.giteaClient()
	if err != nil {
		return nil, err
	}

	resp := &ImportGiteaReposResponse{}
	for _, sel := range req.Repos {
		project, ierr := a.importOneGiteaRepo(client, sel)
		if ierr != nil {
			resp.Errors = append(resp.Errors, ImportError{
				FullName: sel.FullName,
				Error:    ierr.Error(),
			})
			continue
		}
		resp.Imported = append(resp.Imported, project)
	}
	return resp, nil
}

// importOneGiteaRepo 处理单个仓库的导入。失败时回滚（删 store 记录 + 清 dest 目录）。
func (a *App) importOneGiteaRepo(client *gitprovider.GiteaClient, sel ImportGiteaRepo) (*store.Project, error) {
	if strings.TrimSpace(sel.CloneURL) == "" {
		return nil, fmt.Errorf("missing cloneURL for %q", sel.FullName)
	}
	name := strings.TrimSpace(sel.Name)
	if name == "" {
		// owner/repo → repo
		if i := strings.LastIndex(sel.FullName, "/"); i >= 0 {
			name = sel.FullName[i+1:]
		} else {
			name = sel.FullName
		}
	}

	// 先建 store 记录（拿到 ID 才能确定 clone 路径）
	p, err := store.CreateProject(store.CreateProjectInput{
		Name:       name,
		Language:   sel.Language,
		RepoURL:    sel.CloneURL,
		RepoBranch: sel.Branch,
	})
	if err != nil {
		return nil, fmt.Errorf("create project record: %w", err)
	}

	dest := a.repoCloneDir(p.ID)
	// 兜底：如果 dest 已存在（重复导入），先清掉
	_ = os.RemoveAll(dest)

	// 把 token 注入 clone URL
	urlWithToken := client.CloneURLWithToken(sel.CloneURL)

	ctx, cancel := context.WithTimeout(a.ctx, 10*time.Minute)
	defer cancel()
	if err := git.Clone(ctx, git.CloneRequest{
		URL:    urlWithToken,
		Branch: sel.Branch,
		Dest:   dest,
	}); err != nil {
		// 回滚
		_ = store.DeleteProject(p.ID)
		_ = os.RemoveAll(dest)
		return nil, fmt.Errorf("git clone: %w", err)
	}

	// 写回 path 和 lastPullAt
	if _, err := store.UpdateProject(p.ID, store.UpdateProjectInput{
		Name:       p.Name,
		Path:       dest,
		Language:   p.Language,
		RepoURL:    p.RepoURL,
		RepoBranch: p.RepoBranch,
	}); err != nil {
		slog.Warn("update project path after clone failed", "id", p.ID, "err", err)
	}
	_ = store.MarkProjectPulled(p.ID)

	updated, err := store.GetProject(p.ID)
	if err != nil {
		return &p, nil
	}
	return &updated, nil
}

// PullProjectRepo 对 Git 项目执行 git pull。
// 项目必须有 RepoURL；token 从 keyring 取并注入 URL。
func (a *App) PullProjectRepo(projectID string) error {
	if strings.TrimSpace(projectID) == "" {
		return fmt.Errorf("%w: projectId required", ErrBadRequest)
	}
	p, err := store.GetProject(projectID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrProjectNotFound
		}
		return err
	}
	if p.RepoURL == "" {
		return fmt.Errorf("%w: project has no repo URL (not a Git project)", ErrBadRequest)
	}
	if p.Path == "" {
		return fmt.Errorf("%w: project has no local path", ErrBadRequest)
	}

	// token 注入（如有）
	urlWithToken := p.RepoURL
	if token, terr := secret.Get(keyringGiteaToken); terr == nil {
		client := gitprovider.NewGitea("https://placeholder", token) // baseURL 不参与 URL 改写
		urlWithToken = client.CloneURLWithToken(p.RepoURL)
	}

	ctx, cancel := context.WithTimeout(a.ctx, 5*time.Minute)
	defer cancel()
	if err := git.Pull(ctx, p.Path, urlWithToken); err != nil {
		return fmt.Errorf("git pull: %w", err)
	}
	return store.MarkProjectPulled(p.ID)
}

// repoCloneDir 是 git 项目的本地 clone 目录约定：<dataDir>/repos/<projectID>
func (a *App) repoCloneDir(projectID string) string {
	return filepath.Join(a.dataDir, "repos", projectID)
}

// giteaClient 工厂：从 settings + keyring 拼装 client；缺一不可。
func (a *App) giteaClient() (*gitprovider.GiteaClient, error) {
	settings, err := store.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}
	baseURL := settings[settingGiteaBaseURL]
	token, err := secret.Get(keyringGiteaToken)
	if err != nil {
		if errors.Is(err, secret.ErrNotFound) {
			return nil, fmt.Errorf("%w: 请先在 设置 → Gitea 集成 配置 baseURL 和 token", ErrBadRequest)
		}
		return nil, fmt.Errorf("read token from keyring: %w", err)
	}
	client := gitprovider.NewGitea(baseURL, token)
	if !client.Configured() {
		return nil, fmt.Errorf("%w: gitea baseURL 或 token 缺失", ErrBadRequest)
	}
	return client, nil
}
