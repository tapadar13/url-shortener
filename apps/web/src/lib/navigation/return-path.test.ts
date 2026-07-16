import { describe, expect, it } from "vitest"

import { authPath, safeReturnPath } from "./return-path"

describe("safeReturnPath", () => {
  it.each([
    ["/dashboard", "/dashboard"],
    ["/links?page=2#recent", "/links?page=2#recent"],
  ])("accepts application path %s", (value, expected) => {
    expect(safeReturnPath(value)).toBe(expected)
  })

  it.each([
    "https://example.com/steal-session",
    "//example.com/steal-session",
    "/\\example.com/steal-session",
    "not-an-absolute-path",
  ])("rejects unsafe return path %s", (value) => {
    expect(safeReturnPath(value)).toBe("/dashboard")
  })

  it("rejects duplicate query parameters", () => {
    expect(safeReturnPath(["/dashboard", "//example.com"])).toBe("/dashboard")
  })
})

describe("authPath", () => {
  it("encodes a return path as an authentication query parameter", () => {
    expect(authPath("/login", "/links?status=active")).toBe(
      "/login?returnTo=%2Flinks%3Fstatus%3Dactive"
    )
  })
})
