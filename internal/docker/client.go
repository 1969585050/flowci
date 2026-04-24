// Package docker 封装对本机 docker CLI 的调用。
//
// 设计原则：
//  1. 所有对外函数第一参数 context.Context；内部用 CommandContext + 超时
//  2. 返回强类型 struct + error，不返回 map[string]interface{}
//  3. 不做业务级写入（构建记录、审计），由 handler / pipeline 包负责
//
// 本包不依赖 store / handler，可被 handler 与 pipeline 双向使用。
package docker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// 超时常量表（对齐 ipc-spec.md § 6.2）。
const (
	TimeoutQuery     = 10 * time.Second // version / ps / images 等查询
	TimeoutLifecycle = 30 * time.Second // start / stop / rm / tag
	TimeoutPull      = 10 * time.Minute
	TimeoutBuild     = 30 * time.Minute
	TimeoutPush      = 15 * time.Minute
	TimeoutCompose   = 10 * time.Minute
)

// Status 表示 docker daemon 连通性。
type Status struct {
	Connected bool   `json:"connected"`
	Version   string `json:"version"`
}

// Check 探测 docker daemon 是否可用。
// 永不返回 error：无法连通时返回 Status{Connected: false}。
func Check(ctx context.Context) Status {
	ctxTO, cancel := context.WithTimeout(ctx, TimeoutQuery)
	defer cancel()
	cmd := exec.CommandContext(ctxTO, "docker", "version", "--format", "{{.Server.Version}}")
	output, err := cmd.Output()
	if err != nil {
		return Status{Connected: false}
	}
	return Status{Connected: true, Version: strings.TrimSpace(string(output))}
}

// Client 抽象全部依赖 docker CLI 的能力，供 handler / pipeline 依赖注入。
// 包级函数（BuildImage / PushImage 等）保留作为默认实现入口；
// 生产代码通过 NewClient() 构造；测试可用自定义 fake 实现。
//
// GenerateCompose 是纯函数不进 interface；Check 不返 error 也不进。
type Client interface {
	BuildImage(ctx context.Context, req BuildRequest) (BuildResult, error)
	PushImage(ctx context.Context, req PushRequest) (PushResult, error)
	Deploy(ctx context.Context, req DeployRequest) (DeployResult, error)
	DeployWithCompose(ctx context.Context, content, workDir string) (ComposeDeployResult, error)

	ListImages(ctx context.Context) ([]Image, error)
	RemoveImage(ctx context.Context, id string) error

	ListContainers(ctx context.Context) ([]Container, error)
	StartContainer(ctx context.Context, id string) error
	StopContainer(ctx context.Context, id string) error
	RemoveContainer(ctx context.Context, id string) error
	GetContainerLogs(ctx context.Context, id string, tail int) (string, error)

	Check(ctx context.Context) Status
}

// cliClient 是 Client 的默认实现，所有方法委托包级函数。
type cliClient struct{}

// NewClient 构造默认 docker Client（调本机 docker CLI）。
func NewClient() Client { return &cliClient{} }

func (cliClient) BuildImage(ctx context.Context, req BuildRequest) (BuildResult, error) {
	return BuildImage(ctx, req)
}
func (cliClient) PushImage(ctx context.Context, req PushRequest) (PushResult, error) {
	return PushImage(ctx, req)
}
func (cliClient) Deploy(ctx context.Context, req DeployRequest) (DeployResult, error) {
	return Deploy(ctx, req)
}
func (cliClient) DeployWithCompose(ctx context.Context, content, workDir string) (ComposeDeployResult, error) {
	return DeployWithCompose(ctx, content, workDir)
}
func (cliClient) ListImages(ctx context.Context) ([]Image, error) { return ListImages(ctx) }
func (cliClient) RemoveImage(ctx context.Context, id string) error {
	return RemoveImage(ctx, id)
}
func (cliClient) ListContainers(ctx context.Context) ([]Container, error) {
	return ListContainers(ctx)
}
func (cliClient) StartContainer(ctx context.Context, id string) error {
	return StartContainer(ctx, id)
}
func (cliClient) StopContainer(ctx context.Context, id string) error {
	return StopContainer(ctx, id)
}
func (cliClient) RemoveContainer(ctx context.Context, id string) error {
	return RemoveContainer(ctx, id)
}
func (cliClient) GetContainerLogs(ctx context.Context, id string, tail int) (string, error) {
	return GetContainerLogs(ctx, id, tail)
}
func (cliClient) Check(ctx context.Context) Status { return Check(ctx) }

// run 是 docker 子命令执行的内部 helper，统一 CombinedOutput + 错误包装。
// 即使失败也会把输出返回（供错误码分类、日志诊断）。
func run(ctx context.Context, timeout time.Duration, args ...string) ([]byte, error) {
	ctxTO, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctxTO, "docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("docker %s: %w: %s", args[0], err, strings.TrimSpace(string(out)))
	}
	return out, nil
}
