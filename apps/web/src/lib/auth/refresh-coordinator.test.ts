import { describe, expect, it, vi } from "vitest"

import { createRefreshCoordinator } from "./refresh-coordinator"
import type { RefreshResponse } from "./types"

const refreshed: RefreshResponse = {
  accessToken: "new-access-token",
  refreshToken: "new-refresh-token",
  tokenType: "Bearer",
  expiresAt: "2026-07-16T12:00:00Z",
}

describe("createRefreshCoordinator", () => {
  it("shares one in-flight rotation for the same refresh token", async () => {
    const deferred = deferredRefresh()
    const refresh = vi.fn().mockReturnValue(deferred.promise)
    const coordinate = createRefreshCoordinator(refresh)

    const first = coordinate("refresh-token")
    const second = coordinate("refresh-token")
    await Promise.resolve()

    expect(second).toBe(first)
    expect(refresh).toHaveBeenCalledOnce()

    deferred.resolve(refreshed)
    await expect(Promise.all([first, second])).resolves.toEqual([
      refreshed,
      refreshed,
    ])
  })

  it("starts a new rotation after the previous request settles", async () => {
    const refresh = vi.fn().mockResolvedValue(refreshed)
    const coordinate = createRefreshCoordinator(refresh)

    await coordinate("refresh-token")
    await coordinate("refresh-token")

    expect(refresh).toHaveBeenCalledTimes(2)
  })

  it("does not combine rotations for different sessions", async () => {
    const refresh = vi.fn().mockResolvedValue(refreshed)
    const coordinate = createRefreshCoordinator(refresh)

    await Promise.all([coordinate("token-a"), coordinate("token-b")])

    expect(refresh).toHaveBeenCalledTimes(2)
  })
})

function deferredRefresh() {
  let resolve!: (value: RefreshResponse) => void
  const promise = new Promise<RefreshResponse>((resolvePromise) => {
    resolve = resolvePromise
  })

  return { promise, resolve }
}
