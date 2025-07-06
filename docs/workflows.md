# EventShark Workflow Diagrams

This document contains detailed workflow diagrams for the EventShark system, showing the complete flow from API request to Kafka message delivery.

## Complete System Workflow

```mermaid
graph TB
    subgraph "External Layer"
        CLIENT[Client Application]
        BROWSER[Web Browser]
        TESTS[Test Suites]
    end
    
    subgraph "API Gateway Layer"
        LB[Load Balancer<br/>Optional]
        CORS[CORS Middleware]
        HELMET[Security Headers]
        LOGGER[Request Logger]
    end
    
    subgraph "Application Layer"
        API[Fiber REST API<br/>Port 8083]
        HEALTH[Health Check<br/>/health]
        
        subgraph "Routers"
            EXPENSE_ROUTER[Expense Router<br/>/api/expense]
            PAYMENT_ROUTER[Payment Router<br/>/api/payment]
        end
        
        subgraph "Handlers"
            EXPENSE_HANDLER[Expense Handler]
            PAYMENT_HANDLER[Payment Handler]
        end
        
        subgraph "Business Logic"
            VALIDATION[Input Validation]
            TIMESTAMP[Auto Timestamp]
            SERIALIZATION[Avro Serialization]
        end
    end
    
    subgraph "Event Processing Layer"
        PRODUCER[Kafka Producer<br/>franz-go]
        SCHEMA_CLIENT[Schema Registry Client]
        RECORD_BUILDER[Kafka Record Builder]
    end
    
    subgraph "Infrastructure Layer"
        subgraph "Kafka Cluster"
            KAFKA[RedPanda Kafka<br/>Port 9092/29092]
            
            subgraph "Topics"
                EXPENSE_TOPIC[(expense-topic)]
                PAYMENT_TOPIC[(payment-topic)]
                TRANSACTION_TOPIC[(transaction-topic)]
            end
        end
        
        SCHEMA_REGISTRY[Schema Registry<br/>Port 8081]
        
        subgraph "Schemas"
            EXPENSE_SCHEMA[Expense.avsc]
            PAYMENT_SCHEMA[Payment.avsc]
            TRANSACTION_SCHEMA[Transaction.avsc]
        end
    end
    
    subgraph "Monitoring Layer"
        CONSOLE[RedPanda Console<br/>Port 8086]
        METRICS[Application Metrics]
        LOGS[Application Logs]
    end
    
    subgraph "Consumer Layer"
        CONSUMER_APP[Consumer Applications]
        TEST_CONSUMER[Test Consumer<br/>script/consumer]
    end
    
    %% Client connections
    CLIENT --> LB
    BROWSER --> CONSOLE
    TESTS --> API
    
    %% API Gateway flow
    LB --> CORS
    CORS --> HELMET
    HELMET --> LOGGER
    LOGGER --> API
    
    %% API routing
    API --> HEALTH
    API --> EXPENSE_ROUTER
    API --> PAYMENT_ROUTER
    
    %% Handler routing
    EXPENSE_ROUTER --> EXPENSE_HANDLER
    PAYMENT_ROUTER --> PAYMENT_HANDLER
    
    %% Business logic flow
    EXPENSE_HANDLER --> VALIDATION
    PAYMENT_HANDLER --> VALIDATION
    VALIDATION --> TIMESTAMP
    TIMESTAMP --> SERIALIZATION
    
    %% Event processing
    SERIALIZATION --> PRODUCER
    PRODUCER --> SCHEMA_CLIENT
    SCHEMA_CLIENT --> SCHEMA_REGISTRY
    PRODUCER --> RECORD_BUILDER
    
    %% Schema connections
    SCHEMA_REGISTRY --> EXPENSE_SCHEMA
    SCHEMA_REGISTRY --> PAYMENT_SCHEMA
    SCHEMA_REGISTRY --> TRANSACTION_SCHEMA
    
    %% Kafka publishing
    RECORD_BUILDER --> KAFKA
    KAFKA --> EXPENSE_TOPIC
    KAFKA --> PAYMENT_TOPIC
    KAFKA --> TRANSACTION_TOPIC
    
    %% Monitoring
    KAFKA --> CONSOLE
    API --> METRICS
    API --> LOGS
    
    %% Consumer connections
    EXPENSE_TOPIC --> CONSUMER_APP
    PAYMENT_TOPIC --> CONSUMER_APP
    TRANSACTION_TOPIC --> CONSUMER_APP
    EXPENSE_TOPIC --> TEST_CONSUMER
    
    %% Styling
    classDef clientLayer fill:#e1f5fe
    classDef apiLayer fill:#f3e5f5
    classDef businessLayer fill:#e8f5e8
    classDef infraLayer fill:#fff3e0
    classDef monitorLayer fill:#fce4ec
    classDef consumerLayer fill:#f1f8e9
    
    class CLIENT,BROWSER,TESTS clientLayer
    class LB,CORS,HELMET,LOGGER,API,HEALTH apiLayer
    class EXPENSE_ROUTER,PAYMENT_ROUTER,EXPENSE_HANDLER,PAYMENT_HANDLER,VALIDATION,TIMESTAMP,SERIALIZATION businessLayer
    class PRODUCER,SCHEMA_CLIENT,RECORD_BUILDER,KAFKA,SCHEMA_REGISTRY,EXPENSE_TOPIC,PAYMENT_TOPIC,TRANSACTION_TOPIC infraLayer
    class CONSOLE,METRICS,LOGS monitorLayer
    class CONSUMER_APP,TEST_CONSUMER consumerLayer
```

