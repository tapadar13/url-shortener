import { ArrowDown, ArrowRight, ArrowUpRight } from "lucide-react"
import Link from "next/link"

import { AnchorLink } from "@/components/anchor-link"
import { GithubMark } from "@/components/icons"
import { ProductPreview } from "@/components/landing/product-preview"
import { Button } from "@/components/ui/button"
import { siteConfig } from "@/config/site"

export function Hero() {
  return (
    <section className="relative isolate overflow-hidden pb-16 pt-6 sm:pb-24 sm:pt-12 lg:min-h-[calc(100svh-4.5rem)] lg:pb-28 lg:pt-16">
      <div aria-hidden="true" className="noise pointer-events-none absolute inset-0 -z-20 opacity-[0.025]" />
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-[42rem] bg-[linear-gradient(to_right,transparent_0,transparent_calc(50%-0.5px),rgb(30_34_24/0.055)_50%,transparent_calc(50%+0.5px),transparent_100%)] bg-[size:4rem_4rem] [mask-image:linear-gradient(to_bottom,black,transparent_90%)]"
      />
      <div aria-hidden="true" className="absolute -left-32 top-24 -z-10 size-80 rounded-full bg-brand/15 blur-[90px]" />

      <div className="mx-auto grid max-w-[78rem] items-center gap-14 px-4 sm:px-6 lg:grid-cols-[0.86fr_1.14fr] lg:gap-10 lg:px-8">
        <div className="max-w-2xl lg:pb-8">
          <a
            href={siteConfig.repoUrl}
            target="_blank"
            rel="noreferrer"
            className="animate-fade-up group inline-flex items-center gap-2.5 rounded-full border border-foreground/10 bg-card/70 py-1.5 pl-2 pr-3 text-xs font-medium text-muted-foreground shadow-[0_8px_30px_-20px_rgb(20_24_16/0.55)] backdrop-blur-md transition-all hover:-translate-y-0.5 hover:border-foreground/20 hover:text-foreground focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
          >
            <span className="flex size-5 items-center justify-center rounded-full bg-foreground text-brand">
              <GithubMark className="size-3" aria-hidden="true" />
            </span>
            Open-source. Built for control.
            <ArrowUpRight className="size-3 transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5" aria-hidden="true" />
          </a>

          <h1 className="animate-fade-up mt-7 text-[clamp(3.65rem,8.2vw,6.8rem)] font-semibold leading-[0.88] tracking-[-0.075em] text-balance [animation-delay:80ms]">
            One link. <span className="relative whitespace-nowrap text-foreground/42">Total clarity.<span aria-hidden="true" className="absolute inset-x-0 bottom-[0.08em] -z-10 h-[0.16em] -rotate-1 rounded-full bg-brand/80" /></span>
          </h1>

          <p className="animate-fade-up mt-7 max-w-xl text-base leading-7 text-pretty text-muted-foreground [animation-delay:160ms] sm:text-lg sm:leading-8">
            Turn long, forgettable URLs into short links you control. Reroute
            them without breaking a share, and see every visit without digging
            through a marketing suite.
          </p>

          <div className="animate-fade-up mt-9 flex flex-wrap items-center gap-3 [animation-delay:240ms]">
            <Button className="h-12 rounded-2xl px-6 text-[0.95rem]" asChild>
              <Link href="/register">
                Create your first link
                <ArrowRight data-icon="inline-end" aria-hidden="true" />
              </Link>
            </Button>
            <Button variant="outline" className="h-12 rounded-2xl border-foreground/12 bg-card/60 px-5 text-[0.95rem] backdrop-blur-sm" asChild>
              <AnchorLink href="#how-it-works">
                See how it flows
                <ArrowDown data-icon="inline-end" aria-hidden="true" />
              </AnchorLink>
            </Button>
          </div>

          <div className="animate-fade-up mt-9 flex flex-wrap gap-x-5 gap-y-2 border-t border-foreground/8 pt-5 text-xs font-medium text-muted-foreground [animation-delay:320ms]">
            <span className="flex items-center gap-2"><span className="size-1.5 rounded-full bg-brand shadow-[0_0_0_3px_rgb(164_235_92/0.14)]" />No card required</span>
            <span className="flex items-center gap-2"><span className="size-1.5 rounded-full bg-brand shadow-[0_0_0_3px_rgb(164_235_92/0.14)]" />Six-character links</span>
            <span className="flex items-center gap-2"><span className="size-1.5 rounded-full bg-brand shadow-[0_0_0_3px_rgb(164_235_92/0.14)]" />Honest visit counts</span>
          </div>
        </div>

        <div className="animate-fade-up relative [animation-delay:260ms] lg:pl-4">
          <ProductPreview />
        </div>
      </div>
    </section>
  )
}
