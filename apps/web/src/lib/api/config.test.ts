import { describe, expect, it } from "vitest"

import { getAPIBaseURL } from "./config"

describe("getAPIBaseURL", () => {
  it("uses the local API by default", () => {
    expect(getAPIBaseURL(undefined)).toBe("http://localhost:8080")
  })

  it.each([
    ["https://api.example.com/", "https://api.example.com"],
    ["https://api.example.com/v1///", "https://api.example.com/v1"],
    [" http://localhost:9090 ", "http://localhost:9090"],
  ])("normalizes %s", (value, expected) => {
    expect(getAPIBaseURL(value)).toBe(expected)
  })

  it.each([
    "api.example.com",
    "ftp://api.example.com",
    "https://user:password@api.example.com",
    "https://api.example.com?source=web",
    "https://api.example.com#auth",
  ])("rejects unsafe API base URL %s", (value) => {
    expect(() => getAPIBaseURL(value)).toThrow(/API_BASE_URL/)
  })
})
