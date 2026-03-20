import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '30s', target: 10 },
    { duration: '10s', target: 200 },
    { duration: '1m', target: 200 },
    { duration: '10s', target: 10 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'],
    http_req_failed: ['rate<0.2'],
    errors: ['rate<0.25'],
  },
};

export default function () {
  const createPayload = JSON.stringify({
    title: `Spike Book ${__VU}-${__ITER}`,
    author: `Spike Author ${__VU}`,
    isbn: `ISBN-SPIKE-${__VU}-${__ITER}-${Date.now()}`,
  });

  const headers = {
    'Content-Type': 'application/json',
  };

  const createRes = http.post(`${BASE_URL}/books`, createPayload, {
    headers: headers,
  });

  const success = check(createRes, {
    'create status is 201': (r) => r.status === 201,
  });

  errorRate.add(!success);

  sleep(0.1);

  const getAllRes = http.get(`${BASE_URL}/books`);
  const getAllSuccess = check(getAllRes, {
    'get all status is 200': (r) => r.status === 200,
  });

  errorRate.add(!getAllSuccess);

  sleep(0.1);
}
