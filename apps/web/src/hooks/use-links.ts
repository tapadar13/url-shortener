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
} from "@/lib/links/browser-links"
import type {
  CreateLinkInput,
  LinkStats,
  ShortLinkListPage,
} from "@/lib/links/types"

export const linksQueryKey = ["links"] as const

type LinksData = InfiniteData<ShortLinkListPage, string | undefined>

export function useLinks() {
  return useInfiniteQuery({
    queryKey: linksQueryKey,
    queryFn: ({ pageParam, signal }) =>
      listLinks({ cursor: pageParam, limit: 8, signal }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.nextCursor,
    staleTime: 15_000,
  })
}

function mapPages(
  data: LinksData,
  mapItem: (item: LinkStats) => LinkStats | null
): LinksData {
  return {
    ...data,
    pages: data.pages.map((page) => ({
      ...page,
      items: page.items
        .map(mapItem)
        .filter((item): item is LinkStats => item !== null),
    })),
  }
}

export function useCreateLink() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: CreateLinkInput) => createLink(input),
    onSuccess: (created) => {
      if (!queryClient.getQueryData<LinksData>(linksQueryKey)) {
        void queryClient.invalidateQueries({ queryKey: linksQueryKey })
        return
      }
      queryClient.setQueryData<LinksData>(linksQueryKey, (data) => {
        if (!data || data.pages.length === 0) return data
        const [first, ...rest] = data.pages
        return {
          ...data,
          pages: [
            {
              ...first,
              items: [{ ...created, accessCount: 0 }, ...first.items],
            },
            ...rest,
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
      queryClient.setQueryData<LinksData>(linksQueryKey, (data) =>
        data
          ? mapPages(data, (item) =>
              item.shortCode === updated.shortCode
                ? { ...item, ...updated }
                : item
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
      await queryClient.cancelQueries({ queryKey: linksQueryKey })
      const previous = queryClient.getQueryData<LinksData>(linksQueryKey)
      queryClient.setQueryData<LinksData>(linksQueryKey, (data) =>
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
        queryClient.setQueryData(linksQueryKey, context.previous)
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: linksQueryKey })
    },
  })
}

export function useLinkStats(shortCode: string | null) {
  return useQuery({
    queryKey: ["link-stats", shortCode],
    queryFn: ({ signal }) => getLinkStats(shortCode as string, signal),
    enabled: shortCode !== null,
    refetchInterval: 5_000,
  })
}
