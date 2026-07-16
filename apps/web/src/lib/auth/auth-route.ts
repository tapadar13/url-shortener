import { requestAPI } from "@/lib/api/client"
import {
  apiErrorResponse,
  apiRouteErrorResponse,
} from "@/lib/api/route-response"

import { writeAuthSession } from "./session"
import type { AuthCredentials, AuthResponse, AuthTokens } from "./types"

type AuthEndpoint = "/auth/login" | "/auth/register"

interface AuthRouteDependencies {
  authenticate: (
    endpoint: AuthEndpoint,
    credentials: AuthCredentials,
    signal: AbortSignal
  ) => Promise<AuthResponse>
  writeSession: (tokens: AuthTokens) => Promise<void>
}

const defaultDependencies: AuthRouteDependencies = {
  authenticate: (endpoint, credentials, signal) =>
    requestAPI<AuthResponse>(endpoint, {
      method: "POST",
      body: JSON.stringify(credentials),
      signal,
    }),
  writeSession: writeAuthSession,
}

export function createAuthRoute(
  endpoint: AuthEndpoint,
  successStatus: 200 | 201,
  dependencies: AuthRouteDependencies = defaultDependencies
): (request: Request) => Promise<Response> {
  return async (request) => {
    let body: unknown
    try {
      body = await request.json()
    } catch {
      return apiErrorResponse(
        400,
        "invalid_request",
        "request body must be valid JSON"
      )
    }

    if (!isAuthCredentials(body)) {
      return apiErrorResponse(
        400,
        "invalid_request",
        "email and password are required"
      )
    }

    const credentials: AuthCredentials = {
      email: body.email,
      password: body.password,
    }

    try {
      const auth = await dependencies.authenticate(
        endpoint,
        credentials,
        request.signal
      )
      await dependencies.writeSession(auth)

      return Response.json(
        { user: auth.user },
        {
          status: successStatus,
          headers: { "Cache-Control": "no-store" },
        }
      )
    } catch (error) {
      return apiRouteErrorResponse(error, "authentication failed")
    }
  }
}

function isAuthCredentials(
  value: unknown
): value is { email: string; password: string } {
  if (typeof value !== "object" || value === null) {
    return false
  }

  const candidate = value as Record<string, unknown>
  return (
    typeof candidate.email === "string" &&
    typeof candidate.password === "string"
  )
}
