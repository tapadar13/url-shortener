import { siteConfig } from "@/config/site"
import { cn } from "@/lib/utils"

export function Brand({ className }: { className?: string }) {
  return (
    <span className={cn("flex items-center gap-2.5", className)}>
      <span className="relative flex size-7 items-center justify-center overflow-hidden rounded-[0.6rem] bg-foreground shadow-[inset_0_0_0_1px_rgb(255_255_255/0.08)]">
        <span className="absolute h-1.5 w-4 -rotate-45 rounded-full bg-brand" />
        <span className="absolute h-1.5 w-4 translate-x-1.5 translate-y-1.5 -rotate-45 rounded-full border border-brand" />
      </span>
      <span className="text-[0.98rem] font-semibold tracking-[-0.03em]">
        {siteConfig.name}
      </span>
    </span>
  )
}
