test:
	go test ./... -v --tags=integration

build:
	docker compose up -d

clean:
	docker compose down --rmi all --volumes

code-gen:
	go generate ./...

schema-gen:
	go run script/avsc2json/main.go schema/avro/expense.avsc > docker/schema/expense.json