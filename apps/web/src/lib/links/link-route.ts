import {
  apiErrorResponse,
  apiRouteErrorResponse,
} from "@/lib/api/route-response"
import { requestAuthenticatedAPI } from "@/lib/auth/authenticated-client"
import { isSameOriginRequest } from "@/lib/security/request-origin"

import {
  invalidShortCodeResponse,
  readShortCode,
  shortLinkAPIPath,
  type LinkRouteContext,
} from "./route-params"
import type { ShortLink } from "./types"

interface LinkRouteDependencies {
  get: (shortCode: string, signal: AbortSignal) => Promise<ShortLink>
  update: (
    shortCode: string,
    url: string,
    signal: AbortSignal
  ) => Promise<ShortLink>
  delete: (shortCode: string, signal: AbortSignal) => Promise<void>
}

const defaultDependencies: LinkRouteDependencies = {
  get: (shortCode, signal) =>
    requestAuthenticatedAPI<ShortLink>(shortLinkAPIPath(shortCode), { signal }),
  update: (shortCode, url, signal) =>
    requestAuthenticatedAPI<ShortLink>(shortLinkAPIPath(shortCode), {
      method: "PUT",
      body: JSON.stringify({ url }),
      signal,
    }),
  delete: (shortCode, signal) =>
    requestAuthenticatedAPI<void>(shortLinkAPIPath(shortCode), {
      method: "DELETE",
      signal,
    }),
}

export function createLinkRoute(
  dependencies: LinkRouteDependencies = defaultDependencies
) {
  return {
    GET: async (
      request: Request,
      context: LinkRouteContext
    ): Promise<Response> => {
      const shortCode = await readShortCode(context)
      if (!shortCode) {
        return invalidShortCodeResponse()
      }

      try {
        const link = await dependencies.get(shortCode, request.signal)
        return noStoreJSON(link)
      } catch (error) {
        return apiRouteErrorResponse(error, "could not load link")
      }
    },

    PUT: async (
      request: Request,
      context: LinkRouteContext
    ): Promise<Response> => {
      if (!isSameOriginRequest(request)) {
        return originNotAllowed()
      }

      const shortCode = await readShortCode(context)
      if (!shortCode) {
        return invalidShortCodeResponse()
      }

      let body: unknown
      try {
        body = await request.json()
      } catch {
        return invalidUpdateRequest()
      }
      if (!isUpdateLinkInput(body)) {
        return invalidUpdateRequest()
      }

      try {
        const updated = await dependencies.update(
          shortCode,
          body.url,
          request.signal
        )
        return noStoreJSON(updated)
      } catch (error) {
        return apiRouteErrorResponse(error, "could not update link")
      }
    },

    DELETE: async (
      request: Request,
      context: LinkRouteContext
    ): Promise<Response> => {
      if (!isSameOriginRequest(request)) {
        return originNotAllowed()
      }

      const shortCode = await readShortCode(context)
      if (!shortCode) {
        return invalidShortCodeResponse()
      }

      try {
        await dependencies.delete(shortCode, request.signal)
        return new Response(null, {
          status: 204,
          headers: { "Cache-Control": "no-store" },
        })
      } catch (error) {
        return apiRouteErrorResponse(error, "could not delete link")
      }
    },
  }
}

function isUpdateLinkInput(value: unknown): value is { url: string } {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    return false
  }

  const candidate = value as Record<string, unknown>
  return (
    Object.keys(candidate).length === 1 && typeof candidate.url === "string"
  )
}

function originNotAllowed(): Response {
  return apiErrorResponse(
    403,
    "origin_not_allowed",
    "request origin is not allowed"
  )
}

function invalidUpdateRequest(): Response {
  return apiErrorResponse(
    400,
    "invalid_request",
    "request body must contain a URL"
  )
}

function noStoreJSON(body: unknown): Response {
  return Response.json(body, {
    status: 200,
    headers: { "Cache-Control": "no-store" },
  })
}
