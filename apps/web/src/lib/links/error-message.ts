import { APIConnectionError, APIRequestError } from "@/lib/api/errors"

export function linkErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof APIRequestError) {
    return error.message
  }
  if (error instanceof APIConnectionError) {
    return "Unable to connect. Try again in a moment."
  }
  return fallback
}
