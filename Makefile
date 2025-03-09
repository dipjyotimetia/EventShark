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