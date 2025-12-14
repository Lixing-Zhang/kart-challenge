# GitHub Actions CI/CD

This repository uses GitHub Actions for continuous integration and deployment.

## Workflows

### Test Workflow (`.github/workflows/test.yml`)

Runs on every push and pull request to `main` branch.

**Jobs:**

1. **Unit Tests**
   - Runs all unit tests with race detector
   - Generates coverage report
   - Uploads coverage to Codecov (optional)
   - Fast execution (~30 seconds)

2. **Integration Tests**
   - Downloads real S3 coupon files (1.8GB)
   - Tests with actual production data
   - 15-minute timeout for large file downloads
   - Continues even if first test fails (for robustness)

3. **Build**
   - Compiles the server binary
   - Uploads artifact for download
   - Validates that code compiles successfully

### Lint Workflow (`.github/workflows/lint.yml`)

Ensures code quality and consistency.

**Jobs:**

1. **golangci-lint** - Comprehensive Go linting
2. **gofmt** - Code formatting check
3. **go vet** - Static analysis

## Running Locally

### Unit Tests Only
```bash
cd backend-challenge
go test -short ./...
```

### All Tests (Including Integration)
```bash
cd backend-challenge
go test -v -timeout=15m ./...
```

### Specific Integration Test
```bash
cd backend-challenge
go test -v -timeout=15m -run TestValidator_RealS3Files_Sample ./internal/coupon/
```

### Linting
```bash
cd backend-challenge
go vet ./...
gofmt -l .
```

### Build
```bash
cd backend-challenge
go build -o bin/server ./cmd/server
```

## Badges

Add these badges to your README:

```markdown
[![Test](https://github.com/Lixing-Zhang/kart-challenge/actions/workflows/test.yml/badge.svg)](https://github.com/Lixing-Zhang/kart-challenge/actions/workflows/test.yml)
[![Lint](https://github.com/Lixing-Zhang/kart-challenge/actions/workflows/lint.yml/badge.svg)](https://github.com/Lixing-Zhang/kart-challenge/actions/workflows/lint.yml)
```

## Notes

- Integration tests may take 2-5 minutes due to large file downloads
- Unit tests complete in under 30 seconds
- All tests run in parallel where possible
- Coverage reports are generated automatically
