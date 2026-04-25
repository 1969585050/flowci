// Package gitprovider 抽象 Git 托管平台 (Gitea / GitHub / GitLab) 的 REST API。
//
// MVP 只实现 Gitea (gitea.go)；后续按 Provider 接口扩。
//
// 与 internal/git 的区别：
//   - internal/git    - 调本机 git CLI (clone / pull)
//   - internal/gitprovider - 调远程 platform REST API (列仓库、Verify token、webhook 等)
package gitprovider

// UserInfo 已认证用户的基本信息。
type UserInfo struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarUrl"`
}

// Repo 一条仓库元数据。
// CloneURL 通常是 HTTPS（用户后续 clone 时会注入 token）。
type Repo struct {
	Name          string `json:"name"`          // repo 短名
	FullName      string `json:"fullName"`      // owner/repo
	CloneURL      string `json:"cloneUrl"`      // https://gitea.example.com/owner/repo.git
	HTMLURL       string `json:"htmlUrl"`       // 浏览器打开页面
	DefaultBranch string `json:"defaultBranch"` // 通常是 main / master
	Description   string `json:"description"`
	Private       bool   `json:"private"`
	UpdatedAt     string `json:"updatedAt"` // ISO 8601 字符串
}
