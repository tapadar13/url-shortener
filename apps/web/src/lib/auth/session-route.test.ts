import { describe, expect, it, vi } from "vitest"

import {
  APIConnectionError,
  APIRequestError,
} from "@/lib/api/client"

import { createSessionRoute } from "./session-route"

describe("createSessionRoute", () => {
  it("returns the current user without token data", async () => {
    const currentUser = vi
      .fn()
      .mockResolvedValue({ id: "user-1", email: "user@example.com" })
    const response = await createSessionRoute({ currentUser })(sessionRequest())

    expect(response.status).toBe(200)
    expect(response.headers.get("Cache-Control")).toBe("no-store")
    await expect(response.json()).resolves.toEqual({
      user: { id: "user-1", email: "user@example.com" },
    })
    expect(currentUser).toHaveBeenCalledWith(expect.any(AbortSignal))
  })

  it("preserves unauthenticated session responses", async () => {
    const response = await createSessionRoute({
      currentUser: vi
        .fn()
        .mockRejectedValue(
          new APIRequestError(401, "unauthorized", "authentication is required")
        ),
    })(sessionRequest())

    expect(response.status).toBe(401)
    await expect(response.json()).resolves.toMatchObject({
      error: { code: "unauthorized" },
    })
  })

  it("maps API connection failures to bad gateway", async () => {
    const response = await createSessionRoute({
      currentUser: vi.fn().mockRejectedValue(new APIConnectionError("offline")),
    })(sessionRequest())

    expect(response.status).toBe(502)
    await expect(response.json()).resolves.toMatchObject({
      error: { code: "api_unavailable" },
    })
  })

  it("sanitizes unexpected session failures", async () => {
    const response = await createSessionRoute({
      currentUser: vi.fn().mockRejectedValue(new Error("cookie failure")),
    })(sessionRequest())

    expect(response.status).toBe(500)
    await expect(response.json()).resolves.toEqual({
      error: { code: "internal_error", message: "could not restore session" },
    })
  })
})

function sessionRequest(): Request {
  return new Request("http://localhost/api/auth/session")
}
