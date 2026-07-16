import { getAPIBaseURL } from "./config"

interface APIErrorPayload {
  error?: {
    code?: unknown
    message?: unknown
  }
}

export class APIRequestError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    message: string
  ) {
    super(message)
    this.name = "APIRequestError"
  }
}

export class APIConnectionError extends Error {
  constructor(cause: unknown) {
    super("Unable to reach the API", { cause })
    this.name = "APIConnectionError"
  }
}

export async function requestAPI<T>(
  path: string,
  init: RequestInit = {}
): Promise<T> {
  if (!path.startsWith("/") || path.startsWith("//")) {
    throw new TypeError("API request path must start with a single slash")
  }

  const headers = new Headers(init.headers)
  if (!headers.has("Accept")) {
    headers.set("Accept", "application/json")
  }
  if (init.body !== undefined && init.body !== null && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json")
  }

  let response: Response
  try {
    response = await fetch(`${getAPIBaseURL()}${path}`, {
      ...init,
      cache: "no-store",
      headers,
    })
  } catch (error) {
    if (error instanceof Error && error.name === "AbortError") {
      throw error
    }
    throw new APIConnectionError(error)
  }

  if (!response.ok) {
    throw await requestError(response)
  }
  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
}

async function requestError(response: Response): Promise<APIRequestError> {
  let payload: APIErrorPayload | undefined
  try {
    payload = (await response.json()) as APIErrorPayload
  } catch {
    // Upstream proxies may return non-JSON errors.
  }

  const code =
    typeof payload?.error?.code === "string"
      ? payload.error.code
      : "upstream_error"
  const message =
    typeof payload?.error?.message === "string"
      ? payload.error.message
      : response.statusText || "API request failed"

  return new APIRequestError(response.status, code, message)
}
