package pipeline

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"flowci/internal/docker"
	"flowci/internal/store"
)

// fakeDockerClient 是 docker.Client 的测试用 fake 实现。
// 计数各方法调用次数、记录参数、按预设的 buildFailTimes 顺序返回错误。
type fakeDockerClient struct {
	mu sync.Mutex

	buildCalls  []docker.BuildRequest
	pushCalls   []docker.PushRequest
	deployCalls []docker.DeployRequest

	// BuildImage 前 buildFailTimes 次调用返回 buildErr；之后成功返回 BuildResult{Log: "ok-log"}
	buildFailTimes int32
	buildErr       error
	buildLog       string

	pushErr   error
	deployErr error

	// Build 调用时的 hook；用于测试并发锁（让 build 阻塞）
	buildHook func()
}

func (f *fakeDockerClient) BuildImage(ctx context.Context, req docker.BuildRequest) (docker.BuildResult, error) {
	f.mu.Lock()
	f.buildCalls = append(f.buildCalls, req)
	callIdx := int32(len(f.buildCalls))
	f.mu.Unlock()

	if f.buildHook != nil {
		f.buildHook()
	}

	if callIdx <= atomic.LoadInt32(&f.buildFailTimes) {
		return docker.BuildResult{Log: f.buildLog}, f.buildErr
	}
	return docker.BuildResult{ImageName: "img", ImageTag: req.Tag, Log: f.buildLog}, nil
}

func (f *fakeDockerClient) PushImage(ctx context.Context, req docker.PushRequest) (docker.PushResult, error) {
	f.mu.Lock()
	f.pushCalls = append(f.pushCalls, req)
	f.mu.Unlock()
	if f.pushErr != nil {
		return docker.PushResult{}, f.pushErr
	}
	return docker.PushResult{Log: "pushed"}, nil
}

func (f *fakeDockerClient) Deploy(ctx context.Context, req docker.DeployRequest) (docker.DeployResult, error) {
	f.mu.Lock()
	f.deployCalls = append(f.deployCalls, req)
	f.mu.Unlock()
	if f.deployErr != nil {
		return docker.DeployResult{}, f.deployErr
	}
	return docker.DeployResult{ID: "cid", Message: "ok"}, nil
}

// 以下方法 executor 不使用，空实现即可
func (f *fakeDockerClient) DeployWithCompose(context.Context, string, string) (docker.ComposeDeployResult, error) {
	return docker.ComposeDeployResult{}, nil
}
func (f *fakeDockerClient) ListImages(context.Context) ([]docker.Image, error) { return nil, nil }
func (f *fakeDockerClient) RemoveImage(context.Context, string) error          { return nil }
func (f *fakeDockerClient) ListContainers(context.Context) ([]docker.Container, error) {
	return nil, nil
}
func (f *fakeDockerClient) StartContainer(context.Context, string) error  { return nil }
func (f *fakeDockerClient) StopContainer(context.Context, string) error   { return nil }
func (f *fakeDockerClient) RemoveContainer(context.Context, string) error { return nil }
func (f *fakeDockerClient) GetContainerLogs(context.Context, string, int) (string, error) {
	return "", nil
}
func (f *fakeDockerClient) Check(context.Context) docker.Status { return docker.Status{Connected: true} }

