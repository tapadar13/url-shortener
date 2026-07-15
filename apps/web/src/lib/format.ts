export function formatCount(value: number): string {
  return new Intl.NumberFormat("en-US").format(value)
}

export function formatDate(iso: string): string {
  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  }).format(new Date(iso))
}

export function timeAgo(iso: string | undefined): string {
  if (!iso) return "never"
  const seconds = Math.max(0, (Date.now() - new Date(iso).getTime()) / 1000)
  if (seconds < 45) return "just now"
  const minutes = seconds / 60
  if (minutes < 60) return `${Math.round(minutes)}m ago`
  const hours = minutes / 60
  if (hours < 24) return `${Math.round(hours)}h ago`
  const days = hours / 24
  if (days < 30) return `${Math.round(days)}d ago`
  const months = days / 30
  if (months < 12) return `${Math.round(months)}mo ago`
  return `${Math.round(months / 12)}y ago`
}

export function displayUrl(raw: string): string {
  try {
    const parsed = new URL(raw)
    const path = parsed.pathname === "/" ? "" : parsed.pathname
    return `${parsed.hostname}${path}${parsed.search}`
  } catch {
    return raw
  }
}
