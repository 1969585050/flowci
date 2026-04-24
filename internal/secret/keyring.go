// Package secret 封装敏感凭证的存储与日志遮蔽。
//
// 设计原则（见 data-spec.md § 5）:
//  - 密码/token/OAuth secret 等不进 SQLite；只进 OS keyring
//  - Windows 用 DPAPI（Credentials Manager），macOS 用 Keychain，Linux 用 Secret Service
//  - DB 端仅持有引用元数据（registry_url + username + has_password 标记位）
//
// 阶段 3 MVP：只提供 Set/Get/Delete 接口；具体的 registry_credentials 表与 handler
// Bind 方法在后续 iteration 再加（需前端 UI 配合）。
package secret

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

// serviceName 是 keyring 中用于归组本应用条目的命名空间；
// 不同用户共用同一 OS keyring 时用于区分 FlowCI 自己的条目。
const serviceName = "FlowCI"

// Set 写入一条凭证。key 通常形如 "registry:docker.io"。
// 已存在同 key 的条目会被覆盖（keyring 天然幂等）。
func Set(key, value string) error {
	if err := keyring.Set(serviceName, key, value); err != nil {
		return fmt.Errorf("keyring set %q: %w", key, err)
	}
	return nil
}

// Get 读取凭证；不存在时返回 keyring.ErrNotFound（由调用方 errors.Is 判断）。
func Get(key string) (string, error) {
	v, err := keyring.Get(serviceName, key)
	if err != nil {
		return "", fmt.Errorf("keyring get %q: %w", key, err)
	}
	return v, nil
}

// Delete 删除指定 key 的凭证；不存在也视为成功（幂等）。
func Delete(key string) error {
	if err := keyring.Delete(serviceName, key); err != nil {
		return fmt.Errorf("keyring delete %q: %w", key, err)
	}
	return nil
}

// ErrNotFound 重导出 go-keyring 的哨兵错误，方便调用方统一 errors.Is 匹配。
var ErrNotFound = keyring.ErrNotFound
