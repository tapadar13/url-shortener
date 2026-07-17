import { afterEach, describe, expect, it, vi } from "vitest"

import {
  createLink,
  deleteLink,
  getLink,
  getLinkAnalytics,
  getLinkStats,
  listLinks,
  updateLink,
} from "./browser-links"

afterEach(() => {
  vi.unstubAllGlobals()
})

describe("browser links client", () => {
  it("lists a cursor page through the BFF", async () => {
    const fetchMock = jsonFetch({ items: [], nextCursor: "next-page" })
    vi.stubGlobal("fetch", fetchMock)

    await expect(
      listLinks({ limit: 8, cursor: "opaque+cursor" })
    ).resolves.toEqual({ items: [], nextCursor: "next-page" })
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/links?limit=8&cursor=opaque%2Bcursor",
      expect.objectContaining({ credentials: "same-origin" })
    )
  })

  it("creates a link through the same-origin BFF", async () => {
    const created = linkResponse()
    const fetchMock = jsonFetch(created, 201)
    vi.stubGlobal("fetch", fetchMock)
    const input = {
      url: created.url,
      shortCode: created.shortCode,
      expiresAt: "2026-08-17T08:00:00Z",
    }

    await expect(createLink(input)).resolves.toEqual(created)
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/links",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify(input),
      })
    )
  })

  it("retrieves a link", async () => {
    const link = linkResponse()
    const fetchMock = jsonFetch(link)
    vi.stubGlobal("fetch", fetchMock)

    await expect(getLink(link.shortCode)).resolves.toEqual(link)
    expect(fetchMock).toHaveBeenCalledWith(
      `/api/links/${link.shortCode}`,
      expect.any(Object)
    )
  })

  it("updates a destination", async () => {
    const link = linkResponse()
    const fetchMock = jsonFetch(link)
    vi.stubGlobal("fetch", fetchMock)

    await expect(updateLink(link.shortCode, link.url)).resolves.toEqual(link)
    expect(fetchMock).toHaveBeenCalledWith(
      `/api/links/${link.shortCode}`,
      expect.objectContaining({
        method: "PUT",
        body: JSON.stringify({ url: link.url }),
      })
    )
  })

  it("deletes a link", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }))
    vi.stubGlobal("fetch", fetchMock)

    await expect(deleteLink("AbC1234")).resolves.toBeUndefined()
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/links/AbC1234",
      expect.objectContaining({ method: "DELETE" })
    )
  })

  it("retrieves link statistics", async () => {
    const stats = { ...linkResponse(), accessCount: 42 }
    const fetchMock = jsonFetch(stats)
    vi.stubGlobal("fetch", fetchMock)

    await expect(getLinkStats(stats.shortCode)).resolves.toEqual(stats)
    expect(fetchMock).toHaveBeenCalledWith(
      `/api/links/${stats.shortCode}/stats`,
      expect.any(Object)
    )
  })

  it("retrieves an explicit analytics range", async () => {
    const analytics = {
      shortCode: "AbC1234",
      from: "2026-07-15",
      to: "2026-07-17",
      totalClicks: 12,
      daily: [{ date: "2026-07-17", clicks: 7 }],
    }
    const fetchMock = jsonFetch(analytics)
    vi.stubGlobal("fetch", fetchMock)

    await expect(
      getLinkAnalytics(analytics.shortCode, {
        from: analytics.from,
        to: analytics.to,
      })
    ).resolves.toEqual(analytics)
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/links/AbC1234/analytics?from=2026-07-15&to=2026-07-17",
      expect.any(Object)
    )
  })
})

function linkResponse() {
  return {
    id: "507f1f77bcf86cd799439011",
    url: "https://example.com/articles/123",
    shortCode: "AbC1234",
    shortUrl: "https://rly.to/AbC1234",
    createdAt: "2026-07-17T08:00:00Z",
    updatedAt: "2026-07-17T08:00:00Z",
  }
}

function jsonFetch(body: unknown, status = 200) {
  return vi.fn().mockResolvedValue(
    new Response(JSON.stringify(body), {
      status,
      headers: { "Content-Type": "application/json" },
    })
  )
}
