import { apiErrorResponse } from "@/lib/api/route-response"

export interface LinkRouteContext {
  params: Promise<{ shortCode: string }>
}

export async function readShortCode(
  context: LinkRouteContext
): Promise<string | undefined> {
  const { shortCode } = await context.params
  return /^[A-Za-z0-9]{4,32}$/.test(shortCode) ? shortCode : undefined
}

export function shortLinkAPIPath(shortCode: string, suffix = ""): string {
  return `/shorten/${encodeURIComponent(shortCode)}${suffix}`
}

export function allowedQuery(
  requestURL: string,
  names: readonly string[]
): string {
  const source = new URL(requestURL).searchParams
  const target = new URLSearchParams()

  for (const name of names) {
    const value = source.get(name)
    if (value !== null) {
      target.set(name, value)
    }
  }

  const query = target.toString()
  return query ? `?${query}` : ""
}

export function invalidShortCodeResponse(): Response {
  return apiErrorResponse(400, "invalid_short_code", "short code is invalid")
}
