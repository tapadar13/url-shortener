import http from "k6/http"
import { check, sleep } from "k6"

import {
  apiBaseURL,
  duration,
  positiveInteger,
} from "./lib/config.js"

const virtualUsers = positiveInteger("SMOKE_VUS", 1)
const p95LimitMilliseconds = positiveInteger("SMOKE_P95_MS", 250)

export const options = {
  scenarios: {
    probe_smoke: {
      executor: "constant-vus",
      vus: virtualUsers,
      duration: duration("SMOKE_DURATION", "15s"),
      gracefulStop: "5s",
    },
  },
  thresholds: {
    checks: ["rate==1"],
    http_req_failed: ["rate==0"],
    http_req_duration: [`p(95)<${p95LimitMilliseconds}`],
  },
}

export default function probeSmoke() {
  const health = http.get(`${apiBaseURL}/healthz`, {
    tags: { name: "GET /healthz" },
  })
  check(health, {
    "health probe returns 200": (response) => response.status === 200,
    "health probe reports ok": (response) =>
      response.json("status") === "ok",
  })

  const readiness = http.get(`${apiBaseURL}/readyz`, {
    tags: { name: "GET /readyz" },
  })
  check(readiness, {
    "readiness probe returns 200": (response) => response.status === 200,
    "readiness probe reports ready": (response) =>
      response.json("status") === "ready",
  })

  sleep(0.5)
}
