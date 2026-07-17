import { apiRouteErrorResponse } from "@/lib/api/route-response"
import { requestAuthenticatedAPI } from "@/lib/auth/authenticated-client"

import {
  invalidShortCodeResponse,
  readShortCode,
  shortLinkAPIPath,
  type LinkRouteContext,
} from "./route-params"
import type { LinkStats } from "./types"

interface LinkStatsRouteDependencies {
  getStats: (shortCode: string, signal: AbortSignal) => Promise<LinkStats>
}

const defaultDependencies: LinkStatsRouteDependencies = {
  getStats: (shortCode, signal) =>
    requestAuthenticatedAPI<LinkStats>(shortLinkAPIPath(shortCode, "/stats"), {
      signal,
    }),
}

export function createLinkStatsRoute(
  dependencies: LinkStatsRouteDependencies = defaultDependencies
) {
  return async (
    request: Request,
    context: LinkRouteContext
  ): Promise<Response> => {
    const shortCode = await readShortCode(context)
    if (!shortCode) {
      return invalidShortCodeResponse()
    }

    try {
      const stats = await dependencies.getStats(shortCode, request.signal)
      return Response.json(stats, {
        status: 200,
        headers: { "Cache-Control": "no-store" },
      })
    } catch (error) {
      return apiRouteErrorResponse(error, "could not load link statistics")
    }
  }
}
