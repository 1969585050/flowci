---
name: "code-quality-gate"
description: "Code quality validator with lint, typecheck, and test requirements. Invoke before any code merge or when user asks to validate code quality."
---

# Code Quality Gate

Automated code quality validation system.

## Quality Requirements

### Must Pass Before Merge

| Check | Tool | Requirement |
|-------|------|-------------|
| Lint | golangci-lint / ESLint | 0 errors |
| Type Check | go vet / tsc --noEmit | 0 errors |
| Unit Tests | go test / vitest | ≥80% coverage |
| Security | trivy / snyk | 0 critical issues |

### Go Quality Gates

```bash
# Lint
golangci-lint run ./...

# Format check
gofmt -s -l .
test -z "$(gofmt -s -l .)" || exit 1

# Vet
go vet ./...

# Test with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total
```

### TypeScript Quality Gates

```bash
# ESLint
npm run lint

# Type check
npm run typecheck

# Unit tests
npm run test:unit -- --coverage

# E2E tests
npm run test:e2e
```

---

## Scoring System

| Score | Grade | Action |
|-------|-------|--------|
| 95-100 | A | Excellent, can merge |
| 85-94 | B | Good, minor suggestions |
| 70-84 | C | Needs improvement |
| <70 | F | Block merge |

### Score Components

| Component | Weight | Measures |
|-----------|--------|----------|
| Correctness | 30% | Tests pass, no panics |
| Maintainability | 25% | Code complexity, duplication |
| Performance | 20% | Efficient algorithms |
| Security | 15% | No vulnerabilities |
| Documentation | 10% | Code comments, docs |

---

## Common Issues & Fixes

### Go Issues

| Issue | Fix |
|-------|-----|
| Error not handled | Use `_ =` or proper handling |
| Missing error return | Add `if err != nil { return err }` |
| Uncleared mutex | Defer unlock |
| SQL injection | Use parameterized queries |
| Password in code | Use environment variables |

### TypeScript Issues

| Issue | Fix |
|-------|-----|
| `any` type | Use specific type or `unknown` |
| Missing null check | Add optional chaining or null check |
| Unused variable | Remove or prefix with `_` |
| Memory leak | Clean up event listeners |
| XSS vulnerability | Sanitize user input |

---

## Usage

### Validate Single File

```
/quality check --file "src/handler.go"
```

### Validate Module

```
/quality check --module "internal/api"
```

### Full Project Validation

```
/quality gate --strict
```

### Git Hook Setup

```bash
# .git/hooks/pre-commit
#!/bin/bash
npm run lint
npm run typecheck
npm run test:unit
```

---

## Integration

- **GitHub Actions**: Runs on every PR
- **Pre-commit**: Runs before commit
- **Pre-push**: Runs before push
- **Manual**: `/quality gate --strict`

---

## Reports

Quality gate generates:

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "passed": true,
  "score": 92,
  "checks": [
    {"name": "lint", "passed": true, "errors": 0},
    {"name": "typecheck", "passed": true, "errors": 0},
    {"name": "test", "passed": true, "coverage": 85},
    {"name": "security", "passed": true, "issues": 0}
  ],
  "issues": [
    {"severity": "low", "file": "src/utils.go", "message": "Consider adding comment"}
  ]
}
```
