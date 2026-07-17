"use client"

import { useState } from "react"
import {
  ArrowDown,
  Check,
  Copy,
  ExternalLink,
  Link2,
  Loader2,
  MoreHorizontal,
  Pencil,
  Trash2,
  TrendingUp,
} from "lucide-react"
import { toast } from "sonner"

import {
  EditDestinationDialog,
  StatsDialog,
} from "@/components/links/link-dialogs"
import { LinkExpirationStatus } from "@/components/links/link-expiration-status"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Skeleton } from "@/components/ui/skeleton"
import { useDeleteLink, type useLinks } from "@/hooks/use-links"
import { displayShortUrl, displayUrl, formatCount, timeAgo } from "@/lib/format"
import type { LinkStats } from "@/lib/links/types"
import { cn } from "@/lib/utils"

function CopyButton({ shortUrl }: { shortUrl: string }) {
  const [copied, setCopied] = useState(false)
  const displayValue = displayShortUrl(shortUrl)

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(shortUrl)
      setCopied(true)
      toast.success("Copied to clipboard")
      setTimeout(() => setCopied(false), 1800)
    } catch {
      toast.error("Couldn't access the clipboard")
    }
  }

  return (
    <Button
      variant="ghost"
      size="icon-sm"
      onClick={copy}
      aria-label={`Copy ${displayValue}`}
      className={cn(
        "text-muted-foreground/60 hover:text-foreground",
        copied && "text-brand hover:text-brand"
      )}
    >
      {copied ? <Check aria-hidden="true" /> : <Copy aria-hidden="true" />}
    </Button>
  )
}

function LinkRow({ link, index }: { link: LinkStats; index: number }) {
  const [editOpen, setEditOpen] = useState(false)
  const [statsOpen, setStatsOpen] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const deleteLink = useDeleteLink()

  const handleDelete = () => {
    setConfirmDelete(false)
    deleteLink.mutate(link.shortCode, {
      onSuccess: () =>
        toast.success(`Deleted ${displayShortUrl(link.shortUrl)}`),
      onError: () => toast.error("Couldn't delete that link. It's back."),
    })
  }

  return (
    <li
      className="animate-fade-up group flex items-center gap-3 px-4 py-3.5 transition-colors [animation-duration:0.5s] hover:bg-brand-muted/25 sm:px-5"
      style={{ animationDelay: `${Math.min(index, 8) * 45}ms` }}
    >
      <span
        aria-hidden="true"
        className="hidden size-8 shrink-0 items-center justify-center rounded-lg border border-foreground/8 bg-background text-muted-foreground/70 shadow-[0_8px_20px_-16px_rgb(20_24_16/0.5)] sm:flex"
      >
        <Link2 className="size-3.5" />
      </span>

      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-1.5">
          <button
            type="button"
            onClick={() => setStatsOpen(true)}
            className="truncate rounded-md font-mono text-[13px] font-medium underline-offset-4 outline-none hover:underline focus-visible:ring-3 focus-visible:ring-ring"
          >
            {displayShortUrl(link.shortUrl)}
          </button>
          <CopyButton shortUrl={link.shortUrl} />
        </div>
        <p className="mt-0.5 flex items-baseline gap-1.5 font-mono text-[11px] text-muted-foreground/70">
          <span className="truncate">{displayUrl(link.url)}</span>
          <span className="shrink-0 md:hidden">
            {`· ${formatCount(link.accessCount)} visits`}
          </span>
        </p>
        {link.expiresAt && (
          <LinkExpirationStatus
            expiresAt={link.expiresAt}
            className="mt-1 font-mono"
          />
        )}
      </div>

      <div className="hidden shrink-0 text-right md:block">
        <p className="font-mono text-sm font-semibold tabular-nums">
          {formatCount(link.accessCount)}
        </p>
        <p className="text-[10px] text-muted-foreground/60">visits</p>
      </div>

      <div className="hidden w-20 shrink-0 text-right lg:block">
        <p className="text-xs text-muted-foreground">
          {timeAgo(link.lastAccessedAt)}
        </p>
        <p className="text-[10px] text-muted-foreground/60">last visit</p>
      </div>

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={`Actions for ${displayShortUrl(link.shortUrl)}`}
            className="shrink-0 text-muted-foreground/60 hover:text-foreground"
          >
            <MoreHorizontal aria-hidden="true" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-52">
          <DropdownMenuItem onSelect={() => setStatsOpen(true)}>
            <TrendingUp aria-hidden="true" />
            View stats
          </DropdownMenuItem>
          <DropdownMenuItem onSelect={() => setEditOpen(true)}>
            <Pencil aria-hidden="true" />
            Edit destination
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <a href={link.url} target="_blank" rel="noreferrer">
              <ExternalLink aria-hidden="true" />
              Open destination
            </a>
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            variant="destructive"
            onSelect={() => setConfirmDelete(true)}
          >
            <Trash2 aria-hidden="true" />
            Delete link
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <EditDestinationDialog
        link={link}
        open={editOpen}
        onOpenChange={setEditOpen}
      />
      <StatsDialog
        link={statsOpen ? link : null}
        onOpenChange={setStatsOpen}
      />
      <AlertDialog open={confirmDelete} onOpenChange={setConfirmDelete}>
        <AlertDialogContent className="overflow-hidden rounded-[1.5rem] border-foreground/10 bg-background/95 shadow-[0_30px_100px_-35px_rgb(20_24_16/0.7)] backdrop-blur-xl">
          <AlertDialogHeader>
            <AlertDialogTitle>
              Delete {displayShortUrl(link.shortUrl)}?
            </AlertDialogTitle>
            <AlertDialogDescription>
              {`Anyone opening this short link afterwards will hit a dead end, and its ${formatCount(link.accessCount)} recorded visits go with it. This can't be undone.`}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Keep it</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={handleDelete}>
              Delete link
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </li>
  )
}

