import { ArrowUpRight } from "lucide-react"
import Link from "next/link"

import { Reveal } from "@/components/landing/reveal"
import { Button } from "@/components/ui/button"
import { siteConfig } from "@/config/site"

export function FinalCta() {
  return (
    <section aria-label="Get started" className="bg-foreground">
      <Reveal className="mx-auto max-w-6xl px-4 py-20 text-center sm:px-6 sm:py-24">
        <h2 className="mx-auto max-w-2xl text-3xl font-semibold text-balance text-background sm:text-4xl">
          Give your links a proper workspace
        </h2>
        <p className="mx-auto mt-4 max-w-xl text-pretty text-background/70">
          Create a workspace, shorten your first URL, and keep every destination
          and click signal within reach.
        </p>
        <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
          <Button className="h-10 bg-background px-5 text-foreground hover:bg-background/85" asChild>
            <Link href="/register">
              Get started
            </Link>
          </Button>
          <Button
            variant="outline"
            className="h-10 border-background/25 bg-transparent px-5 text-background hover:bg-background/10 hover:text-background"
            asChild
          >
            <a href={siteConfig.repoUrl} target="_blank" rel="noreferrer">
              View on GitHub
              <ArrowUpRight data-icon="inline-end" aria-hidden="true" />
            </a>
          </Button>
        </div>
      </Reveal>
    </section>
  )
}