## Request Processing Workflow

```mermaid
sequenceDiagram
    participant C as Client
    participant API as REST API
    participant MW as Middleware
    participant R as Router
    participant H as Handler
    participant P as Producer
    participant SR as Schema Registry
    participant K as Kafka
    participant CON as Consumer
    
    Note over C,CON: Expense Event Publishing Flow
    
    C->>+API: POST /api/expense
    Note right of C: JSON payload with<br/>expense data
    
    API->>+MW: Apply middleware
    MW->>MW: CORS validation
    MW->>MW: Security headers
    MW->>MW: Request logging
    MW->>-API: Continue
    
    API->>+R: Route to expense endpoint
    R->>+H: ExpenseHandler
    
    H->>H: Parse JSON body
    Note right of H: Unmarshal into<br/>gen.Expense struct
    
    H->>H: Validate required fields
    alt Validation fails
        H->>C: 400 Bad Request
    end
    
    H->>H: Set timestamp if empty
    Note right of H: time.Now().UnixNano()<br/>/ int64(time.Millisecond)
    
    H->>+P: Create Kafka record
    P->>+SR: Get schema for expense-topic-value
    SR->>-P: Return Avro schema
    
    P->>P: Parse Avro schema
    P->>P: Marshal data with Avro
    P->>P: Create kgo.Record
    P->>-H: Return record
    
    H->>+P: Produce message
    P->>+K: ProduceSync(record)
    K->>K: Write to expense-topic
    K->>-P: Confirm delivery
    Note right of K: Returns offset,<br/>partition info
    
    P->>P: Log success
    Note right of P: "Message sent: topic: expense-topic,<br/>offset: X, partition: Y"
    P->>-H: Success
    
    H->>-R: 200 OK
    R->>-API: Response
    API->>-C: "expense created successfully"
    
    Note over K,CON: Async consumption
    K->>CON: Deliver event
    CON->>CON: Process expense event
```

## Error Handling Workflow

