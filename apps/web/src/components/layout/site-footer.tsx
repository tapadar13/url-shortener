import { ArrowUpRight } from "lucide-react"

import { AnchorLink } from "@/components/anchor-link"
import { Brand } from "@/components/layout/brand"
import { siteConfig } from "@/config/site"

export function SiteFooter() {
  return (
    <footer className="border-t">
      <div className="mx-auto max-w-6xl px-4 py-12 sm:px-6">
        <div className="flex flex-col justify-between gap-10 sm:flex-row">
          <div className="max-w-xs">
            <Brand />
            <p className="mt-3 text-sm text-pretty text-muted-foreground">
              A focused link-management platform: compact short links, managed
              destinations, and honest visit counts.
            </p>
          </div>

          <div className="flex gap-16">
            <nav aria-label="Product">
              <p className="text-sm font-medium">Product</p>
              <ul className="mt-3 space-y-2">
                {siteConfig.nav.map((item) => (
                  <li key={item.href}>
                    <AnchorLink
                      href={item.href}
                      className="rounded-sm text-sm text-muted-foreground transition-colors outline-none hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
                    >
                      {item.label}
                    </AnchorLink>
                  </li>
                ))}
              </ul>
            </nav>

            <nav aria-label="Project">
              <p className="text-sm font-medium">Project</p>
              <ul className="mt-3 space-y-2">
                <li>
                  <a
                    href={siteConfig.repoUrl}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex items-center gap-1 rounded-sm text-sm text-muted-foreground transition-colors outline-none hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
                  >
                    GitHub repository
                    <ArrowUpRight className="size-3" aria-hidden="true" />
                  </a>
                </li>
              </ul>
            </nav>
          </div>
        </div>

        <p className="mt-12 border-t pt-6 text-xs text-muted-foreground">
          © {new Date().getFullYear()} {siteConfig.name}. An open-source URL
          shortener.
        </p>
      </div>
    </footer>
  )
}
