import { ArrowDown, Check, Copy, Link2 } from "lucide-react"

import { MockWindow } from "@/components/landing/mock-window"
import { Reveal } from "@/components/landing/reveal"
import { siteConfig } from "@/config/site"
import { cn } from "@/lib/utils"

function EndpointChip({ children }: { children: React.ReactNode }) {
  return (
    <span className="inline-flex items-center rounded-md border bg-muted/50 px-2 py-0.5 font-mono text-[11px] text-muted-foreground">
      {children}
    </span>
  )
}

function FieldLabel({ children }: { children: React.ReactNode }) {
  return (
    <p className="text-[11px] font-medium text-muted-foreground">{children}</p>
  )
}

function CreateMockup() {
  return (
    <MockWindow url={`app.${siteConfig.shortHost}/new`}>
      <div className="space-y-3 p-4">
        <div className="flex items-center justify-between">
          <p className="text-sm font-medium">Shorten a link</p>
          <EndpointChip>POST /shorten</EndpointChip>
        </div>
        <div>
          <FieldLabel>Destination</FieldLabel>
          <div className="mt-1 flex h-9 items-center gap-2 rounded-md border bg-background px-3">
            <Link2
              className="size-3.5 shrink-0 text-muted-foreground"
              aria-hidden="true"
            />
            <span className="truncate font-mono text-xs">
              https://docs.example.com/changelog/2026-june-release-notes
            </span>
          </div>
          <p className="mt-1.5 text-[11px] text-muted-foreground">
            HTTP and HTTPS destinations are validated before a code is issued.
          </p>
        </div>
        <div className="flex items-center justify-between rounded-md border border-brand/30 bg-brand-muted/60 px-3 py-2.5">
          <span className="font-mono text-xs font-medium text-brand">
            {siteConfig.shortHost}/Q9mBf3
          </span>
          <span className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
            <Copy className="size-3" aria-hidden="true" />
            Copy
          </span>
        </div>
        <p className="font-mono text-[11px] text-muted-foreground">
          6-char Base62 · unique index checked · 201 Created
        </p>
      </div>
    </MockWindow>
  )
}

function ManageMockup() {
  return (
    <MockWindow url={`app.${siteConfig.shortHost}/links/x7Kd2a`}>
      <div className="space-y-3 p-4">
        <div className="flex items-center justify-between">
          <p className="font-mono text-sm font-medium">
            {siteConfig.shortHost}/x7Kd2a
          </p>
          <EndpointChip>PUT /shorten/x7Kd2a</EndpointChip>
        </div>
        <div>
          <FieldLabel>Current destination</FieldLabel>
          <div className="mt-1 flex h-9 items-center rounded-md border bg-muted/40 px-3">
            <span className="truncate font-mono text-xs text-muted-foreground line-through">
              example.com/spring-launch/announcement
            </span>
          </div>
        </div>
        <div className="flex justify-center" aria-hidden="true">
          <ArrowDown className="size-3.5 text-muted-foreground/60" />
        </div>
        <div>
          <FieldLabel>New destination</FieldLabel>
          <div className="mt-1 flex h-9 items-center justify-between gap-2 rounded-md border border-brand/40 bg-background px-3">
            <span className="truncate font-mono text-xs">
              example.com/summer-launch/announcement
            </span>
            <span className="shrink-0 rounded bg-primary px-2 py-0.5 text-[11px] font-medium text-primary-foreground">
              Save
            </span>
          </div>
        </div>
        <p className="font-mono text-[11px] text-muted-foreground">
          same short code · updated_at refreshed · visits preserved
        </p>
      </div>
    </MockWindow>
  )
}

