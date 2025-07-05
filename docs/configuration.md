# Project Configuration

This document provides detailed information on configuring the Event Shark project.

## Configuration Options

Event Shark can be configured using environment variables. The following environment variables are available:

- `BROKERS`: The Kafka brokers to connect to (default: `localhost:9092`)
- `TOPICS`: The Kafka topics to produce messages to (default: `expense-topic,payment-topic,transaction-topic`)
- `SCHEMAREGISTRY`: The URL of the schema registry (default: `localhost:8081`)

## Environment Variables

### BROKERS

The `BROKERS` environment variable specifies the Kafka brokers to connect to. It should be a comma-separated list of broker addresses. The default value is `localhost:9092`.

Example:
```sh
export BROKERS="broker1:9092,broker2:9092"
```

### TOPICS

The `TOPICS` environment variable specifies the Kafka topics to produce messages to. It should be a comma-separated list of topic names. The default value is `expense-topic,payment-topic,transaction-topic`.

Example:
```sh
export TOPICS="expense-topic,payment-topic,transaction-topic"
```

### SCHEMAREGISTRY

The `SCHEMAREGISTRY` environment variable specifies the URL of the schema registry. The default value is `localhost:8081`.

Example:
```sh
export SCHEMAREGISTRY="http://localhost:8081"
```

## Configuration Files

Event Shark can also be configured using configuration files. The configuration files should be in JSON format and should contain the same configuration options as the environment variables.

Example configuration file:
```json
{
  "brokers": "broker1:9092,broker2:9092",
  "topics": "expense-topic,payment-topic,transaction-topic",
  "schemaRegistry": "http://localhost:8081"
}
```

To use a configuration file, set the `CONFIG_FILE` environment variable to the path of the configuration file.

Example:
```sh
export CONFIG_FILE="/path/to/config.json"
```

## Docker Compose Environment Variables

When running with Docker Compose, you can override environment variables in the `docker-compose.yml` file under the `environment` section for each service.

## Examples

### Example 1: Using Environment Variables

```sh
export BROKERS="broker1:9092,broker2:9092"
export TOPICS="expense-topic,payment-topic,transaction-topic"
export SCHEMAREGISTRY="http://localhost:8081"

docker-compose up -d
```

### Example 2: Using a Configuration File

Create a configuration file `config.json` with the following content:
```json
{
  "brokers": "broker1:9092,broker2:9092",
  "topics": "expense-topic,payment-topic,transaction-topic",
  "schemaRegistry": "http://localhost:8081"
}
```

Set the `CONFIG_FILE` environment variable and start the services:
```sh
export CONFIG_FILE="/path/to/config.json"
docker-compose up -d
```