function RowSkeleton() {
  return (
    <li className="flex items-center gap-3 px-4 py-3.5 sm:px-5">
      <Skeleton className="hidden size-8 rounded-lg sm:block" />
      <div className="flex-1 space-y-2">
        <Skeleton className="h-3.5 w-36" />
        <Skeleton className="h-2.5 w-64 max-w-full" />
      </div>
      <Skeleton className="hidden h-6 w-12 md:block" />
      <Skeleton className="hidden h-6 w-14 lg:block" />
      <Skeleton className="size-7 rounded-lg" />
    </li>
  )
}

export function LinksList({
  query,
}: {
  query: ReturnType<typeof useLinks>
}) {
  const pages = query.data?.pages ?? []
  const links = pages.flatMap((page) => page.items)

  return (
    <section aria-label="All links">
      <div className="flex items-baseline justify-between px-1">
        <h2 className="text-sm font-semibold">
          All links
          {query.isSuccess && (
            <span className="ml-2 font-mono text-xs font-normal text-muted-foreground/70">
              {formatCount(links.length)}{query.hasNextPage ? "+" : ""}
            </span>
          )}
        </h2>
        <p className="text-[11px] text-muted-foreground/60">
          Sorted by newest first
        </p>
      </div>

      <div className="mt-3 overflow-hidden rounded-[1.75rem] border border-foreground/9 bg-card/75 shadow-[0_20px_60px_-48px_rgb(20_24_16/0.48)] backdrop-blur-sm">
        {query.isPending ? (
          <ul aria-label="Loading links">
            {Array.from({ length: 6 }, (_, index) => (
              <RowSkeleton key={index} />
            ))}
          </ul>
        ) : query.isError ? (
          <div className="px-5 py-14 text-center">
            <p className="text-sm text-muted-foreground">
              Couldn&apos;t load your links.
            </p>
            <Button
              variant="outline"
              size="sm"
              onClick={() => query.refetch()}
              className="mt-3"
            >
              Try again
            </Button>
          </div>
        ) : links.length === 0 ? (
          <div className="px-5 py-16 text-center">
            <span
              aria-hidden="true"
              className="mx-auto flex size-11 items-center justify-center rounded-xl border border-foreground/8 bg-background shadow-[0_10px_30px_-22px_rgb(20_24_16/0.5)]"
            >
              <Link2 className="size-5 text-muted-foreground/60" />
            </span>
            <p className="mt-4 text-sm font-medium">No links yet</p>
            <p className="mx-auto mt-1 max-w-60 text-xs text-muted-foreground/70">
              Paste a long URL above and it will show up here, ready to share.
            </p>
          </div>
        ) : (
          <ul className="divide-y divide-foreground/6">
            {links.map((link, index) => (
              <LinkRow key={link.id} link={link} index={index} />
            ))}
          </ul>
        )}
      </div>

      {query.hasNextPage && (
        <div className="mt-4 flex justify-center">
          <Button
            variant="outline"
            onClick={() => query.fetchNextPage()}
            disabled={query.isFetchingNextPage}
            className="gap-1.5 rounded-full border-foreground/10 bg-card/75 px-4 text-xs shadow-[0_10px_30px_-24px_rgb(20_24_16/0.5)] backdrop-blur-sm"
          >
            {query.isFetchingNextPage ? (
              <Loader2
                className="animate-spin"
                data-icon="inline-start"
                aria-hidden="true"
              />
            ) : (
              <ArrowDown data-icon="inline-start" aria-hidden="true" />
            )}
            Show more
          </Button>
        </div>
      )}
    </section>
  )
}
