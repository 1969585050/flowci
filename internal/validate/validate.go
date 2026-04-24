// Package validate 提供用户输入的白名单校验。
//
// 设计原则：
//  1. handler 层在调业务之前调用本包，拦在最外层
//  2. 失败返回 error（handler 包会转成 ErrBadRequest 子类）
//  3. 正则表达式预编译为包级变量
//
// 对应规范：ipc-spec.md § 6.3 白名单校验表。
package validate

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// 业务错误；handler 层通常用 fmt.Errorf("%w: ...", ErrBadRequest) 包一层。
var (
	ErrInvalidContainerName = errors.New("invalid container name")
	ErrInvalidImageRef      = errors.New("invalid image reference")
	ErrInvalidPort          = errors.New("invalid port")
	ErrInvalidRegistryHost  = errors.New("invalid registry host")
	ErrInvalidEnvKey        = errors.New("invalid env key")
	ErrEnvLineMalformed     = errors.New("env line must be KEY=VALUE")
)

// container name：docker 规范 https://docs.docker.com/reference/cli/docker/container/rename/
// 首字符字母或数字；后续可含 _/./-；长度 1-63。
var containerNameRE = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,62}$`)

// image reference：简化版，允许 lowercase/digits，可带 / 分隔的路径和 :tag。
// 完整 OCI 规范复杂，此处采纳宽松白名单覆盖 99% 场景。
var imageRefRE = regexp.MustCompile(
	`^[a-z0-9]+(?:[._-][a-z0-9]+)*(?:/[a-z0-9]+(?:[._-][a-z0-9]+)*)*(?::[a-zA-Z0-9_.-]+)?$`)

// registry host: domain/IP[:port]，禁止 scheme 前缀。
var registryHostRE = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]*(?::\d{1,5})?$`)

// env key: shell env 变量名惯例，大写字母/数字/下划线，不能以数字开头。
var envKeyRE = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

// ContainerName 校验容器名。
func ContainerName(name string) error {
	if !containerNameRE.MatchString(name) {
		return fmt.Errorf("%w: %q", ErrInvalidContainerName, name)
	}
	return nil
}

// ImageRef 校验镜像引用（name[:tag] 或 registry/name[:tag]）。
func ImageRef(ref string) error {
	if !imageRefRE.MatchString(ref) {
		return fmt.Errorf("%w: %q", ErrInvalidImageRef, ref)
	}
	return nil
}

// Port 校验端口号字符串（"1"-"65535"）。空串允许（调用方用来表示"未指定"）。
func Port(p string) error {
	if p == "" {
		return nil
	}
	n, err := strconv.Atoi(p)
	if err != nil {
		return fmt.Errorf("%w: %q not a number", ErrInvalidPort, p)
	}
	if n < 1 || n > 65535 {
		return fmt.Errorf("%w: %d out of range [1, 65535]", ErrInvalidPort, n)
	}
	return nil
}

// RegistryHost 校验 Registry 主机名（不含 scheme）。空串允许（默认 Docker Hub）。
func RegistryHost(h string) error {
	if h == "" || h == "docker.io" {
		return nil
	}
	if !registryHostRE.MatchString(h) {
		return fmt.Errorf("%w: %q", ErrInvalidRegistryHost, h)
	}
	return nil
}

// EnvMultiline 校验多行 KEY=VALUE 字符串；空串允许。
// 每行 trim 后若非空则必须形如 KEY=VALUE，且 KEY 符合 envKeyRE。
func EnvMultiline(multi string) error {
	if multi == "" {
		return nil
	}
	for i, raw := range strings.Split(multi, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx <= 0 {
			return fmt.Errorf("%w (line %d): %q", ErrEnvLineMalformed, i+1, line)
		}
		key := line[:idx]
		if !envKeyRE.MatchString(key) {
			return fmt.Errorf("%w (line %d): %q", ErrInvalidEnvKey, i+1, key)
		}
	}
	return nil
}
