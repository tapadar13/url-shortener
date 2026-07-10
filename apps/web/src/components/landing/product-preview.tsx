import {
  ArrowRight,
  BarChart2,
  CheckCircle2,
  Clock,
  Copy,
  CornerDownRight,
  Link2,
  MousePointerClick,
} from "lucide-react"

import { MockWindow } from "@/components/landing/mock-window"
import { siteConfig } from "@/config/site"

const sampleLinks = [
  {
    code: "x7Kd2a",
    destination: "example.com/spring-launch/announcement?utm_source=news…",
    visits: "1,284",
    lastAccessed: "2m ago",
  },
  {
    code: "Q9mBf3",
    destination: "docs.example.com/changelog/2026-june-release-notes",
    visits: "962",
    lastAccessed: "18m ago",
  },
  {
    code: "aT5nWc",
    destination: "github.com/tapadar13/url-shortener",
    visits: "417",
    lastAccessed: "1h ago",
  },
] as const

function ShortenBar() {
  return (
    <div className="flex items-center gap-2">
      <div className="flex h-9 min-w-0 flex-1 items-center gap-2 rounded-md border bg-background px-3">
        <Link2
          className="size-3.5 shrink-0 text-muted-foreground"
          aria-hidden="true"
        />
        <span className="truncate font-mono text-xs text-foreground">
          https://example.com/spring-launch/announcement?utm_source=newsletter
        </span>
      </div>
      <div className="flex h-9 shrink-0 items-center rounded-md bg-primary px-3.5 text-xs font-medium text-primary-foreground">
        Shorten
      </div>
    </div>
  )
}

function FreshResult() {
  return (
    <div className="flex flex-wrap items-center gap-x-2.5 gap-y-1.5 rounded-md border border-brand/30 bg-brand-muted/60 px-3 py-2.5">
      <CheckCircle2 className="size-3.5 text-brand" aria-hidden="true" />
      <span className="font-mono text-xs font-medium text-brand">
        {siteConfig.shortHost}/x7Kd2a
      </span>
      <Copy className="size-3.5 text-muted-foreground" aria-hidden="true" />
      <span className="flex min-w-0 flex-1 items-center gap-1.5 text-[11px] text-muted-foreground">
        <span className="shrink-0">redirects to</span>
        <span className="truncate font-mono text-xs">
          example.com/spring-launch/announcement…
        </span>
      </span>
      <span className="text-[11px] text-muted-foreground">
        created just now
      </span>
    </div>
  )
}

function LinkTable() {
  return (
    <div className="overflow-hidden rounded-md border">
      <div className="grid grid-cols-[1fr_auto] items-center gap-2 border-b bg-muted/50 px-3 py-1.5 text-[11px] font-medium text-muted-foreground sm:grid-cols-[8rem_1fr_4rem_5rem]">
        <span>Short link</span>
        <span className="hidden sm:block">Destination</span>
        <span className="text-right">Visits</span>
        <span className="hidden text-right sm:block">Last visit</span>
      </div>
      {sampleLinks.map((link) => (
        <div
          key={link.code}
          className="grid grid-cols-[1fr_auto] items-center gap-2 border-b px-3 py-2 transition-colors last:border-b-0 hover:bg-muted/40 sm:grid-cols-[8rem_1fr_4rem_5rem]"
        >
          <span className="truncate font-mono text-xs font-medium">
            /{link.code}
          </span>
          <span className="hidden truncate font-mono text-[11px] text-muted-foreground sm:block">
            {link.destination}
          </span>
          <span className="text-right font-mono text-xs tabular-nums">
            {link.visits}
          </span>
          <span className="hidden text-right text-[11px] text-muted-foreground sm:block">
            {link.lastAccessed}
          </span>
        </div>
      ))}
    </div>
  )
}

function DetailPanel() {
  return (
    <aside className="hidden w-64 shrink-0 flex-col gap-3 border-l bg-muted/30 p-4 lg:flex">
      <div>
        <p className="text-[11px] font-medium text-muted-foreground">
          Link details
        </p>
        <p className="mt-1 font-mono text-sm font-medium text-brand">
          {siteConfig.shortHost}/x7Kd2a
        </p>
      </div>

      <div className="rounded-md border bg-background p-2.5">
        <p className="text-[11px] font-medium text-muted-foreground">
          Destination
        </p>
        <p className="mt-1 truncate font-mono text-[11px]">
          example.com/spring-launch/announcement…
        </p>
      </div>

      <div className="grid grid-cols-2 gap-2">
        <div className="rounded-md border bg-background p-2.5">
          <p className="flex items-center gap-1 text-[11px] text-muted-foreground">
            <BarChart2 className="size-3" aria-hidden="true" />
            Total visits
          </p>
          <p className="mt-1 font-mono text-lg font-semibold tabular-nums">
            1,284
          </p>
        </div>
        <div className="rounded-md border bg-background p-2.5">
          <p className="flex items-center gap-1 text-[11px] text-muted-foreground">
            <Clock className="size-3" aria-hidden="true" />
            Last visit
          </p>
          <p className="mt-1 font-mono text-lg font-semibold">2m</p>
        </div>
      </div>

      <div className="rounded-md border bg-background p-2.5">
        <p className="flex items-center gap-1.5 text-[11px] font-medium text-muted-foreground">
          <MousePointerClick className="size-3" aria-hidden="true" />
          Latest visit
        </p>
        <div className="mt-2 space-y-1 font-mono text-[11px]">
          <p className="text-muted-foreground">
            someone opens&nbsp;
            <span className="text-foreground">/x7Kd2a</span>
          </p>
          <p className="flex items-center gap-1 text-muted-foreground">
            <CornerDownRight className="size-3" aria-hidden="true" />
            redirected in one hop
          </p>
          <p className="text-brand">visit count +1 · recorded instantly</p>
        </div>
      </div>
    </aside>
  )
}

const legend = [
  { icon: Link2, label: "Paste a long URL" },
  { icon: ArrowRight, label: "Get a short link" },
  { icon: BarChart2, label: "Every visit counted" },
] as const

export function ProductPreview() {
  return (
    <figure className="mx-auto w-full max-w-4xl">
      <MockWindow url={`app.${siteConfig.shortHost}/links`}>
        <div className="flex">
          <div className="min-w-0 flex-1 space-y-3 p-4">
            <ShortenBar />
            <FreshResult />
            <LinkTable />
          </div>
          <DetailPanel />
        </div>
      </MockWindow>
      <figcaption className="mt-4 space-y-2 text-center">
        <span className="flex flex-wrap items-center justify-center gap-x-5 gap-y-1.5 text-xs font-medium text-muted-foreground">
          {legend.map((item, index) => (
            <span key={item.label} className="flex items-center gap-1.5">
              <span
                className="flex size-4 items-center justify-center rounded-full bg-muted font-mono text-[10px] text-muted-foreground"
                aria-hidden="true"
              >
                {index + 1}
              </span>
              {item.label}
            </span>
          ))}
        </span>
        <span className="block text-xs text-muted-foreground/70">
          Illustrative preview — not a live workspace yet.
        </span>
      </figcaption>
    </figure>
  )
}
