export function isSameOriginRequest(request: Request): boolean {
  const fetchSite = request.headers.get("Sec-Fetch-Site")
  if (fetchSite !== null && fetchSite !== "same-origin") {
    return false
  }

  const origin = request.headers.get("Origin")
  if (!origin) {
    return false
  }

  try {
    return new URL(origin).origin === new URL(request.url).origin
  } catch {
    return false
  }
}
