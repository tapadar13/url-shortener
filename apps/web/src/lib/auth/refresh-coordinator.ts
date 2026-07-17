import type { RefreshResponse } from "./types"

type RefreshSession = (refreshToken: string) => Promise<RefreshResponse>

export function createRefreshCoordinator(
  refresh: RefreshSession
): RefreshSession {
  const inFlight = new Map<string, Promise<RefreshResponse>>()

  return (refreshToken) => {
    const pending = inFlight.get(refreshToken)
    if (pending) {
      return pending
    }

    const operation = Promise.resolve().then(() => refresh(refreshToken))
    inFlight.set(refreshToken, operation)

    const clear = () => {
      if (inFlight.get(refreshToken) === operation) {
        inFlight.delete(refreshToken)
      }
    }
    void operation.then(clear, clear)

    return operation
  }
}
