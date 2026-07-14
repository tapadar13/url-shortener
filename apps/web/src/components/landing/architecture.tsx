import { ArrowDown, Braces, Check, Database, Gauge, ShieldCheck } from "lucide-react"

import { Reveal } from "@/components/landing/reveal"
import { siteConfig } from "@/config/site"

const assurances = [
  { icon: ShieldCheck, title: "Collision-safe by default", body: "Every generated code is protected by a unique database index, with bounded retries if a collision occurs." },
  { icon: Gauge, title: "Counting you can trust", body: "A visit is recorded atomically at redirect time, keeping the number honest when traffic arrives together." },
  { icon: Database, title: "Durable link history", body: "Created, updated, and last-visited timestamps keep the full lifecycle visible without extra bookkeeping." },
  { icon: Braces, title: "Clean boundaries", body: "Transport, business logic, and persistence stay separate—easier to test, audit, and evolve." },
] as const

const flow = [
  { label: "Request", code: `GET ${siteConfig.shortHost}/x7Kd2a`, tone: "muted" },
  { label: "Resolve", code: "validate · find · count", tone: "muted" },
  { label: "Relay", code: "302 → destination", tone: "brand" },
] as const

export function Architecture() {
  return (
    <section id="architecture" className="relative overflow-hidden bg-foreground text-background">
      <div aria-hidden="true" className="noise absolute inset-0 opacity-[0.045]" />
      <div aria-hidden="true" className="absolute -left-40 top-1/3 size-96 rounded-full bg-brand/10 blur-[110px]" />
      <div className="relative mx-auto max-w-[78rem] px-4 py-24 sm:px-6 sm:py-32 lg:px-8">
        <Reveal className="grid gap-10 lg:grid-cols-[1.04fr_0.96fr] lg:items-end">
          <div>
            <p className="text-xs font-semibold tracking-[0.16em] text-brand uppercase">Built different</p>
            <h2 className="mt-5 max-w-3xl text-4xl font-semibold leading-[0.98] tracking-[-0.055em] text-balance sm:text-5xl lg:text-6xl">
              Simple on the surface. <span className="text-background/32">Serious underneath.</span>
            </h2>
          </div>
          <p className="max-w-xl text-base leading-7 text-pretty text-background/50 lg:justify-self-end">
            A short link is tiny infrastructure with a very public job. Relay
            is engineered so every redirect stays predictable, every count
            stays accurate, and every layer can be understood.
          </p>
        </Reveal>

        <div className="mt-16 grid gap-5 lg:grid-cols-[0.92fr_1.08fr] lg:gap-8">
          <Reveal>
            <div className="relative h-full overflow-hidden rounded-[2rem] border border-white/10 bg-white/[0.035] p-5 shadow-[inset_0_1px_0_rgb(255_255_255/0.04)] sm:p-7">
              <div className="flex items-center justify-between border-b border-white/8 pb-5">
                <div className="flex items-center gap-2 text-[10px] font-medium text-background/35"><span className="size-1.5 rounded-full bg-brand" />LIVE REDIRECT TRACE</div>
                <span className="font-mono text-[9px] text-background/20">relay-edge-01</span>
              </div>
              <div className="mt-7 space-y-2">
                {flow.map((item, index) => (
                  <div key={item.label}>
                    {index > 0 && <div className="flex h-8 justify-center"><ArrowDown className="size-3 text-background/18" aria-hidden="true" /></div>}
                    <div className={`flex items-center gap-4 rounded-xl border px-4 py-4 ${item.tone === "brand" ? "border-brand/25 bg-brand/10" : "border-white/8 bg-black/10"}`}>
                      <span className={`flex size-6 items-center justify-center rounded-lg ${item.tone === "brand" ? "bg-brand text-foreground" : "bg-white/7 text-background/34"}`}>
                        {item.tone === "brand" ? <Check className="size-3" /> : <span className="font-mono text-[9px]">0{index + 1}</span>}
                      </span>
                      <div><p className="text-xs font-medium">{item.label}</p><p className={`mt-1 font-mono text-[9px] ${item.tone === "brand" ? "text-brand" : "text-background/30"}`}>{item.code}</p></div>
                      {item.tone === "brand" && <span className="ml-auto rounded-full bg-brand/10 px-2 py-1 font-mono text-[8px] text-brand">complete</span>}
                    </div>
                  </div>
                ))}
              </div>
              <p className="mt-6 text-center font-mono text-[9px] text-background/22">visit_count +1 · last_accessed_at updated</p>
            </div>
          </Reveal>

          <div className="grid gap-px overflow-hidden rounded-[2rem] border border-white/10 bg-white/10 sm:grid-cols-2">
            {assurances.map((item, index) => (
              <Reveal key={item.title} delay={80 + index * 60} className="h-full">
                <article className="group h-full bg-[#171b14] p-6 transition-colors duration-500 hover:bg-white/[0.045] sm:p-7">
                  <span className="flex size-10 items-center justify-center rounded-xl border border-white/9 bg-white/[0.04] text-brand transition-transform duration-500 group-hover:-rotate-3 group-hover:scale-105"><item.icon className="size-4.5" aria-hidden="true" /></span>
                  <h3 className="mt-6 text-base font-semibold tracking-[-0.025em]">{item.title}</h3>
                  <p className="mt-2 text-sm leading-6 text-background/42">{item.body}</p>
                </article>
              </Reveal>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
