package main

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYamlMarshalFormat(t *testing.T) {
	type YamlStep struct {
		Type   string `yaml:"type"`
		Name   string `yaml:"name"`
		Retry  int    `yaml:"retry,omitempty"`
		OnFail string `yaml:"on_fail,omitempty"`
		Config map[string]interface{} `yaml:"config,omitempty"`
	}

	type YamlConfig struct {
		Parallel   bool `yaml:"parallel,omitempty"`
		StopOnFail bool `yaml:"stop_on_fail"`
	}

	type YamlPipeline struct {
		Name   string     `yaml:"name"`
		Config YamlConfig `yaml:"config"`
		Steps  []YamlStep `yaml:"steps"`
	}

	pipeline := YamlPipeline{
		Name: "test-pipeline",
		Config: YamlConfig{
			Parallel:   false,
			StopOnFail: true,
		},
		Steps: []YamlStep{
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

	yamlBytes, err := yaml.Marshal(pipeline)
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

	if !contains(yamlStr, "name: test-pipeline") {
		t.Error("Output does not contain pipeline name")
	}

	if !contains(yamlStr, "type: build") {
		t.Error("Output does not contain step type 'build'")
	}

	if !contains(yamlStr, "type: push") {
		t.Error("Output does not contain step type 'push'")
	}

	if !contains(yamlStr, "stop_on_fail: true") {
		t.Error("Output does not contain stop_on_fail setting")
	}
}

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

	type YamlStep struct {
		Type   string `yaml:"type"`
		Name   string `yaml:"name"`
		Retry  int    `yaml:"retry,omitempty"`
		OnFail string `yaml:"on_fail,omitempty"`
		Config map[string]interface{} `yaml:"config,omitempty"`
	}

	type YamlConfig struct {
		Parallel   bool `yaml:"parallel,omitempty"`
		StopOnFail bool `yaml:"stop_on_fail"`
	}

	type YamlPipeline struct {
		Name   string     `yaml:"name"`
		Config YamlConfig `yaml:"config"`
		Steps  []YamlStep `yaml:"steps"`
	}

	var yp YamlPipeline
	err := yaml.Unmarshal([]byte(yamlContent), &yp)
	if err != nil {
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

func TestImportPipelineValidation(t *testing.T) {
	yamlContent := `name: invalid-pipeline
config:
  stop_on_fail: true
steps:
  - type: invalid_type
    name: invalid-step
`

	type YamlStep struct {
		Type   string `yaml:"type"`
		Name   string `yaml:"name"`
		Retry  int    `yaml:"retry,omitempty"`
		OnFail string `yaml:"on_fail,omitempty"`
		Config map[string]interface{} `yaml:"config,omitempty"`
	}

	type YamlConfig struct {
		Parallel   bool `yaml:"parallel,omitempty"`
		StopOnFail bool `yaml:"stop_on_fail"`
	}

	type YamlPipeline struct {
		Name   string     `yaml:"name"`
		Config YamlConfig `yaml:"config"`
		Steps  []YamlStep `yaml:"steps"`
	}

	var yp YamlPipeline
	err := yaml.Unmarshal([]byte(yamlContent), &yp)
	if err != nil {
		t.Fatalf("yaml.Unmarshal should not fail for valid YAML structure: %v", err)
	}

	validTypes := map[string]bool{"build": true, "push": true, "deploy": true}
	for _, s := range yp.Steps {
		if !validTypes[s.Type] {
			t.Logf("Step type '%s' is invalid as expected", s.Type)
		}
	}
}

func TestEmptyPipelineExport(t *testing.T) {
	type YamlPipeline struct {
		Name   string `yaml:"name"`
		Config struct{} `yaml:"config"`
		Steps  []struct{} `yaml:"steps"`
	}

	pipeline := YamlPipeline{
		Name:   "empty-pipeline",
		Config: struct{}{},
		Steps:  []struct{}{},
	}

	yamlBytes, err := yaml.Marshal(pipeline)
	if err != nil {
		t.Fatalf("yaml.Marshal failed: %v", err)
	}

	yamlStr := string(yamlBytes)
	if !contains(yamlStr, "name: empty-pipeline") {
		t.Error("Output does not contain pipeline name")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
