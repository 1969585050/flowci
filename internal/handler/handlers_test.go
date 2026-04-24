package handler

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"flowci/internal/docker"
	"flowci/internal/store"
)

// setupApp 在临时目录初始化 store，返回 (App, projectID, cleanup)。
func setupApp(t *testing.T, fake *fakeDockerClient) (*App, string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "flowci-handler-test-*")
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
	app := NewAppWithClient(tmpDir, fake)
	return app, proj.ID, func() {
		store.Close()
		_ = os.RemoveAll(tmpDir)
	}
}

// ---- CheckDocker ----

func TestCheckDocker_ReturnsClientStatus(t *testing.T) {
	fake := &fakeDockerClient{status: docker.Status{Connected: true, Version: "25.0.3"}}
	app, _, cleanup := setupApp(t, fake)
	defer cleanup()

	got := app.CheckDocker(context.Background())
	if !got.Connected || got.Version != "25.0.3" {
		t.Errorf("unexpected: %+v", got)
	}
}

// ---- BuildImage ----

func TestBuildImage_WritesBuildRecord(t *testing.T) {
	fake := &fakeDockerClient{buildResult: docker.BuildResult{ImageName: "nginx", ImageTag: "v1", Log: "build log"}}
	app, pid, cleanup := setupApp(t, fake)
	defer cleanup()

	res, err := app.BuildImage(context.Background(), &BuildImageRequest{
		ProjectID: pid, Tag: "nginx:v1", ContextPath: ".",
	})
	if err != nil {
		t.Fatalf("BuildImage: %v", err)
	}
	if res.ImageName != "nginx" || res.ImageTag != "v1" {
		t.Errorf("unexpected result: %+v", res)
	}
	if len(fake.buildCalls) != 1 {
		t.Fatalf("expected 1 build call, got %d", len(fake.buildCalls))
	}

	records, _ := store.ListBuildRecords(pid)
	if len(records) != 1 {
		t.Fatalf("expected 1 build record, got %d", len(records))
	}
	if records[0].Status != "success" {
		t.Errorf("expected success status, got %q", records[0].Status)
	}
}

func TestBuildImage_FailedRecordStatus(t *testing.T) {
	fake := &fakeDockerClient{
		buildResult: docker.BuildResult{Log: "error log"},
		buildErr:    errors.New("docker build failed"),
	}
	app, pid, cleanup := setupApp(t, fake)
	defer cleanup()

	_, err := app.BuildImage(context.Background(), &BuildImageRequest{
		ProjectID: pid, Tag: "x:v1",
	})
	if err == nil {
		t.Fatal("expected build error")
	}

	records, _ := store.ListBuildRecords(pid)
	if len(records) != 1 || records[0].Status != "failed" {
		t.Errorf("expected 1 failed record, got %+v", records)
	}
}

