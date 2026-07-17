import { apiRouteErrorResponse } from "@/lib/api/route-response"

import { requestAuthenticatedAPI } from "./authenticated-client"
import type { AuthUser } from "./types"

interface SessionRouteDependencies {
  currentUser: (signal: AbortSignal) => Promise<AuthUser>
}

const defaultDependencies: SessionRouteDependencies = {
  currentUser: (signal) =>
    requestAuthenticatedAPI<AuthUser>("/auth/me", { signal }),
}

export function createSessionRoute(
  dependencies: SessionRouteDependencies = defaultDependencies
): (request: Request) => Promise<Response> {
  return async (request) => {
    try {
      const user = await dependencies.currentUser(request.signal)
      return Response.json(
        { user },
        { status: 200, headers: { "Cache-Control": "no-store" } }
      )
    } catch (error) {
      return apiRouteErrorResponse(error, "could not restore session")
    }
  }
}
