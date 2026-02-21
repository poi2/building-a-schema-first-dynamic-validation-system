# Development Guide

This document is a guide for developers contributing to this project.

## Table of Contents

- [Development Environment Setup](#development-environment-setup)
- [Git Hooks](#git-hooks)
- [Commit Conventions](#commit-conventions)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [PR Creation and Review](#pr-creation-and-review)

## Development Environment Setup

### Required Tools

Please install the following tools:

- **Docker & Docker Compose**: Container environment
- **Node.js 20+**: For BFF/FE development
- **Go 1.24+**: For Backend/ISR development
- **Buf CLI**: Proto code generation
  - Installation: <https://docs.buf.build/installation>

### Initial Setup

```bash
# 1. Clone the repository
git clone https://github.com/poi2/building-a-schema-first-dynamic-validation-system.git
cd building-a-schema-first-dynamic-validation-system

# 2. Set up Git hooks (Important!)
bash .github/git-hooks/setup-hooks.sh

# 3. Install dependencies
npm install

# 4. Generate Proto code
make proto-generate

# 5. Start Docker services
docker compose up -d

# 6. Verify services are running
docker compose ps
```

### Directory Structure

```text
.
├── .github/
│   ├── git-hooks/           # Git hooks (commit-msg, setup script)
│   └── workflows/           # GitHub Actions CI/CD
├── docs/                    # Design documents
├── proto/                   # Proto definitions (Single Source of Truth)
├── pkg/gen/                 # Generated code (shared modules)
│   ├── go/                  # Go generated code
│   └── ts/                  # TypeScript generated code
└── services/
    ├── isr/                 # Internal Schema Registry (Go)
    ├── be/                  # Backend service (Go) - planned
    ├── bff/                 # Backend for Frontend (Node.js) - planned
    └── fe/                  # Frontend (React) - planned
```

## Git Hooks

### Setup

**Please run this first:**

```bash
bash .github/git-hooks/setup-hooks.sh
```

This script copies `.github/git-hooks/commit-msg` to `.git/hooks/` and grants execute permissions.

### commit-msg Hook

Checks the following during commit:

1. **Single-line rule**: Commit messages must be single-line
2. **Conventional Commits**: Must follow the specified format

Commits will be rejected if these rules are violated.

## Commit Conventions

### Conventional Commits

This project adopts [Conventional Commits](https://www.conventionalcommits.org/).

#### Format

```text
<type>(<scope>): <description>
```

- **type**: Required
- **scope**: Optional
- **description**: Required (brief and concise)

#### Type List

| Type | Description | Example |
|------|-------------|---------|
| `feat` | New feature | `feat: Add protovalidate interceptor` |
| `fix` | Bug fix | `fix: Update Go version to 1.23` |
| `docs` | Documentation changes | `docs: Add developer guide` |
| `style` | Code style changes (no behavior changes) | `style: Format code with gofmt` |
| `refactor` | Refactoring | `refactor: Extract test helper function` |
| `perf` | Performance improvements | `perf: Optimize schema validation` |
| `test` | Test additions/modifications | `test: Add validation integration tests` |
| `build` | Build system changes | `build: Update Dockerfile` |
| `ci` | CI configuration changes | `ci: Add linter workflow` |
| `chore` | Other changes | `chore: Update dependencies` |
| `revert` | Revert | `revert: Revert "feat: Add feature"` |

#### Examples

##### ✅ Good

```bash
git commit -m "feat: Add protovalidate interceptor to ISR service"
git commit -m "fix: Update Go version to 1.23 for Docker compatibility"
git commit -m "test: Add validation tests for 10MB size limit"
git commit -m "docs: Update setup instructions in README"
```

##### ❌ Bad

```bash
# Multiple lines are not allowed
git commit -m "feat: Add feature

This is a detailed description"

# Missing type
git commit -m "Added new feature"

# Incorrect format
git commit -m "feat Add feature"  # Missing colon
git commit -m "feature: Add feature"  # Incorrect type
```

## Development Workflow

### Branch Strategy

1. **main**: Stable branch corresponding to production
2. **feature/***: Feature addition branches
3. **fix/***: Bug fix branches
4. **docs/***: Documentation update branches

### Work Flow

```bash
# 1. Update main branch
git checkout main
git pull

# 2. Create working branch
git checkout -b feat/add-new-feature

# 3. Code changes and testing
# ... development work ...
make test

# 4. Commit (git hook automatically validates)
git add .
git commit -m "feat: Add new feature"

# 5. Push
git push -u origin feat/add-new-feature

# 6. Create PR
gh pr create --title "feat: Add new feature" --body "..."
```

### CI/CD Checks

When a PR is created, the following checks run automatically:

- **Proto Lint**: `buf lint`
- **Go Test**: `go test -v -race -coverprofile=coverage.out ./...`
- **Go Lint**: `go vet` + `staticcheck`
- **Build Docker Image**: `docker compose build isr`

All checks must pass before merging.

## Coding Standards

### Go

- **Format**: Use `gofmt`
- **Linter**: `go vet` + `staticcheck`
- **Naming conventions**: Follow Go standards
  - Exported: PascalCase
  - Unexported: camelCase
- **Error handling**: Properly wrap and return errors

```go
// Good
if err != nil {
    return fmt.Errorf("failed to upload schema: %w", err)
}

// Bad
if err != nil {
    return err  // Missing context
}
```

### TypeScript

- **Format**: Prettier
- **Linter**: ESLint
- **Naming conventions**:
  - Variables/functions: camelCase
  - Types/interfaces: PascalCase
  - Constants: UPPER_SNAKE_CASE

### Proto

- **Style Guide**: [Buf Style Guide](https://buf.build/docs/best-practices/style-guide)
- **Lint**: Run `buf lint`
- **Naming conventions**:
  - Message: PascalCase
  - Field: snake_case
  - Service: PascalCase
  - RPC: PascalCase

## Testing

### Go

#### Unit Tests

```go
func TestSchemaHandler_UploadSchema_Success(t *testing.T) {
    // Arrange
    mockRepo := &mockSchemaRepository{...}
    handler := NewSchemaHandler(mockRepo)

    // Act
    resp, err := handler.UploadSchema(ctx, req)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Msg.Version != "1.0.0" {
        t.Errorf("got %v, want 1.0.0", resp.Msg.Version)
    }
}
```

#### Integration Tests

```go
func TestUploadSchema_ValidationError(t *testing.T) {
    // Start server with httptest.NewServer
    client, cleanup := newTestClient(t, handler)
    defer cleanup()

    // Run test
    _, err := client.UploadSchema(ctx, req)

    // Verify error code
    var connectErr *connect.Error
    if !errors.As(err, &connectErr) {
        t.Fatalf("expected connect.Error")
    }
}
```

#### Running Tests

```bash
# Run all tests
go test ./...

# With coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Specific package only
go test ./internal/handler -v
```

## PR Creation and Review

### Pre-PR Checklist

- [ ] All tests pass (`make test`)
- [ ] Code is formatted (`make fmt`)
- [ ] No lint errors (`make lint`)
- [ ] All CI checks pass
- [ ] Commit messages follow Conventional Commits

### PR Template

```markdown
## Summary
Brief overview of this change

## Changes
- Change 1
- Change 2

## Test Plan
- [ ] Test item 1
- [ ] Test item 2

## Related Issues
- Closes #123
```

### Review Process

1. **Self-review**: Re-check your code after creating the PR
2. **CI checks**: Fix until all CI checks are green
3. **Copilot Review**: Address automated review comments
4. **Merge**: Merge to main after approval

### PR Size Guidelines

- **Small**: ~100 lines (ideal)
- **Medium**: 100-300 lines
- **Large**: 300-500 lines (should be avoided)
- **Too Large**: 500+ lines (consider splitting)

It's recommended to split large features into multiple PRs.

## Troubleshooting

### Git hook not working

```bash
# Re-setup hooks
bash .github/git-hooks/setup-hooks.sh

# Check permissions
ls -la .git/hooks/commit-msg

# Expected output: -rwxr-xr-x
```

### Proto code generation errors

```bash
# Check Buf version
buf --version

# Clear cache and regenerate
rm -rf pkg/gen/
make proto-generate
```

### Docker services won't start

```bash
# Check logs
docker compose logs

# Clean up and restart
make docker-clean
make docker-up
```

## References

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Buf Documentation](https://buf.build/docs)
- [Connect Documentation](https://connectrpc.com/docs/introduction)
- [protovalidate](https://github.com/bufbuild/protovalidate)
- [GitHub Flow](https://docs.github.com/en/get-started/quickstart/github-flow)
