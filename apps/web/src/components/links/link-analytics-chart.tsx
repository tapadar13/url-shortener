import { BarChart3 } from "lucide-react"

import { formatCount } from "@/lib/format"
import type { LinkAnalytics } from "@/lib/links/types"

const dateFormatter = new Intl.DateTimeFormat("en-US", {
  month: "short",
  day: "numeric",
  timeZone: "UTC",
})

export function LinkAnalyticsChart({
  analytics,
}: {
  analytics: LinkAnalytics
}) {
  const maximum = Math.max(...analytics.daily.map((point) => point.clicks), 0)
  const middleIndex = Math.floor((analytics.daily.length - 1) / 2)
  const first = analytics.daily[0]
  const middle = analytics.daily[middleIndex]
  const last = analytics.daily.at(-1)

  return (
    <div className="rounded-xl border border-foreground/8 bg-card/70 p-3.5">
      <div className="flex items-end justify-between gap-3">
        <div>
          <p className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
            <BarChart3 className="size-3" aria-hidden="true" />
            Clicks in range
          </p>
          <p className="mt-1 font-mono text-xl font-semibold tabular-nums">
            {formatCount(analytics.totalClicks)}
          </p>
        </div>
        <p className="text-right text-[10px] text-muted-foreground/70">
          Daily UTC
        </p>
      </div>

      <div
        role="img"
        aria-label={`${clickLabel(analytics.totalClicks)} from ${formatAnalyticsDate(analytics.from)} to ${formatAnalyticsDate(analytics.to)}`}
        className="relative mt-4"
      >
        <div className="flex h-28 items-end gap-px border-b border-foreground/10">
          {analytics.daily.map((point) => (
            <div
              key={point.date}
              className="flex h-full min-w-0 flex-1 items-end"
              title={`${formatAnalyticsDate(point.date)}: ${formatCount(point.clicks)} ${point.clicks === 1 ? "click" : "clicks"}`}
              aria-hidden="true"
            >
              <span
                className="block w-full min-w-0 rounded-t-[2px] bg-brand/80 transition-[height,opacity] duration-300"
                style={{
                  height:
                    point.clicks === 0
                      ? "2px"
                      : `${Math.max(8, (point.clicks / maximum) * 100)}%`,
                  opacity: point.clicks === 0 ? 0.2 : 1,
                }}
              />
            </div>
          ))}
        </div>

        {analytics.totalClicks === 0 && (
          <p className="pointer-events-none absolute inset-x-0 top-1/2 -translate-y-1/2 text-center text-xs text-muted-foreground">
            No clicks in this range
          </p>
        )}

        {first && middle && last && (
          <div className="mt-2 grid grid-cols-3 font-mono text-[9px] text-muted-foreground/60">
            <span>{formatAnalyticsDate(first.date)}</span>
            <span className="text-center">{formatAnalyticsDate(middle.date)}</span>
            <span className="text-right">{formatAnalyticsDate(last.date)}</span>
          </div>
        )}
      </div>
    </div>
  )
}

function formatAnalyticsDate(date: string): string {
  return dateFormatter.format(new Date(`${date}T00:00:00Z`))
}

function clickLabel(clicks: number): string {
  return `${formatCount(clicks)} ${clicks === 1 ? "click" : "clicks"}`
}
