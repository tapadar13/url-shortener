"use client"

import { useState, type RefObject } from "react"
import { Check, Copy, Link2, Loader2, Settings2, Sparkles, X } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useCreateLink } from "@/hooks/use-links"
import { siteConfig } from "@/config/site"
import { ApiError } from "@/lib/links/types"
import { cn } from "@/lib/utils"

interface ShortenPanelProps {
  inputRef: RefObject<HTMLInputElement | null>
}

export function ShortenPanel({ inputRef }: ShortenPanelProps) {
  const [url, setUrl] = useState("")
  const [customCode, setCustomCode] = useState("")
  const [customizeOpen, setCustomizeOpen] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [lastCreated, setLastCreated] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const createLink = useCreateLink()

  const shortUrl = lastCreated ? `${siteConfig.shortHost}/${lastCreated}` : null

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault()
    if (createLink.isPending) return
    setError(null)

    createLink.mutate(
      { url, shortCode: customizeOpen ? customCode : undefined },
      {
        onSuccess: (created) => {
          setLastCreated(created.shortCode)
          setCopied(false)
          setUrl("")
          setCustomCode("")
          setCustomizeOpen(false)
          toast.success(`${siteConfig.shortHost}/${created.shortCode} is live`, {
            description: "Ready to share — every visit will be counted.",
          })
        },
        onError: (mutationError) => {
          setError(
            mutationError instanceof ApiError
              ? mutationError.message
              : "Something went wrong. Try again."
          )
        },
      }
    )
  }

  const copyShortUrl = async () => {
    if (!shortUrl) return
    try {
      await navigator.clipboard.writeText(`https://${shortUrl}`)
      setCopied(true)
      toast.success("Copied to clipboard")
      setTimeout(() => setCopied(false), 2000)
    } catch {
      toast.error("Couldn't access the clipboard")
    }
  }

  return (
    <section
      aria-label="Shorten a link"
      className="rounded-[1.75rem] border border-foreground/9 bg-card/75 p-4 shadow-[0_20px_60px_-48px_rgb(20_24_16/0.48)] backdrop-blur-sm sm:p-5"
    >
      <div className="flex items-center justify-between">
        <p className="flex items-center gap-2 text-sm font-medium">
          <span
            className="flex size-6 items-center justify-center rounded-lg bg-brand text-foreground"
            aria-hidden="true"
          >
            <Sparkles className="size-3.5" />
          </span>
          Make a long URL useful
        </p>
        <span
          className="hidden rounded-md border border-foreground/10 bg-background px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground sm:block"
          aria-hidden="true"
        >
          ⌘K
        </span>
      </div>

      <form onSubmit={handleSubmit} className="mt-4">
        <div className="flex flex-col gap-2 sm:flex-row">
          <div className="relative min-w-0 flex-1">
            <Link2
              className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground/60"
              aria-hidden="true"
            />
            <Input
              ref={inputRef}
              type="url"
              inputMode="url"
              value={url}
              onChange={(event) => {
                setUrl(event.target.value)
                if (error) setError(null)
              }}
              placeholder="https://paste-something-long-and-messy.com/here"
              aria-label="Destination URL"
              aria-invalid={error ? true : undefined}
              className="h-11 rounded-xl border-foreground/12 bg-background pl-9 font-mono text-xs shadow-none sm:text-[13px]"
            />
          </div>
          <Button
            type="submit"
            disabled={createLink.isPending}
            className="h-11 shrink-0 rounded-xl px-6 text-sm font-semibold"
          >
            {createLink.isPending ? (
              <>
                <Loader2
                  className="animate-spin"
                  data-icon="inline-start"
                  aria-hidden="true"
                />
                Shortening…
              </>
            ) : (
              "Shorten"
            )}
          </Button>
        </div>

        {error && (
          <p role="alert" className="mt-2 text-xs text-destructive">
            {error}
          </p>
        )}

        <div className="mt-3 flex items-center justify-between">
          <button
            type="button"
            onClick={() => setCustomizeOpen((open) => !open)}
            aria-expanded={customizeOpen}
            className="flex items-center gap-1.5 rounded-md text-xs text-muted-foreground transition-colors outline-none hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring"
          >
            {customizeOpen ? (
              <X className="size-3" aria-hidden="true" />
            ) : (
              <Settings2 className="size-3" aria-hidden="true" />
            )}
            {customizeOpen ? "Use a random code" : "Customize the code"}
          </button>
        </div>

        {customizeOpen && (
          <div className="animate-fade-up mt-3 flex flex-wrap items-end gap-x-3 gap-y-2 [animation-duration:0.35s]">
            <div className="min-w-0">
              <Label
                htmlFor="custom-code"
                className="text-xs text-muted-foreground"
              >
                Custom code
              </Label>
              <div className="mt-1.5 flex items-center gap-1.5">
                <span className="shrink-0 font-mono text-xs text-muted-foreground">
                  {siteConfig.shortHost}/
                </span>
                <Input
                  id="custom-code"
                  value={customCode}
                  onChange={(event) => {
                    setCustomCode(event.target.value)
                    if (error) setError(null)
                  }}
                  placeholder="launch-week"
                  maxLength={20}
                  className="h-9 w-44 rounded-lg border-foreground/12 bg-background font-mono text-xs shadow-none"
                />
              </div>
            </div>
            <p className="pb-2 text-[11px] text-muted-foreground/70">
              4–20 letters or numbers. Leave empty for a random one.
            </p>
          </div>
        )}
      </form>

      {shortUrl && (
        <div className="animate-fade-up mt-4 flex flex-wrap items-center gap-2.5 rounded-2xl border border-brand/45 bg-brand-muted/60 px-3.5 py-2.5 [animation-duration:0.45s]">
          <span
            className="flex size-5 items-center justify-center rounded-full bg-brand text-foreground"
            aria-hidden="true"
          >
            <Check className="size-3" />
          </span>
          <span className="font-mono text-sm font-semibold">{shortUrl}</span>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={copyShortUrl}
            className={cn(
              "gap-1.5 text-xs hover:bg-brand/15",
              copied && "text-foreground"
            )}
          >
            {copied ? (
              <Check data-icon="inline-start" aria-hidden="true" />
            ) : (
              <Copy data-icon="inline-start" aria-hidden="true" />
            )}
            {copied ? "Copied" : "Copy"}
          </Button>
          <span className="ml-auto text-[11px] text-muted-foreground/80">
            ready to share
          </span>
        </div>
      )}
    </section>
  )
}
