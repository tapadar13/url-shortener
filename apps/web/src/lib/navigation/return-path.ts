const applicationOrigin = "https://relay.local"

export function safeReturnPath(
  value: string | string[] | undefined,
  fallback = "/dashboard"
): string {
  if (typeof value !== "string" || !value.startsWith("/")) {
    return fallback
  }

  try {
    const target = new URL(value, applicationOrigin)
    if (target.origin !== applicationOrigin) {
      return fallback
    }

    return `${target.pathname}${target.search}${target.hash}`
  } catch {
    return fallback
  }
}

export function authPath(path: "/login" | "/register", returnTo: string) {
  return `${path}?returnTo=${encodeURIComponent(returnTo)}`
}