// setupExecutor 准备一套 store + executor + 一个 project + logsDir。
// 返回 (executor, projectID, cleanup)。
func setupExecutor(t *testing.T, fake *fakeDockerClient) (*Executor, string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "flowci-executor-test-*")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	if err := store.Init(tmpDir); err != nil {
		t.Fatalf("store.Init: %v", err)
	}
	proj, err := store.CreateProject(store.CreateProjectInput{Name: "p", Path: "/t", Language: "go"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	e := NewExecutorWithClient(tmpDir+"/logs", fake)
	cleanup := func() {
		store.Close()
		_ = os.RemoveAll(tmpDir)
	}
	return e, proj.ID, cleanup
}

// createPipeline 在 store 里创建一条 pipeline，返回其 ID。
func createPipelineFor(t *testing.T, pid string, steps []store.PipelineStep, stopOnFail bool) string {
	t.Helper()
	p, err := store.CreatePipeline(store.CreatePipelineInput{
		ProjectID: pid,
		Name:      "test-pipe",
		Steps:     steps,
		Config:    store.PipelineConfig{StopOnFail: stopOnFail},
	})
	if err != nil {
		t.Fatalf("CreatePipeline: %v", err)
	}
	return p.ID
}

func TestExecute_AllStepsSucceed(t *testing.T) {
	fake := &fakeDockerClient{buildLog: "ok"}
	e, pid, cleanup := setupExecutor(t, fake)
	defer cleanup()

	plid := createPipelineFor(t, pid, []store.PipelineStep{
		{Type: "build", Name: "b", Config: map[string]interface{}{"tag": "v1"}, OnFail: "stop"},
		{Type: "push", Name: "p", Config: map[string]interface{}{"imageName": "img:v1"}, OnFail: "stop"},
		{Type: "deploy", Name: "d", Config: map[string]interface{}{"imageName": "img:v1", "name": "web"}, OnFail: "stop"},
	}, true)

	res, err := e.Execute(context.Background(), plid, pid)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !res.Success {
		t.Errorf("expected success, got logs=%+v", res.Logs)
	}
	if len(res.Logs) != 3 {
		t.Errorf("expected 3 logs, got %d", len(res.Logs))
	}
	for i, log := range res.Logs {
		if log.Status != StepSuccess {
			t.Errorf("step[%d] status=%s, want success", i, log.Status)
		}
	}
	if len(fake.buildCalls) != 1 || len(fake.pushCalls) != 1 || len(fake.deployCalls) != 1 {
		t.Errorf("expected 1 each; got b=%d p=%d d=%d",
			len(fake.buildCalls), len(fake.pushCalls), len(fake.deployCalls))
	}
}

func TestExecute_RetryThenSucceed(t *testing.T) {
	fake := &fakeDockerClient{
		buildErr:       errors.New("transient docker failure"),
		buildFailTimes: 2, // 前 2 次失败，第 3 次成功
	}
	e, pid, cleanup := setupExecutor(t, fake)
	defer cleanup()

	plid := createPipelineFor(t, pid, []store.PipelineStep{
		{Type: "build", Name: "b", Config: map[string]interface{}{"tag": "v1"}, Retry: 2, OnFail: "stop"},
	}, true)

	res, err := e.Execute(context.Background(), plid, pid)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !res.Success {
		t.Errorf("expected success with retry, got logs=%+v", res.Logs)
	}
	if len(fake.buildCalls) != 3 {
		t.Errorf("expected 3 build attempts (initial+2 retries), got %d", len(fake.buildCalls))
	}
}

func TestExecute_FailAllRetries_OnFailStop(t *testing.T) {
	fake := &fakeDockerClient{
		buildErr:       errors.New("persistent failure"),
		buildFailTimes: 100, // 所有尝试都失败
	}
	e, pid, cleanup := setupExecutor(t, fake)
	defer cleanup()

	plid := createPipelineFor(t, pid, []store.PipelineStep{
		{Type: "build", Name: "b", Config: map[string]interface{}{"tag": "v1"}, Retry: 1, OnFail: "stop"},
		{Type: "push", Name: "p", Config: map[string]interface{}{"imageName": "img"}, OnFail: "stop"},
	}, true)

	res, err := e.Execute(context.Background(), plid, pid)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.Success {
		t.Error("expected Success=false")
	}
	if len(fake.buildCalls) != 2 {
		t.Errorf("expected 2 build attempts (1 retry), got %d", len(fake.buildCalls))
	}
	if len(fake.pushCalls) != 0 {
		t.Errorf("push should not run after build failure with OnFail=stop, got %d calls", len(fake.pushCalls))
	}
	if len(res.Logs) != 1 {
		t.Errorf("expected logs to stop at first failed step, got %d", len(res.Logs))
	}
	if res.Logs[0].Status != StepFailed {
		t.Errorf("first step status=%s, want failed", res.Logs[0].Status)
	}
}

func TestExecute_FailContinues_OnFailContinue(t *testing.T) {
	fake := &fakeDockerClient{
		buildErr:       errors.New("build failed but continue"),
		buildFailTimes: 100,
	}
	e, pid, cleanup := setupExecutor(t, fake)
	defer cleanup()

	// 用 pipeline-level StopOnFail=false + step OnFail=continue 跳过失败步骤
	plid := createPipelineFor(t, pid, []store.PipelineStep{
		{Type: "build", Name: "b", Config: map[string]interface{}{"tag": "v1"}, OnFail: "continue"},
		{Type: "push", Name: "p", Config: map[string]interface{}{"imageName": "img"}, OnFail: "stop"},
	}, false)

	res, err := e.Execute(context.Background(), plid, pid)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.Success {
		t.Error("expected Success=false (had failed step even with continue)")
	}
	if len(res.Logs) != 2 {
		t.Errorf("expected 2 log entries (build failed + push ran), got %d", len(res.Logs))
	}
	if len(fake.pushCalls) != 1 {
		t.Errorf("push should run after build failure with OnFail=continue, got %d calls", len(fake.pushCalls))
	}
}

func TestExecute_UnknownStepType(t *testing.T) {
	fake := &fakeDockerClient{}
	e, pid, cleanup := setupExecutor(t, fake)
	defer cleanup()

	plid := createPipelineFor(t, pid, []store.PipelineStep{
		{Type: "unknown-type", Name: "x", OnFail: "stop"},
	}, true)

	res, err := e.Execute(context.Background(), plid, pid)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.Success {
		t.Error("expected Success=false for unknown step type")
	}
	if len(res.Logs) != 1 || res.Logs[0].Status != StepFailed {
		t.Errorf("expected one failed log, got %+v", res.Logs)
	}
	if !strings.Contains(res.Logs[0].Error, "unknown step type") {
		t.Errorf("expected 'unknown step type' in error, got %q", res.Logs[0].Error)
	}
}

func TestExecute_PipelineNotFound(t *testing.T) {
	fake := &fakeDockerClient{}
	e, _, cleanup := setupExecutor(t, fake)
	defer cleanup()

	_, err := e.Execute(context.Background(), "nonexistent-id", "some-project")
	if err == nil {
		t.Fatal("expected error for nonexistent pipeline")
	}
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected wrap of store.ErrNotFound, got %v", err)
	}
}

