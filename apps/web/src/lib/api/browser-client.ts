import { APIConnectionError, parseAPIError } from "./errors"

export async function requestBFF<T>(
  path: string,
  init: RequestInit = {}
): Promise<T> {
  if (!path.startsWith("/api/") || path.startsWith("//")) {
    throw new TypeError("BFF request path must start with /api/")
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
    response = await fetch(path, {
      ...init,
      cache: "no-store",
      credentials: "same-origin",
      headers,
    })
  } catch (error) {
    if (error instanceof Error && error.name === "AbortError") {
      throw error
    }
    throw new APIConnectionError(error)
  }

  if (!response.ok) {
    throw await parseAPIError(response)
  }
  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
}
