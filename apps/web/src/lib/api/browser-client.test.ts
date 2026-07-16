import { afterEach, describe, expect, it, vi } from "vitest"

import { requestBFF } from "./browser-client"
import { APIConnectionError } from "./errors"

afterEach(() => {
  vi.unstubAllGlobals()
})

describe("requestBFF", () => {
  it("requests a same-origin route with browser credentials", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ user: { id: "user-1" } }), {
        headers: { "Content-Type": "application/json" },
      })
    )
    vi.stubGlobal("fetch", fetchMock)

    await expect(requestBFF("/api/auth/session")).resolves.toEqual({
      user: { id: "user-1" },
    })

    const [path, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(path).toBe("/api/auth/session")
    expect(init.credentials).toBe("same-origin")
    expect(init.cache).toBe("no-store")
    expect(new Headers(init.headers).get("Accept")).toBe("application/json")
  })

  it("sets the JSON content type for mutation bodies", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ user: { id: "user-1" } }), {
        headers: { "Content-Type": "application/json" },
      })
    )
    vi.stubGlobal("fetch", fetchMock)

    await requestBFF("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ email: "user@example.com" }),
    })

    const init = fetchMock.mock.calls[0]?.[1] as RequestInit
    expect(new Headers(init.headers).get("Content-Type")).toBe(
      "application/json"
    )
  })

  it("translates structured BFF errors", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: { code: "invalid_credentials", message: "invalid login" },
          }),
          { status: 401 }
        )
      )
    )

    await expect(requestBFF("/api/auth/login")).rejects.toMatchObject({
      status: 401,
      code: "invalid_credentials",
      message: "invalid login",
    })
  })

  it("distinguishes browser network failures", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")))

    await expect(requestBFF("/api/auth/session")).rejects.toBeInstanceOf(
      APIConnectionError
    )
  })

  it("supports successful empty responses", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(null, { status: 204 }))
    )

    await expect(requestBFF<void>("/api/auth/logout")).resolves.toBeUndefined()
  })

  it("rejects paths outside the same-origin BFF", async () => {
    await expect(requestBFF("/auth/login")).rejects.toBeInstanceOf(TypeError)
    await expect(requestBFF("https://example.com/api/auth/login")).rejects.toBeInstanceOf(
      TypeError
    )
  })
})
