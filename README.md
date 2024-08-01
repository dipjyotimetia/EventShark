# Event Shark: A Serverless Kafka Event Publisher Testing Framework

* Accelerate your Kafka-driven development: Event Shark makes it easy to test consumer applications in isolation.
* Gain confidence in your consumer code: Simulate real-world publisher behavior for comprehensive testing.
* Optimize performance: Run targeted performance tests on consumer applications without complex publisher dependencies.

<img src="./docs/assets/Architecture.png" width="800">

Event Shark is a Serverless framework which fill in the gap for kafka event publisher by exposing the eventing as simple json request.
Using this the consumer applcations can simulate the publisher behaviour and test their systems independently, not just the integration testing this can also help to run the performance testing for the consumer driven applications.

## Installation

To install Event Shark, follow these steps:

1. Clone the repository:
   ```sh
   git clone https://github.com/dipjyotimetia/EventShark.git
   cd EventShark
   ```

2. Build the Docker images:
   ```sh
   docker-compose build
   ```

3. Start the services:
   ```sh
   docker-compose up -d
   ```

## Configuration

Event Shark can be configured using environment variables. The following environment variables are available:

- `BROKERS`: The Kafka brokers to connect to (default: `localhost:9092`)
- `TOPICS`: The Kafka topics to produce messages to (default: `expense-topic,payment-topic,transaction-topic`)
- `SCHEMAREGISTRY`: The URL of the schema registry (default: `localhost:8081`)

## Usage

To use Event Shark, send a POST request to the appropriate endpoint with the event data in JSON format. The following endpoints are available:

- `/api/expense`: Create an expense event
- `/api/payment`: Create a payment event

### Example

```sh
curl -X POST http://localhost:8083/api/expense -H "Content-Type: application/json" -d '{
  "id": "123",
  "amount": 100.0,
  "timestamp": 1620000000000
}'
```

## Running Tests

To run the tests, use the following command:

```sh
make test
```

## Building the Project

To build the project, use the following command:

```sh
make build
```

## Contributing

Contributions are welcome! Please read the [contributing guidelines](docs/contributing.md) for more information on how to contribute to this project.
