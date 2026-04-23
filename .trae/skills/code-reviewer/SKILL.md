---
name: "code-reviewer"
description: "Reviews code for standards compliance, best practices, and potential issues. Invoke when user asks for code review, before merge, or during PR review."
---

# Code Reviewer

Professional code review skill with focus on standards compliance and best practices.

## Review Scope

### 1. Code Style Compliance

**Go Checkpoints**:
- [ ] Follows `gofmt` formatting
- [ ] Variable names are descriptive (no `x`, `tmp`)
- [ ] Functions are appropriately sized (<50 lines ideal)
- [ ] No nested depth > 3 levels
- [ ] Error handling on every error path
- [ ] No `panic` in production code
- [ ] Context properly propagated

**TypeScript Checkpoints**:
- [ ] No `any` type usage
- [ ] Strict null checks
- [ ] Proper async/await usage
- [ ] No unused variables
- [ ] Components use PascalCase
- [ ] Imports sorted correctly

### 2. Security Review

**Common Issues**:
| Issue | Severity | Fix |
|-------|----------|-----|
| SQL Injection | Critical | Use parameterized queries |
| XSS | High | Sanitize user input |
| Hardcoded secrets | Critical | Use environment variables |
| Insecure random | Medium | Use crypto/rand |
| Missing auth check | High | Verify permissions |
| Weak crypto | High | Use standard libraries |

### 3. Performance Review

**Checkpoints**:
- [ ] No N+1 queries
- [ ] Proper indexing (database)
- [ ] Streaming for large responses
- [ ] Caching where appropriate
- [ ] Connection pooling
- [ ] No memory leaks

### 4. Error Handling

**Best Practices**:
```go
// ❌ Bad
func doSomething() error {
    return nil
}

// ✅ Good
func doSomething() error {
    if err := validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    return nil
}
```

```typescript
// ❌ Bad
async function fetch() {
    const data = await getData();
    return data;
}

// ✅ Good
async function fetch(): Promise<Data> {
    try {
        const data = await getData();
        return data;
    } catch (error) {
        throw new Error(`Fetch failed: ${error}`);
    }
}
```

---

## Review Checklist

### Before Review

- [ ] Code compiles/ builds
- [ ] Tests pass locally
- [ ] No lint errors
- [ ] Self-review done

### During Review

#### Logic & correctness
- [ ] Business logic is correct
- [ ] Edge cases handled
- [ ] Boundary conditions checked
- [ ] No off-by-one errors

#### Security
- [ ] Input validation
- [ ] Output sanitization
- [ ] Authentication/authorization
- [ ] Secrets management

#### Error Handling
- [ ] Errors not swallowed
- [ ] Meaningful error messages
- [ ] Errors logged appropriately
- [ ] Graceful degradation

#### Performance
- [ ] No unnecessary allocations
- [ ] Proper use of concurrency
- [ ] Database queries optimized
- [ ] Caching considered

#### Maintainability
- [ ] Code is self-documenting
- [ ] Complex logic commented
- [ ] No duplication
- [ ] Single responsibility

---

## Review Report Template

```markdown
## Code Review Report

**Date**: YYYY-MM-DD
**Reviewer**: [Name]
**Author**: [Author]
**PR/Commit**: [Link]

### Summary
<!-- One paragraph summary of the change -->

### Files Reviewed
- `src/handler.go`
- `src/service.go`

### Issues Found

| Severity | File | Line | Issue | Suggestion |
|----------|------|------|-------|------------|
| 🔴 Critical | handler.go | 42 | SQL injection | Use parameterized query |
| 🟡 Medium | service.go | 88 | Missing error check | Add nil check |
| 🟢 Low | utils.go | 15 | Consider using | Extract to constant |

### Approval Status

| Criterion | Status |
|-----------|--------|
| Correctness | ✅ Pass |
| Security | ⚠️ Issues |
| Performance | ✅ Pass |
| Style | ✅ Pass |
| Tests | ✅ Pass |

### Verdict
- [ ] **Approved** - Can merge
- [ ] **Request Changes** - Fix issues and re-review
- [ ] **Blocked** - Major problems

### Comments
<!-- Detailed comments for author -->
```

---

## Severity Levels

| Level | Symbol | Description | Action |
|-------|--------|-------------|--------|
| Critical | 🔴 | Security vulnerability, data loss risk | Must fix before merge |
| High | 🟠 | Could cause bugs, outages | Should fix before merge |
| Medium | 🟡 | Code smell, maintainability issue | Consider fixing |
| Low | 🟢 | Style, preferences | Nitpick, optional |

---

## Quick Review Commands

### Go
```bash
# Format check
gofmt -s -l .
gofmt -s -l . | grep -v vendor

# Lint
golangci-lint run ./...

# Vet
go vet ./...

# Static analysis
staticcheck ./...
```

### TypeScript
```bash
# ESLint
npm run lint

# Type check
npm run typecheck

# Import order
eslint --fix src/**/*.ts
```

---

## Usage

### Review a PR
```
/review pr --url "https://github.com/org/repo/pull/123"
```

### Review specific files
```
/review files --pattern "src/**/*.go"
```

### Review changes since last commit
```
/review diff --since "HEAD~1"
```

### Quick style check
```
/review style --language "go"
```

---

## Integration

This skill should be invoked:
1. **Before merge** - Required approval
2. **During PR creation** - Automated checks
3. **During development** - Self-review

Combine with `/code-quality-gate` for complete validation.
