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
    return new URL(origin).origin === requestOrigin(request)
  } catch {
    return false
  }
}

function requestOrigin(request: Request): string {
  const requestURL = new URL(request.url)
  const host = request.headers.get("Host")
  if (!host) {
    return requestURL.origin
  }

  const hostURL = new URL(`${requestURL.protocol}//${host}`)
  if (
    hostURL.username ||
    hostURL.password ||
    hostURL.pathname !== "/" ||
    hostURL.search ||
    hostURL.hash
  ) {
    throw new Error("invalid host header")
  }

  return hostURL.origin
}
