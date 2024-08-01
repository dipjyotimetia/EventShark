# Running Tests

This document provides detailed information on how to run tests for the Event Shark project.

## Unit Tests

To run the unit tests, use the following command:

```sh
go test ./... -v
```

This command will run all the unit tests in the project and display detailed output.

## Integration Tests

To run the integration tests, use the following command:

```sh
make test
```

This command will run the integration tests defined in the `tests/integration` directory.

## Performance Tests

To run the performance tests, follow these steps:

1. Install the necessary dependencies:
   ```sh
   cd tests/performance
   npm install
   ```

2. Run the performance tests:
   ```sh
   npm test
   ```

This will execute the performance tests defined in the `tests/performance/tests/perf.js` file.

## Example Test Commands and Expected Outputs

### Unit Test Example

```sh
go test ./... -v
```

Expected output:
```
=== RUN   TestExample
--- PASS: TestExample (0.00s)
PASS
ok      github.com/dipjyotimetia/event-shark/pkg/example   0.002s
```

### Integration Test Example

```sh
make test
```

Expected output:
```
go test ./... -v --tags=integration -count=1
=== RUN   TestExpenseAPI
--- PASS: TestExpenseAPI (0.00s)
PASS
ok      github.com/dipjyotimetia/event-shark/tests/integration   0.002s
```

### Performance Test Example

```sh
npm test
```

Expected output:
```
> performance@1.0.0 test
> k6 run tests/perf.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: tests/perf.js
     output: -

  scenarios: (100.00%) 1 scenario, 50 max VUs, 1m30s max duration (incl. graceful stop):
           * default: Up to 50 looping VUs for 30s over 1 stages (gracefulRampDown: 30s, gracefulStop: 30s)


running (0m30.0s), 00/50 VUs, 1000 complete and 0 interrupted iterations
default ✓ [======================================] 50 VUs  30s

     data_received..................: 1.2 MB 40 kB/s
     data_sent......................: 1.1 MB 37 kB/s
     http_req_blocked...............: avg=1.2ms    min=0s      med=1.1ms   max=2.5ms    p(90)=1.3ms    p(95)=1.4ms
     http_req_connecting............: avg=0s       min=0s      med=0s      max=0s       p(90)=0s       p(95)=0s
     http_req_duration..............: avg=50ms     min=45ms    med=48ms    max=60ms     p(90)=55ms     p(95)=57ms
     http_req_receiving.............: avg=1ms      min=0s      med=1ms     max=2ms      p(90)=1ms      p(95)=1ms
     http_req_sending...............: avg=0.5ms    min=0s      med=0.5ms   max=1ms      p(90)=0.6ms    p(95)=0.7ms
     http_req_tls_handshaking.......: avg=0s       min=0s      med=0s      max=0s       p(90)=0s       p(95)=0s
     http_req_waiting...............: avg=48ms     min=44ms    med=47ms    max=58ms     p(90)=53ms     p(95)=55ms
     http_reqs......................: 1000    33.333333/s
     iteration_duration.............: avg=50ms     min=45ms    med=48ms    max=60ms     p(90)=55ms     p(95)=57ms
     iterations.....................: 1000    33.333333/s
     vus............................: 50      min=50    max=50
     vus_max........................: 50      min=50    max=50
```
