"use client"

import { useEffect, useState } from "react"
import { CalendarClock } from "lucide-react"

import { expirationStatus } from "@/lib/links/expiration"
import { cn } from "@/lib/utils"

export function LinkExpirationStatus({
  expiresAt,
  className,
}: {
  expiresAt: string
  className?: string
}) {
  const [now, setNow] = useState(() => new Date())
  const status = expirationStatus(expiresAt, now)

  useEffect(() => {
    const interval = window.setInterval(() => setNow(new Date()), 60_000)
    return () => window.clearInterval(interval)
  }, [])

  return (
    <span
      title={status.title}
      suppressHydrationWarning
      className={cn(
        "inline-flex w-fit items-center gap-1 text-[10px]",
        status.state === "expired" && "text-destructive",
        status.state === "expiring" && "text-foreground/75",
        status.state === "scheduled" && "text-muted-foreground/65",
        className
      )}
    >
      <CalendarClock className="size-3" aria-hidden="true" />
      {status.label}
    </span>
  )
}
