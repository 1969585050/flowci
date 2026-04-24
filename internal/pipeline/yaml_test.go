package pipeline

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestYamlMarshalFormat 验证 YamlPipeline 能正确 Marshal 为含 name/type/stop_on_fail 的 YAML。
func TestYamlMarshalFormat(t *testing.T) {
	p := YamlPipeline{
		Name:   "test-pipeline",
		Config: YamlConfig{Parallel: false, StopOnFail: true},
		Steps: []YamlStep{
			{Type: "build", Name: "build-image", Retry: 0, OnFail: "stop", Config: map[string]interface{}{"tag": "latest"}},
			{Type: "push", Name: "push-image", Retry: 2, OnFail: "continue", Config: map[string]interface{}{}},
		},
	}

	yamlBytes, err := yaml.Marshal(p)
	if err != nil {
		t.Fatalf("yaml.Marshal failed: %v", err)
	}
	yamlStr := string(yamlBytes)

	for _, want := range []string{"name: test-pipeline", "type: build", "type: push", "stop_on_fail: true"} {
		if !strings.Contains(yamlStr, want) {
			t.Errorf("output missing %q:\n%s", want, yamlStr)
		}
	}
}

// TestYamlUnmarshalImport 验证 YAML 文本能反序列化为 YamlPipeline。
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

	var yp YamlPipeline
	if err := yaml.Unmarshal([]byte(yamlContent), &yp); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}

	if yp.Name != "import-test-pipeline" {
		t.Errorf("expected name 'import-test-pipeline', got %q", yp.Name)
	}
	if len(yp.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(yp.Steps))
	}
	if yp.Steps[0].Type != "build" || yp.Steps[1].Type != "deploy" {
		t.Errorf("unexpected step types: %q %q", yp.Steps[0].Type, yp.Steps[1].Type)
	}
	if yp.Steps[1].Retry != 1 {
		t.Errorf("expected second step retry=1, got %d", yp.Steps[1].Retry)
	}
	if !yp.Config.StopOnFail {
		t.Error("expected StopOnFail=true")
	}
}

// TestValidateYamlRejectsInvalidStepType 验证 ValidateYaml 拒绝非法 step type。
func TestValidateYamlRejectsInvalidStepType(t *testing.T) {
	yp := YamlPipeline{
		Name:   "invalid-pipeline",
		Config: YamlConfig{StopOnFail: true},
		Steps:  []YamlStep{{Type: "invalid_type", Name: "x"}},
	}
	err := ValidateYaml(yp)
	if err == nil {
		t.Fatal("expected error for invalid step type")
	}
}

// TestValidateYamlRejectsEmptyName 验证空 pipeline name 会被拒绝。
func TestValidateYamlRejectsEmptyName(t *testing.T) {
	yp := YamlPipeline{
		Name:  "",
		Steps: []YamlStep{{Type: "build", Name: "build-image"}},
	}
	if err := ValidateYaml(yp); err == nil {
		t.Fatal("expected error for empty pipeline name")
	}
}

// TestValidateYamlRejectsNegativeRetry 验证负 retry 会被拒绝。
func TestValidateYamlRejectsNegativeRetry(t *testing.T) {
	yp := YamlPipeline{
		Name:  "ok-name",
		Steps: []YamlStep{{Type: "build", Name: "step", Retry: -1}},
	}
	if err := ValidateYaml(yp); err == nil {
		t.Fatal("expected error for negative retry")
	}
}

// TestEmptyPipelineMarshal 空 pipeline 仍能 Marshal 出带 name 字段的 YAML。
func TestEmptyPipelineMarshal(t *testing.T) {
	p := YamlPipeline{Name: "empty-pipeline", Steps: []YamlStep{}}
	bs, err := yaml.Marshal(p)
	if err != nil {
		t.Fatalf("yaml.Marshal failed: %v", err)
	}
	if !strings.Contains(string(bs), "name: empty-pipeline") {
		t.Error("output missing name field")
	}
}
