# CI Pipeline Update Instructions

Due to GitHub App permissions, the CI pipeline changes need to be applied manually.

## Current CI Issue

The current CI pipeline (`.github/workflows/tests.yml`) has a limitation:
- Runs all tests in a single job
- No separation between unit and integration tests
- No code coverage tracking
- No linting or build verification
- Limited error reporting

## Updated CI Pipeline

Replace the contents of `.github/workflows/tests.yml` with the following:

```yaml
name: Tests

on:
  pull_request:
    branches: [ master ]
  push:
    branches: [ master ]

permissions:
  pull-requests: write
  contents: read
  issues: write

jobs:
  unit-tests:
    name: Unit Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: ./go.mod
          cache: true

      - name: Install dependencies
        run: |
          go mod download
          go mod tidy

      - name: Run unit tests
        run: |
          go test -v -race -coverprofile=coverage.out -covermode=atomic ./pkg/... -json > unit-test-report.json

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          flags: unittests
          name: unit-test-coverage
        continue-on-error: true

      - name: Generate Unit Test Report
        if: always()
        uses: dipjyotimetia/gotest-report@main
        with:
          test-json-file: unit-test-report.json
          output-file: unit-test-report.md
          comment-pr: false
        continue-on-error: true

  integration-tests:
    name: Integration Tests
    runs-on: ubuntu-latest
    env:
      DOCKER_BUILDKIT: 1
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: ./go.mod
          cache: true

      - name: Install dependencies
        run: |
          go mod download
          go mod tidy

      - name: Start Docker services
        run: |
          docker compose up -d
          echo "Waiting for services to be ready..."
          sleep 30

      - name: Check service health
        run: |
          docker compose ps
          curl -f http://localhost:8083/health || (docker compose logs && exit 1)

      - name: Run integration tests
        run: |
          go test -v -tags=integration ./tests/integration/... -count=1 -json > integration-test-report.json

      - name: Show service logs on failure
        if: failure()
        run: |
          docker compose logs

      - name: Generate Integration Test Report
        if: always()
        uses: dipjyotimetia/gotest-report@main
        with:
          test-json-file: integration-test-report.json
          output-file: integration-test-report.md
          comment-pr: true
        continue-on-error: true

      - name: Cleanup
        if: always()
        run: docker compose down --rmi all --volumes

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: ./go.mod
          cache: true

      - name: Install dependencies
        run: go mod download

      - name: Run go vet
        run: go vet ./...

      - name: Run go fmt
        run: |
          if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
            echo "The following files need formatting:"
            gofmt -s -l .
            exit 1
          fi

  build:
    name: Build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: ./go.mod
          cache: true

      - name: Install dependencies
        run: go mod download

      - name: Build application
        run: go build -v ./cmd/main.go

      - name: Build CLI
        run: go build -v -o eventshark-cli ./cmd/cli/main.go

      - name: Upload CLI artifact
        uses: actions/upload-artifact@v3
        with:
          name: eventshark-cli
          path: eventshark-cli
```

## Key Improvements

### 1. Parallel Job Execution
- **4 jobs run in parallel**: unit-tests, integration-tests, lint, build
- Faster overall CI time (runs concurrently)
- Better isolation between test types

### 2. Unit Tests Job
- ✅ Runs all unit tests with race detection
- ✅ Generates code coverage reports
- ✅ Uploads to Codecov (optional)
- ✅ Fast feedback (<2 minutes)
- ✅ No Docker dependencies

### 3. Integration Tests Job
- ✅ Starts Docker compose services
- ✅ Health check verification before tests
- ✅ Runs only integration tests (--tags=integration)
- ✅ Captures service logs on failure
- ✅ Proper cleanup with volumes

### 4. Lint Job
- ✅ Static analysis with `go vet`
- ✅ Code formatting with `go fmt`
- ✅ Fast quality checks

### 5. Build Job
- ✅ Verifies main application builds
- ✅ Verifies CLI tool builds
- ✅ Uploads CLI as artifact
- ✅ Catches compilation errors early

### 6. Additional Features
- ✅ Go module caching for faster builds
- ✅ Runs on both PRs and pushes to master
- ✅ Separate test reports for unit and integration
- ✅ Detailed error reporting with logs
- ✅ Continue on error for non-critical steps

## How to Apply

### Option 1: Manual Update
1. Open `.github/workflows/tests.yml` in your repository
2. Replace the entire contents with the YAML above
3. Commit directly to master or create a new PR

### Option 2: Via GitHub Web Interface
1. Navigate to `.github/workflows/tests.yml` on GitHub
2. Click "Edit this file" (pencil icon)
3. Replace all contents with the YAML above
4. Commit changes with message: "ci: Update CI pipeline with parallel jobs and comprehensive testing"

### Option 3: Via Git Command Line
```bash
# Make sure you're on the correct branch
git checkout claude/analyze-project-features-011CUpNVY8s4efkN81zbDyRN

# Update the workflow file with the new content
# (paste the YAML content into .github/workflows/tests.yml)

git add .github/workflows/tests.yml
git commit -m "ci: Update CI pipeline with parallel jobs and comprehensive testing"
git push
```

## Testing the Pipeline

After applying the changes:

1. **Create a test PR** to verify all jobs run correctly
2. **Check the Actions tab** to see all 4 jobs running in parallel
3. **Verify coverage reports** are generated
4. **Check test reports** are commented on PR

## Expected CI Behavior

### On Pull Request
- All 4 jobs run in parallel
- Unit tests complete in ~2 minutes
- Integration tests complete in ~3 minutes
- Lint and build complete in ~1 minute
- Test report commented on PR

### On Push to Master
- Same as PR but without PR comments
- All artifacts saved for download

## Troubleshooting

### If unit tests fail
```bash
# Run locally
go test -v -race ./pkg/...
```

### If integration tests fail
```bash
# Check services are running
docker compose up -d
docker compose ps
curl http://localhost:8083/health

# Run integration tests
go test -v -tags=integration ./tests/integration/...
```

### If lint fails
```bash
# Check formatting
gofmt -s -l .

# Fix formatting
gofmt -s -w .

# Run static analysis
go vet ./...
```

### If build fails
```bash
# Try building locally
go build -v ./cmd/main.go
go build -v -o eventshark-cli ./cmd/cli/main.go
```

## Benefits Summary

| Aspect | Before | After |
|--------|--------|-------|
| Jobs | 1 (sequential) | 4 (parallel) |
| CI Time | ~5 minutes | ~3 minutes |
| Unit Tests | Mixed with integration | Separate, fast |
| Coverage | None | Tracked & reported |
| Lint | None | go vet + go fmt |
| Build Verification | Basic | Full with artifacts |
| Error Logs | Limited | Comprehensive |

## Support

If you encounter issues:
1. Check GitHub Actions logs
2. Review the test documentation in `docs/testing.md`
3. Run tests locally first to verify they pass
4. Ensure all dependencies are in go.mod

---

**Note**: This CI pipeline is designed to work with the new test suite added in the previous commits. Make sure those changes are merged first.
