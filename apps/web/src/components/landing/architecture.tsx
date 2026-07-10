import {
  ArrowDown,
  Database,
  FileText,
  Layers,
  ShieldCheck,
  TimerReset,
} from "lucide-react"

import { siteConfig } from "@/config/site"

const practices = [
  {
    icon: Layers,
    title: "Layered by design",
    body: "HTTP transport, business logic, and persistence live in separate layers, so each piece can be tested and evolved on its own.",
  },
  {
    icon: TimerReset,
    title: "Context-aware requests",
    body: "Every request carries a context with deadlines, and the service shuts down gracefully instead of dropping in-flight work.",
  },
  {
    icon: ShieldCheck,
    title: "Collisions handled, not hoped away",
    body: "A unique MongoDB index guarantees no two links share a code; generation retries within a strict bound if a clash ever occurs.",
  },
  {
    icon: FileText,
    title: "Structured operations",
    body: "Requests are logged with structured fields rather than free-form strings, built for the day this runs somewhere that matters.",
  },
] as const

const flow = [
  {
    label: "Browser",
    detail: `GET ${siteConfig.shortHost}/x7Kd2a`,
  },
  {
    label: "Go API — chi router",
    detail: "validate short code · request context & deadline",
  },
  {
    label: "URL service",
    detail: "look up the link · record the visit",
  },
  {
    label: "MongoDB",
    detail: "unique index on code · atomic visit increment",
  },
] as const

export function Architecture() {
  return (
    <section id="architecture">
      <div className="mx-auto max-w-6xl px-4 py-20 sm:px-6 sm:py-28">
        <div className="grid gap-12 lg:grid-cols-2 lg:gap-16">
          <div>
            <p className="text-sm font-medium text-brand">Architecture</p>
            <h2 className="mt-2 text-3xl font-semibold text-balance sm:text-4xl">
              A small service, engineered like a big one
            </h2>
            <p className="mt-4 text-pretty text-muted-foreground">
              {`${siteConfig.name} is a Go service in front of MongoDB — no framework magic, no hidden layers. The parts you'd audit first are the parts that got the most attention.`}
            </p>

            <dl className="mt-8 space-y-6">
              {practices.map((practice) => (
                <div key={practice.title} className="flex gap-3.5">
                  <practice.icon
                    className="mt-0.5 size-4.5 shrink-0 text-brand"
                    aria-hidden="true"
                  />
                  <div>
                    <dt className="text-sm font-semibold">{practice.title}</dt>
                    <dd className="mt-1 text-sm text-pretty text-muted-foreground">
                      {practice.body}
                    </dd>
                  </div>
                </div>
              ))}
            </dl>
          </div>

          <div className="lg:pt-10">
            <div className="rounded-lg border bg-muted/40 p-5 sm:p-6">
              <p className="flex items-center gap-2 text-[11px] font-medium text-muted-foreground">
                <Database className="size-3.5" aria-hidden="true" />
                One redirect, end to end
              </p>
              <div className="mt-4 flex flex-col items-stretch">
                {flow.map((node, index) => (
                  <div key={node.label} className="flex flex-col items-center">
                    {index > 0 && (
                      <ArrowDown
                        className="my-1.5 size-3.5 text-muted-foreground/60"
                        aria-hidden="true"
                      />
                    )}
                    <div className="w-full rounded-md border bg-background px-4 py-3">
                      <p className="text-sm font-medium">{node.label}</p>
                      <p className="mt-0.5 font-mono text-[11px] text-muted-foreground">
                        {node.detail}
                      </p>
                    </div>
                  </div>
                ))}
                <ArrowDown
                  className="my-1.5 self-center size-3.5 text-muted-foreground/60"
                  aria-hidden="true"
                />
                <div className="w-full rounded-md border border-brand/30 bg-brand-muted/60 px-4 py-3">
                  <p className="text-sm font-medium">
                    302 → original destination
                  </p>
                  <p className="mt-0.5 font-mono text-[11px] text-muted-foreground">
                    visit recorded · access_count +1
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
