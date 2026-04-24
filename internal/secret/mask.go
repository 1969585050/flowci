package secret

import (
	"reflect"
	"strings"
)

// Mask 将敏感字符串遮蔽为固定形态（非空 → "***"，空 → ""）。
// 单字段日志用：slog.Info("login", "password", secret.Mask(pwd))
func Mask(s string) string {
	if s == "" {
		return ""
	}
	return "***"
}

// MaskStruct 遍历 struct，把 tag `mask:"true"` 的 string 字段复制替换为 "***"，
// 返回的是**同类型的新值**（原值不受影响）。用于 IPC 日志中间件。
//
// 仅支持顶层 struct；嵌套 struct 的 mask tag 不递归处理（阶段 3 够用）。
func MaskStruct(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	// 对指针，先解引用再操作
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return v
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return v
	}

	// 复制一份到新 struct
	out := reflect.New(rv.Type()).Elem()
	out.Set(rv)

	t := rv.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := strings.TrimSpace(f.Tag.Get("mask"))
		if tag != "true" {
			continue
		}
		if out.Field(i).Kind() != reflect.String {
			continue
		}
		if !out.Field(i).CanSet() {
			continue
		}
		if out.Field(i).String() == "" {
			continue
		}
		out.Field(i).SetString("***")
	}
	return out.Interface()
}
