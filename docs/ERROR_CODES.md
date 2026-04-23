# FlowCI Error Codes Specification

## Error Code Overview

FlowCI 使用统一的错误码体系，所有 API 响应遵循以下格式：

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

## Error Code Ranges

| Range | Category | Description |
|-------|----------|-------------|
| 0 | Success | 操作成功 |
| 1000-1999 | Request Error | 请求参数错误 |
| 2000-2999 | Business Error | 业务逻辑错误 |
| 3000-3999 | System Error | 系统级错误 |
| 4000-4999 | External Error | 外部服务错误 |

---

## Request Errors (1000-1999)

### 1001 - Invalid Parameter

**HTTP Status**: 400 Bad Request

**Description**: 请求参数无效或缺失

**Example**:
```json
{
  "code": 1001,
  "message": "Invalid parameter: project_id is required",
  "data": {
    "field": "project_id",
    "reason": "required"
  }
}
```

### 1002 - Invalid Format

**HTTP Status**: 400 Bad Request

**Description**: 参数格式错误

**Example**:
```json
{
  "code": 1002,
  "message": "Invalid format: image_tag must match registry.example.com/namespace:tag",
  "data": {
    "field": "image_tag",
    "expected": "registry格式"
  }
}
```

### 1003 - Resource Not Found

**HTTP Status**: 404 Not Found

**Description**: 请求的资源不存在

**Example**:
```json
{
  "code": 1003,
  "message": "Project not found: proj-abc123",
  "data": {
    "resource_type": "project",
    "resource_id": "proj-abc123"
  }
}
```

### 1004 - Duplicate Resource

**HTTP Status**: 409 Conflict

**Description**: 资源已存在

**Example**:
```json
{
  "code": 1004,
  "message": "Project already exists: my-project",
  "data": {
    "resource_type": "project",
    "name": "my-project"
  }
}
```

### 1005 - Unsupported Operation

**HTTP Status**: 400 Bad Request

**Description**: 不支持的操作

**Example**:
```json
{
  "code": 1005,
  "message": "Unsupported operation: rollback is not supported for single-container deployment",
  "data": {
    "operation": "rollback",
    "deployment_type": "single"
  }
}
```

---

## Business Errors (2000-2999)

### 2001 - Docker Connection Failed

**HTTP Status**: 503 Service Unavailable

**Description**: 无法连接到 Docker daemon

**Example**:
```json
{
  "code": 2001,
  "message": "Docker connection failed: cannot connect to docker daemon",
  "data": {
    "reason": "connection_refused",
    "socket": "/var/run/docker.sock"
  }
}
```

### 2002 - Build Failed

**HTTP Status**: 500 Internal Server Error

**Description**: Docker 镜像构建失败

**Example**:
```json
{
  "code": 2002,
  "message": "Build failed: exit code 1",
  "data": {
    "build_id": "build-xyz789",
    "exit_code": 1,
    "stage": "RUN pip install -r requirements.txt",
    "logs": ["ERROR: Could not find package..."]
  }
}
```

### 2003 - Deploy Failed

**HTTP Status**: 500 Internal Server Error

**Description**: 容器部署失败

**Example**:
```json
{
  "code": 2003,
  "message": "Deploy failed: container exited with code 1",
  "data": {
    "deploy_id": "deploy-abc123",
    "container_id": "def456",
    "exit_code": 1,
    "reason": "health_check_failed"
  }
}
```

### 2004 - Image Not Found

**HTTP Status**: 404 Not Found

**Description**: 指定的镜像不存在

**Example**:
```json
{
  "code": 2004,
  "message": "Image not found: registry.example.com/namespace/app:v1.0.0",
  "data": {
    "image_tag": "registry.example.com/namespace/app:v1.0.0"
  }
}
```

### 2005 - Container Not Running

**HTTP Status**: 500 Internal Server Error

**Description**: 容器未运行

**Example**:
```json
{
  "code": 2005,
  "message": "Container not running: myapp",
  "data": {
    "container_name": "myapp",
    "state": "exited"
  }
}
```

### 2006 - Rollback Failed

**HTTP Status**: 500 Internal Server Error

**Description**: 回滚失败

