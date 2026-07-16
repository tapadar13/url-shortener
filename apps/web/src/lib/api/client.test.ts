import { afterEach, describe, expect, it, vi } from "vitest"

import {
  APIConnectionError,
  APIRequestError,
  requestAPI,
} from "./client"

afterEach(() => {
  vi.unstubAllGlobals()
  vi.unstubAllEnvs()
})

describe("requestAPI", () => {
  it("requests the configured API and decodes JSON", async () => {
    vi.stubEnv("API_BASE_URL", "https://api.example.com/v1/")
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ status: "ok" }), {
        headers: { "Content-Type": "application/json" },
      })
    )
    vi.stubGlobal("fetch", fetchMock)

    await expect(requestAPI<{ status: string }>("/healthz")).resolves.toEqual({
      status: "ok",
    })

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toBe("https://api.example.com/v1/healthz")
    expect(init.cache).toBe("no-store")
    expect(new Headers(init.headers).get("Accept")).toBe("application/json")
  })

  it("sets the JSON content type when a body is present", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ id: "user-1" }), {
        headers: { "Content-Type": "application/json" },
      })
    )
    vi.stubGlobal("fetch", fetchMock)

    await requestAPI("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email: "user@example.com" }),
    })

    const init = fetchMock.mock.calls[0]?.[1] as RequestInit
    expect(new Headers(init.headers).get("Content-Type")).toBe(
      "application/json"
    )
  })

  it("translates structured API errors", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: {
              code: "invalid_credentials",
              message: "email or password is invalid",
            },
          }),
          { status: 401 }
        )
      )
    )

    await expect(requestAPI("/auth/login")).rejects.toMatchObject({
      name: "APIRequestError",
      status: 401,
      code: "invalid_credentials",
      message: "email or password is invalid",
    } satisfies Partial<APIRequestError>)
  })

  it("uses a stable fallback for non-JSON upstream errors", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response("Bad gateway", { status: 502, statusText: "Bad Gateway" })
      )
    )

    await expect(requestAPI("/healthz")).rejects.toMatchObject({
      status: 502,
      code: "upstream_error",
      message: "Bad Gateway",
    })
  })

  it("distinguishes connection failures", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("ECONNREFUSED")))

    await expect(requestAPI("/healthz")).rejects.toBeInstanceOf(
      APIConnectionError
    )
  })

  it("supports successful empty responses", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(null, { status: 204 }))
    )

    await expect(requestAPI<void>("/auth/logout")).resolves.toBeUndefined()
  })

  it("rejects absolute and protocol-relative request paths", async () => {
    await expect(requestAPI("https://example.com")).rejects.toBeInstanceOf(
      TypeError
    )
    await expect(requestAPI("//example.com")).rejects.toBeInstanceOf(TypeError)
  })
})
