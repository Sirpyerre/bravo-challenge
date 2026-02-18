/**
 * Applications CRUD integration test.
 * Run: k6 run tests/k6/applications.js
 */
import http from "k6/http";
import { check, group, sleep } from "k6";
import { randomString } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<1000"],
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8000";

function register(email, password) {
  const res = http.post(
    `${BASE_URL}/auth/register`,
    JSON.stringify({ email, password, country: "MX" }),
    { headers: { "Content-Type": "application/json" } }
  );
  if (res.status !== 201) {
    console.error(`register failed: ${res.status} ${res.body}`);
    return null;
  }
  return JSON.parse(res.body).token;
}

function authHeaders(token, idempotencyKey) {
  const headers = {
    "Content-Type": "application/json",
    Authorization: `Bearer ${token}`,
  };
  if (idempotencyKey) {
    headers["Idempotency-Key"] = idempotencyKey;
  }
  return { headers };
}

export default function () {
  const email = `app_${randomString(8)}@bravo.test`;
  const token = register(email, "Test1234!");
  if (!token) return;

  let appID;

  group("create application - MX valid", () => {
    const idempotencyKey = `create-${randomString(16)}`;
    const res = http.post(
      `${BASE_URL}/api/v1/applications`,
      JSON.stringify({
        country: "MX",
        full_name: "Juan Pérez",
        identity_document: "LOOE890101HDFPSL09",
        monthly_income: 10000,
        requested_amount: 25000,
      }),
      authHeaders(token, idempotencyKey)
    );

    check(res, {
      "create status 201": (r) => r.status === 201,
      "create returns id": (r) => JSON.parse(r.body).id !== undefined,
      "create status is PENDING": (r) => JSON.parse(r.body).status === "PENDING",
    });

    if (res.status === 201) {
      appID = JSON.parse(res.body).id;
    }
  });

  group("create application - idempotency replay", () => {
    const idempotencyKey = `replay-${randomString(16)}`;
    const payload = JSON.stringify({
      country: "MX",
      full_name: "Juan Pérez",
      identity_document: "LOOE890101HDFPSL09",
      monthly_income: 10000,
      requested_amount: 25000,
    });

    const res1 = http.post(
      `${BASE_URL}/api/v1/applications`,
      payload,
      authHeaders(token, idempotencyKey)
    );
    const res2 = http.post(
      `${BASE_URL}/api/v1/applications`,
      payload,
      authHeaders(token, idempotencyKey)
    );

    check(res1, { "first call 201": (r) => r.status === 201 });
    check(res2, {
      "replay returns same status": (r) => r.status === res1.status,
      "replay returns same id": (r) => JSON.parse(r.body).id === JSON.parse(res1.body).id,
    });
  });

  group("create application - MX invalid CURP", () => {
    const params = authHeaders(token, `bad-${randomString(16)}`);
    params.responseCallback = http.expectedStatuses({ min: 400, max: 499 });
    const res = http.post(
      `${BASE_URL}/api/v1/applications`,
      JSON.stringify({
        country: "MX",
        full_name: "Juan Pérez",
        identity_document: "INVALID",
        monthly_income: 10000,
        requested_amount: 5000,
      }),
      params
    );
    check(res, { "invalid CURP returns 4xx": (r) => r.status >= 400 && r.status < 500 });
  });

  group("create application - missing idempotency key", () => {
    const params = authHeaders(token, null);
    params.responseCallback = http.expectedStatuses(400);
    const res = http.post(
      `${BASE_URL}/api/v1/applications`,
      JSON.stringify({
        country: "MX",
        full_name: "Juan",
        identity_document: "LOOE890101HDFPSL09",
        monthly_income: 10000,
        requested_amount: 5000,
      }),
      params
    );
    check(res, { "missing key returns 400": (r) => r.status === 400 });
  });

  if (!appID) return;

  group("get application by id", () => {
    const res = http.get(
      `${BASE_URL}/api/v1/applications/${appID}`,
      authHeaders(token, null)
    );
    check(res, {
      "get status 200": (r) => r.status === 200,
      "get returns correct id": (r) => JSON.parse(r.body).id === appID,
    });
  });

  group("list applications", () => {
    const res = http.get(
      `${BASE_URL}/api/v1/applications`,
      authHeaders(token, null)
    );
    check(res, {
      "list status 200": (r) => r.status === 200,
      "list returns array": (r) => Array.isArray(JSON.parse(r.body).applications),
      "list total > 0": (r) => JSON.parse(r.body).total > 0,
    });
  });

  group("list applications with filters", () => {
    const res = http.get(
      `${BASE_URL}/api/v1/applications?country=MX&limit=5`,
      authHeaders(token, null)
    );
    check(res, {
      "filtered list status 200": (r) => r.status === 200,
    });
  });

  group("get non-existent application", () => {
    const params = authHeaders(token, null);
    params.responseCallback = http.expectedStatuses({ min: 400, max: 499 });
    const res = http.get(
      `${BASE_URL}/api/v1/applications/00000000-0000-0000-0000-000000000000`,
      params
    );
    check(res, { "not found returns 4xx": (r) => r.status >= 400 });
  });

  // Wait briefly for worker to process and potentially update status
  sleep(3);

  group("check application status after worker", () => {
    const res = http.get(
      `${BASE_URL}/api/v1/applications/${appID}`,
      authHeaders(token, null)
    );
    check(res, {
      "status updated by worker": (r) => {
        const body = JSON.parse(r.body);
        return ["APPROVED", "DENIED", "PENDING", "VALIDATING"].includes(body.status);
      },
    });
  });
}
