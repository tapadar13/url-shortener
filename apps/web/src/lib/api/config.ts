const defaultAPIBaseURL = "http://localhost:8080"

export function getAPIBaseURL(value = process.env.API_BASE_URL): string {
  const candidate = value?.trim() || defaultAPIBaseURL

  let parsed: URL
  try {
    parsed = new URL(candidate)
  } catch {
    throw new Error(
      "API_BASE_URL must be a valid HTTP or HTTPS URL without credentials, a query, or a fragment"
    )
  }

  if (
    !["http:", "https:"].includes(parsed.protocol) ||
    parsed.username !== "" ||
    parsed.password !== "" ||
    parsed.search !== "" ||
    parsed.hash !== ""
  ) {
    throw new Error(
      "API_BASE_URL must be a valid HTTP or HTTPS URL without credentials, a query, or a fragment"
    )
  }

  return `${parsed.origin}${parsed.pathname.replace(/\/+$/, "")}`
}
