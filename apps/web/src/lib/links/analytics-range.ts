export const analyticsRangeOptions = [7, 30, 90] as const

export type AnalyticsRangeDays = (typeof analyticsRangeOptions)[number]

interface AnalyticsDateRange {
  from: string
  to: string
}

export function analyticsDateRange(
  days: AnalyticsRangeDays,
  now = new Date()
): AnalyticsDateRange {
  const to = new Date(
    Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate())
  )
  const from = new Date(to)
  from.setUTCDate(from.getUTCDate() - days + 1)

  return {
    from: isoDate(from),
    to: isoDate(to),
  }
}

function isoDate(date: Date): string {
  return date.toISOString().slice(0, 10)
}
