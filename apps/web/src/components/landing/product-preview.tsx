import {
  ArrowUpRight,
  BarChart3,
  Check,
  ChevronDown,
  Copy,
  Link2,
  MoreHorizontal,
  MousePointer2,
  Plus,
  Sparkles,
} from "lucide-react"

import { siteConfig } from "@/config/site"

const chart = [34, 43, 38, 58, 52, 72, 61, 83, 74, 92, 84, 100] as const

export function ProductPreview() {
  return (
    <figure className="relative mx-auto w-full max-w-[43rem] lg:mr-0">
      <div aria-hidden="true" className="absolute -inset-5 -z-10 rounded-[2.5rem] bg-[linear-gradient(145deg,rgb(171_237_94/0.26),transparent_45%,rgb(219_199_154/0.2))] blur-2xl" />
      <div className="animate-soft-float overflow-hidden rounded-[1.65rem] border border-white/10 bg-[#171b14] text-[#f4f3e9] shadow-[0_35px_100px_-28px_rgb(20_25_16/0.58),0_0_0_1px_rgb(22_25_18/0.14)]">
        <div className="flex h-12 items-center justify-between border-b border-white/8 px-4 sm:px-5">
          <div className="flex items-center gap-2.5">
            <span className="flex gap-1.5" aria-hidden="true"><span className="size-2 rounded-full bg-white/14" /><span className="size-2 rounded-full bg-white/14" /><span className="size-2 rounded-full bg-brand/70" /></span>
            <span className="hidden h-4 w-px bg-white/8 sm:block" />
            <span className="hidden text-[10px] font-medium text-white/34 sm:block">app.{siteConfig.shortHost}</span>
          </div>
          <div className="flex items-center gap-2 text-[10px] text-white/45"><span className="size-1.5 rounded-full bg-brand shadow-[0_0_0_3px_rgb(171_237_94/0.08)]" />All systems clear</div>
        </div>

        <div className="grid min-h-[32rem] sm:grid-cols-[8.5rem_1fr]">
          <aside className="hidden border-r border-white/8 p-3.5 sm:flex sm:flex-col">
            <div className="flex items-center gap-2 px-1 py-2">
              <span className="relative flex size-6 items-center justify-center overflow-hidden rounded-lg bg-brand"><span className="h-1 w-3.5 -rotate-45 rounded-full bg-[#171b14]" /></span>
              <span className="text-xs font-semibold">Relay</span>
            </div>
            <div className="mt-5 space-y-1 text-[11px]">
              <div className="flex items-center gap-2 rounded-lg bg-white/8 px-2.5 py-2 font-medium text-white/90"><Link2 className="size-3.5 text-brand" />Links</div>
              <div className="flex items-center gap-2 rounded-lg px-2.5 py-2 text-white/35"><BarChart3 className="size-3.5" />Insights</div>
            </div>
            <div className="mt-auto rounded-xl border border-white/8 bg-white/[0.035] p-3">
              <p className="text-[9px] font-semibold tracking-[0.12em] text-brand/80 uppercase">Workspace</p>
              <p className="mt-1.5 truncate text-[11px] font-medium">Launch team</p>
              <p className="mt-0.5 text-[9px] text-white/30">3 active links</p>
            </div>
          </aside>

          <div className="min-w-0 p-4 sm:p-5">
            <div className="flex items-center justify-between">
              <div><p className="text-[10px] text-white/35">Good morning, Maya</p><h2 className="mt-0.5 text-base font-semibold tracking-[-0.025em]">Your links</h2></div>
              <span className="flex h-8 items-center gap-1.5 rounded-lg bg-brand px-2.5 text-[10px] font-semibold text-[#171b14]"><Plus className="size-3" />New link</span>
            </div>

            <div className="mt-5 rounded-xl border border-brand/20 bg-[linear-gradient(115deg,rgb(171_237_94/0.12),rgb(255_255_255/0.025))] p-3.5 shadow-[inset_0_1px_0_rgb(255_255_255/0.05)]">
              <div className="flex items-center justify-between"><p className="flex items-center gap-1.5 text-[10px] font-medium text-brand"><Sparkles className="size-3" />Make a long URL useful</p><span className="text-[9px] text-white/25">⌘ K</span></div>
              <div className="mt-2.5 flex gap-2">
                <div className="flex h-9 min-w-0 flex-1 items-center gap-2 rounded-lg border border-white/10 bg-[#10130f] px-3"><Link2 className="size-3 shrink-0 text-white/25" /><span className="truncate font-mono text-[9px] text-white/54">https://relay.so/summer-campaign/launch</span></div>
                <span className="flex h-9 shrink-0 items-center rounded-lg bg-brand px-3 text-[10px] font-semibold text-[#151812]">Shorten</span>
              </div>
              <div className="mt-2.5 flex items-center gap-2 rounded-lg border border-brand/20 bg-brand/8 px-2.5 py-2">
                <span className="flex size-4 items-center justify-center rounded-full bg-brand text-[#151812]"><Check className="size-2.5" /></span>
                <span className="font-mono text-[10px] font-medium text-brand">{siteConfig.shortHost}/launch</span>
                <Copy className="size-3 text-white/30" />
                <span className="ml-auto hidden text-[9px] text-white/26 xs:block">ready to share</span>
              </div>
            </div>

            <div className="mt-4 grid gap-3 md:grid-cols-[1.12fr_0.88fr]">
              <div className="rounded-xl border border-white/8 bg-white/[0.025] p-3.5">
                <div className="flex items-center justify-between"><span className="text-[10px] font-medium text-white/45">Visits this week</span><span className="flex items-center gap-1 text-[9px] text-white/30">7 days <ChevronDown className="size-2.5" /></span></div>
                <div className="mt-3 flex items-end gap-2"><strong className="text-2xl font-semibold tracking-[-0.05em]">1,284</strong><span className="mb-1 rounded-full bg-brand/10 px-1.5 py-0.5 text-[8px] font-medium text-brand">+18.4%</span></div>
                <div className="mt-4 flex h-16 items-end gap-1" aria-label="Visits trending upward over twelve periods">
                  {chart.map((height, index) => <span key={index} className="flex-1 rounded-t-sm bg-brand/75 transition-colors hover:bg-brand" style={{ height: `${height}%`, opacity: 0.35 + index * 0.045 }} />)}
                </div>
              </div>

              <div className="rounded-xl border border-white/8 bg-white/[0.025] p-3.5">
                <p className="text-[10px] font-medium text-white/45">Latest signal</p>
                <div className="mt-3 flex items-center gap-2"><span className="flex size-7 items-center justify-center rounded-lg bg-brand/10 text-brand"><MousePointer2 className="size-3.5" /></span><div><p className="text-[10px] font-medium">Link opened</p><p className="mt-0.5 text-[8px] text-white/28">just now · Bengaluru</p></div></div>
                <div className="mt-3 space-y-2 border-l border-white/10 pl-3 text-[9px] text-white/35"><p>{siteConfig.shortHost}/launch</p><p className="flex items-center gap-1 text-brand/75">Redirected in one hop <ArrowUpRight className="size-2.5" /></p><p>Visit count updated</p></div>
              </div>
            </div>

            <div className="mt-3 flex items-center justify-between rounded-xl border border-white/8 bg-white/[0.025] px-3.5 py-3">
              <div className="min-w-0"><p className="font-mono text-[10px] font-medium text-white/78">{siteConfig.shortHost}/launch</p><p className="mt-1 truncate text-[8px] text-white/26">relay.so/summer-campaign/launch</p></div>
              <div className="flex items-center gap-3"><span className="text-right"><span className="block text-[10px] font-semibold">1,284</span><span className="text-[8px] text-white/25">visits</span></span><MoreHorizontal className="size-3.5 text-white/25" /></div>
            </div>
          </div>
        </div>
      </div>

      <figcaption className="mt-4 text-center text-[10px] font-medium tracking-[0.08em] text-muted-foreground/60 uppercase">
        Illustrative preview — not a live workspace yet.
      </figcaption>
    </figure>
  )
}
