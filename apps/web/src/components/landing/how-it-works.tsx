import { ArrowDown, ArrowRight, Check, Link2, MousePointer2, Send } from "lucide-react"

import { Reveal } from "@/components/landing/reveal"
import { siteConfig } from "@/config/site"

const steps = [
  {
    number: "01",
    icon: Link2,
    title: "Paste the messy link",
    body: "Drop in any HTTP or HTTPS destination. Relay validates it before a code is ever created.",
    detail: "yourcompany.com/campaigns/summer/launch?utm…",
  },
  {
    number: "02",
    icon: Send,
    title: "Relay makes it shareable",
    body: "A unique six-character code is checked, stored, and returned as a clean link you can use everywhere.",
    detail: `${siteConfig.shortHost}/x7Kd2a`,
  },
  {
    number: "03",
    icon: MousePointer2,
    title: "Every open becomes a signal",
    body: "Visitors reach the destination in one hop. Relay updates the visit count at the same moment.",
    detail: "1,283 → 1,284 visits",
  },
] as const

export function HowItWorks() {
  return (
    <section id="how-it-works" className="relative overflow-hidden">
      <div aria-hidden="true" className="absolute left-1/2 top-24 -z-10 h-[36rem] w-[52rem] -translate-x-1/2 rounded-full bg-brand/8 blur-[100px]" />
      <div className="mx-auto max-w-[78rem] px-4 py-24 sm:px-6 sm:py-32 lg:px-8">
        <Reveal className="mx-auto max-w-3xl text-center">
          <p className="text-xs font-semibold tracking-[0.16em] text-foreground/42 uppercase">How it flows</p>
          <h2 className="mt-5 text-4xl font-semibold leading-[1] tracking-[-0.055em] text-balance sm:text-5xl lg:text-6xl">
            Long URL in. <span className="text-foreground/35">Clear signal out.</span>
          </h2>
          <p className="mx-auto mt-5 max-w-2xl text-base leading-7 text-pretty text-muted-foreground">
            One simple flow, designed end to end. No detours, no hidden state,
            and no wondering what happened after you pressed share.
          </p>
        </Reveal>

        <div className="relative mt-16 lg:mt-20">
          <div aria-hidden="true" className="absolute left-[16.67%] right-[16.67%] top-[3.25rem] hidden h-px bg-foreground/10 lg:block">
            <span className="absolute inset-y-0 left-0 w-1/3 origin-left bg-gradient-to-r from-transparent via-brand to-transparent [animation:signal-pulse_4.5s_ease-in-out_infinite]" />
            <span className="absolute inset-y-0 left-1/3 w-1/3 origin-left bg-gradient-to-r from-transparent via-brand to-transparent [animation:signal-pulse_4.5s_1.5s_ease-in-out_infinite]" />
          </div>

          <ol className="grid gap-4 lg:grid-cols-3">
            {steps.map((step, index) => (
              <Reveal key={step.number} delay={index * 100}>
                <li className="group relative h-full rounded-[2rem] border border-foreground/9 bg-card/75 p-5 shadow-[0_20px_60px_-48px_rgb(20_24_16/0.48)] backdrop-blur-sm transition-all duration-500 hover:-translate-y-1 hover:border-foreground/15 hover:shadow-[0_28px_70px_-42px_rgb(20_24_16/0.58)] sm:p-6">
                  <div className="relative z-10 flex items-center justify-between">
                    <span className="flex size-[4.5rem] items-center justify-center rounded-2xl border border-foreground/8 bg-background shadow-[0_10px_30px_-22px_rgb(20_24_16/0.5)] transition-transform duration-500 group-hover:rotate-[-3deg] group-hover:scale-105">
                      <step.icon className="size-5 text-foreground" aria-hidden="true" />
                    </span>
                    <span className="font-mono text-[10px] font-semibold tracking-[0.16em] text-foreground/30">STEP {step.number}</span>
                  </div>
                  <h3 className="mt-7 text-xl font-semibold tracking-[-0.035em]">{step.title}</h3>
                  <p className="mt-2 min-h-20 text-sm leading-6 text-muted-foreground">{step.body}</p>
                  <div className="mt-6 flex h-12 items-center gap-2 overflow-hidden rounded-xl border border-foreground/8 bg-muted/55 px-3">
                    {index === 1 && <Check className="size-3.5 shrink-0 text-brand-foreground" aria-hidden="true" />}
                    <span className="truncate font-mono text-[10px] text-foreground/58">{step.detail}</span>
                    {index < steps.length - 1 && <ArrowRight className="ml-auto size-3 shrink-0 text-foreground/25" aria-hidden="true" />}
                  </div>
                </li>
              </Reveal>
            ))}
          </ol>
        </div>

        <Reveal delay={180} className="mt-8 flex justify-center">
          <p className="flex items-center gap-2 rounded-full border border-foreground/8 bg-card/70 px-4 py-2 text-[10px] font-medium text-muted-foreground backdrop-blur-sm">
            <ArrowDown className="size-3 text-brand-foreground" aria-hidden="true" />
            Edit the destination later. The short link and its history stay exactly where they are.
          </p>
        </Reveal>
      </div>
    </section>
  )
}
