"use client"

import { useCallback, useEffect, useRef } from "react"
import Link from "next/link"
import {
  ArrowLeft,
  BarChart3,
  Link2,
  Plus,
  Settings2,
} from "lucide-react"

import { GithubMark } from "@/components/icons"
import { LinksList } from "@/components/links/links-list"
import { ShortenPanel } from "@/components/links/shorten-panel"
import { Button } from "@/components/ui/button"
import { useLinks } from "@/hooks/use-links"
import { siteConfig } from "@/config/site"
import { formatCount } from "@/lib/format"

function SidebarNavItem({
  icon: Icon,
  label,
  active,
  hint,
}: {
  icon: typeof Link2
  label: string
  active?: boolean
  hint?: string
}) {
  return (
    <div
      aria-current={active ? "page" : undefined}
      className={
        active
          ? "flex items-center gap-2.5 rounded-lg bg-foreground/8 px-3 py-2 text-sm font-medium"
          : "flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-muted-foreground/70"
      }
    >
      <Icon
        className={active ? "size-4 text-brand" : "size-4"}
        aria-hidden="true"
      />
      {label}
      {hint && (
        <span className="ml-auto rounded-full border border-border px-1.5 py-0.5 text-[9px] font-semibold tracking-[0.08em] text-muted-foreground/60 uppercase">
          {hint}
        </span>
      )}
    </div>
  )
}

export function Workspace() {
  const shortenInputRef = useRef<HTMLInputElement>(null)
  const linksQuery = useLinks()
  const total = linksQuery.data?.pages[0]?.total

  const focusShortenInput = useCallback(() => {
    const input = shortenInputRef.current
    if (!input) return
    input.scrollIntoView({ block: "center" })
    input.focus()
  }, [])

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault()
        focusShortenInput()
      }
    }
    window.addEventListener("keydown", onKeyDown)
    return () => window.removeEventListener("keydown", onKeyDown)
  }, [focusShortenInput])

  return (
    <div className="flex min-h-dvh w-full">
      <aside className="sticky top-0 hidden h-dvh w-60 shrink-0 flex-col border-r bg-sidebar p-4 lg:flex">
        <Link
          href="/"
          className="flex items-center gap-2.5 rounded-lg px-2 py-1.5 outline-none focus-visible:ring-3 focus-visible:ring-ring"
        >
          <span
            className="relative flex size-7 items-center justify-center overflow-hidden rounded-lg bg-brand"
            aria-hidden="true"
          >
            <span className="h-1 w-4 -rotate-45 rounded-full bg-primary-foreground" />
          </span>
          <span className="text-sm font-semibold">{siteConfig.name}</span>
        </Link>

        <nav aria-label="Workspace" className="mt-8 space-y-1">
          <SidebarNavItem icon={Link2} label="Links" active />
          <SidebarNavItem icon={BarChart3} label="Insights" hint="soon" />
          <SidebarNavItem icon={Settings2} label="Settings" hint="soon" />
        </nav>

        <div className="mt-auto space-y-3">
          <div className="rounded-xl border bg-card/60 p-3.5">
            <p className="text-[10px] font-semibold tracking-[0.12em] text-brand uppercase">
              Workspace
            </p>
            <p className="mt-1.5 text-sm font-medium">Launch team</p>
            <p className="mt-0.5 text-xs text-muted-foreground/70">
              {total === undefined ? "…" : `${formatCount(total)} active links`}
            </p>
          </div>
          <Link
            href="/"
            className="flex items-center gap-2 rounded-lg px-2 py-1.5 text-xs text-muted-foreground/70 transition-colors outline-none hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring"
          >
            <ArrowLeft className="size-3.5" aria-hidden="true" />
            Back to the landing page
          </Link>
        </div>
      </aside>

      <div className="min-w-0 flex-1">
        <header className="sticky top-0 z-40 border-b bg-background/85 backdrop-blur-md">
          <div className="mx-auto flex h-14 max-w-4xl items-center justify-between px-4 sm:px-6">
            <Link
              href="/"
              className="flex items-center gap-2 rounded-lg outline-none focus-visible:ring-3 focus-visible:ring-ring lg:hidden"
            >
              <span
                className="relative flex size-6 items-center justify-center overflow-hidden rounded-md bg-brand"
                aria-hidden="true"
              >
                <span className="h-0.5 w-3 -rotate-45 rounded-full bg-primary-foreground" />
              </span>
              <span className="text-sm font-semibold">{siteConfig.name}</span>
            </Link>
            <p className="hidden text-sm text-muted-foreground/80 lg:block">
              <span className="font-medium text-foreground">Links</span>
              <span className="mx-2 text-muted-foreground/40">/</span>
              app.{siteConfig.shortHost}
            </p>
            <div className="flex items-center gap-3">
              <span className="flex items-center gap-2 rounded-full border bg-card/60 px-3 py-1 text-[11px] text-muted-foreground">
                <span
                  className="size-1.5 rounded-full bg-brand shadow-[0_0_0_3px_var(--brand-muted)]"
                  aria-hidden="true"
                />
                Preview workspace
              </span>
              <a
                href={siteConfig.repoUrl}
                target="_blank"
                rel="noreferrer"
                aria-label="View the repository on GitHub"
                className="rounded-lg p-1.5 text-muted-foreground transition-colors outline-none hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring"
              >
                <GithubMark className="size-4" />
              </a>
            </div>
          </div>
        </header>

        <main className="mx-auto max-w-4xl px-4 pt-8 pb-24 sm:px-6 sm:pt-10">
          <div className="animate-fade-up flex flex-wrap items-end justify-between gap-4">
            <div>
              <p className="text-sm text-muted-foreground/80">
                Good to see you, Maya
              </p>
              <h1 className="mt-1 text-2xl font-semibold tracking-[-0.02em] sm:text-3xl">
                Your links
              </h1>
            </div>
            <Button onClick={focusShortenInput} className="gap-1.5">
              <Plus data-icon="inline-start" aria-hidden="true" />
              New link
            </Button>
          </div>

          <div className="animate-fade-up mt-7 [animation-delay:90ms]">
            <ShortenPanel inputRef={shortenInputRef} />
          </div>

          <div className="animate-fade-up mt-10 [animation-delay:180ms]">
            <LinksList query={linksQuery} />
          </div>
        </main>
      </div>
    </div>
  )
}