```mermaid
flowchart TD
    START([API Request Received])
    
    PARSE{Parse JSON Body}
    PARSE -->|Success| VALIDATE
    PARSE -->|Error| ERROR_400[Return 400 Bad Request]
    
    VALIDATE{Validate Fields}
    VALIDATE -->|Valid| TIMESTAMP
    VALIDATE -->|Invalid| ERROR_400
    
    TIMESTAMP[Set Timestamp if Empty]
    TIMESTAMP --> SCHEMA_GET
    
    SCHEMA_GET{Get Schema from Registry}
    SCHEMA_GET -->|Success| SERIALIZE
    SCHEMA_GET -->|Error| ERROR_500[Return 500 Internal Error]
    
    SERIALIZE{Serialize with Avro}
    SERIALIZE -->|Success| PRODUCE
    SERIALIZE -->|Error| ERROR_500
    
    PRODUCE{Produce to Kafka}
    PRODUCE -->|Success| SUCCESS[Return 200 OK]
    PRODUCE -->|Error| ERROR_500
    
    ERROR_400 --> LOG_ERROR[Log Error Details]
    ERROR_500 --> LOG_ERROR
    LOG_ERROR --> END([End Request])
    
    SUCCESS --> LOG_SUCCESS[Log Success Message]
    LOG_SUCCESS --> END
    
    %% Styling
    classDef startEnd fill:#c8e6c9
    classDef decision fill:#fff3e0
    classDef process fill:#e3f2fd
    classDef error fill:#ffebee
    classDef success fill:#e8f5e8
    
    class START,END startEnd
    class PARSE,VALIDATE,SCHEMA_GET,SERIALIZE,PRODUCE decision
    class TIMESTAMP,LOG_ERROR,LOG_SUCCESS process
    class ERROR_400,ERROR_500 error
    class SUCCESS success
```

## Development Workflow

```mermaid
gitGraph
    commit id: "Initial project setup"
    
    branch feature/schema-design
    checkout feature/schema-design
    commit id: "Define Avro schemas"
    commit id: "Generate Go structs"
    commit id: "Schema validation tests"
    
    checkout main
    merge feature/schema-design
    commit id: "Merge schema design"
    
    branch feature/api-development
    checkout feature/api-development
    commit id: "Setup Fiber framework"
    commit id: "Implement handlers"
    commit id: "Add middleware"
    commit id: "Integration tests"
    
    checkout main
    merge feature/api-development
    commit id: "Merge API development"
    
    branch feature/kafka-integration
    checkout feature/kafka-integration
    commit id: "Kafka producer setup"
    commit id: "Schema registry integration"
    commit id: "Error handling"
    commit id: "Performance optimization"
    
    checkout main
    merge feature/kafka-integration
    commit id: "Merge Kafka integration"
    
    branch feature/testing
    checkout feature/testing
    commit id: "Unit tests"
    commit id: "Integration tests"
    commit id: "Performance tests"
    commit id: "Test automation"
    
    checkout main
    merge feature/testing
    commit id: "Release v1.0.0"
```

## Docker Container Workflow

```mermaid
graph LR
    subgraph "Build Phase"
        SOURCE[Source Code]
        DOCKERFILE[Dockerfile]
        BUILD[Docker Build]
        IMAGE[EventShark Image]
    end
    
    subgraph "Infrastructure Setup"
        COMPOSE[docker-compose.yml]
        KAFKA_IMG[RedPanda Image]
        INIT_IMG[Topic Init Image]
    end
    
    subgraph "Runtime Phase"
        NETWORK[Docker Network]
        
        subgraph "Containers"
            APP_CONTAINER[EventShark Container<br/>:8083]
            KAFKA_CONTAINER[Kafka Container<br/>:9092,:29092,:8081]
            CONSOLE_CONTAINER[Console Container<br/>:8086]
            INIT_CONTAINER[Init Container]
        end
        
        subgraph "Volumes"
            KAFKA_DATA[(Kafka Data)]
            LOGS[(Application Logs)]
        end
    end
    
    subgraph "Health Checks"
        HEALTH_API[API Health Check]
        HEALTH_KAFKA[Kafka Health Check]
        HEALTH_CONSOLE[Console Health Check]
    end
    
    %% Build flow
    SOURCE --> BUILD
    DOCKERFILE --> BUILD
    BUILD --> IMAGE
    
    %% Infrastructure setup
    COMPOSE --> NETWORK
    KAFKA_IMG --> KAFKA_CONTAINER
    INIT_IMG --> INIT_CONTAINER
    IMAGE --> APP_CONTAINER
    
    %% Container dependencies
    INIT_CONTAINER -.->|depends_on| KAFKA_CONTAINER
    APP_CONTAINER -.->|depends_on| INIT_CONTAINER
    CONSOLE_CONTAINER -.->|depends_on| KAFKA_CONTAINER
    
    %% Volume mounts
    KAFKA_CONTAINER --> KAFKA_DATA
    APP_CONTAINER --> LOGS
    
    %% Health checks
    APP_CONTAINER --> HEALTH_API
    KAFKA_CONTAINER --> HEALTH_KAFKA
    CONSOLE_CONTAINER --> HEALTH_CONSOLE
    
    %% Network communication
    APP_CONTAINER <--> KAFKA_CONTAINER
    CONSOLE_CONTAINER <--> KAFKA_CONTAINER
```

