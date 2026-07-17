"use client"

import { useState } from "react"
import {
  BarChart3,
  Clock,
  CornerDownRight,
  Link2,
  Loader2,
} from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Skeleton } from "@/components/ui/skeleton"
import { useLinkStats, useUpdateLink } from "@/hooks/use-links"
import { displayShortUrl, formatCount, formatDate, timeAgo } from "@/lib/format"
import { linkErrorMessage } from "@/lib/links/error-message"
import type { LinkStats } from "@/lib/links/types"

interface EditDestinationDialogProps {
  link: LinkStats
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function EditDestinationDialog({
  link,
  open,
  onOpenChange,
}: EditDestinationDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="overflow-hidden rounded-[1.5rem] border-foreground/10 bg-background/95 p-6 shadow-[0_30px_100px_-35px_rgb(20_24_16/0.7)] backdrop-blur-xl sm:max-w-lg">
        <div aria-hidden="true" className="absolute inset-x-0 top-0 h-1 bg-brand" />
        {/* Content unmounts on close, so the form resets on every open. */}
        <EditDestinationForm link={link} onOpenChange={onOpenChange} />
      </DialogContent>
    </Dialog>
  )
}

function EditDestinationForm({
  link,
  onOpenChange,
}: {
  link: LinkStats
  onOpenChange: (open: boolean) => void
}) {
  const [url, setUrl] = useState(link.url)
  const [error, setError] = useState<string | null>(null)
  const updateLink = useUpdateLink()

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault()
    if (updateLink.isPending) return
    setError(null)
    updateLink.mutate(
      { shortCode: link.shortCode, url },
      {
        onSuccess: () => {
          onOpenChange(false)
          toast.success(
            `${displayShortUrl(link.shortUrl)} now points somewhere new`,
            { description: "Everyone with the link lands on the fresh destination." }
          )
        },
        onError: (mutationError) => {
          setError(
            linkErrorMessage(
              mutationError,
              "Something went wrong. Try again."
            )
          )
        },
      }
    )
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle className="font-mono text-base">
          {displayShortUrl(link.shortUrl)}
        </DialogTitle>
        <DialogDescription>
          Point this short link somewhere new. The code stays the same, and
          its visit history carries over.
        </DialogDescription>
      </DialogHeader>
      <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="edit-destination" className="text-xs">
              Destination
            </Label>
            <Input
              id="edit-destination"
              type="url"
              value={url}
              onChange={(event) => {
                setUrl(event.target.value)
                if (error) setError(null)
              }}
              aria-invalid={error ? true : undefined}
              className="mt-1.5 font-mono text-xs"
              autoFocus
            />
            {error && (
              <p role="alert" className="mt-1.5 text-xs text-destructive">
                {error}
              </p>
            )}
          </div>
          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline">
                Cancel
              </Button>
            </DialogClose>
            <Button
              type="submit"
              disabled={updateLink.isPending || url.trim() === link.url}
            >
              {updateLink.isPending && (
                <Loader2
                  className="animate-spin"
                  data-icon="inline-start"
                  aria-hidden="true"
                />
              )}
              Save destination
            </Button>
          </DialogFooter>
      </form>
    </>
  )
}

interface StatsDialogProps {
  link: LinkStats | null
  onOpenChange: (open: boolean) => void
}

export function StatsDialog({ link, onOpenChange }: StatsDialogProps) {
  const stats = useLinkStats(link?.shortCode ?? null)

  return (
    <Dialog open={link !== null} onOpenChange={onOpenChange}>
      <DialogContent className="overflow-hidden rounded-[1.5rem] border-foreground/10 bg-background/95 p-6 shadow-[0_30px_100px_-35px_rgb(20_24_16/0.7)] backdrop-blur-xl sm:max-w-md">
        <div aria-hidden="true" className="absolute inset-x-0 top-0 h-1 bg-brand" />
        <DialogHeader>
          <DialogTitle className="font-mono text-base">
            {link ? displayShortUrl(link.shortUrl) : "Link statistics"}
          </DialogTitle>
          <DialogDescription>
            Live numbers straight from the counter — no sampling, no delay.
          </DialogDescription>
        </DialogHeader>

        {stats.isError ? (
          <div className="rounded-xl border border-foreground/8 bg-card/70 px-5 py-8 text-center">
            <p className="text-sm text-muted-foreground">
              Couldn&apos;t load this link&apos;s statistics.
            </p>
            <Button
              variant="outline"
              size="sm"
              onClick={() => void stats.refetch()}
              className="mt-3"
            >
              Try again
            </Button>
          </div>
        ) : stats.isPending || !stats.data ? (
          <div className="grid grid-cols-2 gap-2.5">
            <Skeleton className="h-20 rounded-xl" />
            <Skeleton className="h-20 rounded-xl" />
            <Skeleton className="col-span-2 h-24 rounded-xl" />
          </div>
        ) : (
          <div className="grid grid-cols-2 gap-2.5">
            <div className="rounded-xl border border-foreground/8 bg-card/70 p-3.5">
              <p className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
                <BarChart3 className="size-3" aria-hidden="true" />
                Total visits
              </p>
              <p className="mt-1.5 font-mono text-2xl font-semibold tabular-nums">
                {formatCount(stats.data.accessCount)}
              </p>
            </div>
            <div className="rounded-xl border border-foreground/8 bg-card/70 p-3.5">
              <p className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
                <Clock className="size-3" aria-hidden="true" />
                Last visit
              </p>
              <p className="mt-1.5 font-mono text-2xl font-semibold">
                {timeAgo(stats.data.lastAccessedAt)}
              </p>
            </div>
            <div className="col-span-2 space-y-2 rounded-xl border border-foreground/8 bg-card/70 p-3.5">
              <p className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
                <Link2 className="size-3" aria-hidden="true" />
                Destination
              </p>
              <p className="truncate font-mono text-xs">{stats.data.url}</p>
              <div className="flex flex-wrap gap-x-4 gap-y-1 border-t pt-2 text-[11px] text-muted-foreground/80">
                <span>Created {formatDate(stats.data.createdAt)}</span>
                <span className="flex items-center gap-1">
                  <CornerDownRight className="size-3" aria-hidden="true" />
                  Updated {formatDate(stats.data.updatedAt)}
                </span>
              </div>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
