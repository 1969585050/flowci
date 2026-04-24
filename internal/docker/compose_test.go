package docker

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGenerateCompose_BasicFields(t *testing.T) {
	out, err := GenerateCompose(ComposeSpec{
		Image:         "nginx:latest",
		Name:          "web",
		HostPort:      "8080",
		ContainerPort: "80",
		RestartPolicy: "always",
	})
	if err != nil {
		t.Fatalf("GenerateCompose: %v", err)
	}

	wants := []string{
		"version:", "\"3.8\"",
		"services:",
		"web:",
		"image: nginx:latest",
		"container_name: web",
		"restart: always",
		"ports:",
		"8080:80",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("output missing %q:\n%s", w, out)
		}
	}
}

func TestGenerateCompose_DefaultRestartPolicy(t *testing.T) {
	out, err := GenerateCompose(ComposeSpec{Image: "app", Name: "a"})
	if err != nil {
		t.Fatalf("GenerateCompose: %v", err)
	}
	if !strings.Contains(out, "restart: unless-stopped") {
		t.Errorf("expected default restart: unless-stopped:\n%s", out)
	}
}

func TestGenerateCompose_SkipsPortsWhenIncomplete(t *testing.T) {
	// HostPort 有但 ContainerPort 为空：应不输出 ports 节
	out, _ := GenerateCompose(ComposeSpec{
		Image: "app", Name: "a", HostPort: "8080",
	})
	if strings.Contains(out, "ports:") {
		t.Errorf("ports should be skipped when ContainerPort empty:\n%s", out)
	}
}

func TestGenerateCompose_EnvMultiline(t *testing.T) {
	out, err := GenerateCompose(ComposeSpec{
		Image:        "app",
		Name:         "a",
		EnvMultiline: "FOO=bar\nBAZ=qux\n\n  DB_HOST=localhost  ",
	})
	if err != nil {
		t.Fatalf("GenerateCompose: %v", err)
	}
	for _, want := range []string{"environment:", "FOO=bar", "BAZ=qux", "DB_HOST=localhost"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output:\n%s", want, out)
		}
	}
}

func TestGenerateCompose_EmptyEnvSkipsSection(t *testing.T) {
	out, _ := GenerateCompose(ComposeSpec{Image: "a", Name: "a", EnvMultiline: "\n\n  \n"})
	if strings.Contains(out, "environment:") {
		t.Errorf("environment section should be skipped for all-whitespace env:\n%s", out)
	}
}

// TestGenerateCompose_ValidYaml 确保产出能被 yaml.Unmarshal 往回解析（结构完整性）。
func TestGenerateCompose_ValidYaml(t *testing.T) {
	out, err := GenerateCompose(ComposeSpec{
		Image: "nginx:1.25", Name: "web",
		HostPort: "80", ContainerPort: "80",
		EnvMultiline: "FOO=bar",
	})
	if err != nil {
		t.Fatalf("GenerateCompose: %v", err)
	}

	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("yaml.Unmarshal: %v\noutput:\n%s", err, out)
	}
	services, ok := parsed["services"].(map[string]interface{})
	if !ok {
		t.Fatalf("services node missing or wrong type:\n%s", out)
	}
	if _, ok := services["web"]; !ok {
		t.Errorf("services.web missing:\n%s", out)
	}
}

// TestGenerateCompose_SpecialCharsInImageTag 验证 yaml.Marshal 正确处理特殊字符，
// 不会像原 fmt.Sprintf 拼接版本那样破坏 YAML。
func TestGenerateCompose_SpecialCharsInImageTag(t *testing.T) {
	// 冒号在 docker image 里合法（tag 分隔符）；yaml.Marshal 会自动加引号或转义
	out, err := GenerateCompose(ComposeSpec{
		Image: "registry.example.com:5000/my-app:v1.2-beta",
		Name:  "myapp",
	})
	if err != nil {
		t.Fatalf("GenerateCompose: %v", err)
	}

	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("image with colons broke YAML: %v\n%s", err, out)
	}
	services := parsed["services"].(map[string]interface{})
	svc := services["myapp"].(map[string]interface{})
	if svc["image"] != "registry.example.com:5000/my-app:v1.2-beta" {
		t.Errorf("image tag lost in round-trip: %v", svc["image"])
	}
}
