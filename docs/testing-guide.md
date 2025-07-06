# EventShark Testing Guide

This document provides comprehensive testing guidelines for the EventShark project.

## Test Types

### Unit Tests
Unit tests are located in the `pkg/` directory alongside the source code. They test individual components in isolation.

**Running Unit Tests:**
```bash
make test-unit
```

**Coverage Report:**
```bash
make test-coverage
```

### Integration Tests
Integration tests are located in `tests/integration/` and test the complete API endpoints with real services.

**Running Integration Tests:**
```bash
make test-integration
```

**Prerequisites:**
- Docker and Docker Compose installed
- Services running (`make build`)

### Performance Tests
Performance tests use k6 and are located in `tests/performance/`.

**Running Performance Tests:**
```bash
make perf-test
```

## Test Structure

### Unit Test Example
```go
func TestValidateExpense(t *testing.T) {
    validator := NewValidator()
    
    tests := []struct {
        name        string
        expense     gen.Expense
        expectError bool
    }{
        {
            name: "valid expense",
            expense: gen.Expense{
                ExpenseID: "exp-001",
                UserID:    "user-001",
                Category:  "food",
                Amount:    25.99,
                Currency:  "USD",
            },
            expectError: false,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.ValidateExpense(tt.expense)
            // ... assertions
        })
    }
}
```

### Integration Test Example
```go
func TestExpenseAPI(t *testing.T) {
    // Wait for service to be ready
    if !waitForService(t, baseURL+"/health", 30*time.Second) {
        t.Fatal("Service did not become ready in time")
    }
    
    expense := gen.Expense{
        ExpenseID: "exp-001",
        UserID:    "user-001",
        Category:  "food",
        Amount:    25.99,
        Currency:  "USD",
    }
    
    jsonData, _ := json.Marshal(expense)
    resp, err := http.Post(baseURL+"/api/expense", "application/json", bytes.NewBuffer(jsonData))
    // ... assertions
}
```

## Test Coverage

The project aims for:
- **Unit Tests**: 80%+ code coverage
- **Integration Tests**: Cover all API endpoints
- **Performance Tests**: Test under expected load

## Test Data

### Valid Test Data
```go
validExpense := gen.Expense{
    ExpenseID:   "exp-001",
    UserID:      "user-001",
    Category:    "food",
    Amount:      25.99,
    Currency:    "USD",
    Timestamp:   time.Now().UnixNano() / int64(time.Millisecond),
    Description: stringPtr("Test expense"),
}

validPayment := gen.Payment{
    TransactionID: "txn-001",
    UserID:        "user-001",
    Amount:        100.00,
    Currency:      "USD",
    PaymentMethod: "CREDIT_CARD",
    Status:        "COMPLETED",
}
```

### Invalid Test Data
```go
invalidExpense := gen.Expense{
    // Missing required fields
    Amount:   -25.99, // Negative amount
    Currency: "INVALID", // Invalid currency
}
```

## Mocking

### Mock Producer
```go
type MockProducer struct {
    shouldFailSetRecord bool
    shouldFailProduce   bool
    records             []*kgo.Record
}

func (m *MockProducer) ProduceSync(ctx context.Context, record *kgo.Record) error {
    if m.shouldFailProduce {
        return errors.New("producer error")
    }
    m.records = append(m.records, record)
    return nil
}
```

### Mock Validator
```go
type MockValidator struct {
    shouldFailValidation bool
}

func (m *MockValidator) ValidateExpense(expense gen.Expense) error {
    if m.shouldFailValidation {
        return errors.ErrValidation("validation failed", nil)
    }
    return nil
}
```

## Best Practices

1. **Test Names**: Use descriptive test names that explain what is being tested
2. **Table-Driven Tests**: Use table-driven tests for multiple test cases
3. **Test Isolation**: Each test should be independent and not rely on other tests
4. **Mocking**: Use mocks to isolate units under test
5. **Error Testing**: Test both success and failure scenarios
6. **Edge Cases**: Test boundary conditions and edge cases
7. **Test Data**: Use realistic test data that represents actual usage

## Continuous Integration

Tests are run automatically on every commit and pull request:

```bash
make ci  # Runs the complete CI pipeline
```

The CI pipeline includes:
1. Dependency download
2. Code formatting check
3. Linting
4. Security scanning
5. Unit tests
6. Integration tests

## Test Reports

- **Unit Test Coverage**: `coverage.html`
- **Integration Test Results**: `test-report.json`
- **Performance Test Results**: k6 HTML report

## Troubleshooting

### Common Issues

1. **Service Not Ready**: Ensure services are running before integration tests
2. **Port Conflicts**: Check that required ports are available
3. **Schema Errors**: Verify Avro schemas are valid
4. **Kafka Connection**: Ensure Kafka is running and accessible

### Debug Tips

1. **Verbose Output**: Use `-v` flag for verbose test output
2. **Individual Tests**: Run specific tests using `go test -run TestName`
3. **Race Detection**: Use `-race` flag to detect race conditions
4. **Logs**: Check application logs for error details

## Performance Testing

Performance tests validate the system under load:

```javascript
export const options = {
    thresholds: {
        'http_req_failed': ['rate<0.01'], // Less than 1% errors
        'http_req_duration': ['p(95)<200'], // 95% under 200ms
    },
    scenarios: {
        constant_load: {
            executor: 'constant-arrival-rate',
            rate: 1000, // 1000 requests per minute
            duration: '1m',
        },
    },
};
```

The performance tests validate:
- **Throughput**: Requests per second
- **Latency**: Response times (p95, p99)
- **Error Rate**: Percentage of failed requests
- **Resource Usage**: Memory and CPU utilization
