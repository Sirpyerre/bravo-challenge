/**
 * Load test — simulates concurrent users creating applications.
 * Run: k6 run tests/k6/load_test.js
 * Run with higher load: k6 run --vus 20 --duration 30s tests/k6/load_test.js
 */
import http from "k6/http";
import { check, sleep } from "k6";
import { Counter, Rate, Trend } from "k6/metrics";
import { randomString } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

// Custom metrics
const applicationCreated = new Counter("applications_created");
const applicationFailed = new Counter("applications_failed");
const applicationCreateDuration = new Trend("application_create_duration", true);
const errorRate = new Rate("error_rate");

export const options = {
  stages: [
    { duration: "10s", target: 5 },   // ramp up
    { duration: "20s", target: 10 },  // sustain
    { duration: "10s", target: 0 },   // ramp down
  ],
  thresholds: {
    http_req_failed: ["rate<0.05"],
    http_req_duration: ["p(95)<2000", "p(99)<5000"],
    error_rate: ["rate<0.05"],
    application_create_duration: ["p(95)<1500"],
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8000";

// Países y documentos válidos para cada uno
const TEST_CASES = [
  {
    country: "MX",
    identity_document: "LOOE890101HDFPSL09",
    monthly_income: 10000,
    requested_amount: 20000,
  },
  {
    country: "BR",
    identity_document: "52998224725",
    monthly_income: 8000,
    requested_amount: 15000,
  },
];

function registerAndLogin() {
  const email = `load_${randomString(12)}@bravo.test`;
  const password = "Test1234!";

  const res = http.post(
    `${BASE_URL}/auth/register`,
    JSON.stringify({ email, password, country: "MX" }),
    { headers: { "Content-Type": "application/json" } }
  );

  if (res.status !== 201) return null;
  return JSON.parse(res.body).token;
}

export function setup() {
  // Verify server is up before load test
  const res = http.get(`${BASE_URL}/health`);
  if (res.status !== 200) {
    console.error(`Server health check failed: ${res.status}`);
  }
}

export default function () {
  const token = registerAndLogin();
  if (!token) {
    errorRate.add(1);
    return;
  }

  const tc = TEST_CASES[Math.floor(Math.random() * TEST_CASES.length)];
  const idempotencyKey = `load-${randomString(20)}`;

  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/applications`,
    JSON.stringify({
      country: tc.country,
      full_name: "Test User",
      identity_document: tc.identity_document,
      monthly_income: tc.monthly_income,
      requested_amount: tc.requested_amount,
    }),
    {
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
        "Idempotency-Key": idempotencyKey,
      },
    }
  );

  applicationCreateDuration.add(Date.now() - start);

  const ok = check(res, {
    "application created": (r) => r.status === 201,
  });

  if (ok) {
    applicationCreated.add(1);
    errorRate.add(0);
  } else {
    applicationFailed.add(1);
    errorRate.add(1);
    console.error(`Create failed [${tc.country}]: ${res.status} ${res.body}`);
  }

  sleep(1);
}
