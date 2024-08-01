# Usage

This document provides detailed information on using the Event Shark project. It includes examples of API endpoints and their usage, as well as explanations of the functionality and features.

## API Endpoints

### Create an Expense Event

Endpoint: `/api/expense`

Method: `POST`

Description: This endpoint is used to create an expense event.

Example Request:
```sh
curl -X POST http://localhost:8083/api/expense -H "Content-Type: application/json" -d '{
  "id": "123",
  "amount": 100.0,
  "timestamp": 1620000000000
}'
```

### Create a Payment Event

Endpoint: `/api/payment`

Method: `POST`

Description: This endpoint is used to create a payment event.

Example Request:
```sh
curl -X POST http://localhost:8083/api/payment -H "Content-Type: application/json" -d '{
  "id": "456",
  "amount": 200.0,
  "timestamp": 1620000000000
}'
```

## Functionality and Features

Event Shark provides a serverless framework for Kafka event publishing. It allows consumer applications to simulate publisher behavior and test their systems independently. The key features include:

- **Ease of Use**: Simple JSON requests to create events.
- **Isolation**: Test consumer applications without complex publisher dependencies.
- **Performance Testing**: Run targeted performance tests on consumer applications.

By using Event Shark, you can accelerate your Kafka-driven development, gain confidence in your consumer code, and optimize performance.
