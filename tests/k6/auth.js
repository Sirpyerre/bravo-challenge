/**
 * Auth smoke test — register + login flow.
 * Run: k6 run tests/k6/auth.js
 */
import http from "k6/http";
import { check, group } from "k6";
import { randomString } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<500"],
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8000";

export default function () {
  const email = `test_${randomString(8)}@bravo.test`;
  const password = "Test1234!";
  let token;

  group("register", () => {
    const res = http.post(
      `${BASE_URL}/auth/register`,
      JSON.stringify({ email, password, country: "MX" }),
      { headers: { "Content-Type": "application/json" } }
    );

    check(res, {
      "register status 201": (r) => r.status === 201,
      "register returns token": (r) => JSON.parse(r.body).token !== undefined,
    });

    if (res.status === 201) {
      token = JSON.parse(res.body).token;
    }
  });

  group("register duplicate email", () => {
    const res = http.post(
      `${BASE_URL}/auth/register`,
      JSON.stringify({ email, password, country: "MX" }),
      {
        headers: { "Content-Type": "application/json" },
        responseCallback: http.expectedStatuses({ min: 400, max: 499 }),
      }
    );
    check(res, {
      "duplicate register returns 4xx": (r) => r.status >= 400 && r.status < 500,
    });
  });

  group("login", () => {
    const res = http.post(
      `${BASE_URL}/auth/login`,
      JSON.stringify({ email, password }),
      { headers: { "Content-Type": "application/json" } }
    );

    check(res, {
      "login status 200": (r) => r.status === 200,
      "login returns token": (r) => JSON.parse(r.body).token !== undefined,
    });
  });

  group("login wrong password", () => {
    const res = http.post(
      `${BASE_URL}/auth/login`,
      JSON.stringify({ email, password: "wrongpass" }),
      {
        headers: { "Content-Type": "application/json" },
        responseCallback: http.expectedStatuses({ min: 400, max: 499 }),
      }
    );
    check(res, {
      "wrong password returns 4xx": (r) => r.status >= 400 && r.status < 500,
    });
  });

  group("protected route without token", () => {
    const res = http.get(`${BASE_URL}/api/v1/applications`, {
      responseCallback: http.expectedStatuses(401),
    });
    check(res, {
      "no token returns 401": (r) => r.status === 401,
    });
  });
}
