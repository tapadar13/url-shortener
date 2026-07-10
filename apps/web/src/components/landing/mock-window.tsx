import { cn } from "@/lib/utils"

interface MockWindowProps {
  url: string
  children: React.ReactNode
  className?: string
}

/**
 * Browser-style frame used by the product mockups. Purely illustrative —
 * nothing inside a MockWindow is interactive.
 */
export function MockWindow({ url, children, className }: MockWindowProps) {
  return (
    <div
      className={cn(
        "overflow-hidden rounded-lg border bg-background shadow-[0_1px_2px_rgb(0_0_0/0.04),0_8px_24px_rgb(0_0_0/0.06)]",
        className
      )}
    >
      <div className="flex items-center gap-3 border-b bg-muted/50 px-3 py-2">
        <div className="flex gap-1.5" aria-hidden="true">
          <span className="size-2.5 rounded-full bg-border" />
          <span className="size-2.5 rounded-full bg-border" />
          <span className="size-2.5 rounded-full bg-border" />
        </div>
        <div className="flex h-6 min-w-0 flex-1 items-center justify-center rounded-md border bg-background px-3 sm:mx-8">
          <span className="truncate font-mono text-[11px] text-muted-foreground">
            {url}
          </span>
        </div>
        <div className="hidden w-8 sm:block" aria-hidden="true" />
      </div>
      {children}
    </div>
  )
}
