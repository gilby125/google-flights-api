import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 20 }, // ramp up to 20 users
    { duration: '1m', target: 20 },  // stay at 20 users
    { duration: '30s', target: 0 },  // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests <500ms
    http_req_failed: ['rate<0.01'],   // <1% errors
  },
};

const testRoutes = [
  {
    origin: 'WAW',
    destination: 'ATH',
    departureDate: '2025-06-01',
    returnDate: '2025-06-15',
    adults: 2,
  },
  {
    origin: 'JFK',
    destination: 'LAX',
    departureDate: '2025-07-10',
    returnDate: '2025-07-20',
    adults: 1,
  },
  {
    origin: 'LHR',
    destination: 'DXB',
    departureDate: '2025-08-05',
    returnDate: '2025-08-15',
    adults: 3,
  }
];

export default function () {
  // Test POST /api/v1/search
  const searchParams = testRoutes[Math.floor(Math.random() * testRoutes.length)];
  const createRes = http.post('http://localhost:8080/api/v1/search', JSON.stringify(searchParams), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(createRes, {
    'POST /search status is 200': (r) => r.status === 200,
    'POST /search has searchId': (r) => JSON.parse(r.body).searchId !== undefined,
  });

  const searchId = JSON.parse(createRes.body).searchId;

  // Test GET /api/v1/search/:id
  const getRes = http.get(`http://localhost:8080/api/v1/search/${searchId}`);
  check(getRes, {
    'GET /search/:id status is 200': (r) => r.status === 200,
    'GET /search/:id has status': (r) => JSON.parse(r.body).status !== undefined,
  });

  // Test GET /api/v1/search/:id/results
  const resultsRes = http.get(`http://localhost:8080/api/v1/search/${searchId}/results`);
  check(resultsRes, {
    'GET /search/:id/results status is 200': (r) => r.status === 200,
    'GET /search/:id/results has results': (r) => JSON.parse(r.body).results !== undefined,
  });

  // Test POST /api/v1/bulk_search with 2 random searches
  const bulkParams = {
    searches: [
      testRoutes[Math.floor(Math.random() * testRoutes.length)],
      testRoutes[Math.floor(Math.random() * testRoutes.length)]
    ]
  };
  const bulkRes = http.post('http://localhost:8080/api/v1/bulk_search', JSON.stringify(bulkParams), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(bulkRes, {
    'POST /bulk_search status is 200': (r) => r.status === 200,
    'POST /bulk_search has bulkSearchId': (r) => JSON.parse(r.body).bulkSearchId !== undefined,
  });

  sleep(1);
}
