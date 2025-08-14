.PHONY: test test-unit test-integration
test: test-unit test-integration

test-unit:
	go test -v -race -timeout=30s ./pkg/...

test-integration:
	go test -v -timeout=60s --tags=integration ./tests/integration/...

.PHONY: build
build:
	docker compose up -d

.PHONY: clean
clean:
	docker compose down --rmi all --volumes

.PHONY: code-gen
code-gen:
	go generate ./...

.PHONY: test-k6 run lint fmt schema-gen
test-k6:
	cd tests/performance && npm run test

run:
	go run ./cmd/main.go

lint:
	golangci-lint run

fmt:
	go fmt ./...

.PHONY: schema-gen
schema-gen:
	go run script/avsc2json/main.go schema/avro/expense.avsc > docker/schema/expense.json