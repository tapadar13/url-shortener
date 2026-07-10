import { ArrowUpRight } from "lucide-react"

import { AnchorLink } from "@/components/anchor-link"
import { GithubMark } from "@/components/icons"
import { AuthDialog } from "@/components/landing/auth-dialog"
import { ProductPreview } from "@/components/landing/product-preview"
import { Button } from "@/components/ui/button"
import { siteConfig } from "@/config/site"

export function Hero() {
  return (
    <section className="relative overflow-hidden">
      {/* Faint engineering-grid backdrop, fading out before the next section. */}
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-0 -z-10 bg-[linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] bg-[size:72px_72px] opacity-40 [mask-image:linear-gradient(to_bottom,black,transparent_75%)]"
      />

      <div className="mx-auto max-w-6xl px-4 pt-16 pb-14 sm:px-6 sm:pt-24 sm:pb-20">
        <div className="mx-auto max-w-2xl text-center">
          <div className="animate-fade-up">
            <a
              href={siteConfig.repoUrl}
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center gap-2 rounded-full border bg-background py-1 pr-2.5 pl-3 text-xs font-medium text-muted-foreground transition-colors outline-none hover:border-foreground/20 hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
            >
              <span className="relative flex size-1.5" aria-hidden="true">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-brand opacity-60 [animation-duration:2.5s]" />
                <span className="relative inline-flex size-1.5 rounded-full bg-brand" />
              </span>
              Built in the open
              <span
                className="h-3.5 w-px bg-border"
                aria-hidden="true"
              />
              <GithubMark className="size-3.5" />
              GitHub
              <ArrowUpRight className="size-3" aria-hidden="true" />
            </a>
          </div>

          <h1 className="animate-fade-up mt-6 text-4xl font-semibold text-balance [animation-delay:80ms] sm:text-5xl md:text-6xl">
            {siteConfig.tagline}
          </h1>

          <p className="animate-fade-up mx-auto mt-5 max-w-xl text-base text-pretty text-muted-foreground [animation-delay:160ms] sm:text-lg">
            Paste a long URL and get a short link you can share anywhere.
            Change where it points at any time, and see exactly how often
            it&apos;s opened.
          </p>

          <div className="animate-fade-up mt-8 flex flex-wrap items-center justify-center gap-3 [animation-delay:240ms]">
            <AuthDialog intent="get-started">
              <Button className="h-10 px-5">Get started</Button>
            </AuthDialog>
            <Button variant="outline" className="h-10 px-5" asChild>
              <AnchorLink href="#product">Explore the platform</AnchorLink>
            </Button>
          </div>
        </div>

        <div className="animate-fade-up mt-14 [animation-delay:320ms] sm:mt-16">
          <ProductPreview />
        </div>
      </div>
    </section>
  )
}
