import { cleanup, render, screen } from "@testing-library/react"
import { afterEach, describe, expect, it } from "vitest"

import type { LinkAnalytics } from "@/lib/links/types"

import { LinkAnalyticsChart } from "./link-analytics-chart"

afterEach(cleanup)

describe("LinkAnalyticsChart", () => {
  it("renders the API series with an accessible summary", () => {
    render(<LinkAnalyticsChart analytics={analytics} />)

    expect(
      screen.getByRole("img", {
        name: "12 clicks from Jul 15 to Jul 17",
      })
    ).toBeDefined()
    expect(screen.getByTitle("Jul 15: 5 clicks")).toBeDefined()
    expect(screen.getByTitle("Jul 16: 0 clicks")).toBeDefined()
    expect(screen.getByTitle("Jul 17: 7 clicks")).toBeDefined()
  })

  it("shows an explicit empty state for a zero-click range", () => {
    render(
      <LinkAnalyticsChart
        analytics={{
          ...analytics,
          totalClicks: 0,
          daily: analytics.daily.map((point) => ({ ...point, clicks: 0 })),
        }}
      />
    )

    expect(screen.getByText("No clicks in this range")).toBeDefined()
  })
})

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
