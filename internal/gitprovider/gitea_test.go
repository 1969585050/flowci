package gitprovider

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGiteaConfigured(t *testing.T) {
	if NewGitea("", "").Configured() {
		t.Error("empty config should not be configured")
	}
	if NewGitea("https://x", "").Configured() {
		t.Error("missing token should not be configured")
	}
	if !NewGitea("https://x", "tok").Configured() {
		t.Error("expected configured")
	}
}

func TestGiteaTokenSettingsURL(t *testing.T) {
	g := NewGitea("https://gitea.example.com/", "tok")
	want := "https://gitea.example.com/user/settings/applications"
	if got := g.TokenSettingsURL(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestVerify_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "token testtok" {
			t.Errorf("auth = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/user/repos":
			_, _ = w.Write([]byte(`[]`))
		case "/api/v1/user":
			_, _ = w.Write([]byte(`{"login":"alice","email":"a@b","avatar_url":"https://x/a.png"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	user, err := NewGitea(srv.URL, "testtok").Verify(context.Background())
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if user.Username != "alice" || user.Email != "a@b" {
		t.Errorf("got %+v", user)
	}
}

// 主验证 (read:repository) 通过、user 接口 403 (缺 read:user) → 仍算成功，username 留空。
func TestVerify_NoUserScopeFallsThrough(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/user/repos":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[]`))
		case "/api/v1/user":
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"message":"missing scope read:user"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	user, err := NewGitea(srv.URL, "tok").Verify(context.Background())
	if err != nil {
		t.Fatalf("Verify should not fail without read:user scope: %v", err)
	}
	if user.Username != "" {
		t.Errorf("expected empty username, got %q", user.Username)
	}
}

func TestVerify_NotConfigured(t *testing.T) {
	_, err := NewGitea("", "").Verify(context.Background())
	if !errors.Is(err, ErrGiteaNotConfigured) {
		t.Errorf("expected ErrGiteaNotConfigured, got %v", err)
	}
}

func TestVerify_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"token revoked"}`))
	}))
	defer srv.Close()

	_, err := NewGitea(srv.URL, "bad").Verify(context.Background())
	if !errors.Is(err, ErrGiteaUnauthorized) {
		t.Errorf("expected ErrGiteaUnauthorized, got %v", err)
	}
}

// 主验证 (列仓库) 收到 403 应当算 ErrGiteaUnauthorized（缺 read:repository）。
func TestVerify_NoRepoScope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/user/repos" {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"message":"missing scope read:repository"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := NewGitea(srv.URL, "tok").Verify(context.Background())
	if !errors.Is(err, ErrGiteaUnauthorized) {
		t.Errorf("expected ErrGiteaUnauthorized, got %v", err)
	}
}

func TestListRepos_PaginatedAndFlat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case "1":
			// 50 条满页（提示还有下一页）
			items := strings.Repeat(`{"name":"r","full_name":"o/r","clone_url":"u","html_url":"h","default_branch":"main","description":"","private":false,"updated_at":"2026-01-01T00:00:00Z"},`, 50)
			items = strings.TrimRight(items, ",")
			_, _ = w.Write([]byte("[" + items + "]"))
		case "2":
			// 1 条不满页（结束）
			_, _ = w.Write([]byte(`[{"name":"last","full_name":"o/last","clone_url":"u","html_url":"h","default_branch":"dev","description":"d","private":true,"updated_at":"2026-01-02T00:00:00Z"}]`))
		default:
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	g := NewGitea(srv.URL, "tok")
	repos, err := g.ListRepos(context.Background())
	if err != nil {
		t.Fatalf("ListRepos: %v", err)
	}
	if len(repos) != 51 {
		t.Errorf("expected 51 repos (50 page1 + 1 page2), got %d", len(repos))
	}
	if repos[50].FullName != "o/last" || repos[50].DefaultBranch != "dev" {
		t.Errorf("last repo wrong: %+v", repos[50])
	}
}

func TestCloneURLWithToken(t *testing.T) {
	g := NewGitea("https://x", "secret")

	cases := []struct {
		in   string
		want string
	}{
		{
			in:   "https://gitea.example.com/owner/repo.git",
			want: "https://oauth2:secret@gitea.example.com/owner/repo.git",
		},
		{
			in:   "ssh://git@gitea.example.com:222/owner/repo.git",
			want: "ssh://git@gitea.example.com:222/owner/repo.git",
		},
	}
	for _, tc := range cases {
		if got := g.CloneURLWithToken(tc.in); got != tc.want {
			t.Errorf("CloneURLWithToken(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestCloneURLWithToken_NoToken(t *testing.T) {
	g := NewGitea("https://x", "")
	in := "https://gitea.example.com/o/r.git"
	if got := g.CloneURLWithToken(in); got != in {
		t.Errorf("expected unchanged, got %q", got)
	}
}
