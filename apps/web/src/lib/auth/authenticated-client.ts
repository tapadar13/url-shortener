import { APIRequestError, requestAPI } from "@/lib/api/client"

import {
  clearAuthSession,
  readAccessToken,
  readRefreshToken,
  writeAuthSession,
} from "./session"
import { createRefreshCoordinator } from "./refresh-coordinator"
import type { AuthTokens, RefreshResponse } from "./types"

interface AuthenticatedClientDependencies {
  request: (path: string, init?: RequestInit) => Promise<unknown>
  readAccessToken: () => Promise<string | undefined>
  readRefreshToken: () => Promise<string | undefined>
  refreshSession: (refreshToken: string) => Promise<RefreshResponse>
  writeSession: (tokens: AuthTokens) => Promise<void>
  clearSession: () => Promise<void>
}

const refreshSession = createRefreshCoordinator((refreshToken) =>
  requestAPI<RefreshResponse>("/auth/refresh", {
    method: "POST",
    body: JSON.stringify({ refreshToken }),
  })
)

const defaultDependencies: AuthenticatedClientDependencies = {
  request: requestAPI,
  readAccessToken,
  readRefreshToken,
  refreshSession,
  writeSession: writeAuthSession,
  clearSession: clearAuthSession,
}

export async function requestAuthenticatedAPI<T>(
  path: string,
  init: RequestInit = {},
  dependencies: AuthenticatedClientDependencies = defaultDependencies
): Promise<T> {
  const accessToken = await dependencies.readAccessToken()
  if (accessToken) {
    try {
      return (await requestWithAccessToken(
        dependencies,
        path,
        accessToken,
        init
      )) as T
    } catch (error) {
      if (!isUnauthorized(error)) {
        throw error
      }
    }
  }

  const refreshToken = await dependencies.readRefreshToken()
  if (!refreshToken) {
    await dependencies.clearSession()
    throw authenticationRequired()
  }

  let refreshed: RefreshResponse
  try {
    refreshed = await dependencies.refreshSession(refreshToken)
  } catch (error) {
    if (isUnauthorized(error)) {
      await dependencies.clearSession()
      throw authenticationRequired()
    }
    throw error
  }

  await dependencies.writeSession(refreshed)

  try {
    return (await requestWithAccessToken(
      dependencies,
      path,
      refreshed.accessToken,
      init
    )) as T
  } catch (error) {
    if (isUnauthorized(error)) {
      await dependencies.clearSession()
      throw authenticationRequired()
    }
    throw error
  }
}

function requestWithAccessToken(
  dependencies: AuthenticatedClientDependencies,
  path: string,
  accessToken: string,
  init: RequestInit
): Promise<unknown> {
  const headers = new Headers(init.headers)
  headers.set("Authorization", `Bearer ${accessToken}`)

  return dependencies.request(path, { ...init, headers })
}

function isUnauthorized(error: unknown): boolean {
  return error instanceof APIRequestError && error.status === 401
}

function authenticationRequired(): APIRequestError {
  return new APIRequestError(
    401,
    "unauthorized",
    "authentication is required"
  )
}
