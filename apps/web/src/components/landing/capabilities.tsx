import {
  ArrowRight,
  BarChart3,
  Braces,
  Check,
  Copy,
  Link2,
  RefreshCw,
  ShieldCheck,
  Sparkles,
} from "lucide-react"

import { Reveal } from "@/components/landing/reveal"
import { siteConfig } from "@/config/site"

export function Capabilities() {
  return (
    <section id="product" className="relative border-y border-foreground/8 bg-card/55">
      <div aria-hidden="true" className="noise pointer-events-none absolute inset-0 opacity-[0.018]" />
      <div className="relative mx-auto max-w-[78rem] px-4 py-24 sm:px-6 sm:py-32 lg:px-8">
        <Reveal className="grid items-end gap-6 lg:grid-cols-[1fr_0.72fr]">
          <div>
            <p className="text-xs font-semibold tracking-[0.16em] text-foreground/42 uppercase">The entire link lifecycle</p>
            <h2 className="mt-5 max-w-3xl text-4xl font-semibold leading-[0.98] tracking-[-0.055em] text-balance sm:text-5xl lg:text-6xl">
              Everything your link needs. <span className="text-foreground/35">Nothing it doesn&apos;t.</span>
            </h2>
          </div>
          <p className="max-w-xl text-base leading-7 text-pretty text-muted-foreground lg:justify-self-end">
            Most link tools bury a simple job under campaign software. Relay
            keeps the three moments that matter—create, control, understand—in
            one exceptionally clear workspace.
          </p>
        </Reveal>

        <div className="mt-14 grid auto-rows-[minmax(15rem,auto)] gap-4 lg:grid-cols-12">
          <Reveal className="lg:col-span-7 lg:row-span-2">
            <article className="group relative flex h-full min-h-[32rem] flex-col overflow-hidden rounded-[2rem] border border-foreground/9 bg-foreground p-6 text-background shadow-[0_24px_80px_-48px_rgb(20_24_16/0.7)] sm:p-8">
              <div aria-hidden="true" className="noise absolute inset-0 opacity-[0.04]" />
              <div aria-hidden="true" className="absolute -right-20 -top-20 size-64 rounded-full bg-brand/12 blur-3xl transition-transform duration-700 group-hover:scale-125" />
              <div className="relative flex items-start justify-between gap-4">
                <div>
                  <span className="inline-flex size-10 items-center justify-center rounded-xl border border-white/10 bg-white/[0.06] text-brand"><Link2 className="size-4.5" aria-hidden="true" /></span>
                  <h3 className="mt-6 text-2xl font-semibold tracking-[-0.04em] sm:text-3xl">From unwieldy to unforgettable.</h3>
                  <p className="mt-3 max-w-lg text-sm leading-6 text-background/58 sm:text-base">
                    Paste any valid destination. Relay issues a compact,
                    collision-safe short code that is ready to share anywhere.
                  </p>
                </div>
                <span className="hidden rounded-full border border-brand/20 bg-brand/8 px-3 py-1 text-[10px] font-semibold tracking-[0.12em] text-brand uppercase sm:inline-flex">Create</span>
              </div>

              <div className="relative mt-auto pt-10">
                <div className="rounded-2xl border border-white/10 bg-[#10130e] p-3 shadow-[0_20px_60px_-28px_rgb(0_0_0/0.8)] sm:p-4">
                  <p className="mb-2.5 text-[10px] font-medium text-white/35">Destination URL</p>
                  <div className="flex gap-2">
                    <div className="flex h-11 min-w-0 flex-1 items-center gap-2 rounded-xl border border-white/10 bg-white/[0.035] px-3"><Link2 className="size-3.5 shrink-0 text-white/25" /><span className="truncate font-mono text-[10px] text-white/56 sm:text-xs">https://yourcompany.com/launch/the-complete-story</span></div>
                    <span className="flex h-11 shrink-0 items-center rounded-xl bg-brand px-4 text-xs font-semibold text-foreground">Shorten</span>
                  </div>
                  <div className="mt-3 flex flex-wrap items-center gap-2 rounded-xl border border-brand/20 bg-brand/8 px-3 py-2.5">
                    <span className="flex size-5 items-center justify-center rounded-full bg-brand text-foreground"><Check className="size-3" /></span>
                    <span className="font-mono text-xs font-medium text-brand">{siteConfig.shortHost}/story</span>
                    <Copy className="size-3.5 text-white/28" />
                    <span className="ml-auto text-[10px] text-white/28">created just now</span>
                  </div>
                </div>
              </div>
            </article>
          </Reveal>

          <Reveal delay={80} className="lg:col-span-5">
            <article className="group relative flex h-full min-h-[15rem] flex-col overflow-hidden rounded-[2rem] border border-foreground/9 bg-[linear-gradient(145deg,var(--card),var(--brand-muted))] p-6 sm:p-7">
              <div className="flex items-center justify-between"><span className="inline-flex size-10 items-center justify-center rounded-xl bg-foreground text-brand"><RefreshCw className="size-4.5 transition-transform duration-700 group-hover:rotate-180" aria-hidden="true" /></span><span className="text-[10px] font-semibold tracking-[0.12em] text-foreground/38 uppercase">Reroute</span></div>
              <h3 className="mt-6 text-xl font-semibold tracking-[-0.035em]">Change the destination. Keep the link.</h3>
              <p className="mt-2 text-sm leading-6 text-muted-foreground">Update a campaign, fix a mistake, or move a resource. Every existing share keeps working and its history stays intact.</p>
              <div className="mt-5 flex items-center gap-2 font-mono text-[10px]"><span className="truncate text-muted-foreground line-through">/spring</span><ArrowRight className="size-3 text-foreground/30" /><span className="truncate font-medium text-foreground">/summer</span></div>
            </article>
          </Reveal>

          <Reveal delay={120} className="lg:col-span-5">
            <article className="group relative flex h-full min-h-[15rem] flex-col overflow-hidden rounded-[2rem] border border-foreground/9 bg-card p-6 sm:p-7">
              <div className="flex items-center justify-between"><span className="inline-flex size-10 items-center justify-center rounded-xl bg-brand-muted text-foreground"><BarChart3 className="size-4.5" aria-hidden="true" /></span><span className="text-[10px] font-semibold tracking-[0.12em] text-foreground/38 uppercase">Measure</span></div>
              <div className="mt-5 flex items-end justify-between gap-4"><div><p className="text-[10px] text-muted-foreground">Total visits</p><p className="mt-1 text-3xl font-semibold tracking-[-0.06em]">1,284</p></div><span className="mb-1 rounded-full bg-brand-muted px-2 py-1 text-[9px] font-semibold text-foreground">+18.4%</span></div>
              <div className="mt-4 flex h-7 items-end gap-1" aria-hidden="true">{[32, 48, 41, 62, 56, 79, 70, 100].map((height, index) => <span key={index} className="flex-1 rounded-sm bg-foreground/12 transition-colors group-hover:bg-brand" style={{ height: `${height}%`, transitionDelay: `${index * 40}ms` }} />)}</div>
            </article>
          </Reveal>

          <Reveal delay={80} className="lg:col-span-4">
            <article className="flex h-full min-h-[15rem] flex-col rounded-[2rem] border border-foreground/9 bg-card p-6 sm:p-7">
              <span className="inline-flex size-10 items-center justify-center rounded-xl bg-muted text-foreground"><Braces className="size-4.5" aria-hidden="true" /></span>
              <h3 className="mt-6 text-xl font-semibold tracking-[-0.035em]">API-first, not API-later.</h3>
              <p className="mt-2 text-sm leading-6 text-muted-foreground">The same focused JSON contract behind the workspace is ready for your scripts and systems.</p>
              <code className="mt-auto pt-5 text-[10px] text-foreground/45"><span className="text-brand-foreground">POST</span> /shorten</code>
            </article>
          </Reveal>

          <Reveal delay={120} className="lg:col-span-4">
            <article className="flex h-full min-h-[15rem] flex-col rounded-[2rem] border border-foreground/9 bg-card p-6 sm:p-7">
              <span className="inline-flex size-10 items-center justify-center rounded-xl bg-muted text-foreground"><ShieldCheck className="size-4.5" aria-hidden="true" /></span>
              <h3 className="mt-6 text-xl font-semibold tracking-[-0.035em]">Built for the boring guarantees.</h3>
              <p className="mt-2 text-sm leading-6 text-muted-foreground">Unique codes, validated destinations, atomic visit counts, and graceful requests. Quietly dependable by design.</p>
            </article>
          </Reveal>

          <Reveal delay={160} className="lg:col-span-4">
            <article className="relative flex h-full min-h-[15rem] flex-col overflow-hidden rounded-[2rem] border border-brand/30 bg-brand p-6 text-foreground sm:p-7">
              <Sparkles className="size-5" aria-hidden="true" />
              <blockquote className="mt-6 text-xl font-semibold leading-snug tracking-[-0.035em]">“Links should feel simple again.”</blockquote>
              <p className="mt-3 text-sm leading-6 text-foreground/65">No vanity layers. No maze of menus. Just a sharper way to create, control, and understand every share.</p>
            </article>
          </Reveal>
        </div>
      </div>
    </section>
  )
}
