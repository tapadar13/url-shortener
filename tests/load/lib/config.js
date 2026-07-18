const DEFAULT_API_BASE_URL = "http://localhost:8080"

export const apiBaseURL = trimTrailingSlashes(
  __ENV.API_BASE_URL || DEFAULT_API_BASE_URL
)

export function positiveInteger(name, fallback) {
  const rawValue = __ENV[name]
  if (rawValue === undefined || rawValue === "") return fallback

  const value = Number(rawValue)
  if (!Number.isInteger(value) || value <= 0) {
    throw new Error(`${name} must be a positive integer`)
  }

  return value
}

export function duration(name, fallback) {
  const value = __ENV[name]
  return value === undefined || value === "" ? fallback : value
}

function trimTrailingSlashes(value) {
  return value.replace(/\/+$/, "")
}
