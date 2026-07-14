import { ArrowUpRight } from "lucide-react"

import { Reveal } from "@/components/landing/reveal"

const signals = [
  { value: "06", label: "characters per code" },
  { value: "01", label: "hop to destination" },
  { value: "24/7", label: "links stay ready" },
] as const

export function EngineeringStrip() {
  return (
    <section aria-label="Relay at a glance" className="relative overflow-hidden bg-foreground text-background">
      <div aria-hidden="true" className="noise absolute inset-0 opacity-[0.04]" />
      <div aria-hidden="true" className="absolute -right-16 -top-28 size-64 rounded-full bg-brand/12 blur-3xl" />
      <div className="relative mx-auto grid max-w-[78rem] items-center gap-10 px-4 py-12 sm:px-6 md:grid-cols-[1fr_auto] lg:px-8">
        <Reveal>
          <p className="flex items-center gap-2 text-sm font-medium text-background/52">
            <span className="size-1.5 rounded-full bg-brand shadow-[0_0_0_4px_rgb(171_237_94/0.09)]" />
            One focused workspace. Zero enterprise theatre.
          </p>
          <p className="mt-2 max-w-2xl text-xl font-medium leading-snug tracking-[-0.035em] text-balance sm:text-2xl">
            Relay is for builders, creators, and small teams who care more about
            a link working beautifully than a dashboard looking busy.
          </p>
        </Reveal>

        <Reveal delay={100} className="grid grid-cols-3 gap-5 sm:gap-9">
          {signals.map((signal) => (
            <div key={signal.label} className="min-w-0">
              <p className="flex items-center gap-1 font-mono text-xl font-medium tracking-[-0.05em] text-brand sm:text-2xl">{signal.value}<ArrowUpRight className="size-3 opacity-45" aria-hidden="true" /></p>
              <p className="mt-1 max-w-24 text-[9px] leading-4 text-background/35 sm:text-[10px]">{signal.label}</p>
            </div>
          ))}
        </Reveal>
      </div>
    </section>
  )
}
