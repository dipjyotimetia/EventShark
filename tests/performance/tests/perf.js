import http from 'k6/http';
import { check } from 'k6';

export const options = {
    tags: {
        test: 'api-performance',
        test_run_id: `api-Load-Testing-${new Date().toISOString()}`,
    },
    thresholds: {
        'http_req_failed{test_type:addExpense}': ['rate<0.1'], // http errors should be less than 10%, availability
        'http_req_duration{test_type:addExpense}': ['p(95)<500'], // 95% of requests should be below 500ms, latency
        'http_req_failed{test_type:addPayments}': ['rate<0.1'], // http errors should be less than 10%, availability
        'http_req_duration{test_type:addPayments}': ['p(95)<500'], // 95% of requests should be below 500ms, latency
    },
    scenarios: {
        // Realistic load testing using K6 ramping arrival rate scenario
        addExpense_realistic: {
            executor: 'ramping-arrival-rate',
            startRate: 1, // start with 1 iteration per time unit
            timeUnit: '1s', // iterations will be per second
            preAllocatedVUs: 2, // the size of the VU (i.e. worker) pool for this scenario
            maxVUs: 10, // if the preAllocatedVUs are not enough, we can initialize more
            stages: [
                { target: 10, duration: '30s' }, // ramp up to 10 iterations per second over 30 seconds
                { target: 10, duration: '60s' }, // stay at 10 iterations per second for 60 seconds
                { target: 0, duration: '30s' },  // ramp down to 0 iterations per second over 30 seconds
            ],
            tags: { test_type: 'addExpense' }, // different extra metric tags for this scenario
            exec: 'addExpense',// Test scenario function to call
        },
        addPayments_realistic: {
          executor: 'ramping-arrival-rate',
          startRate: 1, // start with 1 iteration per time unit
          timeUnit: '1s', // iterations will be per second
          preAllocatedVUs: 2, // the size of the VU (i.e. worker) pool for this scenario
          maxVUs: 10, // if the preAllocatedVUs are not enough, we can initialize more
          stages: [
              { target: 10, duration: '30s' }, // ramp up to 10 iterations per second over 30 seconds
              { target: 10, duration: '60s' }, // stay at 10 iterations per second for 60 seconds
              { target: 0, duration: '30s' },  // ramp down to 0 iterations per second over 30 seconds
          ],
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