// TestExecute_ConcurrentRejectsBusy 验证 per-pipeline TryLock：
// 同一 pipelineID 第二次并发调用应立即返回 ErrPipelineBusy。
func TestExecute_ConcurrentRejectsBusy(t *testing.T) {
	blocked := make(chan struct{})
	release := make(chan struct{})
	fake := &fakeDockerClient{
		buildHook: func() {
			close(blocked)
			<-release
		},
		buildLog: "ok",
	}
	e, pid, cleanup := setupExecutor(t, fake)
	defer cleanup()

	plid := createPipelineFor(t, pid, []store.PipelineStep{
		{Type: "build", Name: "b", Config: map[string]interface{}{"tag": "v1"}, OnFail: "stop"},
	}, true)

	// 第一次调用阻塞在 buildHook
	firstDone := make(chan struct{})
	go func() {
		_, _ = e.Execute(context.Background(), plid, pid)
		close(firstDone)
	}()

	<-blocked // 确保第一次已获锁并在 BuildImage 里挂住

	_, err := e.Execute(context.Background(), plid, pid)
	if !errors.Is(err, ErrPipelineBusy) {
		t.Errorf("expected ErrPipelineBusy, got %v", err)
	}

	close(release)
	select {
	case <-firstDone:
	case <-time.After(2 * time.Second):
		t.Fatal("first Execute did not complete after release")
	}
}

// TestExecute_DifferentPipelinesNotBusy 不同 pipelineID 并发不应互相阻塞。
func TestExecute_DifferentPipelinesNotBusy(t *testing.T) {
	fake := &fakeDockerClient{buildLog: "ok"}
	e, pid, cleanup := setupExecutor(t, fake)
	defer cleanup()

	plid1 := createPipelineFor(t, pid, []store.PipelineStep{
		{Type: "build", Name: "b1", Config: map[string]interface{}{"tag": "v1"}, OnFail: "stop"},
	}, true)

	// 第二条 pipeline（不同 ID，同 project）
	p2, _ := store.CreatePipeline(store.CreatePipelineInput{
		ProjectID: pid, Name: "pipe2",
		Steps:  []store.PipelineStep{{Type: "build", Name: "b2", Config: map[string]interface{}{"tag": "v2"}, OnFail: "stop"}},
		Config: store.PipelineConfig{StopOnFail: true},
	})

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		res, err := e.Execute(context.Background(), plid1, pid)
		if err != nil || !res.Success {
			t.Errorf("p1 failed: %v %+v", err, res)
		}
	}()
	go func() {
		defer wg.Done()
		res, err := e.Execute(context.Background(), p2.ID, pid)
		if err != nil || !res.Success {
			t.Errorf("p2 failed: %v %+v", err, res)
		}
	}()
	wg.Wait()
}
