import http from "k6/http"
import { check, sleep } from "k6"
import exec from "k6/execution"

import { registerLoadUser } from "./lib/auth.js"
import {
  apiBaseURL,
  duration,
  positiveInteger,
} from "./lib/config.js"

const virtualUsers = positiveInteger("MANAGEMENT_VUS", 2)
const p95LimitMilliseconds = positiveInteger("MANAGEMENT_P95_MS", 500)

export const options = {
  scenarios: {
    authenticated_management: {
      executor: "constant-vus",
      vus: virtualUsers,
      duration: duration("MANAGEMENT_DURATION", "30s"),
      gracefulStop: "10s",
    },
  },
  thresholds: {
    checks: ["rate==1"],
    http_req_failed: ["rate==0"],
    "http_req_duration{name:POST /shorten}": [
      `p(95)<${p95LimitMilliseconds}`,
    ],
    "http_req_duration{name:GET /shorten/:shortCode/stats}": [
      `p(95)<${p95LimitMilliseconds}`,
    ],
    "http_req_duration{name:DELETE /shorten/:shortCode}": [
      `p(95)<${p95LimitMilliseconds}`,
    ],
  },
}

export function setup() {
  return registerLoadUser()
}

export default function manageLinks(testUser) {
  const shortCode = `mg${testUser.runID}${exec.scenario.iterationInTest}`
  const authorizationHeaders = {
    Authorization: `Bearer ${testUser.accessToken}`,
  }

  const createResponse = http.post(
    `${apiBaseURL}/shorten`,
    JSON.stringify({
      url: `https://example.com/load/${shortCode}`,
      shortCode,
    }),
    {
      headers: {
        ...authorizationHeaders,
        "Content-Type": "application/json",
      },
      tags: { name: "POST /shorten" },
    }
  )
  const created = check(createResponse, {
    "short link creation returns 201": (response) => response.status === 201,
    "short link creation preserves the code": (response) =>
      response.json("shortCode") === shortCode,
  })
  if (!created) {
    sleep(0.5)
    return
  }

  const statsResponse = http.get(
    `${apiBaseURL}/shorten/${shortCode}/stats`,
    {
      headers: authorizationHeaders,
      tags: { name: "GET /shorten/:shortCode/stats" },
    }
  )
  check(statsResponse, {
    "link statistics return 200": (response) => response.status === 200,
    "new link starts with zero visits": (response) =>
      response.json("accessCount") === 0,
  })

  const deleteResponse = http.del(
    `${apiBaseURL}/shorten/${shortCode}`,
    null,
    {
      headers: authorizationHeaders,
      tags: { name: "DELETE /shorten/:shortCode" },
    }
  )
  check(deleteResponse, {
    "short link deletion returns 204": (response) => response.status === 204,
  })

  sleep(0.5)
}
