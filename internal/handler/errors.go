package handler

import "errors"

// 业务哨兵错误；Wails 会把 return 的 error 作为 Promise reject 传给前端。
// 前端按 error.message 里的子串匹配分类处理（不引入独立错误码表）。
var (
	ErrBadRequest         = errors.New("bad request")
	ErrProjectNotFound    = errors.New("project not found")
	ErrPipelineNotFound   = errors.New("pipeline not found")
	ErrBuildNotFound      = errors.New("build record not found")
	ErrUnsupportedLang    = errors.New("unsupported language")
)
