import { Link2 } from "lucide-react"

import { siteConfig } from "@/config/site"
import { cn } from "@/lib/utils"

export function Brand({ className }: { className?: string }) {
  return (
    <span className={cn("flex items-center gap-2", className)}>
      <span className="flex size-6 items-center justify-center rounded-md bg-foreground text-background">
        <Link2 className="size-3.5" aria-hidden="true" />
      </span>
      <span className="text-[0.95rem] font-semibold">
        {siteConfig.name}
      </span>
    </span>
  )
}
