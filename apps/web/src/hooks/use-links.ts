"use client"

import {
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
  type InfiniteData,
} from "@tanstack/react-query"

import {
  createLink,
  deleteLink,
  getLinkStats,
  listLinks,
  updateLink,
} from "@/lib/links/mock-api"
import type {
  CreateLinkInput,
  LinkListPage,
  LinkRecord,
} from "@/lib/links/types"

const LINKS_KEY = ["links"] as const

type LinksData = InfiniteData<LinkListPage, string | undefined>

export function useLinks() {
  return useInfiniteQuery({
    queryKey: LINKS_KEY,
    queryFn: ({ pageParam }) => listLinks({ cursor: pageParam }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.nextCursor,
    // Ambient refresh keeps the simulated visit counts ticking.
    refetchInterval: 7_000,
  })
}

function mapPages(
  data: LinksData,
  mapItem: (item: LinkRecord) => LinkRecord | null
): LinksData {
  let removed = 0
  const pages = data.pages.map((page) => {
    const items = page.items
      .map(mapItem)
      .filter((item): item is LinkRecord => item !== null)
    removed += page.items.length - items.length
    return { ...page, items }
  })
  return {
    ...data,
    pages: pages.map((page) => ({ ...page, total: page.total - removed })),
  }
}

export function useCreateLink() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: CreateLinkInput) => createLink(input),
    onSuccess: (created) => {
      queryClient.setQueryData<LinksData>(LINKS_KEY, (data) => {
        if (!data || data.pages.length === 0) return data
        const [first, ...rest] = data.pages
        return {
          ...data,
          pages: [
            {
              ...first,
              items: [created, ...first.items],
              total: first.total + 1,
            },
            ...rest.map((page) => ({ ...page, total: page.total + 1 })),
          ],
        }
      })
    },
  })
}

export function useUpdateLink() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (params: { shortCode: string; url: string }) =>
      updateLink(params.shortCode, params.url),
    onSuccess: (updated) => {
      queryClient.setQueryData<LinksData>(LINKS_KEY, (data) =>
        data
          ? mapPages(data, (item) =>
              item.shortCode === updated.shortCode ? updated : item
            )
          : data
      )
    },
  })
}

export function useDeleteLink() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (shortCode: string) => deleteLink(shortCode),
    onMutate: async (shortCode) => {
      await queryClient.cancelQueries({ queryKey: LINKS_KEY })
      const previous = queryClient.getQueryData<LinksData>(LINKS_KEY)
      queryClient.setQueryData<LinksData>(LINKS_KEY, (data) =>
        data
          ? mapPages(data, (item) =>
              item.shortCode === shortCode ? null : item
            )
          : data
      )
      return { previous }
    },
    onError: (_error, _shortCode, context) => {
      if (context?.previous) {
        queryClient.setQueryData(LINKS_KEY, context.previous)
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: LINKS_KEY })
    },
  })
}

export function useLinkStats(shortCode: string | null) {
  return useQuery({
    queryKey: ["link-stats", shortCode],
    queryFn: () => getLinkStats(shortCode as string),
    enabled: shortCode !== null,
    refetchInterval: 5_000,
  })
}
