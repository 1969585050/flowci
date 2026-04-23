---
name: "project-standards"
description: "Creates project standards including error codes, naming conventions, API specs. Invoke when setting up a new project or when user asks for coding standards."
---

# Project Standards Generator

Professional project standards template generator for consistent quality across projects.

## Core Modules

### 1. Error Code System

**Template** (`docs/standards/error-codes.md`):

```markdown
# Error Codes Specification

## Overview
All API responses follow this format:
\`\`\`json
{
  "code": 0,
  "message": "success",
  "data": null
}
\`\`\`

## Code Ranges

| Range | Category | HTTP Status |
|-------|----------|-------------|
| 0 | Success | 200 |
| 1001-1999 | Request Error | 400 |
| 2001-2999 | Business Error | 400/500 |
| 3001-3999 | System Error | 500 |
| 4001-4999 | External Error | 502/503 |

## Detailed Codes

### Request Errors (1001-1999)

| Code | Name | Description |
|------|------|-------------|
| 1001 | INVALID_PARAM | Missing or invalid parameter |
| 1002 | INVALID_FORMAT | Parameter format error |
| 1003 | NOT_FOUND | Resource not found |
| 1004 | DUPLICATE | Resource already exists |
| 1005 | UNSUPPORTED | Unsupported operation |

### Business Errors (2001-2999)

| Code | Name | Description |
|------|------|-------------|
| 2001 | EXTERNAL_SERVICE_ERROR | External service call failed |
| 2002 | BUILD_FAILED | Build operation failed |
| 2003 | DEPLOY_FAILED | Deployment failed |
| 2004 | VALIDATION_FAILED | Business validation failed |

### System Errors (3001-3999)

| Code | Name | Description |
|------|------|-------------|
| 3001 | INTERNAL_ERROR | Internal server error |
| 3002 | DATABASE_ERROR | Database operation failed |
| 3003 | CONFIG_ERROR | Configuration error |

### External Errors (4001-4999)

| Code | Name | Description |
|------|------|-------------|
| 4001 | REGISTRY_AUTH_FAILED | Docker registry auth failed |
| 4002 | NETWORK_ERROR | Network connectivity issue |
```

---

### 2. Naming Conventions

**Template** (`docs/standards/naming.md`):

```markdown
# Naming Conventions

## File Naming

| Language | Convention | Example |
|----------|-----------|---------|
| Go | snake_case | `image_builder.go` |
| TypeScript | camelCase | `buildTypes.ts` |
| Vue Components | PascalCase | `BuildView.vue` |
| Test Files | `*_test.go` | `builder_test.go` |
| Config Files | snake_case | `docker-compose.yml` |

## Code Naming

### Go

| Type | Convention | Example |
|------|------------|---------|
| Package | lowercase | `builder` |
| Function (exported) | PascalCase | `BuildImage()` |
| Function (private) | camelCase | `buildImage()` |
| Variable | camelCase | `imageTag` |
| Constant | PascalCase | `DefaultPort` |
| Interface | PascalCase | `Builder` |
| Error variable | `err` prefix | `err error` |

### TypeScript

| Type | Convention | Example |
|------|------------|---------|
| Variable | camelCase | `imageTag` |
| Function | camelCase | `buildImage()` |
| Class | PascalCase | `BuildService` |
| Interface | PascalCase | `IBuildConfig` |
| Type | PascalCase | `BuildStatus` |
| Enum | PascalCase | `BuildStatus` |
| Constant | UPPER_SNAKE | `API_BASE_URL` |

## Database Naming

| Object | Convention | Example |
|--------|-----------|---------|
| Table | snake_case, plural | `build_logs` |
| Column | snake_case | `created_at` |
| Index | `idx_` prefix | `idx_project_id` |
| Foreign Key | `fk_` prefix | `fk_project_id` |

## API Naming

| Endpoint | Convention | Example |
|----------|-----------|---------|
| URL Path | kebab-case, plural nouns | `/api/v1/builds` |
| Query Param | camelCase | `pageSize` |
| JSON Field | snake_case | `image_tag` |
| HTTP Method | uppercase | GET, POST, PUT, DELETE |
```

---

### 3. API Specification

**Template** (`docs/api/openapi.yaml`):

