#!/bin/bash

# Install curl
apt-get update && apt-get install -y curl

# Connect to kafka and create topics with schema
rpk --brokers kafka:29092 topic create expense-topic
rpk --brokers kafka:29092 topic create payment-topic
rpk --brokers kafka:29092 topic create transaction-topic

# Add schema to the Schema Registry
curl -X POST -H "Content-Type: application/vnd.schemaregistry.v1+json" \
    --data @/expense.json http://kafka:8081/subjects/expense-topic-value/versions

curl -X POST -H "Content-Type: application/vnd.schemaregistry.v1+json" \
    --data @/payment.json http://kafka:8081/subjects/payment-topic-value/versions

curl -X POST -H "Content-Type: application/vnd.schemaregistry.v1+json" \
    --data @/transaction.json http://kafka:8081/subjects/transaction-topic-value/versions

rpk --brokers kafka:29092 topic info expense-topic
rpk --brokers kafka:29092 topic info payment-topic
rpk --brokers kafka:29092 topic info transaction-topic
