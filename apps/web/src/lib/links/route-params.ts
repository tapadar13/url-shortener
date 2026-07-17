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

export function invalidShortCodeResponse(): Response {
  return apiErrorResponse(400, "invalid_short_code", "short code is invalid")
}
