package validate

import (
	"strings"
	"testing"
)

func TestContainerName(t *testing.T) {
	ok := []string{"web", "my-app", "app_01", "A.B.C", "a1"}
	bad := []string{"", "_leading", "-leading", ".leading", "a b", "toolong" + strings.Repeat("x", 60)}

	for _, s := range ok {
		if err := ContainerName(s); err != nil {
			t.Errorf("expected %q valid, got %v", s, err)
		}
	}
	for _, s := range bad {
		if err := ContainerName(s); err == nil {
			t.Errorf("expected %q invalid", s)
		}
	}
}

func TestImageRef(t *testing.T) {
	ok := []string{"nginx", "nginx:latest", "nginx:1.25", "my/app", "my/app:v1", "foo/bar/baz:tag"}
	bad := []string{"", "Nginx", "nginx:", "my//app", "my/app:bad tag"}

	for _, s := range ok {
		if err := ImageRef(s); err != nil {
			t.Errorf("expected %q valid, got %v", s, err)
		}
	}
	for _, s := range bad {
		if err := ImageRef(s); err == nil {
			t.Errorf("expected %q invalid", s)
		}
	}
}

func TestPort(t *testing.T) {
	if err := Port(""); err != nil {
		t.Errorf("empty port should be allowed, got %v", err)
	}
	for _, s := range []string{"1", "80", "65535"} {
		if err := Port(s); err != nil {
			t.Errorf("expected %q valid, got %v", s, err)
		}
	}
	for _, s := range []string{"0", "65536", "-1", "abc", "8080a"} {
		if err := Port(s); err == nil {
			t.Errorf("expected %q invalid", s)
		}
	}
}

func TestRegistryHost(t *testing.T) {
	for _, s := range []string{"", "docker.io", "my-registry.local", "registry.example.com:5000", "192.168.1.10:5000"} {
		if err := RegistryHost(s); err != nil {
			t.Errorf("expected %q valid, got %v", s, err)
		}
	}
	for _, s := range []string{"https://example.com", "example.com/path", "-leading"} {
		if err := RegistryHost(s); err == nil {
			t.Errorf("expected %q invalid", s)
		}
	}
}

func TestEnvMultiline(t *testing.T) {
	okCases := []string{
		"",
		"FOO=bar",
		"FOO=bar\nBAZ=qux",
		"  FOO=bar  \n\n  BAZ=qux",
	}
	for _, s := range okCases {
		if err := EnvMultiline(s); err != nil {
			t.Errorf("expected valid:\n%s\ngot: %v", s, err)
		}
	}

	badCases := []string{
		"foo=bar",       // 小写 key
		"FOO",           // 没有 =
		"=bar",          // 空 key
		"FOO=bar\nfoo=bar",
	}
	for _, s := range badCases {
		if err := EnvMultiline(s); err == nil {
			t.Errorf("expected invalid:\n%s", s)
		}
	}
}
