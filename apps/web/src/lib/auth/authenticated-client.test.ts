import { describe, expect, it, vi } from "vitest"

import { APIRequestError } from "@/lib/api/client"

import { requestAuthenticatedAPI } from "./authenticated-client"
import type { RefreshResponse } from "./types"

const refreshed: RefreshResponse = {
  accessToken: "new-access-token",
  refreshToken: "new-refresh-token",
  tokenType: "Bearer",
  expiresAt: "2026-07-16T12:00:00Z",
}

describe("requestAuthenticatedAPI", () => {
  it("uses the current access token without refreshing", async () => {
    const dependencies = authDependencies({
      request: vi.fn().mockResolvedValue({ id: "user-1" }),
      readAccessToken: vi.fn().mockResolvedValue("access-token"),
    })

    await expect(
      requestAuthenticatedAPI("/auth/me", {}, dependencies)
    ).resolves.toEqual({ id: "user-1" })

    expect(dependencies.request).toHaveBeenCalledTimes(1)
    expect(authorizationHeader(dependencies.request, 0)).toBe(
      "Bearer access-token"
    )
    expect(dependencies.readRefreshToken).not.toHaveBeenCalled()
  })

  it("rotates the refresh token and retries one unauthorized request", async () => {
    const dependencies = authDependencies({
      request: vi
        .fn()
        .mockRejectedValueOnce(
          new APIRequestError(401, "unauthorized", "expired")
        )
        .mockResolvedValueOnce(refreshed)
        .mockResolvedValueOnce({ id: "user-1" }),
      readAccessToken: vi.fn().mockResolvedValue("expired-access-token"),
      readRefreshToken: vi.fn().mockResolvedValue("refresh-token"),
    })

    await expect(
      requestAuthenticatedAPI("/auth/me", {}, dependencies)
    ).resolves.toEqual({ id: "user-1" })

    expect(dependencies.request).toHaveBeenCalledTimes(3)
    expect(dependencies.request).toHaveBeenNthCalledWith(
      2,
      "/auth/refresh",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ refreshToken: "refresh-token" }),
      })
    )
    expect(dependencies.writeSession).toHaveBeenCalledWith(refreshed)
    expect(authorizationHeader(dependencies.request, 2)).toBe(
      "Bearer new-access-token"
    )
  })

  it("refreshes directly when the access cookie has expired", async () => {
    const dependencies = authDependencies({
      request: vi
        .fn()
        .mockResolvedValueOnce(refreshed)
        .mockResolvedValueOnce({ items: [] }),
      readRefreshToken: vi.fn().mockResolvedValue("refresh-token"),
    })

    await expect(
      requestAuthenticatedAPI("/shorten", {}, dependencies)
    ).resolves.toEqual({ items: [] })

    expect(dependencies.request).toHaveBeenCalledTimes(2)
    expect(dependencies.writeSession).toHaveBeenCalledWith(refreshed)
  })

  it("clears a session that has no refresh token", async () => {
    const dependencies = authDependencies()

    await expect(
      requestAuthenticatedAPI("/auth/me", {}, dependencies)
    ).rejects.toMatchObject({ status: 401, code: "unauthorized" })

    expect(dependencies.request).not.toHaveBeenCalled()
    expect(dependencies.clearSession).toHaveBeenCalledOnce()
  })

  it("clears and normalizes an invalid refresh session", async () => {
    const dependencies = authDependencies({
      request: vi
        .fn()
        .mockRejectedValue(
          new APIRequestError(401, "invalid_refresh_token", "invalid")
        ),
      readRefreshToken: vi.fn().mockResolvedValue("invalid-refresh-token"),
    })

    await expect(
      requestAuthenticatedAPI("/auth/me", {}, dependencies)
    ).rejects.toMatchObject({ status: 401, code: "unauthorized" })

    expect(dependencies.clearSession).toHaveBeenCalledOnce()
  })

  it("does not refresh non-authentication failures", async () => {
    const error = new APIRequestError(500, "internal_error", "failed")
    const dependencies = authDependencies({
      request: vi.fn().mockRejectedValue(error),
      readAccessToken: vi.fn().mockResolvedValue("access-token"),
      readRefreshToken: vi.fn().mockResolvedValue("refresh-token"),
    })

    await expect(
      requestAuthenticatedAPI("/auth/me", {}, dependencies)
    ).rejects.toBe(error)

    expect(dependencies.request).toHaveBeenCalledOnce()
    expect(dependencies.readRefreshToken).not.toHaveBeenCalled()
  })
})

function authDependencies(overrides: Record<string, unknown> = {}) {
  return {
    request: vi.fn(),
    readAccessToken: vi.fn().mockResolvedValue(undefined),
    readRefreshToken: vi.fn().mockResolvedValue(undefined),
    writeSession: vi.fn().mockResolvedValue(undefined),
    clearSession: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  }
}

function authorizationHeader(request: ReturnType<typeof vi.fn>, call: number) {
  const init = request.mock.calls[call]?.[1] as RequestInit
  return new Headers(init.headers).get("Authorization")
}
