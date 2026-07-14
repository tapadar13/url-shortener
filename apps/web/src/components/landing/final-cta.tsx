import { ArrowRight, ArrowUpRight } from "lucide-react"

import { GithubMark } from "@/components/icons"
import { AuthDialog } from "@/components/landing/auth-dialog"
import { Reveal } from "@/components/landing/reveal"
import { Button } from "@/components/ui/button"
import { siteConfig } from "@/config/site"

export function FinalCta() {
  return (
    <section aria-label="Get started" className="relative overflow-hidden bg-brand text-foreground">
      <div aria-hidden="true" className="noise absolute inset-0 opacity-[0.055] mix-blend-multiply" />
      <div aria-hidden="true" className="absolute inset-0 bg-[radial-gradient(circle_at_75%_20%,rgb(255_255_255/0.28),transparent_28rem)]" />
      <div className="relative mx-auto max-w-[78rem] px-4 py-24 sm:px-6 sm:py-32 lg:px-8">
        <Reveal>
          <div className="flex items-center gap-3 text-xs font-semibold tracking-[0.16em] uppercase"><span className="size-2 rounded-full bg-foreground" />Your next link can be clearer</div>
          <h2 className="mt-7 max-w-5xl text-[clamp(3.2rem,9vw,7.4rem)] font-semibold leading-[0.84] tracking-[-0.075em] text-balance">
            Make every share count.
          </h2>
          <div className="mt-10 grid gap-8 border-t border-foreground/18 pt-8 md:grid-cols-[1fr_auto] md:items-center">
            <p className="max-w-xl text-base leading-7 text-foreground/65 sm:text-lg">
              Create a short link that stays yours—easy to share, easy to
              change, and refreshingly easy to understand.
            </p>
            <div className="flex flex-wrap gap-3">
              <AuthDialog intent="get-started">
                <Button className="h-12 rounded-2xl px-6 text-[0.95rem]">
                  Start with Relay
                  <ArrowRight data-icon="inline-end" aria-hidden="true" />
                </Button>
              </AuthDialog>
              <Button variant="outline" className="h-12 rounded-2xl border-foreground/20 bg-transparent px-5 text-foreground hover:bg-foreground/8" asChild>
                <a href={siteConfig.repoUrl} target="_blank" rel="noreferrer">
                  <GithubMark aria-hidden="true" />GitHub<ArrowUpRight data-icon="inline-end" aria-hidden="true" />
                </a>
              </Button>
            </div>
          </div>
        </Reveal>
      </div>
    </section>
  )
}
