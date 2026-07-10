import { Check, Copy, Link2 } from "lucide-react"

import { MockWindow } from "@/components/landing/mock-window"
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
        <div className="space-y-1.5 rounded-md border p-3 font-mono text-[11px] text-muted-foreground">
          <p className="flex items-center justify-between">
            <span>created_at</span>
            <span className="text-foreground">2026-06-12 09:41 UTC</span>
          </p>
          <p className="flex items-center justify-between">
            <span>updated_at</span>
            <span className="text-foreground">2026-07-01 14:03 UTC</span>
          </p>
          <p className="flex items-center justify-between">
            <span>last_accessed_at</span>
            <span className="text-foreground">2026-07-10 08:12 UTC</span>
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
    body: "Paste any HTTP or HTTPS destination and get back a compact six-character Base62 code. Codes are random, checked against a unique index, and retried on the rare collision — so every link you mint is genuinely one of a kind.",
    bullets: [
      "Destinations validated before a code is issued",
      "Random Base62 codes with collision retries",
      "A consistent JSON response for every request",
    ],
    mockup: CreateMockup,
  },
  {
    eyebrow: "Manage",
    title: "Destinations you can change your mind about",
    body: "A short link shouldn't be a one-way door. Repoint an existing code to a new destination without breaking anything already shared — the code stays stable, the timestamps record the change, and the visit history carries over.",
    bullets: [
      "Update a destination while the short code stays put",
      "Created and updated timestamps on every link",
      "Delete links you're finished with",
    ],
    mockup: ManageMockup,
  },
  {
    eyebrow: "Measure",
    title: "Track visits without the clutter",
    body: "Each redirect is designed to increment one honest counter, atomically, at the moment it happens. You get the numbers that matter — total visits and when they occurred — without a vanity dashboard in the way.",
    bullets: [
      "Access counts recorded at redirect time",
      "Last-visit and lifecycle timestamps per link",
      "Stats available over the same JSON API",
    ],
    mockup: StatsMockup,
  },
] as const

export function Capabilities() {
  return (
    <section id="product" className="scroll-mt-14">
      <div className="mx-auto max-w-6xl px-4 py-20 sm:px-6 sm:py-28">
        <div className="mx-auto max-w-2xl text-center">
          <p className="text-sm font-medium text-brand">Product</p>
          <h2 className="mt-2 text-3xl font-semibold text-balance sm:text-4xl">
            Everything a link needs. Nothing it doesn&apos;t.
          </h2>
          <p className="mt-4 text-base text-pretty text-muted-foreground">
            {siteConfig.name} keeps the whole lifecycle of a short link — create,
            repoint, measure — in one focused workspace.
          </p>
        </div>

        <div className="mt-16 space-y-20 sm:mt-20 sm:space-y-24">
          {capabilities.map((capability, index) => (
            <div
              key={capability.title}
              className="grid items-center gap-8 lg:grid-cols-2 lg:gap-16"
            >
              <div className={cn(index % 2 === 1 && "lg:order-last")}>
                <p className="text-sm font-medium text-brand">
                  {capability.eyebrow}
                </p>
                <h3 className="mt-2 text-2xl font-semibold text-balance">
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
              </div>
              <capability.mockup />
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
