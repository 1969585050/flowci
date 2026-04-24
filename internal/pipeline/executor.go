package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"flowci/internal/docker"
	"flowci/internal/store"
)

// StepStatus 单步执行状态。
type StepStatus string

const (
	StepSuccess StepStatus = "success"
	StepFailed  StepStatus = "failed"
)

// StepLog 单步执行日志，用于前端展示。
type StepLog struct {
	Step    string     `json:"step"`
	Type    string     `json:"type"`
	Status  StepStatus `json:"status"`
	Message string     `json:"message,omitempty"`
	Error   string     `json:"error,omitempty"`
}

// ExecuteResult 整个 pipeline 的执行结果。
type ExecuteResult struct {
	Success bool      `json:"success"`
	Logs    []StepLog `json:"logs"`
	Message string    `json:"message"`
}

// ErrPipelineBusy 表示目标 pipeline 正在执行中，重复提交会被拒绝。
var ErrPipelineBusy = errors.New("pipeline is busy")

// Executor 流水线执行器，内部持有 per-pipeline 锁防止并发重复提交。
type Executor struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

// NewExecutor 构造新的执行器。全局通常一个实例即可（handler 包持有）。
func NewExecutor() *Executor {
	return &Executor{locks: make(map[string]*sync.Mutex)}
}

// Execute 按顺序执行 pipelineID 对应的流水线。
// 同一 pipelineID 不允许并发进入（TryLock 失败返回 ErrPipelineBusy）。
// TODO(phase-3)：支持 pipeline.Config.Parallel 并行分支。
func (e *Executor) Execute(ctx context.Context, pipelineID, projectID string) (ExecuteResult, error) {
	lock := e.getLock(pipelineID)
	if !lock.TryLock() {
		return ExecuteResult{}, ErrPipelineBusy
	}
	defer lock.Unlock()

	p, err := store.GetPipeline(pipelineID)
	if err != nil {
		return ExecuteResult{}, fmt.Errorf("get pipeline: %w", err)
	}

	logs := make([]StepLog, 0, len(p.Steps))
	allSuccess := true

	for _, step := range p.Steps {
		log, stopOnFailure := runStep(ctx, projectID, step, p.Config.StopOnFail)
		logs = append(logs, log)
		if log.Status == StepFailed && stopOnFailure {
			allSuccess = false
			break
		}
	}

	msg := "Pipeline executed successfully"
	if !allSuccess {
		msg = "Pipeline failed"
	}
	return ExecuteResult{Success: allSuccess, Logs: logs, Message: msg}, nil
}

// getLock 懒创建 pipelineID 对应的 Mutex。
func (e *Executor) getLock(id string) *sync.Mutex {
	e.mu.Lock()
	defer e.mu.Unlock()
	l, ok := e.locks[id]
	if !ok {
		l = &sync.Mutex{}
		e.locks[id] = l
	}
	return l
}

// runStep 执行单步，含 step.Retry 次数的重试。
// 第二个返回值表示"失败时是否整体停止"。
func runStep(ctx context.Context, projectID string, step store.PipelineStep, stopOnFail bool) (StepLog, bool) {
	log := StepLog{Step: step.Name, Type: step.Type}

	var stepErr error
	var message string
	for attempt := 0; attempt <= step.Retry; attempt++ {
		if attempt > 0 {
			slog.Info("retrying pipeline step", "step", step.Name, "attempt", attempt)
		}
		message, stepErr = dispatchStep(ctx, projectID, step)
		if stepErr == nil {
			break
		}
	}

	if stepErr != nil {
		log.Status = StepFailed
		log.Error = stepErr.Error()
		// step.OnFail 为空等同于 "stop"
		stop := step.OnFail != "continue" && (step.OnFail == "stop" || stopOnFail || step.OnFail == "")
		return log, stop
	}
	log.Status = StepSuccess
	log.Message = message
	return log, false
}

// dispatchStep 按 step.Type 分发到 docker 包的具体操作。
func dispatchStep(ctx context.Context, projectID string, step store.PipelineStep) (string, error) {
	switch step.Type {
	case "build":
		return runBuildStep(ctx, projectID, step)
	case "push":
		return runPushStep(ctx, step)
	case "deploy":
		return runDeployStep(ctx, step)
	default:
		return "", fmt.Errorf("unknown step type: %s", step.Type)
	}
}

func runBuildStep(ctx context.Context, projectID string, step store.PipelineStep) (string, error) {
	tag, _ := step.Config["tag"].(string)
	if tag == "" {
		tag = "latest"
	}
	contextPath, _ := step.Config["contextPath"].(string)

	record, err := store.CreateBuildRecord(projectID, tag, "latest")
	if err != nil {
		return "", fmt.Errorf("create build record: %w", err)
	}

	res, buildErr := docker.BuildImage(ctx, docker.BuildRequest{
		Tag:         tag,
		ContextPath: contextPath,
	})

	status := "success"
	if buildErr != nil {
		status = "failed"
	}
	if finishErr := store.FinishBuildRecord(record.ID, status, res.Log); finishErr != nil {
		slog.Error("finish build record failed", "id", record.ID, "err", finishErr)
	}

	if buildErr != nil {
		return "", buildErr
	}
	return fmt.Sprintf("Built: %s:%s", res.ImageName, res.ImageTag), nil
}

func runPushStep(ctx context.Context, step store.PipelineStep) (string, error) {
	imageName, _ := step.Config["imageName"].(string)
	registry, _ := step.Config["registry"].(string)
	username, _ := step.Config["username"].(string)
	password, _ := step.Config["password"].(string)

	if _, err := docker.PushImage(ctx, docker.PushRequest{
		Image:    imageName,
		Registry: registry,
		Username: username,
		Password: password,
	}); err != nil {
		return "", err
	}
	return "Image pushed successfully", nil
}

func runDeployStep(ctx context.Context, step store.PipelineStep) (string, error) {
	image, _ := step.Config["imageName"].(string)
	name, _ := step.Config["name"].(string)
	hostPort, _ := step.Config["hostPort"].(string)
	containerPort, _ := step.Config["containerPort"].(string)
	restartPolicy, _ := step.Config["restartPolicy"].(string)

	if _, err := docker.Deploy(ctx, docker.DeployRequest{
		Image:         image,
		Name:          name,
		HostPort:      hostPort,
		ContainerPort: containerPort,
		RestartPolicy: restartPolicy,
	}); err != nil {
		return "", err
	}
	return "Container deployed successfully", nil
}
