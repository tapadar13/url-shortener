import { describe, expect, it } from "vitest"

import { authCookieNames, getAuthCookieOptions } from "./session"

describe("auth session cookies", () => {
  it("uses stable application-scoped cookie names", () => {
    expect(authCookieNames).toEqual({
      accessToken: "relay_access_token",
      refreshToken: "relay_refresh_token",
    })
  })

  it("protects tokens from browser scripts and cross-site requests", () => {
    const expiresAt = "2026-07-16T12:00:00Z"
    const options = getAuthCookieOptions(expiresAt, true)

    expect(options.accessToken).toMatchObject({
      httpOnly: true,
      sameSite: "lax",
      secure: true,
      path: "/",
      expires: new Date(expiresAt),
    })
    expect(options.refreshToken).toEqual({
      httpOnly: true,
      sameSite: "lax",
      secure: true,
      path: "/",
    })
  })

  it("allows insecure cookies only for local development", () => {
    const options = getAuthCookieOptions("2026-07-16T12:00:00Z", false)

    expect(options.accessToken.secure).toBe(false)
    expect(options.refreshToken.secure).toBe(false)
  })

  it("rejects invalid access token expiries", () => {
    expect(() => getAuthCookieOptions("not-a-date")).toThrow(/expiry/)
  })
})
