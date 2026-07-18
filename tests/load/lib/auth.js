import http from "k6/http"
import { check } from "k6"
import exec from "k6/execution"

import { apiBaseURL } from "./config.js"

const LOAD_TEST_PASSWORD = "correct horse battery staple"

export function registerLoadUser() {
  const runID = `${Date.now()}${Math.floor(Math.random() * 1_000_000)}`
  const response = http.post(
    `${apiBaseURL}/auth/register`,
    JSON.stringify({
      email: `load-${runID}@example.com`,
      password: LOAD_TEST_PASSWORD,
    }),
    {
      headers: { "Content-Type": "application/json" },
      tags: { name: "POST /auth/register" },
    }
  )

  const registered = check(response, {
    "load user registration returns 201": (result) => result.status === 201,
    "load user registration returns an access token": (result) =>
      typeof result.json("accessToken") === "string",
  })
  if (!registered) {
    exec.test.abort(`load user registration failed with ${response.status}`)
  }

  return {
    accessToken: response.json("accessToken"),
    runID,
  }
}
