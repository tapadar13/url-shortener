import { describe, expect, it } from "vitest"

import { isSameOriginRequest } from "./request-origin"

describe("isSameOriginRequest", () => {
  it("accepts a matching browser origin", () => {
    expect(isSameOriginRequest(requestWithOrigin("https://relay.example"))).toBe(
      true
    )
  })

  it("uses the HTTP host when the runtime URL has an internal hostname", () => {
    expect(
      isSameOriginRequest(
        requestWithOrigin("http://127.0.0.1:3100", "same-origin", {
          host: "127.0.0.1:3100",
          url: "http://localhost:3100/api/auth/login",
        })
      )
    ).toBe(true)
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

  it.each([
    ["a different HTTP host", "attacker.example"],
    ["credentials in the HTTP host", "relay.example@attacker.example"],
    ["a path in the HTTP host", "relay.example/path"],
  ])("rejects %s", (_, host) => {
    expect(
      isSameOriginRequest(
        requestWithOrigin("https://relay.example", "same-origin", { host })
      )
    ).toBe(false)
  })
})

function requestWithOrigin(
  origin?: string,
  fetchSite?: string,
  options: { host?: string; url?: string } = {}
): Request {
  const headers = new Headers()
  if (origin) {
    headers.set("Origin", origin)
  }
  if (fetchSite) {
    headers.set("Sec-Fetch-Site", fetchSite)
  }
  if (options.host) {
    headers.set("Host", options.host)
  }

  return new Request(options.url ?? "https://relay.example/api/auth/login", {
    method: "POST",
    headers,
  })
}
