# Event Shark: A Serverless Kafka Event Publisher Testing Framework

* Accelerate your Kafka-driven development: Event Shark makes it easy to test consumer applications in isolation.
* Gain confidence in your consumer code: Simulate real-world publisher behavior for comprehensive testing.
* Optimize performance: Run targeted performance tests on consumer applications without complex publisher dependencies.

<img src="./docs/assets/Architecture.png" width="800">

## Architecture
Event Shark is a serverless framework that exposes Kafka event publishing as simple JSON requests. It allows consumer applications to simulate publisher behavior and test their systems independently. The architecture consists of:
- A REST API server (Fiber, Go) for event publishing
- Kafka and Schema Registry (via Docker Compose)
- Utilities for schema conversion and consumer simulation

## Directory Structure
```
EventShark/
  cmd/                # Main application entrypoint
  docker/             # Docker scripts and schema JSONs
  docs/               # Documentation
  gen/                # Generated Go code from Avro schemas
  pkg/                # Application packages (config, events, handlers, routers)
  schema/             # Avro schemas and codegen
  script/             # Utilities (avsc2json, consumer)
  tests/              # Integration and performance tests
```

## Installation
See [docs/setup.md](docs/setup.md) for detailed setup instructions.

## Configuration
See [docs/configuration.md](docs/configuration.md) for environment variables and config file options.

## Usage
See [docs/usage.md](docs/usage.md) for API endpoints and examples.

## Utilities
- **AVSC to JSON Converter:** Convert Avro schemas to JSON for use with Schema Registry. See [script/avsc2json/readme.md](script/avsc2json/readme.md).
- **Consumer Script:** Example Kafka consumer in Go. See [script/consumer/main.go](script/consumer/main.go).

## Running Tests
See [docs/testing.md](docs/testing.md) for unit, integration, and performance testing instructions.

Test reports are generated as `test-report.json` and (in CI) as `test-report.md`.

## Building the Project
To build the project, use:
```sh
make build
```

## Contributing
See [docs/contributing.md](docs/contributing.md) for guidelines.
