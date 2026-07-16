import { APIConnectionError, APIRequestError } from "./client"

export function apiErrorResponse(
  status: number,
  code: string,
  message: string
): Response {
  return Response.json(
    { error: { code, message } },
    { status, headers: { "Cache-Control": "no-store" } }
  )
}

export function apiRouteErrorResponse(
  error: unknown,
  fallbackMessage: string
): Response {
  if (error instanceof APIRequestError) {
    return apiErrorResponse(error.status, error.code, error.message)
  }
  if (error instanceof APIConnectionError) {
    return apiErrorResponse(502, "api_unavailable", "API is unavailable")
  }

  return apiErrorResponse(500, "internal_error", fallbackMessage)
}
