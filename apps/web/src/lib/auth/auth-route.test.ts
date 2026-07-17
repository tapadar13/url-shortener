import { describe, expect, it, vi } from "vitest"

import {
  APIConnectionError,
  APIRequestError,
} from "@/lib/api/client"

import { createAuthRoute } from "./auth-route"
import type { AuthResponse } from "./types"

const authResponse: AuthResponse = {
  accessToken: "access-token",
  refreshToken: "refresh-token",
  tokenType: "Bearer",
  expiresAt: "2026-07-16T12:00:00Z",
  user: { id: "user-1", email: "user@example.com" },
}

describe("createAuthRoute", () => {
  it.each([
    ["/auth/login" as const, 200 as const],
    ["/auth/register" as const, 201 as const],
  ])("authenticates through %s without exposing tokens", async (endpoint, status) => {
    const authenticate = vi.fn().mockResolvedValue(authResponse)
    const writeSession = vi.fn().mockResolvedValue(undefined)
    const handler = createAuthRoute(endpoint, status, {
      authenticate,
      writeSession,
    })

    const response = await handler(authRequest())

    expect(response.status).toBe(status)
    await expect(response.json()).resolves.toEqual({ user: authResponse.user })
    expect(authenticate).toHaveBeenCalledWith(
      endpoint,
      { email: "user@example.com", password: "correct horse battery staple" },
      expect.any(AbortSignal)
    )
    expect(writeSession).toHaveBeenCalledWith(authResponse)
  })

  it("rejects malformed JSON before calling the API", async () => {
    const authenticate = vi.fn()
    const handler = createAuthRoute("/auth/login", 200, {
      authenticate,
      writeSession: vi.fn(),
    })

    const response = await handler(
      new Request("http://localhost/api/auth/login", {
        method: "POST",
        headers: { Origin: "http://localhost" },
        body: "{",
      })
    )

    expect(response.status).toBe(400)
    expect(authenticate).not.toHaveBeenCalled()
  })

  it("rejects a cross-origin request before handling credentials", async () => {
    const authenticate = vi.fn()
    const writeSession = vi.fn()
    const handler = createAuthRoute("/auth/login", 200, {
      authenticate,
      writeSession,
    })

    const response = await handler(
      authRequest(undefined, "https://attacker.example")
    )

    expect(response.status).toBe(403)
    await expect(response.json()).resolves.toEqual({
      error: {
        code: "origin_not_allowed",
        message: "request origin is not allowed",
      },
    })
    expect(authenticate).not.toHaveBeenCalled()
    expect(writeSession).not.toHaveBeenCalled()
  })

  it.each([
    null,
    {},
    { email: "user@example.com" },
    { email: 42, password: "password" },
  ])("rejects invalid credentials payload %#", async (body) => {
    const authenticate = vi.fn()
    const handler = createAuthRoute("/auth/login", 200, {
      authenticate,
      writeSession: vi.fn(),
    })

    const response = await handler(authRequest(body))

    expect(response.status).toBe(400)
    expect(authenticate).not.toHaveBeenCalled()
  })

  it("preserves structured API errors", async () => {
    const handler = createAuthRoute("/auth/login", 200, {
      authenticate: vi
        .fn()
        .mockRejectedValue(
          new APIRequestError(401, "invalid_credentials", "invalid login")
        ),
      writeSession: vi.fn(),
    })

    const response = await handler(authRequest())

    expect(response.status).toBe(401)
    await expect(response.json()).resolves.toEqual({
      error: { code: "invalid_credentials", message: "invalid login" },
    })
  })

  it("maps API connection failures to bad gateway", async () => {
    const handler = createAuthRoute("/auth/login", 200, {
      authenticate: vi.fn().mockRejectedValue(new APIConnectionError("offline")),
      writeSession: vi.fn(),
    })

    const response = await handler(authRequest())

    expect(response.status).toBe(502)
    await expect(response.json()).resolves.toMatchObject({
      error: { code: "api_unavailable" },
    })
  })

  it("sanitizes unexpected failures", async () => {
    const handler = createAuthRoute("/auth/login", 200, {
      authenticate: vi.fn().mockResolvedValue(authResponse),
      writeSession: vi.fn().mockRejectedValue(new Error("cookie failure")),
    })

    const response = await handler(authRequest())

    expect(response.status).toBe(500)
    await expect(response.json()).resolves.toEqual({
      error: { code: "internal_error", message: "authentication failed" },
    })
  })
})

function authRequest(
  body: unknown = {
    email: "user@example.com",
    password: "correct horse battery staple",
  },
  origin = "http://localhost"
): Request {
  return new Request("http://localhost/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json", Origin: origin },
    body: JSON.stringify(body),
  })
}
