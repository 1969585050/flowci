package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"flowci/internal/pipeline"
	"flowci/internal/store"

	"gopkg.in/yaml.v3"
)

// ListPipelines 列出某项目下全部流水线。
func (a *App) ListPipelines(ctx context.Context, projectID string) ([]store.Pipeline, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, fmt.Errorf("%w: projectId required", ErrBadRequest)
	}
	return store.ListPipelines(projectID)
}

// ListAllPipelines 一次性列出所有项目下的流水线，避免前端 N+1。
func (a *App) ListAllPipelines(ctx context.Context) ([]store.Pipeline, error) {
	return store.ListAllPipelines()
}

// CreatePipeline 新建流水线。
func (a *App) CreatePipeline(ctx context.Context, req *CreatePipelineRequest) (*store.Pipeline, error) {
	if req == nil || strings.TrimSpace(req.ProjectID) == "" || strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("%w: projectId and name are required", ErrBadRequest)
	}
	if len(req.Steps) == 0 {
		return nil, fmt.Errorf("%w: at least one step required", ErrBadRequest)
	}
	return store.CreatePipeline(store.CreatePipelineInput{
		ProjectID: req.ProjectID,
		Name:      req.Name,
		Steps:     req.Steps,
		Config:    req.Config,
	})
}

// UpdatePipeline 更新指定 ID 的流水线。
func (a *App) UpdatePipeline(ctx context.Context, req *UpdatePipelineRequest) (*store.Pipeline, error) {
	if req == nil || strings.TrimSpace(req.ID) == "" {
		return nil, fmt.Errorf("%w: id required", ErrBadRequest)
	}
	p, err := store.UpdatePipeline(req.ID, store.UpdatePipelineInput{
		Name:   req.Name,
		Steps:  req.Steps,
		Config: req.Config,
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrPipelineNotFound
		}
		return nil, err
	}
	return p, nil
}

// DeletePipeline 按 ID 删除流水线。
func (a *App) DeletePipeline(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: id required", ErrBadRequest)
	}
	if err := store.DeletePipeline(id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrPipelineNotFound
		}
		return err
	}
	return nil
}

// ExportPipelineToYaml 导出指定流水线为 YAML 文本。
// 不存在时返回 ErrPipelineNotFound；YAML 解析失败时返回具体错误。
func (a *App) ExportPipelineToYaml(ctx context.Context, id string) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("%w: id required", ErrBadRequest)
	}
	p, err := store.GetPipeline(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return "", ErrPipelineNotFound
		}
		return "", err
	}

	steps := make([]pipeline.YamlStep, len(p.Steps))
	for i, s := range p.Steps {
		steps[i] = pipeline.YamlStep{
			Type:   s.Type,
			Name:   s.Name,
			Retry:  s.Retry,
			OnFail: s.OnFail,
			Config: s.Config,
		}
	}
	yp := pipeline.YamlPipeline{
		Name: p.Name,
		Config: pipeline.YamlConfig{
			Parallel:   p.Config.Parallel,
			StopOnFail: p.Config.StopOnFail,
		},
		Steps: steps,
	}
	bs, err := yaml.Marshal(yp)
	if err != nil {
		return "", fmt.Errorf("marshal pipeline yaml: %w", err)
	}
	return string(bs), nil
}

// ImportPipelineFromYaml 解析 YAML 并新建流水线；失败原因前置在 error 中。
func (a *App) ImportPipelineFromYaml(ctx context.Context, req *ImportPipelineYamlRequest) (*store.Pipeline, error) {
	if req == nil {
		return nil, ErrBadRequest
	}
	if strings.TrimSpace(req.ProjectID) == "" {
		return nil, fmt.Errorf("%w: projectId required", ErrBadRequest)
	}
	if strings.TrimSpace(req.Yaml) == "" {
		return nil, fmt.Errorf("%w: yaml content required", ErrBadRequest)
	}

	var yp pipeline.YamlPipeline
	if err := yaml.Unmarshal([]byte(req.Yaml), &yp); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}
	if err := pipeline.ValidateYaml(yp); err != nil {
		return nil, err
	}

	steps := make([]store.PipelineStep, len(yp.Steps))
	for i, s := range yp.Steps {
		steps[i] = store.PipelineStep{
			Type:   s.Type,
			Name:   s.Name,
			Retry:  s.Retry,
			OnFail: s.OnFail,
			Config: s.Config,
		}
	}
	return store.CreatePipeline(store.CreatePipelineInput{
		ProjectID: req.ProjectID,
		Name:      yp.Name,
		Steps:     steps,
		Config: store.PipelineConfig{
			Parallel:   yp.Config.Parallel,
			StopOnFail: yp.Config.StopOnFail,
		},
	})
}

// ExecutePipeline 触发一次流水线执行；同一 pipelineID 并发提交时返回 pipeline.ErrPipelineBusy。
func (a *App) ExecutePipeline(ctx context.Context, req *ExecutePipelineRequest) (*pipeline.ExecuteResult, error) {
	if req == nil || strings.TrimSpace(req.PipelineID) == "" {
		return nil, fmt.Errorf("%w: pipelineId required", ErrBadRequest)
	}
	res, err := a.executor.Execute(ctx, req.PipelineID, req.ProjectID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrPipelineNotFound
		}
		return nil, err
	}
	return &res, nil
}
