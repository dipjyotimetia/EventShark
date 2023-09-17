#!/bin/sh

# Function to check if a Kafka topic exists
topic_exists() {
  local topic_name="$1"
  local brokers="$2"
  rpk --brokers "$brokers" topic describe "$topic_name" >/dev/null 2>&1
  return $?
}

# Install curl if not already installed
if ! command -v curl >/dev/null 2>&1; then
  apt-get update && apt-get install -y curl
fi

# Define Kafka broker address
KAFKA_BROKERS="kafka:29092"

# Define Schema Registry URL
SCHEMA_REGISTRY_URL="http://kafka:8081"

# Define schemas and topics
schemas=("expense.json" "payment.json" "transaction.json")
topics=("expense-topic" "payment-topic" "transaction-topic")

# Create topics and register schemas
for ((i=0; i<${#topics[@]}; i++)); do
  topic="${topics[i]}"
  schema_file="${schemas[i]}"

  # Check if the topic exists, create it if not
  if ! topic_exists "$topic" "$KAFKA_BROKERS"; then
    rpk --brokers "$KAFKA_BROKERS" topic create "$topic"
  fi

  # Register schema with the Schema Registry
  curl -X POST -H "Content-Type: application/vnd.schemaregistry.v1+json" \
    --data "@/$schema_file" "$SCHEMA_REGISTRY_URL/subjects/${topic}-value/versions"

  # Display topic information
  rpk --brokers "$KAFKA_BROKERS" topic info "$topic"
done

# Clean up
apt-get remove -y curl
