import { requestBFF } from "@/lib/api/browser-client"

import type {
  CreateLinkInput,
  LinkAnalytics,
  LinkStats,
  ShortLink,
  ShortLinkListPage,
} from "./types"

interface ListLinksParams {
  cursor?: string
  limit?: number
  signal?: AbortSignal
}

interface AnalyticsRange {
  from?: string
  to?: string
}

export function listLinks(
  params: ListLinksParams = {}
): Promise<ShortLinkListPage> {
  const query = new URLSearchParams()
  if (params.limit !== undefined) {
    query.set("limit", String(params.limit))
  }
  if (params.cursor) {
    query.set("cursor", params.cursor)
  }

  const suffix = query.size > 0 ? `?${query.toString()}` : ""
  return requestBFF<ShortLinkListPage>(`/api/links${suffix}`, {
    signal: params.signal,
  })
}

export function createLink(input: CreateLinkInput): Promise<ShortLink> {
  return requestBFF<ShortLink>("/api/links", {
    method: "POST",
    body: JSON.stringify(input),
  })
}

export function getLink(shortCode: string): Promise<ShortLink> {
  return requestBFF<ShortLink>(linkPath(shortCode))
}

export function updateLink(
  shortCode: string,
  url: string
): Promise<ShortLink> {
  return requestBFF<ShortLink>(linkPath(shortCode), {
    method: "PUT",
    body: JSON.stringify({ url }),
  })
}

export function deleteLink(shortCode: string): Promise<void> {
  return requestBFF<void>(linkPath(shortCode), { method: "DELETE" })
}

export function getLinkStats(
  shortCode: string,
  signal?: AbortSignal
): Promise<LinkStats> {
  return requestBFF<LinkStats>(`${linkPath(shortCode)}/stats`, { signal })
}

export function getLinkAnalytics(
  shortCode: string,
  range: AnalyticsRange = {}
): Promise<LinkAnalytics> {
  const query = new URLSearchParams()
  if (range.from) {
    query.set("from", range.from)
  }
  if (range.to) {
    query.set("to", range.to)
  }

  const suffix = query.size > 0 ? `?${query.toString()}` : ""
  return requestBFF<LinkAnalytics>(
    `${linkPath(shortCode)}/analytics${suffix}`
  )
}

function linkPath(shortCode: string): string {
  return `/api/links/${encodeURIComponent(shortCode)}`
}
