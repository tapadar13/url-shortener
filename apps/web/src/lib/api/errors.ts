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

export async function parseAPIError(response: Response): Promise<APIRequestError> {
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
