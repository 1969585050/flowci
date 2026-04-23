# FlowCI API Specification

## Base URL

```
http://localhost:3847/api/v1
```

## Common Headers

```
Content-Type: application/json
Accept: application/json
```

## Common Response Format

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

## Error Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1001 | Invalid request parameter |
| 1002 | Resource not found |
| 2001 | Docker connection failed |
| 2002 | Build failed |
| 2003 | Deploy failed |
| 3001 | Internal server error |

---

## Build APIs

### Create Build

**POST** `/builds`

**Request Body:**

```json
{
  "project_id": "proj-001",
  "language": "go",
  "context_path": "/path/to/project",
  "image_tags": ["registry.example.com/namespace/app:v1.0.0"],
  "build_args": {
    "GO_VERSION": "1.21"
  },
  "no_cache": false,
  "pull_base_image": true
}
```

**Response:**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "build-uuid-123",
    "image_id": "sha256:abc123...",
    "tags": ["registry.example.com/namespace/app:v1.0.0"],
    "size": 12345678,
    "duration_ms": 45230,
    "status": "success",
    "logs": ["Step 1/5 : FROM golang:1.21-alpine", "..."],
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### Get Build Status

**GET** `/builds/:id`

**Response:**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "build-uuid-123",
    "image_id": "sha256:abc123...",
    "tags": ["registry.example.com/namespace/app:v1.0.0"],
    "size": 12345678,
    "duration_ms": 45230,
    "status": "success",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### Get Build Logs

**GET** `/builds/:id/logs`

**Response:**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "build_id": "build-uuid-123",
    "logs": [
      {"timestamp": "2024-01-15T10:30:00Z", "level": "info", "message": "Step 1/5 : FROM golang:1.21-alpine"},
      {"timestamp": "2024-01-15T10:30:01Z", "level": "info", "message": "..."}
    ]
  }
}
```

---

## Deploy APIs

### Deploy Application

**POST** `/deploys`

**Request Body:**

```json
{
  "project_id": "proj-001",
  "deployment_type": "compose",
  "image_tag": "registry.example.com/namespace/app:v1.0.0",
  "container_name": "myapp",
  "ports": [
    {"host_port": 8080, "container_port": 8080, "protocol": "tcp"}
  ],
  "env_vars": {
    "DATABASE_URL": "postgres://localhost:5432/app",
    "LOG_LEVEL": "debug"
  },
  "volumes": ["/host/path:/container/path"],
  "restart_policy": "unless-stopped",
  "replicas": 1
}
```

**Response:**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "deploy-uuid-456",
    "container_id": "abc123def456",
    "name": "myapp",
    "status": "running",
    "ports": [
      {"host_port": 8080, "container_port": 8080, "protocol": "tcp"}
    ],
    "created_at": "2024-01-15T11:00:00Z"
  }
}
```

---

### Get Deploy Status

**GET** `/deploys/:id`

**Response:**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "deploy-uuid-456",
    "container_id": "abc123def456",
    "name": "myapp",
    "status": "running",
    "ports": [
      {"host_port": 8080, "container_port": 8080, "protocol": "tcp"}
    ],
    "created_at": "2024-01-15T11:00:00Z"
  }
}
```

---

### Rollback Deploy

**POST** `/deploys/:id/rollback`

**Response:**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "rollback-uuid-789",
    "status": "rolled_back",
    "previous_image_tag": "registry.example.com/namespace/app:v0.9.0",
    "created_at": "2024-01-15T12:00:00Z"
  }
}
```

---

## Docker APIs

### Check Docker Connection

**GET** `/docker/check`

**Response:**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "connected": true,
    "version": "24.0.7",
    "api_version": "1.43",
    "os": "linux",
    "arch": "amd64"
  }
}
```

---

### List Images

**GET** `/docker/images`

**Response:**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "images": [
      {
        "id": "sha256:abc123...",
        "tags": ["registry.example.com/namespace/app:v1.0.0", "latest"],
        "size": 12345678,
        "created": "2024-01-15T10:30:00Z"
      }
    ],
    "total": 1
  }
}
```

---

### List Containers

**GET** `/docker/containers`

**Response:**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "containers": [
      {
        "id": "abc123def456",
        "names": ["/myapp"],
        "image": "registry.example.com/namespace/app:v1.0.0",
        "state": "running",
        "status": "Up 2 hours",
        "ports": [
          {"host_port": 8080, "container_port": 8080, "protocol": "tcp"}
        ],
        "created": "2024-01-15T10:30:00Z"
      }
    ],
    "total": 1
  }
}
```

---

## Project APIs

### List Projects

**GET** `/projects`

**Response:**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "projects": [
      {
        "id": "proj-001",
        "name": "My Application",
        "path": "/path/to/project",
        "language": "go",
        "created_at": "2024-01-15T10:00:00Z",
        "updated_at": "2024-01-15T10:00:00Z"
      }
    ],
    "total": 1
  }
}
```

---

### Create Project

**POST** `/projects`

**Request Body:**

```json
{
  "name": "My Application",
  "path": "/path/to/project",
  "language": "go",
  "build_config": {
    "dockerfile_path": "Dockerfile",
    "context_path": "."
  },
  "deploy_config": {
    "deployment_type": "compose",
    "container_name": "myapp"
  }
}
```

**Response:**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "proj-001",
    "name": "My Application",
    "path": "/path/to/project",
    "language": "go",
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-01-15T10:00:00Z"
  }
}
```

---

### Get Project

**GET** `/projects/:id`

**Response:**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "proj-001",
    "name": "My Application",
    "path": "/path/to/project",
    "language": "go",
    "build_config": {
      "dockerfile_path": "Dockerfile",
      "context_path": "."
    },
    "deploy_config": {
      "deployment_type": "compose",
      "container_name": "myapp"
    },
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-01-15T10:00:00Z"
  }
}
```

---

### Update Project

**PUT** `/projects/:id`

**Request Body:**

```json
{
  "name": "My Application Updated",
  "language": "go",
  "build_config": {
    "dockerfile_path": "Dockerfile.prod",
    "context_path": "."
  },
  "deploy_config": {
    "deployment_type": "compose",
    "container_name": "myapp-updated"
  }
}
```

**Response:**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "proj-001",
    "name": "My Application Updated",
    "updated_at": "2024-01-15T12:00:00Z"
  }
}
```

---

### Delete Project

**DELETE** `/projects/:id`

**Response:**

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## Health Check

### Get Health Status

**GET** `/health`

**Response:**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "healthy",
    "version": "0.1.0",
    "uptime_seconds": 3600,
    "docker_connected": true
  }
}
```
