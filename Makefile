.PHONY: test
test:
	go test ./... -v -count=1 --json > test-report.json

.PHONY: test-unit
test-unit:
	go test ./pkg/... -v -count=1 -race

.PHONY: test-integration
test-integration:
	go test ./tests/integration/... -v --tags=integration -count=1

.PHONY: build
build:
	docker compose up -d

.PHONY: clean
clean:
	docker compose down --rmi all --volumes

.PHONY: code-gen
code-gen:
	go generate ./...

.PHONY: schema-gen
schema-gen:
	go run script/avsc2json/main.go schema/avro/expense.avsc > docker/schema/expense.json

.PHONY: build-cli
build-cli:
	go build -o bin/eventshark-cli ./cmd/cli

.PHONY: install-deps
install-deps:
	go mod download
	go mod tidy

.PHONY: run
run:
	go run cmd/main.go

.PHONY: help
help:
	@echo "EventShark Makefile Commands:"
	@echo "  make build            - Start Docker services"
	@echo "  make clean            - Stop and remove Docker services"
	@echo "  make test             - Run all tests (unit tests only)"
	@echo "  make test-unit        - Run unit tests with race detection"
	@echo "  make test-integration - Run integration tests (requires running server)"
	@echo "  make code-gen         - Generate code from schemas"
	@echo "  make schema-gen       - Generate JSON schemas"
	@echo "  make build-cli        - Build CLI tool"
	@echo "  make install-deps     - Install Go dependencies"
	@echo "  make run              - Run EventShark server"
	@echo "  make help             - Show this help message"