import { APIRequestError, requestAPI } from "@/lib/api/client"
import {
  apiErrorResponse,
  apiRouteErrorResponse,
} from "@/lib/api/route-response"
import { isSameOriginRequest } from "@/lib/security/request-origin"

import { clearAuthSession, readRefreshToken } from "./session"

interface LogoutRouteDependencies {
  readRefreshToken: () => Promise<string | undefined>
  revoke: (refreshToken: string, signal: AbortSignal) => Promise<void>
  clearSession: () => Promise<void>
}

const defaultDependencies: LogoutRouteDependencies = {
  readRefreshToken,
  revoke: (refreshToken, signal) =>
    requestAPI<void>("/auth/logout", {
      method: "POST",
      body: JSON.stringify({ refreshToken }),
      signal,
    }),
  clearSession: clearAuthSession,
}

export function createLogoutRoute(
  dependencies: LogoutRouteDependencies = defaultDependencies
): (request: Request) => Promise<Response> {
  return async (request) => {
    if (!isSameOriginRequest(request)) {
      return apiErrorResponse(
        403,
        "origin_not_allowed",
        "request origin is not allowed"
      )
    }

    let failure: unknown

    try {
      const refreshToken = await dependencies.readRefreshToken()
      if (refreshToken) {
        try {
          await dependencies.revoke(refreshToken, request.signal)
        } catch (error) {
          if (!(error instanceof APIRequestError && error.status === 401)) {
            failure = error
          }
        }
      }
    } catch (error) {
      failure = error
    }

    try {
      await dependencies.clearSession()
    } catch (error) {
      failure ??= error
    }

    if (failure !== undefined) {
      return apiRouteErrorResponse(failure, "could not complete logout")
    }

    return new Response(null, {
      status: 204,
      headers: { "Cache-Control": "no-store" },
    })
  }
}
