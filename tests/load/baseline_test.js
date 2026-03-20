import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  vus: 10,
  duration: '30s',
  thresholds: {
    http_req_duration: ['p(95)<300', 'p(99)<500'],
    http_req_failed: ['rate<0.01'],
    errors: ['rate<0.05'],
  },
};

export default function () {
  const createPayload = JSON.stringify({
    title: `Baseline Book ${__VU}-${__ITER}`,
    author: `Baseline Author ${__VU}`,
    isbn: `ISBN-BASE-${__VU}-${__ITER}-${Date.now()}`,
  });

  const headers = {
    'Content-Type': 'application/json',
  };

  const createRes = http.post(`${BASE_URL}/books`, createPayload, {
    headers: headers,
  });

  const success = check(createRes, {
    'status is 201': (r) => r.status === 201,
  });

  errorRate.add(!success);

  sleep(1);

  const getAllRes = http.get(`${BASE_URL}/books`);
  const getAllSuccess = check(getAllRes, {
    'get all status is 200': (r) => r.status === 200,
    'get all returns paginated response': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.books !== undefined && body.total !== undefined && body.page !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(!getAllSuccess);

  sleep(1);
}
