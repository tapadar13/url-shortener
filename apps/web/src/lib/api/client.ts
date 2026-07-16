import { getAPIBaseURL } from "./config"
import { APIConnectionError, parseAPIError } from "./errors"

export { APIConnectionError, APIRequestError } from "./errors"

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
    throw await parseAPIError(response)
  }
  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
}
