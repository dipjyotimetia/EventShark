.PHONY: test
test:
	go test ./... -v --tags=integration -count=1 --json > test-report.json

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
	@echo "  make build         - Start Docker services"
	@echo "  make clean         - Stop and remove Docker services"
	@echo "  make test          - Run tests"
	@echo "  make code-gen      - Generate code from schemas"
	@echo "  make schema-gen    - Generate JSON schemas"
	@echo "  make build-cli     - Build CLI tool"
	@echo "  make install-deps  - Install Go dependencies"
	@echo "  make run           - Run EventShark server"
	@echo "  make help          - Show this help message"