```yaml
openapi: 3.0.3
info:
  title: ${PROJECT_NAME} API
  description: API specification
  version: 1.0.0

servers:
  - url: http://localhost:${PORT}/api/v1
    description: Development server

paths:
  /health:
    get:
      summary: Health check
      responses:
        '200':
          description: OK

  /${resource}:
    get:
      summary: List ${resource}
      responses:
        '200':
          description: Success
    post:
      summary: Create ${resource}
      responses:
        '201':
          description: Created

  /${resource}/{id}:
    get:
      summary: Get ${resource} by ID
      parameters:
        - name: id
          in: path
          required: true
      responses:
        '200':
          description: Success
    put:
      summary: Update ${resource}
      responses:
        '200':
          description: Success
    delete:
      summary: Delete ${resource}
      responses:
        '200':
          description: Success
```

---

### 4. Git Workflow

**Template** (`docs/standards/git-workflow.md`):

```markdown
# Git Workflow

## Branch Strategy

\`\`\`
main (production)
  └── develop
        ├── feature/xxx
        ├── fix/xxx
        └── release/vx.x.x
\`\`\`

## Branch Naming

| Type | Pattern | Example |
|------|---------|---------|
| Feature | `feature/<id>-<description>` | `feature/123-add-login` |
| Bug Fix | `fix/<id>-<description>` | `fix/456-fix-timeout` |
| Hotfix | `hotfix/<description>` | `hotfix/critical-security` |
| Release | `release/v<version>` | `release/v1.0.0` |
| Chore | `chore/<description>` | `chore/update-deps` |

## Commit Message Format

\`\`\`
<type>(<scope>): <subject>

<body>

<footer>
\`\`\`

### Type

| Type | Description |
|------|-------------|
| feat | New feature |
| fix | Bug fix |
| docs | Documentation |
| style | Formatting |
| refactor | Code refactoring |
| perf | Performance |
| test | Adding tests |
| chore | Build/tooling |

### Examples

\`\`\`bash
feat(auth): add JWT token refresh

-implement token refresh endpoint
-add refresh token to cookie
-update auth middleware

Closes #123
\`\`\`

## PR Requirements

- [ ] Title follows format
- [ ] Description filled
- [ ] Linked issue
- [ ] CI passes
- [ ] Code reviewed
- [ ] Tests added/updated
- [ ] Documentation updated
```

---

### 5. Testing Standards

**Template** (`docs/standards/testing.md`):

```markdown
# Testing Standards

## Coverage Requirements

| Module | Min Coverage |
|--------|--------------|
| API Handlers | 80% |
| Business Logic | 90% |
| Utilities | 80% |
| Critical Paths | 100% |

## Test Naming

| Type | Convention | Example |
|------|------------|---------|
| Unit Test | `Test<FunctionName>` | `TestBuildImage` |
| Test Case | `Test<FunctionName>_<Scenario>` | `TestBuildImage_WithInvalidPath` |
| Benchmark | `Benchmark<FunctionName>` | `BenchmarkBuildImage` |

## Test Structure (Go)

\`\`\`go
func Test<Subject>_<Scenario>(t *testing.T) {
    // Arrange
    input := setupInput()
    
    // Act
    result, err := SubjectUnderTest(input)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
\`\`\`

## Test Structure (TypeScript)

\`\`\`typescript
describe('Subject', () => {
  describe('scenario', () => {
    it('should do something', () => {
      // Arrange
      const input = setupInput();
      
      // Act
      const result = subjectUnderTest(input);
      
      // Assert
      expect(result).toEqual(expected);
    });
  });
});
\`\`\`
```

---

### 6. CI/CD Pipeline

**Template** (`docs/standards/cicd.md`):

```markdown
# CI/CD Pipeline

## Stages

### 1. Lint
\`\`\`yaml
lint:
  script:
    - go fmt ./...
    - golangci-lint run
    - npm run lint
\`\`\`

### 2. Test
\`\`\`yaml
test:
  script:
    - go test -v -cover ./...
    - npm run test:unit
    - npm run test:e2e
\`\`\`

### 3. Build
\`\`\`yaml
build:
  script:
    - go build -o app
    - npm run build
    - cargo build --release
\`\`\`

### 4. Deploy
\`\`\`yaml
deploy:
  stage: deploy
  only:
    - main
    - tags
  script:
    - deploy.sh
\`\`\`
```

---

## Usage

### Generate All Standards

```
/standards create --project-type "web-service" --language "go-typescript"
```

### Create Specific Standard

```
/standards create-error-codes
/standards create-naming-conventions
/standards create-git-workflow
/standards create-testing-standards
```

### Validate Standards Compliance

```
/standards validate --file "src/handler.go"
```

---

## Checklist for New Project

- [ ] Error codes document created
- [ ] Naming conventions documented
- [ ] API specification (OpenAPI) created
- [ ] Git workflow defined
- [ ] Testing standards established
- [ ] CI/CD pipeline configured
- [ ] CONTRIBUTING.md created
- [ ] README.md with setup instructions
