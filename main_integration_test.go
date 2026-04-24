package main

import (
	"context"
	"os"
	"testing"

	"flowci/store"
	"gopkg.in/yaml.v3"
)

func TestIntegrationExportImportPipeline(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "flowci-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := store.Init(tmpDir); err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}
	defer store.Close()

	project, err := store.CreateProject(store.CreateProjectInput{
		Name:     "test-project",
		Path:     "/tmp/test",
		Language: "go",
	})
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	pipeline, err := store.CreatePipeline(store.CreatePipelineInput{
		ProjectID: project.ID,
		Name:     "test-pipeline",
		Steps: []store.PipelineStep{
			{Type: "build", Name: "build-step", Retry: 0, OnFail: "stop", Config: map[string]interface{}{"tag": "latest"}},
			{Type: "push", Name: "push-step", Retry: 1, OnFail: "continue", Config: map[string]interface{}{}},
		},
		Config: store.PipelineConfig{Parallel: false, StopOnFail: true},
	})
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	app := &App{}
	ctx := context.Background()

	yamlOutput := app.ExportPipelineToYaml(ctx, pipeline.ID)

	t.Logf("Export output:\n%s", yamlOutput)

	if yamlOutput == "" {
		t.Fatal("Export returned empty string")
	}

	if yamlOutput == "{}" {
		t.Fatal("Export returned empty JSON-like object")
	}

	if contains(yamlOutput, "# Pipeline not found") {
		t.Fatal("Export failed: pipeline not found")
	}

	type ExportedStep struct {
		Type   string `yaml:"type"`
		Name   string `yaml:"name"`
		Retry  int    `yaml:"retry,omitempty"`
		OnFail string `yaml:"on_fail,omitempty"`
	}
	type ExportedConfig struct {
		Parallel   bool `yaml:"parallel,omitempty"`
		StopOnFail bool `yaml:"stop_on_fail"`
	}
	type ExportedPipeline struct {
		Name   string          `yaml:"name"`
		Config ExportedConfig  `yaml:"config"`
		Steps  []ExportedStep `yaml:"steps"`
	}

	var exported ExportedPipeline
	if err := yaml.Unmarshal([]byte(yamlOutput), &exported); err != nil {
		t.Fatalf("Failed to parse exported YAML: %v\nOutput was:\n%s", err, yamlOutput)
	}

	if exported.Name != "test-pipeline" {
		t.Errorf("Expected name 'test-pipeline', got '%s'", exported.Name)
	}

	if len(exported.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(exported.Steps))
	}

	if exported.Steps[0].Type != "build" {
		t.Errorf("Expected first step type 'build', got '%s'", exported.Steps[0].Type)
	}

	if exported.Steps[1].Type != "push" {
		t.Errorf("Expected second step type 'push', got '%s'", exported.Steps[1].Type)
	}

	if !exported.Config.StopOnFail {
		t.Error("Expected StopOnFail to be true")
	}

	t.Log("YAML export test PASSED!")
}

func TestIntegrationImportPipeline(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "flowci-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := store.Init(tmpDir); err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}
	defer store.Close()

	project, err := store.CreateProject(store.CreateProjectInput{
		Name:     "test-project-2",
		Path:     "/tmp/test",
		Language: "go",
	})
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
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

	app := &App{}
	ctx := context.Background()

	result := app.ImportPipelineFromYaml(ctx, map[string]interface{}{
		"projectId": project.ID,
		"yaml":      yamlContent,
	})

	if result == nil {
		t.Fatal("Import returned nil")
	}

	if errMsg, ok := result["error"]; ok {
		t.Fatalf("Import failed: %v", errMsg)
	}

	importedName, ok := result["name"].(string)
	if !ok || importedName != "imported-pipeline" {
		t.Errorf("Expected imported name 'imported-pipeline', got '%v'", result["name"])
	}

	t.Logf("Import result: %v", result)
	t.Log("YAML import test PASSED!")
}
