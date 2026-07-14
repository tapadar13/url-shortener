import { ArrowRight, ArrowUpRight } from "lucide-react"

import { AnchorLink } from "@/components/anchor-link"
import { Brand } from "@/components/layout/brand"
import { siteConfig } from "@/config/site"

export function SiteFooter() {
  return (
    <footer className="relative overflow-hidden border-t border-white/8 bg-foreground text-background">
      <div aria-hidden="true" className="noise pointer-events-none absolute inset-0 opacity-[0.035]" />
      <div aria-hidden="true" className="absolute -top-32 right-0 size-80 rounded-full bg-brand/10 blur-3xl" />
      <div className="relative mx-auto max-w-[78rem] px-4 py-14 sm:px-6 sm:py-16 lg:px-8">
        <div className="flex flex-col justify-between gap-12 sm:flex-row">
          <div className="max-w-xs">
            <Brand className="text-background" />
            <p className="mt-4 text-sm leading-6 text-pretty text-background/55">
              The focused link workspace for people who want clean URLs,
              flexible destinations, and numbers they can actually understand.
            </p>
            <a
              href={siteConfig.repoUrl}
              target="_blank"
              rel="noreferrer"
              className="mt-6 inline-flex items-center gap-2 text-sm font-medium text-brand transition-colors hover:text-brand/75"
            >
              Built in the open <ArrowRight className="size-3.5" aria-hidden="true" />
            </a>
          </div>

          <div className="flex gap-16">
            <nav aria-label="Product">
              <p className="text-xs font-semibold tracking-[0.12em] text-background/35 uppercase">Explore</p>
              <ul className="mt-3 space-y-2">
                {siteConfig.nav.map((item) => (
                  <li key={item.href}>
                    <AnchorLink
                      href={item.href}
                      className="rounded-sm text-sm text-background/60 transition-colors outline-none hover:text-brand focus-visible:ring-3 focus-visible:ring-ring/50"
                    >
                      {item.label}
                    </AnchorLink>
                  </li>
                ))}
              </ul>
            </nav>

            <nav aria-label="Project">
              <p className="text-xs font-semibold tracking-[0.12em] text-background/35 uppercase">Project</p>
              <ul className="mt-3 space-y-2">
                <li>
                  <a
                    href={siteConfig.repoUrl}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex items-center gap-1 rounded-sm text-sm text-background/60 transition-colors outline-none hover:text-brand focus-visible:ring-3 focus-visible:ring-ring/50"
                  >
                    GitHub repository
                    <ArrowUpRight className="size-3" aria-hidden="true" />
                  </a>
                </li>
              </ul>
            </nav>
          </div>
        </div>

        <div className="mt-14 flex flex-col gap-2 border-t border-white/10 pt-6 text-xs text-background/35 sm:flex-row sm:items-center sm:justify-between">
          <p>© {new Date().getFullYear()} {siteConfig.name}. Open source, by design.</p>
          <p>Made for links worth following.</p>
        </div>
      </div>
    </footer>
  )
}
