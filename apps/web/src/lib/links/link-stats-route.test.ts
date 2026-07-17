import { describe, expect, it, vi } from "vitest"

import { APIRequestError } from "@/lib/api/client"

import { createLinkStatsRoute } from "./link-stats-route"
import type { LinkStats } from "./types"

const stats: LinkStats = {
  id: "507f1f77bcf86cd799439011",
  url: "https://example.com/articles/123",
  shortCode: "AbC1234",
  shortUrl: "https://rly.to/AbC1234",
  accessCount: 42,
  createdAt: "2026-07-17T08:00:00Z",
  updatedAt: "2026-07-17T08:00:00Z",
  lastAccessedAt: "2026-07-17T09:15:00Z",
}

describe("createLinkStatsRoute", () => {
  it("returns authenticated link statistics", async () => {
    const getStats = vi.fn().mockResolvedValue(stats)
    const handler = createLinkStatsRoute({ getStats })

    const response = await handler(
      new Request("http://localhost/api/links/AbC1234/stats"),
      routeContext()
    )

    expect(response.status).toBe(200)
    expect(response.headers.get("Cache-Control")).toBe("no-store")
    await expect(response.json()).resolves.toEqual(stats)
    expect(getStats).toHaveBeenCalledWith(
      stats.shortCode,
      expect.any(AbortSignal)
    )
  })

  it("rejects an invalid short code before accessing the API", async () => {
    const getStats = vi.fn()
    const handler = createLinkStatsRoute({ getStats })

    const response = await handler(
      new Request("http://localhost/api/links/invalid/stats"),
      routeContext("not/valid")
    )

    expect(response.status).toBe(400)
    expect(getStats).not.toHaveBeenCalled()
  })

  it("preserves structured API errors", async () => {
    const getStats = vi
      .fn()
      .mockRejectedValue(
        new APIRequestError(404, "not_found", "short URL was not found")
      )
    const handler = createLinkStatsRoute({ getStats })

    const response = await handler(
      new Request("http://localhost/api/links/AbC1234/stats"),
      routeContext()
    )

    expect(response.status).toBe(404)
    await expect(response.json()).resolves.toEqual({
      error: { code: "not_found", message: "short URL was not found" },
    })
  })
})

function routeContext(shortCode = stats.shortCode) {
  return { params: Promise.resolve({ shortCode }) }
}
