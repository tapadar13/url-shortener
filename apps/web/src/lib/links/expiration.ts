const HOUR = 60 * 60 * 1000
const DAY = 24 * HOUR

export const expirationOptions = [
  { value: "never", label: "Never" },
  { value: "day", label: "In 24 hours" },
  { value: "week", label: "In 7 days" },
  { value: "month", label: "In 30 days" },
  { value: "custom", label: "Custom date" },
] as const

export type ExpirationPreset = (typeof expirationOptions)[number]["value"]

interface ExpirationResolution {
  expiresAt?: string
  error?: string
}

export interface ExpirationStatus {
  state: "scheduled" | "expiring" | "expired"
  label: string
  title: string
}

const presetDuration: Partial<Record<ExpirationPreset, number>> = {
  day: DAY,
  week: 7 * DAY,
  month: 30 * DAY,
}

const expirationDateFormatter = new Intl.DateTimeFormat("en-US", {
  month: "short",
  day: "numeric",
  year: "numeric",
  hour: "numeric",
  minute: "2-digit",
})

const expirationDayFormatter = new Intl.DateTimeFormat("en-US", {
  month: "short",
  day: "numeric",
  year: "numeric",
})

export function resolveExpiration(
  preset: ExpirationPreset,
  customValue: string,
  now = new Date()
): ExpirationResolution {
  if (preset === "never") {
    return {}
  }

  const duration = presetDuration[preset]
  const expiration =
    duration === undefined
      ? new Date(customValue)
      : new Date(now.getTime() + duration)

  if (Number.isNaN(expiration.getTime()) || expiration <= now) {
    return { error: "Choose a future expiration date and time." }
  }

  return { expiresAt: expiration.toISOString() }
}

export function minimumCustomExpiration(now = new Date()): string {
  return dateTimeLocalValue(roundUpToMinute(new Date(now.getTime() + 5 * 60_000)))
}

export function defaultCustomExpiration(now = new Date()): string {
  return dateTimeLocalValue(roundUpToMinute(new Date(now.getTime() + DAY)))
}

export function expirationStatus(
  expiresAt: string,
  now = new Date()
): ExpirationStatus {
  const expiration = new Date(expiresAt)
  const remaining = expiration.getTime() - now.getTime()
  const title = expirationDateFormatter.format(expiration)

  if (remaining <= 0) {
    return { state: "expired", label: "Expired", title }
  }

  const hours = Math.ceil(remaining / HOUR)
  if (hours < 24) {
    return {
      state: "expiring",
      label: `Expires in ${hours}h`,
      title,
    }
  }

  const days = Math.ceil(remaining / DAY)
  if (days <= 7) {
    return {
      state: "expiring",
      label: `Expires in ${days}d`,
      title,
    }
  }

  return {
    state: "scheduled",
    label: `Expires ${expirationDayFormatter.format(expiration)}`,
    title,
  }
}

function roundUpToMinute(date: Date): Date {
  return new Date(Math.ceil(date.getTime() / 60_000) * 60_000)
}

function dateTimeLocalValue(date: Date): string {
  return [
    date.getFullYear(),
    "-",
    pad(date.getMonth() + 1),
    "-",
    pad(date.getDate()),
    "T",
    pad(date.getHours()),
    ":",
    pad(date.getMinutes()),
  ].join("")
}

function pad(value: number): string {
  return String(value).padStart(2, "0")
}
