#!/bin/sh

# Install curl
apt-get update && apt-get install -y curl

BROKER="kafka:29092"
SCHEMA_REGISTRY="http://kafka:8081/subjects"

# Connect to kafka and create topics with schema
for TOPIC in expense payment transaction
do
  rpk --brokers $BROKER topic create ${TOPIC}-topic
  if [ $? -ne 0 ]; then
    echo "Failed to create topic ${TOPIC}-topic"
    exit 1
  fi

  curl -X POST -H "Content-Type: application/vnd.schemaregistry.v1+json" \
      --data @/${TOPIC}.json $SCHEMA_REGISTRY/${TOPIC}-topic-value/versions
  if [ $? -ne 0 ]; then
    echo "Failed to add schema for ${TOPIC}-topic"
    exit 1
  fi

  rpk --brokers $BROKER topic info ${TOPIC}-topic
  if [ $? -ne 0 ]; then
    echo "Failed to get info for ${TOPIC}-topic"
    exit 1
  fi
done