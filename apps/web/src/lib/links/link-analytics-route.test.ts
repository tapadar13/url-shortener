import { describe, expect, it, vi } from "vitest"

import { APIRequestError } from "@/lib/api/client"

import { createLinkAnalyticsRoute } from "./link-analytics-route"
import type { LinkAnalytics } from "./types"

const analytics: LinkAnalytics = {
  shortCode: "AbC1234",
  from: "2026-07-15",
  to: "2026-07-17",
  totalClicks: 12,
  daily: [
    { date: "2026-07-15", clicks: 5 },
    { date: "2026-07-16", clicks: 0 },
    { date: "2026-07-17", clicks: 7 },
  ],
}

describe("createLinkAnalyticsRoute", () => {
  it("forwards supported analytics dates", async () => {
    const getAnalytics = vi.fn().mockResolvedValue(analytics)
    const handler = createLinkAnalyticsRoute({ getAnalytics })

    const response = await handler(
      new Request(
        "http://localhost/api/links/AbC1234/analytics?from=2026-07-15&to=2026-07-17&ignored=value"
      ),
      routeContext()
    )

    expect(response.status).toBe(200)
    expect(response.headers.get("Cache-Control")).toBe("no-store")
    await expect(response.json()).resolves.toEqual(analytics)
    expect(getAnalytics).toHaveBeenCalledWith(
      analytics.shortCode,
      "?from=2026-07-15&to=2026-07-17",
      expect.any(AbortSignal)
    )
  })

  it("uses the API default range when dates are omitted", async () => {
    const getAnalytics = vi.fn().mockResolvedValue(analytics)
    const handler = createLinkAnalyticsRoute({ getAnalytics })

    await handler(
      new Request("http://localhost/api/links/AbC1234/analytics"),
      routeContext()
    )

    expect(getAnalytics).toHaveBeenCalledWith(
      analytics.shortCode,
      "",
      expect.any(AbortSignal)
    )
  })

  it("rejects an invalid short code before accessing the API", async () => {
    const getAnalytics = vi.fn()
    const handler = createLinkAnalyticsRoute({ getAnalytics })

    const response = await handler(
      new Request("http://localhost/api/links/invalid/analytics"),
      routeContext("not/valid")
    )

    expect(response.status).toBe(400)
    expect(getAnalytics).not.toHaveBeenCalled()
  })

  it("preserves analytics range errors", async () => {
    const getAnalytics = vi
      .fn()
      .mockRejectedValue(
        new APIRequestError(
          400,
          "invalid_date_range",
          "analytics date range is invalid"
        )
      )
    const handler = createLinkAnalyticsRoute({ getAnalytics })

    const response = await handler(
      new Request("http://localhost/api/links/AbC1234/analytics"),
      routeContext()
    )

    expect(response.status).toBe(400)
    await expect(response.json()).resolves.toEqual({
      error: {
        code: "invalid_date_range",
        message: "analytics date range is invalid",
      },
    })
  })
})

function routeContext(shortCode = analytics.shortCode) {
  return { params: Promise.resolve({ shortCode }) }
}
