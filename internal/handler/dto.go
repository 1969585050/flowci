// Package handler 是 Wails Bind 层，所有前端可见的方法聚集在此包下。
//
// 设计规则（详见 ipc-spec.md）：
//  1. 方法签名 (ctx, *Request) (*Response, error) 或其允许的简化形式
//  2. 参数/返回值一律强类型 struct，JSON tag 全部 camelCase
//  3. 业务错误 return error，禁止 {success: false, error: ""} 协议
//  4. handler 本身不写业务；调 internal/docker, internal/pipeline, internal/store
package handler

import "flowci/internal/store"

// ---- Project ----

// CreateProjectRequest 新建项目。
type CreateProjectRequest struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Language string `json:"language"`
}

// UpdateProjectRequest 更新项目。
type UpdateProjectRequest struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Language string `json:"language"`
}

// ---- Pipeline ----

// CreatePipelineRequest 新建流水线。
type CreatePipelineRequest struct {
	ProjectID string                `json:"projectId"`
	Name      string                `json:"name"`
	Steps     []store.PipelineStep  `json:"steps"`
	Config    store.PipelineConfig  `json:"config"`
}

// UpdatePipelineRequest 更新流水线。
type UpdatePipelineRequest struct {
	ID     string                `json:"id"`
	Name   string                `json:"name"`
	Steps  []store.PipelineStep  `json:"steps"`
	Config store.PipelineConfig  `json:"config"`
}

// ImportPipelineYamlRequest 通过 YAML 文本导入流水线。
type ImportPipelineYamlRequest struct {
	ProjectID string `json:"projectId"`
	Yaml      string `json:"yaml"`
}

// ExecutePipelineRequest 触发流水线执行。
type ExecutePipelineRequest struct {
	PipelineID string `json:"pipelineId"`
	ProjectID  string `json:"projectId"`
}

// ---- Build ----

// BuildImageRequest 单次构建镜像（非 pipeline）。
type BuildImageRequest struct {
	ProjectID   string `json:"projectId"`
	Tag         string `json:"tag"`
	ContextPath string `json:"contextPath"`
	NoCache     bool   `json:"noCache"`
	PullLatest  bool   `json:"pullLatest"`
}

// ---- Container / Compose ----

// DeployContainerRequest 直接 docker run 启动容器。
type DeployContainerRequest struct {
	Image         string `json:"image"`
	Name          string `json:"name"`
	HostPort      string `json:"hostPort"`
	ContainerPort string `json:"containerPort"`
	RestartPolicy string `json:"restartPolicy"`
	Env           string `json:"env"`
}

// GenerateComposeRequest 依规格产出 docker-compose.yml 文本。
type GenerateComposeRequest struct {
	Image         string `json:"image"`
	Name          string `json:"name"`
	HostPort      string `json:"hostPort"`
	ContainerPort string `json:"containerPort"`
	RestartPolicy string `json:"restartPolicy"`
	Env           string `json:"env"`
}

// DeployWithComposeRequest 由前端传来完整 compose 文本 + 工作目录，走 up -d。
type DeployWithComposeRequest struct {
	Compose string `json:"compose"`
	Workdir string `json:"workdir"`
}

// ---- Push ----

// PushImageRequest 推送镜像。Password 是敏感字段。
// TODO(phase-3)：Password 通过 secret.Mask 遮蔽后写 IPC 日志；
// 凭证改从 OS keyring 读而非前端明文传入。
type PushImageRequest struct {
	Image    string `json:"image"`
	Registry string `json:"registry"`
	Username string `json:"username"`
	Password string `json:"password" mask:"true"`
}

// ---- Settings ----

// SaveSettingsRequest 批量写入 key-value 设置。
type SaveSettingsRequest struct {
	Settings map[string]string `json:"settings"`
}

// DetectDockerEnvRequest 触发一次 Docker 环境探测。
// Host 为空时探测当前 settings 配置的 host（或本地）；非空时一次性覆盖。
type DetectDockerEnvRequest struct {
	Host string `json:"host"`
}

// DiagnoseBuildRequest 触发 AI 诊断指定构建记录的失败原因。
type DiagnoseBuildRequest struct {
	BuildID string `json:"buildId"`
}

// DiagnoseBuildResponse AI 返回 Markdown 格式的诊断报告。
type DiagnoseBuildResponse struct {
	Markdown string `json:"markdown"`
	Model    string `json:"model"`
}

// AIConfig 在 settings 中持久化的 AI 提供方配置。
// APIKey 永远不通过 Settings DTO 来回传，单独 SaveAIKey/GetAIKeyStatus 走 keyring。
type AIConfig struct {
	BaseURL string `json:"baseUrl"`
	Model   string `json:"model"`
}

// SaveAIKeyRequest 把 AI API key 写入 OS keyring。
type SaveAIKeyRequest struct {
	APIKey string `json:"apiKey" mask:"true"`
}

// AIKeyStatus 当前 keyring 是否已配置 AI API key（只暴露布尔，不回传 key 本身）。
type AIKeyStatus struct {
	Configured bool `json:"configured"`
}

// Language 表示"支持的构建语言"选项，用于前端下拉。
type Language struct {
	Language    string `json:"language"`
	DisplayName string `json:"displayName"`
}
