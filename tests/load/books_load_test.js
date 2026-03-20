import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '30s', target: 10 },
    { duration: '1m', target: 50 },
    { duration: '30s', target: 100 },
    { duration: '1m', target: 50 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.05'],
    errors: ['rate<0.1'],
  },
};

export default function () {
  let bookId;

  const createPayload = JSON.stringify({
    title: `Book ${__VU}-${__ITER}`,
    author: `Author ${__VU}`,
    isbn: `ISBN-${__VU}-${__ITER}-${Date.now()}`,
  });

  const createHeaders = {
    'Content-Type': 'application/json',
  };

  const createRes = http.post(`${BASE_URL}/books`, createPayload, {
    headers: createHeaders,
  });

  const createSuccess = check(createRes, {
    'create book status is 201': (r) => r.status === 201,
    'create book has id': (r) => {
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

  if (!createSuccess || !bookId) {
    return;
  }

  sleep(0.1);

  const getByIdRes = http.get(`${BASE_URL}/books/${bookId}`);
  const getByIdSuccess = check(getByIdRes, {
    'get book by id status is 200': (r) => r.status === 200,
    'get book by id returns correct data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.id === bookId;
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(!getByIdSuccess);

  sleep(0.1);

  const getAllRes = http.get(`${BASE_URL}/books`);
  const getAllSuccess = check(getAllRes, {
    'get all books status is 200': (r) => r.status === 200,
    'get all books returns array': (r) => {
      try {
        const body = JSON.parse(r.body);
        return Array.isArray(body);
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(!getAllSuccess);

  sleep(0.1);

  const updatePayload = JSON.stringify({
    title: `Updated Book ${__VU}-${__ITER}`,
    author: `Updated Author ${__VU}`,
    isbn: `ISBN-UPD-${__VU}-${__ITER}-${Date.now()}`,
  });

  const updateRes = http.put(`${BASE_URL}/books/${bookId}`, updatePayload, {
    headers: createHeaders,
  });

  const updateSuccess = check(updateRes, {
    'update book status is 200': (r) => r.status === 200,
    'update book returns updated data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.title.startsWith('Updated Book');
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(!updateSuccess);

  sleep(0.1);

  const deleteRes = http.del(`${BASE_URL}/books/${bookId}`);
  const deleteSuccess = check(deleteRes, {
    'delete book status is 204': (r) => r.status === 204,
  });

  errorRate.add(!deleteSuccess);

  sleep(0.5);
}
