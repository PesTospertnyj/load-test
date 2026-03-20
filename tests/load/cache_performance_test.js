import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

const cachedRequestDuration = new Trend('cached_request_duration');
const uncachedRequestDuration = new Trend('uncached_request_duration');

export const options = {
  scenarios: {
    cache_test: {
      executor: 'constant-vus',
      vus: 20,
      duration: '1m',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<500'],
    cached_request_duration: ['p(95)<100'],
    uncached_request_duration: ['p(95)<300'],
  },
};

export function setup() {
  const headers = { 'Content-Type': 'application/json' };

  const bookIds = [];
  for (let i = 0; i < 10; i++) {
    const payload = JSON.stringify({
      title: `Cache Test Book ${i}`,
      author: `Cache Author ${i}`,
      isbn: `ISBN-CACHE-${i}-${Date.now()}`,
    });

    const res = http.post(`${BASE_URL}/books`, payload, { headers });
    if (res.status === 201) {
      const body = JSON.parse(res.body);
      bookIds.push(body.id);
    }
  }

  return { bookIds };
}

export default function (data) {
  const randomBookId = data.bookIds[Math.floor(Math.random() * data.bookIds.length)];
  const randomPage = Math.floor(Math.random() * 10) + 1;

  // Test paginated list caching
  const getAllRes1 = http.get(`${BASE_URL}/books?page=${randomPage}&limit=25`);
  check(getAllRes1, {
    'first get all status is 200': (r) => r.status === 200,
    'first get all returns paginated response': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.books !== undefined && body.page === randomPage;
      } catch (e) {
        return false;
      }
    },
  });
  uncachedRequestDuration.add(getAllRes1.timings.duration);

  sleep(0.1);

  // Same request should be cached
  const getAllRes2 = http.get(`${BASE_URL}/books?page=${randomPage}&limit=25`);
  check(getAllRes2, {
    'cached get all status is 200': (r) => r.status === 200,
  });
  cachedRequestDuration.add(getAllRes2.timings.duration);

  sleep(0.1);

  // Test individual book caching
  const getByIdRes1 = http.get(`${BASE_URL}/books/${randomBookId}`);
  check(getByIdRes1, {
    'first get by id status is 200': (r) => r.status === 200,
  });
  uncachedRequestDuration.add(getByIdRes1.timings.duration);

  sleep(0.1);

  const getByIdRes2 = http.get(`${BASE_URL}/books/${randomBookId}`);
  check(getByIdRes2, {
    'cached get by id status is 200': (r) => r.status === 200,
  });
  cachedRequestDuration.add(getByIdRes2.timings.duration);

  sleep(0.5);
}

export function teardown(data) {
  const headers = { 'Content-Type': 'application/json' };

  for (const bookId of data.bookIds) {
    http.del(`${BASE_URL}/books/${bookId}`, null, { headers });
  }
}
