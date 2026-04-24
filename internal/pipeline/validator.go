package pipeline

import (
	"errors"
	"fmt"
	"strings"
)

// 校验错误（可用 errors.Is 分类）
var (
	ErrPipelineNameEmpty = errors.New("pipeline name empty")
	ErrStepTypeInvalid   = errors.New("invalid step type")
	ErrStepNameEmpty     = errors.New("step name empty")
	ErrRetryNegative     = errors.New("retry count must be >= 0")
	ErrOnFailInvalid     = errors.New("invalid onFail value")
)

// 合法 step.type 白名单。
var validStepTypes = map[string]bool{
	"build":  true,
	"push":   true,
	"deploy": true,
}

// 合法 step.on_fail 白名单；空串与 "stop" 语义相同（默认失败停止）。
var validOnFail = map[string]bool{
	"":         true,
	"stop":     true,
	"continue": true,
}

// ValidateYaml 校验 YAML 反序列化出来的 pipeline 结构。
// 对 step[i] 报错时会带上索引，便于用户定位。
func ValidateYaml(yp YamlPipeline) error {
	if strings.TrimSpace(yp.Name) == "" {
		return ErrPipelineNameEmpty
	}
	for i, s := range yp.Steps {
		if err := validateStep(s.Type, s.Name, s.Retry, s.OnFail); err != nil {
			return fmt.Errorf("step[%d]: %w", i, err)
		}
	}
	return nil
}

// validateStep 单步校验；空名 / 非法枚举 / 负重试次数都拒绝。
func validateStep(stepType, name string, retry int, onFail string) error {
	if !validStepTypes[stepType] {
		return fmt.Errorf("%w: %q (must be build/push/deploy)", ErrStepTypeInvalid, stepType)
	}
	if strings.TrimSpace(name) == "" {
		return ErrStepNameEmpty
	}
	if retry < 0 {
		return fmt.Errorf("%w: %d", ErrRetryNegative, retry)
	}
	if !validOnFail[onFail] {
		return fmt.Errorf("%w: %q (must be stop/continue)", ErrOnFailInvalid, onFail)
	}
	return nil
}
