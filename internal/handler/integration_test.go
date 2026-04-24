package handler

import (
	"context"
	"os"
	"strings"
	"testing"

	"flowci/internal/pipeline"
	"flowci/internal/store"

	"gopkg.in/yaml.v3"
)

// setupStore 为单个测试初始化临时目录下的 store，返回清理函数。
func setupStore(t *testing.T) func() {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "flowci-test-*")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	if err := store.Init(tmpDir); err != nil {
		t.Fatalf("store.Init: %v", err)
	}
	return func() {
		store.Close()
		_ = os.RemoveAll(tmpDir)
	}
}

// TestExportPipelineToYaml_RoundTrip 集成测试：
// 建项目 + 建流水线 + 调 handler.ExportPipelineToYaml → 反解校验字段。
func TestExportPipelineToYaml_RoundTrip(t *testing.T) {
	cleanup := setupStore(t)
	defer cleanup()

	project, err := store.CreateProject(store.CreateProjectInput{
		Name: "test-project", Path: "/tmp/test", Language: "go",
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	created, err := store.CreatePipeline(store.CreatePipelineInput{
		ProjectID: project.ID,
		Name:      "test-pipeline",
		Steps: []store.PipelineStep{
			{Type: "build", Name: "build-step", Retry: 0, OnFail: "stop", Config: map[string]interface{}{"tag": "latest"}},
			{Type: "push", Name: "push-step", Retry: 1, OnFail: "continue", Config: map[string]interface{}{}},
		},
		Config: store.PipelineConfig{StopOnFail: true},
	})
	if err != nil {
		t.Fatalf("CreatePipeline: %v", err)
	}

	app := NewApp(os.TempDir())
	out, err := app.ExportPipelineToYaml(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("ExportPipelineToYaml: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("export returned empty string")
	}

	var exported pipeline.YamlPipeline
	if err := yaml.Unmarshal([]byte(out), &exported); err != nil {
		t.Fatalf("unmarshal exported yaml: %v\noutput:\n%s", err, out)
	}
	if exported.Name != "test-pipeline" {
		t.Errorf("expected name 'test-pipeline', got %q", exported.Name)
	}
	if len(exported.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(exported.Steps))
	}
	if exported.Steps[0].Type != "build" || exported.Steps[1].Type != "push" {
		t.Errorf("unexpected step types: %q %q", exported.Steps[0].Type, exported.Steps[1].Type)
	}
	if !exported.Config.StopOnFail {
		t.Error("expected StopOnFail=true")
	}
}

// TestExportPipelineToYaml_NotFound 验证不存在的 pipelineId 返回 ErrPipelineNotFound。
func TestExportPipelineToYaml_NotFound(t *testing.T) {
	cleanup := setupStore(t)
	defer cleanup()

	app := NewApp(os.TempDir())
	_, err := app.ExportPipelineToYaml(context.Background(), "no-such-id")
	if err == nil {
		t.Fatal("expected error for nonexistent pipeline")
	}
	if !strings.Contains(err.Error(), "pipeline not found") {
		t.Errorf("expected pipeline-not-found error, got %v", err)
	}
}

// TestImportPipelineFromYaml_Success 验证合法 YAML 能成功导入。
func TestImportPipelineFromYaml_Success(t *testing.T) {
	cleanup := setupStore(t)
	defer cleanup()

	project, err := store.CreateProject(store.CreateProjectInput{
		Name: "test-project-2", Path: "/tmp/test", Language: "go",
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	yamlContent := `name: imported-pipeline
config:
  parallel: false
  stop_on_fail: true
steps:
  - type: build
    name: import-build
    retry: 0
    on_fail: stop
  - type: deploy
    name: import-deploy
    retry: 2
    on_fail: continue
`

	app := NewApp(os.TempDir())
	result, err := app.ImportPipelineFromYaml(context.Background(), &ImportPipelineYamlRequest{
		ProjectID: project.ID,
		Yaml:      yamlContent,
	})
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if result.Name != "imported-pipeline" {
		t.Errorf("expected imported name 'imported-pipeline', got %q", result.Name)
	}
	if len(result.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(result.Steps))
	}
}

// TestImportPipelineFromYaml_InvalidType 验证非法 step type 被拒绝。
func TestImportPipelineFromYaml_InvalidType(t *testing.T) {
	cleanup := setupStore(t)
	defer cleanup()

	project, err := store.CreateProject(store.CreateProjectInput{
		Name: "test-project-3", Path: "/tmp/test", Language: "go",
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	yamlContent := `name: bad-pipeline
config:
  stop_on_fail: true
steps:
  - type: invalid_type
    name: bad-step
`

	app := NewApp(os.TempDir())
	_, err = app.ImportPipelineFromYaml(context.Background(), &ImportPipelineYamlRequest{
		ProjectID: project.ID,
		Yaml:      yamlContent,
	})
	if err == nil {
		t.Fatal("expected error for invalid step type")
	}
}
