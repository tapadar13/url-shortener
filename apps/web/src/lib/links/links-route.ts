import {
  apiErrorResponse,
  apiRouteErrorResponse,
} from "@/lib/api/route-response"
import { requestAuthenticatedAPI } from "@/lib/auth/authenticated-client"
import { isSameOriginRequest } from "@/lib/security/request-origin"

import { allowedQuery } from "./route-params"
import type {
  CreateLinkInput,
  ShortLink,
  ShortLinkListPage,
} from "./types"

interface LinksRouteDependencies {
  list: (query: string, signal: AbortSignal) => Promise<ShortLinkListPage>
  create: (input: CreateLinkInput, signal: AbortSignal) => Promise<ShortLink>
}

const defaultDependencies: LinksRouteDependencies = {
  list: (query, signal) =>
    requestAuthenticatedAPI<ShortLinkListPage>(`/shorten${query}`, { signal }),
  create: (input, signal) =>
    requestAuthenticatedAPI<ShortLink>("/shorten", {
      method: "POST",
      body: JSON.stringify(input),
      signal,
    }),
}

export function createLinksRoute(
  dependencies: LinksRouteDependencies = defaultDependencies
) {
  return {
    GET: async (request: Request): Promise<Response> => {
      const query = allowedQuery(request.url, ["limit", "cursor"])

      try {
        const page = await dependencies.list(query, request.signal)
        return noStoreJSON(page)
      } catch (error) {
        return apiRouteErrorResponse(error, "could not load links")
      }
    },

    POST: async (request: Request): Promise<Response> => {
      if (!isSameOriginRequest(request)) {
        return apiErrorResponse(
          403,
          "origin_not_allowed",
          "request origin is not allowed"
        )
      }

      let body: unknown
      try {
        body = await request.json()
      } catch {
        return invalidCreateRequest()
      }
      if (!isCreateLinkInput(body)) {
        return invalidCreateRequest()
      }

      const input = createLinkInput(body)
      try {
        const created = await dependencies.create(input, request.signal)
        return noStoreJSON(created, 201)
      } catch (error) {
        return apiRouteErrorResponse(error, "could not create link")
      }
    },
  }
}

function isCreateLinkInput(value: unknown): value is CreateLinkInput {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    return false
  }

  const candidate = value as Record<string, unknown>
  const allowed = new Set(["url", "shortCode", "expiresAt"])
  return (
    Object.keys(candidate).every((key) => allowed.has(key)) &&
    typeof candidate.url === "string" &&
    optionalString(candidate.shortCode) &&
    optionalString(candidate.expiresAt)
  )
}

function optionalString(value: unknown): boolean {
  return value === undefined || typeof value === "string"
}

function createLinkInput(value: CreateLinkInput): CreateLinkInput {
  return {
    url: value.url,
    ...(value.shortCode !== undefined && { shortCode: value.shortCode }),
    ...(value.expiresAt !== undefined && { expiresAt: value.expiresAt }),
  }
}

function invalidCreateRequest(): Response {
  return apiErrorResponse(
    400,
    "invalid_request",
    "request body must contain a URL and optional short code or expiration"
  )
}

function noStoreJSON(body: unknown, status = 200): Response {
  return Response.json(body, {
    status,
    headers: { "Cache-Control": "no-store" },
  })
}
