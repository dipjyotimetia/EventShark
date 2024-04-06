import http from 'k6/http';
import { check } from 'k6';

export const options = {
    tags: {
        test: 'api-performance',
        test_run_id: `api-Load-Testing-${new Date().toISOString()}`,
    },
    thresholds: {
        'http_req_failed{test_type:addExpense}': ['rate<0.01'], // http errors should be less than 1%, availability
        'http_req_duration{test_type:addExpense}': ['p(95)<200'], // 95% of requests should be below 200ms, latency
        'http_req_failed{test_type:addPayments}': ['rate<0.01'], // http errors should be less than 1%, availability
        'http_req_duration{test_type:addPayments}': ['p(95)<200'], // 95% of requests should be below 200ms, latency
    },
    scenarios: {
        // Load testing using K6 constant-rate scenario
        addExpense_constant: {
            executor: 'constant-arrival-rate',
            rate: 10000, // number of iterations per time unit
            timeUnit: '1m', // iterations will be per minute
            duration: '1m', // total duration that the test will run for
            preAllocatedVUs: 2, // the size of the VU (i.e. worker) pool for this scenario
            maxVUs: 25, // if the preAllocatedVUs are not enough, we can initialize more
            tags: { test_type: 'addExpense' }, // different extra metric tags for this scenario
            exec: 'addExpense',// Test scenario function to call
        },
        addPayments_constant: {
          executor: 'constant-arrival-rate',
          rate: 10000, // number of iterations per time unit
          timeUnit: '1m', // iterations will be per minute
          duration: '1m', // total duration that the test will run for
          preAllocatedVUs: 2, // the size of the VU (i.e. worker) pool for this scenario
          maxVUs: 25, // if the preAllocatedVUs are not enough, we can initialize more
          tags: { test_type: 'addPayments' }, // different extra metric tags for this scenario
          exec: 'addPayments',// Test scenario function to call
      }
    }
};

export function addExpense() {
  const url = 'http://localhost:8083/api/expense';
  
  const payload = JSON.stringify({
    expense_id: 'test',
    user_id: '10010',
    category: 'kafkaSync',
    amount: 12.5,
    currency: 'AUD',
    timestamp: Date.now(),
    description: 'Any',
    receipt: 'newTest',
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const response = http.post(url, payload, params);

  check(response, {
    'is status 200': (r) => r.status === 200,
  });
}

export function addPayments() {
  const url = 'http://localhost:8083/api/payment';
  
  const payload = JSON.stringify({
    transaction_id: 'test',
    user_id: '10010',
    amount: 12.5,
    currency: 'AUD',
    payment_method:'CREDIT_CARD',
    timestamp: Date.now(),
    status: 'COMPLETED',
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const response = http.post(url, payload, params);

  check(response, {
    'is status 200': (r) => r.status === 200,
  });
}