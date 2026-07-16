import { describe, expect, it } from "vitest"

import { isSameOriginRequest } from "./request-origin"

describe("isSameOriginRequest", () => {
  it("accepts a matching browser origin", () => {
    expect(isSameOriginRequest(requestWithOrigin("https://relay.example"))).toBe(
      true
    )
  })

  it.each([
    ["a cross-origin request", "https://attacker.example", undefined],
    ["a missing origin", undefined, undefined],
    ["an opaque origin", "null", undefined],
    ["a cross-site fetch", "https://relay.example", "cross-site"],
    ["a same-site fetch", "https://relay.example", "same-site"],
  ])("rejects %s", (_, origin, fetchSite) => {
    expect(isSameOriginRequest(requestWithOrigin(origin, fetchSite))).toBe(false)
  })
})

function requestWithOrigin(origin?: string, fetchSite?: string): Request {
  const headers = new Headers()
  if (origin) {
    headers.set("Origin", origin)
  }
  if (fetchSite) {
    headers.set("Sec-Fetch-Site", fetchSite)
  }

  return new Request("https://relay.example/api/auth/login", {
    method: "POST",
    headers,
  })
}
