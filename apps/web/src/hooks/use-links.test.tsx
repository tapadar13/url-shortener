import { act, cleanup, renderHook, waitFor } from "@testing-library/react"
import {
  QueryClient,
  QueryClientProvider,
  type InfiniteData,
} from "@tanstack/react-query"
import { afterEach, describe, expect, it, vi } from "vitest"

import type {
  LinkAnalytics,
  LinkStats,
  ShortLink,
  ShortLinkListPage,
} from "@/lib/links/types"
import { analyticsDateRange } from "@/lib/links/analytics-range"

import {
  linksQueryKey,
  useCreateLink,
  useLinkAnalytics,
  useLinks,
  useUpdateLink,
} from "./use-links"

type LinksData = InfiniteData<ShortLinkListPage, string | undefined>

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

describe("link hooks", () => {
  it("loads the first cursor page through the BFF", async () => {
    const fetchMock = jsonFetch({ items: [existingLink] })
    vi.stubGlobal("fetch", fetchMock)
    const queryClient = testQueryClient()
    const { result } = renderHook(() => useLinks(), {
      wrapper: queryWrapper(queryClient),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.pages[0]?.items).toEqual([existingLink])
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/links?limit=8",
      expect.any(Object)
    )
  })

  it("adds a created link to the first page with zero visits", async () => {
    const created = shortLink({
      id: "link-2",
      shortCode: "New123",
      shortUrl: "https://rly.to/New123",
    })
    vi.stubGlobal("fetch", jsonFetch(created, 201))
    const queryClient = testQueryClient()
    seedLinks(queryClient)
    const { result } = renderHook(() => useCreateLink(), {
      wrapper: queryWrapper(queryClient),
    })

    await act(async () => {
      await result.current.mutateAsync({ url: created.url })
    })

    const links = cachedLinks(queryClient)
    expect(links[0]).toEqual({ ...created, accessCount: 0 })
    expect(links[1]).toEqual(existingLink)
  })

  it("preserves visit statistics when a destination is updated", async () => {
    const updated = shortLink({
      url: "https://example.com/new-destination",
      updatedAt: "2026-07-17T09:00:00Z",
    })
    vi.stubGlobal("fetch", jsonFetch(updated))
    const queryClient = testQueryClient()
    seedLinks(queryClient)
    const { result } = renderHook(() => useUpdateLink(), {
      wrapper: queryWrapper(queryClient),
    })

    await act(async () => {
      await result.current.mutateAsync({
        shortCode: updated.shortCode,
        url: updated.url,
      })
    })

    expect(cachedLinks(queryClient)[0]).toEqual({
      ...existingLink,
      ...updated,
      accessCount: existingLink.accessCount,
      lastAccessedAt: existingLink.lastAccessedAt,
    })
  })

  it("loads an explicit inclusive analytics range", async () => {
    const range = analyticsDateRange(7)
    const analytics: LinkAnalytics = {
      shortCode: existingLink.shortCode,
      ...range,
      totalClicks: 3,
      daily: [{ date: range.to, clicks: 3 }],
    }
    const fetchMock = jsonFetch(analytics)
    vi.stubGlobal("fetch", fetchMock)
    const queryClient = testQueryClient()
    const { result } = renderHook(
      () => useLinkAnalytics(existingLink.shortCode, 7),
      { wrapper: queryWrapper(queryClient) }
    )

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data).toEqual(analytics)
    expect(fetchMock).toHaveBeenCalledWith(
      `/api/links/${existingLink.shortCode}/analytics?from=${range.from}&to=${range.to}`,
      expect.objectContaining({ signal: expect.any(AbortSignal) })
    )
  })
})

const existingLink: LinkStats = {
  ...shortLink(),
  accessCount: 42,
  lastAccessedAt: "2026-07-17T08:30:00Z",
}

function shortLink(overrides: Partial<ShortLink> = {}): ShortLink {
  return {
    id: "link-1",
    url: "https://example.com/articles/123",
    shortCode: "AbC1234",
    shortUrl: "https://rly.to/AbC1234",
    createdAt: "2026-07-17T08:00:00Z",
    updatedAt: "2026-07-17T08:00:00Z",
    ...overrides,
  }
}

function seedLinks(queryClient: QueryClient) {
  queryClient.setQueryData<LinksData>(linksQueryKey, {
    pages: [{ items: [existingLink] }],
    pageParams: [undefined],
  })
}

function cachedLinks(queryClient: QueryClient): LinkStats[] {
  const data = queryClient.getQueryData<LinksData>(linksQueryKey)
  return data?.pages.flatMap((page) => page.items) ?? []
}

function testQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })
}

function queryWrapper(queryClient: QueryClient) {
  return function QueryWrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    )
  }
}

function jsonFetch(body: unknown, status = 200) {
  return vi.fn().mockResolvedValue(
    new Response(JSON.stringify(body), {
      status,
      headers: { "Content-Type": "application/json" },
    })
  )
}
