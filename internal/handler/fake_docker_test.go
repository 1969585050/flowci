package handler

import (
	"context"
	"errors"
	"sync"

	"flowci/internal/docker"
)

// fakeDockerClient 是 docker.Client 的 handler 包测试专用 fake。
// 所有方法都是可脚本化的（预设返回值 + 记录调用历史）。
type fakeDockerClient struct {
	mu sync.Mutex

	// 预设返回
	status          docker.Status
	buildResult     docker.BuildResult
	buildErr        error
	pushResult      docker.PushResult
	pushErr         error
	deployResult    docker.DeployResult
	deployErr       error
	composeResult   docker.ComposeDeployResult
	composeErr      error
	images          []docker.Image
	imagesErr       error
	containers      []docker.Container
	containersErr   error
	removeImageErr  error
	startErr        error
	stopErr         error
	removeCtrErr    error
	logsOutput      string
	logsErr         error

	// 调用历史
	buildCalls      []docker.BuildRequest
	pushCalls       []docker.PushRequest
	deployCalls     []docker.DeployRequest
	composeCalls    []struct{ Content, WorkDir string }
	removeImageIDs  []string
	startIDs        []string
	stopIDs         []string
	removeCtrIDs    []string
	logsCalls       []struct {
		ID   string
		Tail int
	}
}

func (f *fakeDockerClient) Check(ctx context.Context) docker.Status { return f.status }

func (f *fakeDockerClient) BuildImage(ctx context.Context, req docker.BuildRequest) (docker.BuildResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.buildCalls = append(f.buildCalls, req)
	return f.buildResult, f.buildErr
}

func (f *fakeDockerClient) PushImage(ctx context.Context, req docker.PushRequest) (docker.PushResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pushCalls = append(f.pushCalls, req)
	return f.pushResult, f.pushErr
}

func (f *fakeDockerClient) Deploy(ctx context.Context, req docker.DeployRequest) (docker.DeployResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deployCalls = append(f.deployCalls, req)
	return f.deployResult, f.deployErr
}

func (f *fakeDockerClient) DeployWithCompose(ctx context.Context, content, workDir string) (docker.ComposeDeployResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.composeCalls = append(f.composeCalls, struct{ Content, WorkDir string }{content, workDir})
	return f.composeResult, f.composeErr
}

func (f *fakeDockerClient) ListImages(ctx context.Context) ([]docker.Image, error) {
	return f.images, f.imagesErr
}

func (f *fakeDockerClient) RemoveImage(ctx context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.removeImageIDs = append(f.removeImageIDs, id)
	return f.removeImageErr
}

func (f *fakeDockerClient) ListContainers(ctx context.Context) ([]docker.Container, error) {
	return f.containers, f.containersErr
}

func (f *fakeDockerClient) StartContainer(ctx context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.startIDs = append(f.startIDs, id)
	return f.startErr
}

func (f *fakeDockerClient) StopContainer(ctx context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stopIDs = append(f.stopIDs, id)
	return f.stopErr
}

func (f *fakeDockerClient) RemoveContainer(ctx context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.removeCtrIDs = append(f.removeCtrIDs, id)
	return f.removeCtrErr
}

func (f *fakeDockerClient) GetContainerLogs(ctx context.Context, id string, tail int) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.logsCalls = append(f.logsCalls, struct {
		ID   string
		Tail int
	}{id, tail})
	return f.logsOutput, f.logsErr
}

// 编译期断言 fakeDockerClient 实现 docker.Client。
var _ docker.Client = (*fakeDockerClient)(nil)

// 常用错误，测试内部方便引用。
var errFake = errors.New("fake docker failure")
