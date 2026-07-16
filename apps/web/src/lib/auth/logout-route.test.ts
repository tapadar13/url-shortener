import { describe, expect, it, vi } from "vitest"

import {
  APIConnectionError,
  APIRequestError,
} from "@/lib/api/client"

import { createLogoutRoute } from "./logout-route"

describe("createLogoutRoute", () => {
  it("revokes the refresh session and clears cookies", async () => {
    const dependencies = logoutDependencies({
      readRefreshToken: vi.fn().mockResolvedValue("refresh-token"),
    })

    const response = await createLogoutRoute(dependencies)(logoutRequest())

    expect(response.status).toBe(204)
    expect(response.headers.get("Cache-Control")).toBe("no-store")
    expect(dependencies.revoke).toHaveBeenCalledWith(
      "refresh-token",
      expect.any(AbortSignal)
    )
    expect(dependencies.clearSession).toHaveBeenCalledOnce()
  })

  it("remains idempotent without a refresh token", async () => {
    const dependencies = logoutDependencies()

    const response = await createLogoutRoute(dependencies)(logoutRequest())

    expect(response.status).toBe(204)
    expect(dependencies.revoke).not.toHaveBeenCalled()
    expect(dependencies.clearSession).toHaveBeenCalledOnce()
  })

  it("treats an invalid upstream session as already logged out", async () => {
    const dependencies = logoutDependencies({
      readRefreshToken: vi.fn().mockResolvedValue("invalid-token"),
      revoke: vi
        .fn()
        .mockRejectedValue(
          new APIRequestError(401, "invalid_refresh_token", "invalid")
        ),
    })

    const response = await createLogoutRoute(dependencies)(logoutRequest())

    expect(response.status).toBe(204)
    expect(dependencies.clearSession).toHaveBeenCalledOnce()
  })

  it("clears local cookies while reporting an unavailable API", async () => {
    const dependencies = logoutDependencies({
      readRefreshToken: vi.fn().mockResolvedValue("refresh-token"),
      revoke: vi.fn().mockRejectedValue(new APIConnectionError("offline")),
    })

    const response = await createLogoutRoute(dependencies)(logoutRequest())

    expect(response.status).toBe(502)
    expect(dependencies.clearSession).toHaveBeenCalledOnce()
    await expect(response.json()).resolves.toMatchObject({
      error: { code: "api_unavailable" },
    })
  })

  it("sanitizes local cookie cleanup failures", async () => {
    const dependencies = logoutDependencies({
      clearSession: vi.fn().mockRejectedValue(new Error("cookie failure")),
    })

    const response = await createLogoutRoute(dependencies)(logoutRequest())

    expect(response.status).toBe(500)
    await expect(response.json()).resolves.toEqual({
      error: { code: "internal_error", message: "could not complete logout" },
    })
  })
})

function logoutDependencies(overrides: Record<string, unknown> = {}) {
  return {
    readRefreshToken: vi.fn().mockResolvedValue(undefined),
    revoke: vi.fn().mockResolvedValue(undefined),
    clearSession: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  }
}

function logoutRequest(): Request {
  return new Request("http://localhost/api/auth/logout", { method: "POST" })
}
