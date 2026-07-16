import { act, cleanup, renderHook, waitFor } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { afterEach, describe, expect, it, vi } from "vitest"

import {
  authSessionQueryKey,
  useAuthSession,
  useLogin,
  useLogout,
  useRegister,
} from "./use-auth"

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

describe("authentication hooks", () => {
  it("loads the current session into Query cache", async () => {
    vi.stubGlobal("fetch", sessionFetch(200))
    const queryClient = testQueryClient()
    const { result } = renderHook(() => useAuthSession(), {
      wrapper: queryWrapper(queryClient),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data).toEqual(authUser)
    expect(queryClient.getQueryData(authSessionQueryKey)).toEqual(authUser)
  })

  it("stores an anonymous session as null", async () => {
    vi.stubGlobal("fetch", sessionFetch(401))
    const queryClient = testQueryClient()
    const { result } = renderHook(() => useAuthSession(), {
      wrapper: queryWrapper(queryClient),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data).toBeNull()
  })

  it.each([
    ["login", useLogin, 200],
    ["register", useRegister, 201],
  ] as const)("seeds the session after %s", async (_, useMutationHook, status) => {
    vi.stubGlobal("fetch", sessionFetch(status))
    const queryClient = testQueryClient()
    const { result } = renderHook(() => useMutationHook(), {
      wrapper: queryWrapper(queryClient),
    })

    await act(async () => {
      await result.current.mutateAsync({
        email: "user@example.com",
        password: "correct horse battery staple",
      })
    })

    expect(queryClient.getQueryData(authSessionQueryKey)).toEqual(authUser)
  })

  it("clears the cached session even when revocation fails", async () => {
    vi.stubGlobal("fetch", sessionFetch(502))
    const queryClient = testQueryClient()
    queryClient.setQueryData(authSessionQueryKey, authUser)
    const { result } = renderHook(() => useLogout(), {
      wrapper: queryWrapper(queryClient),
    })

    await act(async () => {
      await expect(result.current.mutateAsync()).rejects.toMatchObject({
        status: 502,
      })
    })

    expect(queryClient.getQueryData(authSessionQueryKey)).toBeNull()
  })
})

const authUser = { id: "user-1", email: "user@example.com" }

function sessionFetch(status: number) {
  const body =
    status === 200 || status === 201
      ? { user: authUser }
      : {
          error: {
            code: status === 401 ? "unauthorized" : "api_unavailable",
            message: status === 401 ? "authentication is required" : "offline",
          },
        }

  return vi.fn().mockResolvedValue(
    new Response(JSON.stringify(body), {
      status,
      headers: { "Content-Type": "application/json" },
    })
  )
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
