import { afterEach, describe, expect, it, vi } from "vitest"

import { getSession, login, logout, register } from "./browser-auth"

afterEach(() => {
  vi.unstubAllGlobals()
})

describe("browser authentication", () => {
  it("returns the current session user", async () => {
    vi.stubGlobal("fetch", authFetch({ status: 200 }))

    await expect(getSession()).resolves.toEqual({
      id: "user-1",
      email: "user@example.com",
    })
  })

  it("treats an unauthorized session as anonymous", async () => {
    vi.stubGlobal("fetch", authFetch({ status: 401 }))

    await expect(getSession()).resolves.toBeNull()
  })

  it("preserves non-authentication session failures", async () => {
    vi.stubGlobal("fetch", authFetch({ status: 502 }))

    await expect(getSession()).rejects.toMatchObject({
      status: 502,
      code: "api_unavailable",
    })
  })

  it.each([
    ["login", login, "/api/auth/login"],
    ["register", register, "/api/auth/register"],
  ] as const)("submits credentials for %s", async (_, authenticate, path) => {
    const fetchMock = authFetch({ status: 200 })
    vi.stubGlobal("fetch", fetchMock)
    const credentials = {
      email: "user@example.com",
      password: "correct horse battery staple",
    }

    await expect(authenticate(credentials)).resolves.toEqual({
      id: "user-1",
      email: "user@example.com",
    })

    expect(fetchMock).toHaveBeenCalledWith(
      path,
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify(credentials),
      })
    )
  })

  it("logs out through the same-origin BFF", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }))
    vi.stubGlobal("fetch", fetchMock)

    await expect(logout()).resolves.toBeUndefined()
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/auth/logout",
      expect.objectContaining({ method: "POST" })
    )
  })
})

function authFetch({ status }: { status: number }) {
  const body =
    status === 200
      ? { user: { id: "user-1", email: "user@example.com" } }
      : {
          error: {
            code: status === 401 ? "unauthorized" : "api_unavailable",
            message: status === 401 ? "authentication is required" : "offline",
          },
        }

  return vi.fn().mockResolvedValue(
    new Response(JSON.stringify(body), {
      status,
      headers: { "Content-Type": "application/json" },
    })
  )
}
