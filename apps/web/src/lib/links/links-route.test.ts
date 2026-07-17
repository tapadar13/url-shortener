import { describe, expect, it, vi } from "vitest"

import { APIConnectionError, APIRequestError } from "@/lib/api/client"

import { createLinksRoute } from "./links-route"
import type { ShortLink } from "./types"

const link: ShortLink = {
  id: "507f1f77bcf86cd799439011",
  url: "https://example.com/articles/123",
  shortCode: "AbC1234",
  shortUrl: "https://rly.to/AbC1234",
  createdAt: "2026-07-17T08:00:00Z",
  updatedAt: "2026-07-17T08:00:00Z",
}

describe("createLinksRoute GET", () => {
  it("forwards supported cursor pagination parameters", async () => {
    const dependencies = linksDependencies({
      list: vi.fn().mockResolvedValue({
        items: [link],
        nextCursor: "next-page",
      }),
    })
    const handler = createLinksRoute(dependencies).GET

    const response = await handler(
      new Request(
        "http://localhost/api/links?limit=8&cursor=opaque%2Bcursor&ignored=value"
      )
    )

    expect(response.status).toBe(200)
    expect(response.headers.get("Cache-Control")).toBe("no-store")
    await expect(response.json()).resolves.toEqual({
      items: [link],
      nextCursor: "next-page",
    })
    expect(dependencies.list).toHaveBeenCalledWith(
      "?limit=8&cursor=opaque%2Bcursor",
      expect.any(AbortSignal)
    )
  })

  it("preserves structured API errors", async () => {
    const dependencies = linksDependencies({
      list: vi
        .fn()
        .mockRejectedValue(
          new APIRequestError(400, "invalid_cursor", "cursor is invalid")
        ),
    })

    const response = await createLinksRoute(dependencies).GET(
      new Request("http://localhost/api/links")
    )

    expect(response.status).toBe(400)
    await expect(response.json()).resolves.toEqual({
      error: { code: "invalid_cursor", message: "cursor is invalid" },
    })
  })
})

describe("createLinksRoute POST", () => {
  it("creates a link without exposing authentication tokens", async () => {
    const dependencies = linksDependencies({
      create: vi.fn().mockResolvedValue(link),
    })
    const response = await createLinksRoute(dependencies).POST(
      createRequest({
        url: link.url,
        shortCode: "AbC1234",
        expiresAt: "2026-08-17T08:00:00Z",
      })
    )

    expect(response.status).toBe(201)
    expect(response.headers.get("Cache-Control")).toBe("no-store")
    await expect(response.json()).resolves.toEqual(link)
    expect(dependencies.create).toHaveBeenCalledWith(
      {
        url: link.url,
        shortCode: "AbC1234",
        expiresAt: "2026-08-17T08:00:00Z",
      },
      expect.any(AbortSignal)
    )
  })

  it("rejects cross-origin requests before reading credentials", async () => {
    const dependencies = linksDependencies()

    const response = await createLinksRoute(dependencies).POST(
      createRequest({ url: link.url }, "https://attacker.example")
    )

    expect(response.status).toBe(403)
    expect(dependencies.create).not.toHaveBeenCalled()
  })

  it.each([
    ["malformed JSON", "{"],
    ["an array", "[]"],
    ["a missing URL", JSON.stringify({ shortCode: "AbC1234" })],
    ["unknown properties", JSON.stringify({ url: link.url, ownerID: "user-2" })],
  ])("rejects %s", async (_, body) => {
    const dependencies = linksDependencies()
    const response = await createLinksRoute(dependencies).POST(
      rawCreateRequest(body)
    )

    expect(response.status).toBe(400)
    expect(dependencies.create).not.toHaveBeenCalled()
  })

  it("maps API connection failures to bad gateway", async () => {
    const dependencies = linksDependencies({
      create: vi.fn().mockRejectedValue(new APIConnectionError("offline")),
    })

    const response = await createLinksRoute(dependencies).POST(
      createRequest({ url: link.url })
    )

    expect(response.status).toBe(502)
    await expect(response.json()).resolves.toMatchObject({
      error: { code: "api_unavailable" },
    })
  })
})

function linksDependencies(overrides: Record<string, unknown> = {}) {
  return {
    list: vi.fn(),
    create: vi.fn(),
    ...overrides,
  }
}

function createRequest(body: unknown, origin = "http://localhost") {
  return rawCreateRequest(JSON.stringify(body), origin)
}

function rawCreateRequest(body: string, origin = "http://localhost") {
  return new Request("http://localhost/api/links", {
    method: "POST",
    headers: { "Content-Type": "application/json", Origin: origin },
    body,
  })
}
