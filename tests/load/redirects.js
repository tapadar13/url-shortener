import http from "k6/http"
import { check } from "k6"
import exec from "k6/execution"

import { registerLoadUser } from "./lib/auth.js"
import {
  apiBaseURL,
  duration,
  positiveInteger,
} from "./lib/config.js"

const linkPoolSize = positiveInteger("REDIRECT_LINKS", 20)
const requestRate = positiveInteger("REDIRECT_RATE", 50)
const preAllocatedVUs = positiveInteger("REDIRECT_PREALLOCATED_VUS", 10)
const maxVUs = positiveInteger("REDIRECT_MAX_VUS", 50)
const p95LimitMilliseconds = positiveInteger("REDIRECT_P95_MS", 200)

if (maxVUs < preAllocatedVUs) {
  throw new Error(
    "REDIRECT_MAX_VUS must be greater than or equal to REDIRECT_PREALLOCATED_VUS"
  )
}

export const options = {
  scenarios: {
    cached_redirects: {
      executor: "constant-arrival-rate",
      rate: requestRate,
      timeUnit: "1s",
      duration: duration("REDIRECT_DURATION", "30s"),
      preAllocatedVUs,
      maxVUs,
      gracefulStop: "10s",
    },
  },
  thresholds: {
    checks: ["rate==1"],
    dropped_iterations: ["count==0"],
    http_req_failed: ["rate==0"],
    "http_req_duration{name:GET /:shortCode}": [
      `p(95)<${p95LimitMilliseconds}`,
    ],
  },
}

export function setup() {
  const testUser = registerLoadUser()
  const authorizationHeaders = {
    Authorization: `Bearer ${testUser.accessToken}`,
    "Content-Type": "application/json",
  }
  const links = []

  for (let index = 0; index < linkPoolSize; index += 1) {
    const shortCode = `rd${testUser.runID}${index}`
    const destination = `https://example.com/load/redirect/${shortCode}`
    const response = http.post(
      `${apiBaseURL}/shorten`,
      JSON.stringify({ url: destination, shortCode }),
      {
        headers: authorizationHeaders,
        tags: { name: "SETUP POST /shorten" },
      }
    )
    const created = check(response, {
      "redirect seed creation returns 201": (result) =>
        result.status === 201,
    })
    if (!created) {
      exec.test.abort(
        `redirect seed ${shortCode} failed with ${response.status}`
      )
    }

    links.push({ shortCode, destination })
  }

  return { ...testUser, links }
}

export default function followShortLinks(testData) {
  const link =
    testData.links[exec.scenario.iterationInTest % testData.links.length]
  const response = http.get(`${apiBaseURL}/${link.shortCode}`, {
    redirects: 0,
    tags: { name: "GET /:shortCode" },
  })

  check(response, {
    "short link returns 302": (result) => result.status === 302,
    "short link preserves its destination": (result) =>
      result.headers.Location === link.destination,
  })
}

export function teardown(testData) {
  const authorizationHeaders = {
    Authorization: `Bearer ${testData.accessToken}`,
  }

  for (const link of testData.links) {
    const response = http.del(
      `${apiBaseURL}/shorten/${link.shortCode}`,
      null,
      {
        headers: authorizationHeaders,
        tags: { name: "TEARDOWN DELETE /shorten/:shortCode" },
      }
    )
    check(response, {
      "redirect seed deletion returns 204": (result) =>
        result.status === 204,
    })
  }
}