function StatsMockup() {
  return (
    <MockWindow url={`app.${siteConfig.shortHost}/links/x7Kd2a/stats`}>
      <div className="space-y-3 p-4">
        <div className="flex items-center justify-between">
          <p className="font-mono text-sm font-medium">
            {siteConfig.shortHost}/x7Kd2a
          </p>
          <EndpointChip>GET /shorten/x7Kd2a/stats</EndpointChip>
        </div>
        <div className="grid grid-cols-2 gap-2">
          <div className="rounded-md border p-3">
            <FieldLabel>Total visits</FieldLabel>
            <p className="mt-1 font-mono text-2xl font-semibold tabular-nums">
              1,284
            </p>
          </div>
          <div className="rounded-md border p-3">
            <FieldLabel>Last visit</FieldLabel>
            <p className="mt-1 font-mono text-2xl font-semibold">2m ago</p>
          </div>
        </div>
        <div className="space-y-1.5 rounded-md border p-3 text-[11px] text-muted-foreground">
          <p className="flex items-center justify-between">
            <span>Created</span>
            <span className="font-mono text-foreground">Jun 12, 09:41 UTC</span>
          </p>
          <p className="flex items-center justify-between">
            <span>Destination updated</span>
            <span className="font-mono text-foreground">Jul 1, 14:03 UTC</span>
          </p>
          <p className="flex items-center justify-between">
            <span>Last visited</span>
            <span className="font-mono text-foreground">Jul 10, 08:12 UTC</span>
          </p>
        </div>
      </div>
    </MockWindow>
  )
}

const capabilities = [
  {
    eyebrow: "Create",
    title: "Clean short links, on the first try",
    body: "Paste any web address and get back a compact six-character link. Every code is checked for uniqueness before it's issued — so the link you share is yours, and yours alone.",
    bullets: [
      "Destinations validated before a link is created",
      "Unique codes, guaranteed — collisions retried automatically",
      "The same clean JSON API behind every action",
    ],
    mockup: CreateMockup,
  },
  {
    eyebrow: "Manage",
    title: "Change where a link points — anytime",
    body: "A short link shouldn't be a one-way door. Point an existing link at a new destination without breaking anything you've already shared: the short link stays the same, and its visit history carries over.",
    bullets: [
      "Update the destination; the short link never changes",
      "Every change is timestamped",
      "Delete links you're finished with",
    ],
    mockup: ManageMockup,
  },
  {
    eyebrow: "Measure",
    title: "Know exactly how often it's opened",
    body: "Every visit bumps one honest counter the moment the redirect happens. You get the numbers that matter — total visits and when they happened — without a vanity dashboard in the way.",
    bullets: [
      "Visits counted at the moment of redirect",
      "Last-visit and lifecycle timestamps per link",
      "The same stats, available over the API",
    ],
    mockup: StatsMockup,
  },
] as const

export function Capabilities() {
  return (
    <section id="product">
      <div className="mx-auto max-w-6xl px-4 py-20 sm:px-6 sm:py-28">
        <Reveal className="mx-auto max-w-2xl text-center">
          <p className="text-xs font-semibold tracking-[0.14em] text-brand uppercase">
            Product
          </p>
          <h2 className="mt-3 text-3xl font-semibold text-balance sm:text-4xl">
            Everything a link needs. Nothing it doesn&apos;t.
          </h2>
          <p className="mt-4 text-base text-pretty text-muted-foreground">
            Create a link, change where it points, and see how it performs —
            the entire lifecycle in one focused workspace.
          </p>
        </Reveal>

        <div className="mt-16 space-y-20 sm:mt-20 sm:space-y-24">
          {capabilities.map((capability, index) => (
            <div
              key={capability.title}
              className="grid items-center gap-8 lg:grid-cols-2 lg:gap-16"
            >
              <Reveal className={cn(index % 2 === 1 && "lg:order-last")}>
                <p className="text-xs font-semibold tracking-[0.14em] text-brand uppercase">
                  {capability.eyebrow}
                </p>
                <h3 className="mt-3 text-2xl font-semibold text-balance">
                  {capability.title}
                </h3>
                <p className="mt-3 text-pretty text-muted-foreground">
                  {capability.body}
                </p>
                <ul className="mt-5 space-y-2.5">
                  {capability.bullets.map((bullet) => (
                    <li key={bullet} className="flex items-start gap-2.5">
                      <Check
                        className="mt-0.5 size-4 shrink-0 text-brand"
                        aria-hidden="true"
                      />
                      <span className="text-sm">{bullet}</span>
                    </li>
                  ))}
                </ul>
              </Reveal>
              <Reveal delay={120}>
                <capability.mockup />
              </Reveal>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
