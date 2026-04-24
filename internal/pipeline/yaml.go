// Package pipeline 流水线 YAML 导入导出类型定义。
//
// 本包现阶段（阶段 1）只提供 YAML 类型的单一定义，
// 用于替代 main.go / main_test.go / main_integration_test.go 中重复声明的
// YamlPipeline / YamlStep / YamlConfig 结构体。
//
// 阶段 2 将继续扩展：
//   - executor.go  流水线执行器（串行/并行 + per-pipeline 锁）
//   - validator.go 步骤与配置校验
//   - pipeline.go  业务类型（目前仍暂存 internal/store/pipelines.go）
package pipeline

// YamlStep 对应 pipeline YAML 中的单个步骤。
// 字段命名沿用 YAML snake_case 约定（on_fail），与存储层 PipelineStep 的 JSON 字段分离。
type YamlStep struct {
	Type   string                 `yaml:"type"`
	Name   string                 `yaml:"name"`
	Retry  int                    `yaml:"retry,omitempty"`
	OnFail string                 `yaml:"on_fail,omitempty"`
	Config map[string]interface{} `yaml:"config,omitempty"`
}

// YamlConfig 对应 pipeline YAML 顶层 config 节。
// StopOnFail 不带 omitempty：默认就应显式落盘为 false 或 true，避免导出歧义。
type YamlConfig struct {
	Parallel   bool `yaml:"parallel,omitempty"`
	StopOnFail bool `yaml:"stop_on_fail"`
}

// YamlPipeline 是完整 YAML 文档的顶层结构。
type YamlPipeline struct {
	Name   string     `yaml:"name"`
	Config YamlConfig `yaml:"config"`
	Steps  []YamlStep `yaml:"steps"`
}