**Example**:
```json
{
  "code": 2006,
  "message": "Rollback failed: no previous deployment found",
  "data": {
    "deploy_id": "deploy-current",
    "reason": "no_history"
  }
}
```

---

## System Errors (3000-3999)

### 3001 - Internal Server Error

**HTTP Status**: 500 Internal Server Error

**Description**: 服务器内部错误

**Example**:
```json
{
  "code": 3001,
  "message": "Internal server error",
  "data": {
    "request_id": "req-abc123",
    "stack_trace": "..."
  }
}
```

### 3002 - Database Error

**HTTP Status**: 500 Internal Server Error

**Description**: 数据库操作错误

**Example**:
```json
{
  "code": 3002,
  "message": "Database error: failed to save project",
  "data": {
    "operation": "save",
    "table": "projects",
    "reason": "connection_timeout"
  }
}
```

### 3003 - Configuration Error

**HTTP Status**: 500 Internal Server Error

**Description**: 配置错误

**Example**:
```json
{
  "code": 3003,
  "message": "Configuration error: registry credentials not found",
  "data": {
    "config_key": "registries.aliyun.credentials",
    "reason": "not_configured"
  }
}
```

---

## External Errors (4000-4999)

### 4001 - Registry Authentication Failed

**HTTP Status**: 401 Unauthorized

**Description**: 镜像仓库认证失败

**Example**:
```json
{
  "code": 4001,
  "message": "Registry authentication failed",
  "data": {
    "registry": "registry.example.com",
    "reason": "invalid_credentials"
  }
}
```

### 4002 - Registry Connection Failed

**HTTP Status**: 503 Service Unavailable

**Description**: 无法连接到镜像仓库

**Example**:
```json
{
  "code": 4002,
  "message": "Registry connection failed: connection timeout",
  "data": {
    "registry": "registry.example.com",
    "reason": "connection_timeout"
  }
}
```

### 4003 - Network Error

**HTTP Status**: 503 Service Unavailable

**Description**: 网络错误

**Example**:
```json
{
  "code": 4003,
  "message": "Network error: failed to reach docker hub",
  "data": {
    "target": "docker.io",
    "reason": "dns_resolution_failed"
  }
}
```

---

## Error Response Format

### Standard Error Response

```json
{
  "code": 2002,
  "message": "Build failed: exit code 1",
  "data": {
    "build_id": "build-xyz789",
    "exit_code": 1,
    "stage": "RUN pip install",
    "logs": ["ERROR: Could not find package"]
  }
}
```

### Validation Error Response

```json
{
  "code": 1001,
  "message": "Validation failed",
  "data": {
    "errors": [
      {
        "field": "image_tag",
        "message": "image_tag is required"
      },
      {
        "field": "language",
        "message": "language must be one of: go, java-maven, java-gradle, nodejs, python, php, ruby, dotnet"
      }
    ]
  }
}
```

---

## Client-Side Error Handling

### TypeScript Error Types

```typescript
export class FlowCIError extends Error {
  constructor(
    public code: number,
    message: string,
    public data?: unknown
  ) {
    super(message);
    this.name = 'FlowCIError';
  }
}

export class RequestError extends FlowCIError {
  constructor(message: string, data?: unknown) {
    super(1001, message, data);
    this.name = 'RequestError';
  }
}

export class BuildError extends FlowCIError {
  constructor(message: string, data?: unknown) {
    super(2002, message, data);
    this.name = 'BuildError';
  }
}

export class DeployError extends FlowCIError {
  constructor(message: string, data?: unknown) {
    super(2003, message, data);
    this.name = 'DeployError';
  }
}
```

### Error Handler Example

```typescript
async function handleApiError(response: Response) {
  const { code, message, data } = await response.json();

  switch (code) {
    case 1001:
      throw new RequestError(message, data);
    case 2002:
      throw new BuildError(message, data);
    case 2003:
      throw new DeployError(message, data);
    default:
      throw new FlowCIError(code, message, data);
  }
}
```

---

## Log Level Mapping

| Error Code Range | Log Level | Action |
|------------------|-----------|--------|
| 0 | INFO | Success operation |
| 1001-1005 | WARN | Client request issue |
| 2001-2006 | ERROR | Business logic failure |
| 3001-3003 | ERROR | System failure |
| 4001-4003 | WARN | External service issue |
