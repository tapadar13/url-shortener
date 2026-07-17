"use client"

import { useEffect, useState } from "react"
import Link from "next/link"
import { Menu } from "lucide-react"

import { AnchorLink } from "@/components/anchor-link"
import { Brand } from "@/components/layout/brand"
import { Button } from "@/components/ui/button"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet"
import { siteConfig } from "@/config/site"
import { useAuthSession } from "@/hooks/use-auth"
import { cn } from "@/lib/utils"

export function SiteHeader() {
  const [scrolled, setScrolled] = useState(false)
  const [menuOpen, setMenuOpen] = useState(false)
  const session = useAuthSession()

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8)
    onScroll()
    window.addEventListener("scroll", onScroll, { passive: true })
    return () => window.removeEventListener("scroll", onScroll)
  }, [])

  // Native hash navigation is swallowed by the sheet's scroll lock, so close
  // the menu and scroll programmatically instead. The URL stays clean.
  const handleMobileNav = (
    event: React.MouseEvent<HTMLAnchorElement>,
    href: string
  ) => {
    event.preventDefault()
    setMenuOpen(false)
    // Jump instantly: the closing sheet covers the transition, and a smooth
    // scroll would race against its scroll lock.
    document.querySelector(href)?.scrollIntoView({ behavior: "instant" })
  }

  return (
    <header
      className={cn(
        "sticky top-0 z-50 transition-all duration-500",
        scrolled
          ? "bg-background/78 shadow-[0_1px_0_rgb(30_34_24/0.08)] backdrop-blur-xl"
          : "bg-transparent"
      )}
    >
      <div className="mx-auto flex h-[4.5rem] max-w-[78rem] items-center justify-between px-4 sm:px-6 lg:px-8">
        <Link
          href="/"
          className="rounded-md outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
          aria-label={`${siteConfig.name} home`}
        >
          <Brand />
        </Link>

        <nav aria-label="Main" className="hidden items-center rounded-full border border-foreground/8 bg-card/55 p-1 shadow-[0_8px_30px_-20px_rgb(20_24_16/0.5)] backdrop-blur-md md:flex">
          {siteConfig.nav.map((item) => (
            <AnchorLink
              key={item.href}
              href={item.href}
              className="rounded-full px-3.5 py-1.5 text-[0.8rem] font-medium text-muted-foreground transition-all outline-none hover:bg-background/75 hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
            >
              {item.label}
            </AnchorLink>
          ))}
        </nav>

        <AccountActions
          authenticated={Boolean(session.data)}
          pending={session.isPending}
          className="hidden md:flex"
        />

        <Sheet open={menuOpen} onOpenChange={setMenuOpen}>
          <SheetTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="md:hidden"
              aria-label="Open menu"
            >
              <Menu aria-hidden="true" />
            </Button>
          </SheetTrigger>
          <SheetContent side="right" className="w-[88vw] max-w-sm bg-background/95 backdrop-blur-xl">
            <SheetHeader>
              <SheetTitle>
                <Brand />
              </SheetTitle>
              <SheetDescription className="sr-only">
                Site navigation
              </SheetDescription>
            </SheetHeader>
            <nav
              aria-label="Mobile"
              className="flex flex-col gap-1 px-4"
            >
              {siteConfig.nav.map((item) => (
                <a
                  key={item.href}
                  href={item.href}
                  onClick={(event) => handleMobileNav(event, item.href)}
                className="rounded-xl px-3 py-3 text-base font-medium text-foreground transition-colors outline-none hover:bg-muted focus-visible:ring-3 focus-visible:ring-ring/50"
                >
                  {item.label}
                </a>
              ))}
            </nav>
            <AccountActions
              authenticated={Boolean(session.data)}
              pending={session.isPending}
              mobile
              onNavigate={() => setMenuOpen(false)}
            />
          </SheetContent>
        </Sheet>
      </div>
    </header>
  )
}

interface AccountActionsProps {
  authenticated: boolean
  pending: boolean
  mobile?: boolean
  className?: string
  onNavigate?: () => void
}

function AccountActions({
  authenticated,
  pending,
  mobile = false,
  className,
  onNavigate,
}: AccountActionsProps) {
  if (pending) {
    return (
      <div
        role="status"
        aria-label="Checking session"
        className={cn(
          "min-h-8 min-w-40 items-center justify-end gap-2",
          mobile ? "mt-auto flex p-4" : "items-center",
          className
        )}
      >
        <span className="h-8 w-16 animate-pulse rounded-lg bg-muted" />
        <span className="h-8 w-24 animate-pulse rounded-lg bg-muted" />
      </div>
    )
  }

  if (authenticated) {
    return (
      <div
        className={cn(
          "min-w-40 items-center justify-end",
          mobile ? "mt-auto flex p-4" : "items-center",
          className
        )}
      >
        <Button className={cn(mobile && "w-full")} asChild>
          <Link href="/dashboard" onClick={onNavigate}>
            Dashboard
          </Link>
        </Button>
      </div>
    )
  }

  return (
    <div
      className={cn(
        "min-w-40 items-center justify-end gap-2",
        mobile ? "mt-auto flex flex-col p-4" : "items-center",
        className
      )}
    >
      <Button
        variant={mobile ? "outline" : "ghost"}
        className={cn(mobile ? "w-full" : "px-3.5")}
        asChild
      >
        <Link href="/login" onClick={onNavigate}>
          Log in
        </Link>
      </Button>
      <Button className={cn(mobile ? "w-full" : "h-9 px-4")} asChild>
        <Link href="/register" onClick={onNavigate}>
          Get started
        </Link>
      </Button>
    </div>
  )
}