func TestBuildImage_InvalidTagRejected(t *testing.T) {
	fake := &fakeDockerClient{}
	app, pid, cleanup := setupApp(t, fake)
	defer cleanup()

	_, err := app.BuildImage(context.Background(), &BuildImageRequest{
		ProjectID: pid, Tag: "NGINX:UPPERCASE", // 大写不符合 ImageRef 白名单
	})
	if !errors.Is(err, ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
	if len(fake.buildCalls) != 0 {
		t.Error("docker should not be called when validation fails")
	}
}

func TestBuildImage_ProjectNotFound(t *testing.T) {
	fake := &fakeDockerClient{}
	app, _, cleanup := setupApp(t, fake)
	defer cleanup()

	_, err := app.BuildImage(context.Background(), &BuildImageRequest{
		ProjectID: "nonexistent", Tag: "x:v1",
	})
	if !errors.Is(err, ErrProjectNotFound) {
		t.Errorf("expected ErrProjectNotFound, got %v", err)
	}
}

// ---- RemoveImage: docker 错误分类为中文消息 ----

func TestRemoveImage_NotFound(t *testing.T) {
	fake := &fakeDockerClient{removeImageErr: docker.ErrImageNotFound}
	app, _, cleanup := setupApp(t, fake)
	defer cleanup()

	err := app.RemoveImage(context.Background(), "abc")
	if err == nil || !strings.Contains(err.Error(), "镜像不存在") {
		t.Errorf("expected 镜像不存在, got %v", err)
	}
}

func TestRemoveImage_InUse(t *testing.T) {
	fake := &fakeDockerClient{removeImageErr: docker.ErrImageInUse}
	app, _, cleanup := setupApp(t, fake)
	defer cleanup()

	err := app.RemoveImage(context.Background(), "abc")
	if err == nil || !strings.Contains(err.Error(), "镜像正在使用中") {
		t.Errorf("expected 镜像正在使用中, got %v", err)
	}
}

func TestRemoveImage_EmptyID(t *testing.T) {
	fake := &fakeDockerClient{}
	app, _, cleanup := setupApp(t, fake)
	defer cleanup()

	err := app.RemoveImage(context.Background(), "")
	if !errors.Is(err, ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
}

// ---- Containers ----

func TestListContainers_PassThroughFromClient(t *testing.T) {
	fake := &fakeDockerClient{
		containers: []docker.Container{
			{ID: "c1", Names: []string{"web"}, Image: "nginx"},
		},
	}
	app, _, cleanup := setupApp(t, fake)
	defer cleanup()

	list, err := app.ListContainers(context.Background())
	if err != nil {
		t.Fatalf("ListContainers: %v", err)
	}
	if len(list) != 1 || list[0].ID != "c1" {
		t.Errorf("unexpected containers: %+v", list)
	}
}

func TestStartStopRemoveContainer_CallsClient(t *testing.T) {
	fake := &fakeDockerClient{}
	app, _, cleanup := setupApp(t, fake)
	defer cleanup()

	ctx := context.Background()
	if err := app.StartContainer(ctx, "id1"); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := app.StopContainer(ctx, "id1"); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if err := app.RemoveContainer(ctx, "id1"); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	if len(fake.startIDs) != 1 || fake.startIDs[0] != "id1" {
		t.Errorf("start history wrong: %v", fake.startIDs)
	}
	if len(fake.stopIDs) != 1 || len(fake.removeCtrIDs) != 1 {
		t.Errorf("stop/remove history wrong")
	}
}

func TestStartContainer_EmptyID(t *testing.T) {
	fake := &fakeDockerClient{}
	app, _, cleanup := setupApp(t, fake)
	defer cleanup()

	if err := app.StartContainer(context.Background(), ""); !errors.Is(err, ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
}

func TestGetContainerLogs_DefaultTail(t *testing.T) {
	fake := &fakeDockerClient{logsOutput: "some log"}
	app, _, cleanup := setupApp(t, fake)
	defer cleanup()

	out, err := app.GetContainerLogs(context.Background(), "id1", 0)
	if err != nil {
		t.Fatalf("GetContainerLogs: %v", err)
	}
	if out != "some log" {
		t.Errorf("unexpected output: %q", out)
	}
	if fake.logsCalls[0].Tail != 0 {
		// handler 不做 tail 默认化（默认化在 docker 包），只 pass-through
		t.Logf("(informational) tail passed through as %d", fake.logsCalls[0].Tail)
	}
}

// ---- DeployContainer: 白名单拒绝 ----

func TestDeployContainer_InvalidName(t *testing.T) {
	fake := &fakeDockerClient{}
	app, _, cleanup := setupApp(t, fake)
	defer cleanup()

	_, err := app.DeployContainer(context.Background(), &DeployContainerRequest{
		Image: "nginx:latest",
		Name:  "has space", // 含空格，白名单拒绝
	})
	if !errors.Is(err, ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
	if len(fake.deployCalls) != 0 {
		t.Error("docker.Deploy should not be called on validation failure")
	}
}

func TestDeployContainer_Success(t *testing.T) {
	fake := &fakeDockerClient{deployResult: docker.DeployResult{ID: "cid", Message: "ok"}}
	app, _, cleanup := setupApp(t, fake)
	defer cleanup()

	res, err := app.DeployContainer(context.Background(), &DeployContainerRequest{
		Image: "nginx:latest", Name: "web",
		HostPort: "8080", ContainerPort: "80",
	})
	if err != nil {
		t.Fatalf("DeployContainer: %v", err)
	}
	if res.ID != "cid" {
		t.Errorf("unexpected: %+v", res)
	}
	if len(fake.deployCalls) != 1 {
		t.Fatalf("expected 1 deploy call, got %d", len(fake.deployCalls))
	}
	got := fake.deployCalls[0]
	if got.Name != "web" || got.HostPort != "8080" {
		t.Errorf("request fields lost: %+v", got)
	}
}

// ---- PushImage: 白名单 + keyring 回退 ----

func TestPushImage_ValidRequestPassesThrough(t *testing.T) {
	fake := &fakeDockerClient{pushResult: docker.PushResult{Log: "pushed"}}
	app, _, cleanup := setupApp(t, fake)
	defer cleanup()

	res, err := app.PushImage(context.Background(), &PushImageRequest{
		Image: "myimg:v1", Registry: "docker.io",
		Username: "alice", Password: "pw123",
	})
	if err != nil {
		t.Fatalf("PushImage: %v", err)
	}
	if res.Log != "pushed" {
		t.Errorf("unexpected log: %q", res.Log)
	}
	if len(fake.pushCalls) != 1 {
		t.Fatalf("expected 1 push call, got %d", len(fake.pushCalls))
	}
	got := fake.pushCalls[0]
	// 前端传的 password 应被直接传给 docker（未走 keyring）
	if got.Password != "pw123" {
		t.Errorf("password lost: %+v", got)
	}
}

func TestPushImage_InvalidRegistryRejected(t *testing.T) {
	fake := &fakeDockerClient{}
	app, _, cleanup := setupApp(t, fake)
	defer cleanup()

	_, err := app.PushImage(context.Background(), &PushImageRequest{
		Image: "myimg:v1", Registry: "https://evil.example.com", // scheme 前缀被白名单拒
	})
	if !errors.Is(err, ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
	if len(fake.pushCalls) != 0 {
		t.Error("docker should not be called")
	}
}

// ---- DeployWithCompose: workdir 白名单 ----

func TestDeployWithCompose_WorkdirOutsideDataDirRejected(t *testing.T) {
	fake := &fakeDockerClient{}
	app, _, cleanup := setupApp(t, fake)
	defer cleanup()

	// app.dataDir 是 setupApp 里的 tmpDir；用完全不相关的 /tmp 外部路径（视平台）
	outside, err := filepath.Abs("/totally-outside-path")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	_, err = app.DeployWithCompose(context.Background(), &DeployWithComposeRequest{
		Compose: "version: '3'",
		Workdir: outside,
	})
	if !errors.Is(err, ErrBadRequest) {
		t.Errorf("expected ErrBadRequest for outside workdir, got %v", err)
	}
	if len(fake.composeCalls) != 0 {
		t.Error("docker.DeployWithCompose should not be called")
	}
}

func TestDeployWithCompose_DefaultWorkdir(t *testing.T) {
	fake := &fakeDockerClient{composeResult: docker.ComposeDeployResult{Output: "ok"}}
	app, _, cleanup := setupApp(t, fake)
	defer cleanup()

	res, err := app.DeployWithCompose(context.Background(), &DeployWithComposeRequest{
		Compose: "version: '3.8'\nservices:\n  web: {image: nginx}",
		Workdir: "", // 空 → 默认 dataDir/tmp/compose
	})
	if err != nil {
		t.Fatalf("DeployWithCompose: %v", err)
	}
	if res.Output != "ok" {
		t.Errorf("unexpected: %+v", res)
	}
	if len(fake.composeCalls) != 1 {
		t.Fatalf("expected 1 compose call, got %d", len(fake.composeCalls))
	}
	wd := fake.composeCalls[0].WorkDir
	if !strings.Contains(wd, "tmp") || !strings.Contains(wd, "compose") {
		t.Errorf("expected default workdir under data dir/tmp/compose, got %q", wd)
	}
}

// ---- ExecutePipeline: 走完 executor + fake docker ----

func TestExecutePipeline_CallsExecutor(t *testing.T) {
	fake := &fakeDockerClient{buildResult: docker.BuildResult{ImageName: "x", ImageTag: "v1", Log: "ok"}}
	app, pid, cleanup := setupApp(t, fake)
	defer cleanup()

	pl, err := store.CreatePipeline(store.CreatePipelineInput{
		ProjectID: pid, Name: "auto",
		Steps:  []store.PipelineStep{{Type: "build", Name: "b", Config: map[string]interface{}{"tag": "x:v1"}, OnFail: "stop"}},
		Config: store.PipelineConfig{StopOnFail: true},
	})
	if err != nil {
		t.Fatalf("CreatePipeline: %v", err)
	}

	res, err := app.ExecutePipeline(context.Background(), &ExecutePipelineRequest{
		PipelineID: pl.ID, ProjectID: pid,
	})
	if err != nil {
		t.Fatalf("ExecutePipeline: %v", err)
	}
	if !res.Success {
		t.Errorf("expected success, got %+v", res)
	}
	if len(fake.buildCalls) != 1 {
		t.Errorf("expected 1 build call, got %d", len(fake.buildCalls))
	}
}

func TestExecutePipeline_NotFound(t *testing.T) {
	fake := &fakeDockerClient{}
	app, _, cleanup := setupApp(t, fake)
	defer cleanup()

	_, err := app.ExecutePipeline(context.Background(), &ExecutePipelineRequest{
		PipelineID: "nonexistent", ProjectID: "nonexistent",
	})
	if !errors.Is(err, ErrPipelineNotFound) {
		t.Errorf("expected ErrPipelineNotFound, got %v", err)
	}
}
