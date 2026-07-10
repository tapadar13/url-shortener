import { ArrowUpRight } from "lucide-react"

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
          <a
            href={siteConfig.repoUrl}
            target="_blank"
            rel="noreferrer"
            className="inline-flex items-center gap-1.5 rounded-full border bg-background px-3 py-1 text-xs font-medium text-muted-foreground transition-colors outline-none hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
          >
            <span className="size-1.5 rounded-full bg-brand" aria-hidden="true" />
            Built in the open — Go, MongoDB &amp; Next.js
            <ArrowUpRight className="size-3" aria-hidden="true" />
          </a>

          <h1 className="mt-6 text-4xl font-semibold text-balance sm:text-5xl md:text-6xl">
            {siteConfig.tagline}
          </h1>

          <p className="mx-auto mt-5 max-w-xl text-base text-pretty text-muted-foreground sm:text-lg">
            Create compact links, manage where they lead, and understand every
            visit — from one focused workspace backed by an API-first Go
            service.
          </p>

          <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
            <AuthDialog intent="get-started">
              <Button className="h-10 px-5">Get started</Button>
            </AuthDialog>
            <Button variant="outline" className="h-10 px-5" asChild>
              <a href="#product">Explore the platform</a>
            </Button>
          </div>
        </div>

        <div className="mt-14 sm:mt-16">
          <ProductPreview />
        </div>
      </div>
    </section>
  )
}
