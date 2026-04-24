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
