import { describe, expect, it, vi } from "vitest"

import { APIRequestError } from "@/lib/api/client"

import { createLinkRoute } from "./link-route"
import type { ShortLink } from "./types"

const link: ShortLink = {
  id: "507f1f77bcf86cd799439011",
  url: "https://example.com/articles/123",
  shortCode: "AbC1234",
  shortUrl: "https://rly.to/AbC1234",
  createdAt: "2026-07-17T08:00:00Z",
  updatedAt: "2026-07-17T08:00:00Z",
}

describe("createLinkRoute GET", () => {
  it("retrieves an authenticated link", async () => {
    const dependencies = linkDependencies({
      get: vi.fn().mockResolvedValue(link),
    })

    const response = await createLinkRoute(dependencies).GET(
      linkRequest(),
      routeContext()
    )

    expect(response.status).toBe(200)
    expect(response.headers.get("Cache-Control")).toBe("no-store")
    await expect(response.json()).resolves.toEqual(link)
    expect(dependencies.get).toHaveBeenCalledWith(
      link.shortCode,
      expect.any(AbortSignal)
    )
  })

  it("preserves a missing-link API error", async () => {
    const dependencies = linkDependencies({
      get: vi
        .fn()
        .mockRejectedValue(
          new APIRequestError(404, "not_found", "short URL was not found")
        ),
    })

    const response = await createLinkRoute(dependencies).GET(
      linkRequest(),
      routeContext()
    )

    expect(response.status).toBe(404)
    await expect(response.json()).resolves.toMatchObject({
      error: { code: "not_found" },
    })
  })
})

describe("createLinkRoute PUT", () => {
  it("updates a link destination", async () => {
    const updated = { ...link, url: "https://example.com/new-destination" }
    const dependencies = linkDependencies({
      update: vi.fn().mockResolvedValue(updated),
    })

    const response = await createLinkRoute(dependencies).PUT(
      linkRequest({ method: "PUT", body: { url: updated.url } }),
      routeContext()
    )

    expect(response.status).toBe(200)
    await expect(response.json()).resolves.toEqual(updated)
    expect(dependencies.update).toHaveBeenCalledWith(
      link.shortCode,
      updated.url,
      expect.any(AbortSignal)
    )
  })

  it.each([
    ["malformed JSON", "{"],
    ["a missing URL", "{}"],
    ["unknown properties", JSON.stringify({ url: link.url, ownerID: "user-2" })],
  ])("rejects %s", async (_, body) => {
    const dependencies = linkDependencies()
    const response = await createLinkRoute(dependencies).PUT(
      rawLinkRequest("PUT", body),
      routeContext()
    )

    expect(response.status).toBe(400)
    expect(dependencies.update).not.toHaveBeenCalled()
  })
})

describe("createLinkRoute DELETE", () => {
  it("deletes a link without a response body", async () => {
    const dependencies = linkDependencies({
      delete: vi.fn().mockResolvedValue(undefined),
    })

    const response = await createLinkRoute(dependencies).DELETE(
      linkRequest({ method: "DELETE" }),
      routeContext()
    )

    expect(response.status).toBe(204)
    expect(await response.text()).toBe("")
    expect(dependencies.delete).toHaveBeenCalledWith(
      link.shortCode,
      expect.any(AbortSignal)
    )
  })
})

describe("link route request protection", () => {
  it.each(["PUT", "DELETE"] as const)(
    "rejects a cross-origin %s before accessing the API",
    async (method) => {
      const dependencies = linkDependencies()
      const handler = createLinkRoute(dependencies)[method]
      const response = await handler(
        linkRequest({
          method,
          body: method === "PUT" ? { url: link.url } : undefined,
          origin: "https://attacker.example",
        }),
        routeContext()
      )

      expect(response.status).toBe(403)
      expect(dependencies.update).not.toHaveBeenCalled()
      expect(dependencies.delete).not.toHaveBeenCalled()
    }
  )

  it.each(["GET", "PUT", "DELETE"] as const)(
    "rejects an invalid short code for %s",
    async (method) => {
      const dependencies = linkDependencies()
      const handler = createLinkRoute(dependencies)[method]
      const response = await handler(
        linkRequest({
          method,
          body: method === "PUT" ? { url: link.url } : undefined,
        }),
        routeContext("not/valid")
      )

      expect(response.status).toBe(400)
      expect(dependencies.get).not.toHaveBeenCalled()
      expect(dependencies.update).not.toHaveBeenCalled()
      expect(dependencies.delete).not.toHaveBeenCalled()
    }
  )
})

function linkDependencies(overrides: Record<string, unknown> = {}) {
  return {
    get: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    ...overrides,
  }
}

function routeContext(shortCode = link.shortCode) {
  return { params: Promise.resolve({ shortCode }) }
}

function linkRequest(
  options: {
    method?: "GET" | "PUT" | "DELETE"
    body?: unknown
    origin?: string
  } = {}
) {
  return rawLinkRequest(
    options.method ?? "GET",
    options.body === undefined ? undefined : JSON.stringify(options.body),
    options.origin
  )
}

function rawLinkRequest(
  method: "GET" | "PUT" | "DELETE",
  body?: string,
  origin = "http://localhost"
) {
  return new Request(`http://localhost/api/links/${link.shortCode}`, {
    method,
    headers: { "Content-Type": "application/json", Origin: origin },
    body,
  })
}
