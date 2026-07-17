import { describe, expect, it } from "vitest"

import { analyticsDateRange } from "./analytics-range"

describe("analyticsDateRange", () => {
  it("returns an inclusive UTC range across a year boundary", () => {
    expect(
      analyticsDateRange(7, new Date("2026-01-03T12:00:00Z"))
    ).toEqual({
      from: "2025-12-28",
      to: "2026-01-03",
    })
  })

  it("uses the UTC date instead of the browser's local date", () => {
    expect(
      analyticsDateRange(30, new Date("2026-07-18T00:30:00+05:30"))
    ).toEqual({
      from: "2026-06-18",
      to: "2026-07-17",
    })
  })

  it("handles leap days in a 30-day range", () => {
    expect(
      analyticsDateRange(30, new Date("2024-03-01T08:00:00Z"))
    ).toEqual({
      from: "2024-02-01",
      to: "2024-03-01",
    })
  })
})
