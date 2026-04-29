package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// 操作超时（git 操作可能慢，timeout 大方点）
const (
	cloneTimeout = 10 * time.Minute
	pullTimeout  = 5 * time.Minute
	statusQueryTimeout = 10 * time.Second
)

// CloneRequest 克隆参数。
//   - URL  必须含凭证（如 https://oauth2:<token>@host/...），本包不做 URL 改写
//   - Branch 空则用远端默认分支
//   - Dest 必须是不存在或空的目录
type CloneRequest struct {
	URL    string
	Branch string
	Dest   string
}

// Clone 执行 git clone（浅克隆，depth=1 提速）。
// 失败时 dest 目录可能残留半成品，调用方应清理。
func Clone(ctx context.Context, req CloneRequest) error {
	if strings.TrimSpace(req.URL) == "" {
		return fmt.Errorf("clone: URL required")
	}
	if strings.TrimSpace(req.Dest) == "" {
		return fmt.Errorf("clone: Dest required")
	}
	if err := os.MkdirAll(parentDir(req.Dest), 0755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	args := []string{"clone", "--depth=1"}
	if strings.TrimSpace(req.Branch) != "" {
		args = append(args, "--branch", req.Branch)
	}
	args = append(args, req.URL, req.Dest)

	ctxTO, cancel := context.WithTimeout(ctx, cloneTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctxTO, "git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone: %w: %s", err, sanitize(string(out), req.URL))
	}
	return nil
}

// Pull 执行 git -C <dir> pull。url 可选：非空时用作"重置 origin"避免 token 变化。
func Pull(ctx context.Context, dir, url string) error {
	if strings.TrimSpace(dir) == "" {
		return fmt.Errorf("pull: dir required")
	}

	// 如有新 url（token 可能更新过），先 set-url
	if url != "" {
		ctxQ, cancel := context.WithTimeout(ctx, statusQueryTimeout)
		setURL := exec.CommandContext(ctxQ, "git", "-C", dir, "remote", "set-url", "origin", url)
		_, _ = setURL.CombinedOutput() // best-effort，错误不致命
		cancel()
	}

	ctxTO, cancel := context.WithTimeout(ctx, pullTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctxTO, "git", "-C", dir, "pull", "--ff-only")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull: %w: %s", err, sanitize(string(out), url))
	}
	return nil
}

// HeadCommit 返回 dir 当前 HEAD 的短 SHA + 提交标题（best-effort）。
func HeadCommit(ctx context.Context, dir string) (sha, subject string, err error) {
	if dir == "" {
		return "", "", fmt.Errorf("head: dir required")
	}
	ctxTO, cancel := context.WithTimeout(ctx, statusQueryTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctxTO, "git", "-C", dir, "log", "-1", "--pretty=%h%n%s")
	out, err := cmd.Output()
	if err != nil {
		return "", "", err
	}
	parts := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)
	if len(parts) >= 1 {
		sha = parts[0]
	}
	if len(parts) >= 2 {
		subject = parts[1]
	}
	return sha, subject, nil
}

var reTokenInURL = regexp.MustCompile(`(https?://)oauth2:[^@]*(@)`)

// sanitize 把 URL 里的 token 遮蔽掉，避免泄漏到错误信息。
// 使用正则匹配 `scheme://oauth2:<token>@`，不依赖精确字符串替换，
// 能正确处理 git 输出中的引号、尾部斜杠等变体。
func sanitize(msg, url string) string {
	if url == "" {
		return msg
	}
	if !strings.Contains(url, "@") || !strings.Contains(url, "//") {
		return msg
	}
	return reTokenInURL.ReplaceAllString(msg, "${1}oauth2:***${2}")
}

func parentDir(p string) string {
	if i := strings.LastIndexAny(p, "/\\"); i > 0 {
		return p[:i]
	}
	return "."
}
