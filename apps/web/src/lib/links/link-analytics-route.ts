import { apiRouteErrorResponse } from "@/lib/api/route-response"
import { requestAuthenticatedAPI } from "@/lib/auth/authenticated-client"

import {
  allowedQuery,
  invalidShortCodeResponse,
  readShortCode,
  shortLinkAPIPath,
  type LinkRouteContext,
} from "./route-params"
import type { LinkAnalytics } from "./types"

interface LinkAnalyticsRouteDependencies {
  getAnalytics: (
    shortCode: string,
    query: string,
    signal: AbortSignal
  ) => Promise<LinkAnalytics>
}

const defaultDependencies: LinkAnalyticsRouteDependencies = {
  getAnalytics: (shortCode, query, signal) =>
    requestAuthenticatedAPI<LinkAnalytics>(
      `${shortLinkAPIPath(shortCode, "/analytics")}${query}`,
      { signal }
    ),
}

export function createLinkAnalyticsRoute(
  dependencies: LinkAnalyticsRouteDependencies = defaultDependencies
) {
  return async (
    request: Request,
    context: LinkRouteContext
  ): Promise<Response> => {
    const shortCode = await readShortCode(context)
    if (!shortCode) {
      return invalidShortCodeResponse()
    }

    const query = allowedQuery(request.url, ["from", "to"])
    try {
      const analytics = await dependencies.getAnalytics(
        shortCode,
        query,
        request.signal
      )
      return Response.json(analytics, {
        status: 200,
        headers: { "Cache-Control": "no-store" },
      })
    } catch (error) {
      return apiRouteErrorResponse(error, "could not load link analytics")
    }
  }
}
