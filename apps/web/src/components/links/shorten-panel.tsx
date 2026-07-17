"use client"

import { useState, type RefObject } from "react"
import {
  CalendarClock,
  Check,
  ChevronDown,
  Copy,
  Link2,
  Loader2,
  Settings2,
  Sparkles,
  X,
} from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useCreateLink } from "@/hooks/use-links"
import { siteConfig } from "@/config/site"
import { displayShortUrl } from "@/lib/format"
import { linkErrorMessage } from "@/lib/links/error-message"
import {
  defaultCustomExpiration,
  expirationOptions,
  expirationStatus,
  minimumCustomExpiration,
  resolveExpiration,
  type ExpirationPreset,
} from "@/lib/links/expiration"
import type { ShortLink } from "@/lib/links/types"
import { cn } from "@/lib/utils"

interface ShortenPanelProps {
  inputRef: RefObject<HTMLInputElement | null>
}

export function ShortenPanel({ inputRef }: ShortenPanelProps) {
  const [url, setUrl] = useState("")
  const [customCode, setCustomCode] = useState("")
  const [optionsOpen, setOptionsOpen] = useState(false)
  const [expirationPreset, setExpirationPreset] =
    useState<ExpirationPreset>("never")
  const [customExpiration, setCustomExpiration] = useState("")
  const [formError, setFormError] = useState<string | null>(null)
  const [expirationError, setExpirationError] = useState<string | null>(null)
  const [lastCreated, setLastCreated] = useState<ShortLink | null>(null)
  const [copied, setCopied] = useState(false)
  const createLink = useCreateLink()

  const shortUrl = lastCreated ? displayShortUrl(lastCreated.shortUrl) : null
  const createdExpiration = lastCreated?.expiresAt
    ? expirationStatus(lastCreated.expiresAt)
    : null
  const selectedExpiration = expirationOptions.find(
    (option) => option.value === expirationPreset
  )

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault()
    if (createLink.isPending) return
    setFormError(null)
    setExpirationError(null)

    const expiration = resolveExpiration(
      expirationPreset,
      customExpiration
    )
    if (expiration.error) {
      setExpirationError(expiration.error)
      return
    }

    const normalizedCode = customCode.trim()

    createLink.mutate(
      {
        url: url.trim(),
        ...(normalizedCode ? { shortCode: normalizedCode } : {}),
        ...(expiration.expiresAt ? { expiresAt: expiration.expiresAt } : {}),
      },
      {
        onSuccess: (created) => {
          setLastCreated(created)
          setCopied(false)
          setUrl("")
          setCustomCode("")
          setOptionsOpen(false)
          setExpirationPreset("never")
          setCustomExpiration("")
          toast.success(`${displayShortUrl(created.shortUrl)} is live`, {
            description: "Ready to share — every visit will be counted.",
          })
        },
        onError: (mutationError) => {
          setFormError(
            linkErrorMessage(
              mutationError,
              "Something went wrong. Try again."
            )
          )
        },
      }
    )
  }

  const copyShortUrl = async () => {
    if (!lastCreated) return
    try {
      await navigator.clipboard.writeText(lastCreated.shortUrl)
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
              required
              inputMode="url"
              value={url}
              onChange={(event) => {
                setUrl(event.target.value)
                if (formError) setFormError(null)
              }}
              placeholder="https://paste-something-long-and-messy.com/here"
              aria-label="Destination URL"
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

        {formError && (
          <p role="alert" className="mt-2 text-xs text-destructive">
            {formError}
          </p>
        )}

        <div className="mt-3 flex items-center justify-between">
          <button
            type="button"
            onClick={() => setOptionsOpen((open) => !open)}
            aria-expanded={optionsOpen}
            className="flex items-center gap-1.5 rounded-md text-xs text-muted-foreground transition-colors outline-none hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring"
          >
            {optionsOpen ? (
              <X className="size-3" aria-hidden="true" />
            ) : (
              <Settings2 className="size-3" aria-hidden="true" />
            )}
            {optionsOpen ? "Hide link options" : "Link options"}
          </button>
        </div>

        {optionsOpen && (
          <div className="animate-fade-up mt-3 grid gap-4 rounded-xl border border-foreground/8 bg-background/55 p-3.5 [animation-duration:0.35s] sm:grid-cols-2">
            <div className="min-w-0 space-y-1.5">
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
                    if (formError) setFormError(null)
                  }}
                  placeholder="launch2026"
                  pattern="[A-Za-z0-9]{4,32}"
                  maxLength={32}
                  className="h-9 min-w-0 flex-1 rounded-lg border-foreground/12 bg-background font-mono text-xs shadow-none"
                />
              </div>
              <p className="text-[11px] text-muted-foreground/70">
                4–32 letters or numbers. Empty generates a random code.
              </p>
            </div>

            <div className="min-w-0 space-y-1.5">
              <Label
                htmlFor={
                  expirationPreset === "custom"
                    ? "custom-expiration"
                    : undefined
                }
                className="text-xs text-muted-foreground"
              >
                Expiration
              </Label>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    type="button"
                    variant="outline"
                    className="h-9 w-full justify-start bg-background px-2.5 text-xs font-normal"
                  >
                    <CalendarClock data-icon="inline-start" aria-hidden="true" />
                    <span className="truncate">
                      {selectedExpiration?.label ?? "Never"}
                    </span>
                    <ChevronDown className="ml-auto" aria-hidden="true" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start">
                  <DropdownMenuRadioGroup
                    value={expirationPreset}
                    onValueChange={(value) => {
                      const preset = value as ExpirationPreset
                      setExpirationPreset(preset)
                      setExpirationError(null)
                      if (preset === "custom" && !customExpiration) {
                        setCustomExpiration(defaultCustomExpiration())
                      }
                    }}
                  >
                    {expirationOptions.map((option) => (
                      <DropdownMenuRadioItem
                        key={option.value}
                        value={option.value}
                      >
                        {option.label}
                      </DropdownMenuRadioItem>
                    ))}
                  </DropdownMenuRadioGroup>
                </DropdownMenuContent>
              </DropdownMenu>

              {expirationPreset === "custom" && (
                <Input
                  id="custom-expiration"
                  type="datetime-local"
                  required
                  min={minimumCustomExpiration()}
                  value={customExpiration}
                  onChange={(event) => {
                    setCustomExpiration(event.target.value)
                    setExpirationError(null)
                  }}
                  aria-invalid={expirationError ? true : undefined}
                  aria-describedby={
                    expirationError ? "custom-expiration-error" : undefined
                  }
                  className="h-9 rounded-lg border-foreground/12 bg-background font-mono text-xs shadow-none"
                />
              )}

              {expirationError ? (
                <p
                  id="custom-expiration-error"
                  role="alert"
                  className="text-[11px] text-destructive"
                >
                  {expirationError}
                </p>
              ) : (
                <p className="text-[11px] text-muted-foreground/70">
                  Redirects stop after this time.
                </p>
              )}
            </div>
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
          <span
            className="ml-auto text-[11px] text-muted-foreground/80"
            title={createdExpiration?.title}
          >
            {createdExpiration?.label ?? "ready to share"}
          </span>
        </div>
      )}
    </section>
  )
}
