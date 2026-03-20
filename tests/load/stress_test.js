import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '2m', target: 50 },
    { duration: '3m', target: 100 },
    { duration: '2m', target: 150 },
    { duration: '3m', target: 200 },
    { duration: '2m', target: 100 },
    { duration: '2m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000', 'p(99)<2000'],
    http_req_failed: ['rate<0.1'],
    errors: ['rate<0.15'],
  },
};

export default function () {
  const createPayload = JSON.stringify({
    title: `Stress Book ${__VU}-${__ITER}`,
    author: `Stress Author ${__VU}`,
    isbn: `ISBN-STRESS-${__VU}-${__ITER}-${Date.now()}`,
  });

  const headers = {
    'Content-Type': 'application/json',
  };

  const createRes = http.post(`${BASE_URL}/books`, createPayload, {
    headers: headers,
  });

  let bookId;
  const createSuccess = check(createRes, {
    'create status is 201': (r) => r.status === 201,
    'has book id': (r) => {
      try {
        const body = JSON.parse(r.body);
        bookId = body.id;
        return body.id !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(!createSuccess);

  sleep(0.2);

  if (bookId) {
    const getRes = http.get(`${BASE_URL}/books/${bookId}`);
    const getSuccess = check(getRes, {
      'get status is 200': (r) => r.status === 200,
    });
    errorRate.add(!getSuccess);
  }

  sleep(0.1);

  // Sequential pagination: cycle through pages
  const page = Math.floor(__ITER / 10) + 1;
  const getAllRes = http.get(`${BASE_URL}/books?page=${page}&limit=25`);
  const getAllSuccess = check(getAllRes, {
    'get all status is 200': (r) => r.status === 200,
    'get all returns paginated response': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.books !== undefined && body.total !== undefined && body.page === page;
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(!getAllSuccess);

  sleep(0.2);
}
