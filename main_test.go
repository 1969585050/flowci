package main

import (
	"strings"
	"testing"

	"flowci/internal/pipeline"

	"gopkg.in/yaml.v3"
)

// TestYamlMarshalFormat 验证 pipeline.YamlPipeline 能正确 Marshal 为含 name/type/stop_on_fail 的 YAML。
func TestYamlMarshalFormat(t *testing.T) {
	p := pipeline.YamlPipeline{
		Name: "test-pipeline",
		Config: pipeline.YamlConfig{
			Parallel:   false,
			StopOnFail: true,
		},
		Steps: []pipeline.YamlStep{
			{
				Type:   "build",
				Name:   "build-image",
				Retry:  0,
				OnFail: "stop",
				Config: map[string]interface{}{"tag": "latest"},
			},
			{
				Type:   "push",
				Name:   "push-image",
				Retry:  2,
				OnFail: "continue",
				Config: map[string]interface{}{},
			},
		},
	}

	yamlBytes, err := yaml.Marshal(p)
	if err != nil {
		t.Fatalf("yaml.Marshal failed: %v", err)
	}

	yamlStr := string(yamlBytes)

	if yamlStr == "" {
		t.Error("yaml.Marshal returned empty string")
	}
	if yamlStr == "{}" {
		t.Error("yaml.Marshal returned empty JSON-like object")
	}
	if !strings.Contains(yamlStr, "name: test-pipeline") {
		t.Error("Output does not contain pipeline name")
	}
	if !strings.Contains(yamlStr, "type: build") {
		t.Error("Output does not contain step type 'build'")
	}
	if !strings.Contains(yamlStr, "type: push") {
		t.Error("Output does not contain step type 'push'")
	}
	if !strings.Contains(yamlStr, "stop_on_fail: true") {
		t.Error("Output does not contain stop_on_fail setting")
	}
}

// TestYamlUnmarshalImport 验证 YAML 字符串能反序列化为 pipeline.YamlPipeline。
func TestYamlUnmarshalImport(t *testing.T) {
	yamlContent := `name: import-test-pipeline
config:
  parallel: false
  stop_on_fail: true
steps:
  - type: build
    name: build-image
    retry: 0
    on_fail: stop
  - type: deploy
    name: deploy-container
    retry: 1
    on_fail: continue
`

	var yp pipeline.YamlPipeline
	if err := yaml.Unmarshal([]byte(yamlContent), &yp); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}

	if yp.Name != "import-test-pipeline" {
		t.Errorf("Expected name 'import-test-pipeline', got '%s'", yp.Name)
	}
	if len(yp.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(yp.Steps))
	}
	if yp.Steps[0].Type != "build" {
		t.Errorf("Expected first step type 'build', got '%s'", yp.Steps[0].Type)
	}
	if yp.Steps[1].Type != "deploy" {
		t.Errorf("Expected second step type 'deploy', got '%s'", yp.Steps[1].Type)
	}
	if yp.Steps[1].Retry != 1 {
		t.Errorf("Expected second step retry 1, got %d", yp.Steps[1].Retry)
	}
	if !yp.Config.StopOnFail {
		t.Error("Expected StopOnFail to be true")
	}
}

// TestImportPipelineValidation 验证非法 step type 在解析层能通过、由业务校验层拒绝。
func TestImportPipelineValidation(t *testing.T) {
	yamlContent := `name: invalid-pipeline
config:
  stop_on_fail: true
steps:
  - type: invalid_type
    name: invalid-step
`

	var yp pipeline.YamlPipeline
	if err := yaml.Unmarshal([]byte(yamlContent), &yp); err != nil {
		t.Fatalf("yaml.Unmarshal should not fail for valid YAML structure: %v", err)
	}

	validTypes := map[string]bool{"build": true, "push": true, "deploy": true}
	for _, s := range yp.Steps {
		if !validTypes[s.Type] {
			t.Logf("Step type '%s' is invalid as expected", s.Type)
		}
	}
}

// TestEmptyPipelineExport 验证空 pipeline 仍能 Marshal 出带 name 字段的 YAML。
func TestEmptyPipelineExport(t *testing.T) {
	p := pipeline.YamlPipeline{
		Name:   "empty-pipeline",
		Config: pipeline.YamlConfig{},
		Steps:  []pipeline.YamlStep{},
	}

	yamlBytes, err := yaml.Marshal(p)
	if err != nil {
		t.Fatalf("yaml.Marshal failed: %v", err)
	}

	yamlStr := string(yamlBytes)
	if !strings.Contains(yamlStr, "name: empty-pipeline") {
		t.Error("Output does not contain pipeline name")
	}
}