## Testing Workflow

```mermaid
graph TB
    subgraph "Test Preparation"
        ENV_SETUP[Environment Setup]
        DATA_PREP[Test Data Preparation]
        SCHEMA_LOAD[Schema Loading]
    end
    
    subgraph "Unit Testing"
        UNIT_CONFIG[Config Tests]
        UNIT_HANDLER[Handler Tests]
        UNIT_PRODUCER[Producer Tests]
        UNIT_SCHEMA[Schema Tests]
    end
    
    subgraph "Integration Testing"
        INT_API[API Endpoint Tests]
        INT_KAFKA[Kafka Integration Tests]
        INT_SCHEMA_REG[Schema Registry Tests]
        INT_E2E[End-to-End Tests]
    end
    
    subgraph "Performance Testing"
        PERF_LOAD[Load Testing]
        PERF_STRESS[Stress Testing]
        PERF_VOLUME[Volume Testing]
        PERF_ENDURANCE[Endurance Testing]
    end
    
    subgraph "Test Reporting"
        RESULTS[Test Results Collection]
        REPORT_JSON[JSON Report Generation]
        REPORT_MD[Markdown Report]
        METRICS[Performance Metrics]
    end
    
    ENV_SETUP --> UNIT_CONFIG
    DATA_PREP --> UNIT_HANDLER
    SCHEMA_LOAD --> UNIT_SCHEMA
    
    UNIT_CONFIG --> INT_API
    UNIT_HANDLER --> INT_API
    UNIT_PRODUCER --> INT_KAFKA
    UNIT_SCHEMA --> INT_SCHEMA_REG
    
    INT_API --> INT_E2E
    INT_KAFKA --> INT_E2E
    INT_SCHEMA_REG --> INT_E2E
    
    INT_E2E --> PERF_LOAD
    PERF_LOAD --> PERF_STRESS
    PERF_STRESS --> PERF_VOLUME
    PERF_VOLUME --> PERF_ENDURANCE
    
    UNIT_CONFIG --> RESULTS
    INT_E2E --> RESULTS
    PERF_ENDURANCE --> RESULTS
    
    RESULTS --> REPORT_JSON
    RESULTS --> REPORT_MD
    RESULTS --> METRICS
    
    %% Test execution commands
    ENV_SETUP -.->|make build| ENV_SETUP
    UNIT_CONFIG -.->|go test ./pkg/config| UNIT_CONFIG
    INT_API -.->|go test --tags=integration| INT_API
    PERF_LOAD -.->|k6 run tests/perf.js| PERF_LOAD
```

## Schema Evolution Workflow

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant Schema as Schema File
    participant Gen as Code Generator
    participant Reg as Schema Registry
    participant App as Application
    participant Test as Tests
    
    Note over Dev,Test: Schema Evolution Process
    
    Dev->>Schema: Update .avsc file
    Note right of Dev: Add new optional field<br/>or modify existing
    
    Dev->>Gen: Run code generation
    Note right of Gen: make code-gen
    Gen->>Gen: Parse Avro schema
    Gen->>Gen: Generate Go structs
    
    Dev->>App: Update application code
    Note right of App: Handle new fields<br/>in handlers
    
    Dev->>Test: Update tests
    Note right of Test: Test new fields<br/>and backward compatibility
    
    Dev->>Reg: Deploy new schema
    Note right of Reg: Schema registry validates<br/>compatibility
    
    alt Schema Compatible
        Reg->>App: Schema accepted
        App->>Test: Run integration tests
        Test->>Dev: Tests pass
    else Schema Incompatible
        Reg->>Dev: Reject schema
        Dev->>Schema: Fix compatibility issues
    end
    
    Note over Dev,Test: Schema successfully evolved
```

This workflow documentation provides comprehensive visual representations of how EventShark operates at different levels, from high-level system architecture to detailed request processing flows.
