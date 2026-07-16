import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { afterEach, describe, expect, it, vi } from "vitest"

import { AuthPage } from "./auth-page"

const navigation = vi.hoisted(() => ({ replace: vi.fn() }))

vi.mock("next/navigation", () => ({
  useRouter: () => navigation,
}))

afterEach(() => {
  cleanup()
  navigation.replace.mockReset()
  vi.unstubAllGlobals()
})

describe("AuthPage", () => {
  it("redirects an authenticated visitor to the requested destination", async () => {
    vi.stubGlobal("fetch", sessionFetch(200))
    renderAuthPage()

    await waitFor(() =>
      expect(navigation.replace).toHaveBeenCalledWith("/links?page=2")
    )
    expect(screen.queryByRole("form", { name: "Log in" })).toBeNull()
  })

  it("shows the form for an anonymous visitor", async () => {
    vi.stubGlobal("fetch", sessionFetch(401))
    renderAuthPage()

    expect(await screen.findByRole("form", { name: "Log in" })).toBeDefined()
    expect(navigation.replace).not.toHaveBeenCalled()
  })

  it("redirects after a successful login", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(authResponse(401))
      .mockResolvedValueOnce(authResponse(200))
    vi.stubGlobal("fetch", fetchMock)
    renderAuthPage()

    await screen.findByRole("form", { name: "Log in" })
    fireEvent.change(screen.getByLabelText("Email"), {
      target: { value: "user@example.com" },
    })
    fireEvent.change(screen.getByLabelText("Password"), {
      target: { value: "correct horse battery staple" },
    })
    fireEvent.submit(screen.getByRole("form", { name: "Log in" }))

    await waitFor(() =>
      expect(navigation.replace).toHaveBeenCalledWith("/links?page=2")
    )
  })

  it("keeps the form available when session restoration fails", async () => {
    vi.stubGlobal("fetch", sessionFetch(502))
    renderAuthPage()

    expect(await screen.findByRole("form", { name: "Log in" })).toBeDefined()
    expect(navigation.replace).not.toHaveBeenCalled()
  })
})

function renderAuthPage() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <AuthPage mode="login" returnTo="/links?page=2" />
    </QueryClientProvider>
  )
}

function sessionFetch(status: number) {
  return vi.fn().mockResolvedValue(authResponse(status))
}

function authResponse(status: number) {
  const body =
    status === 200
      ? { user: { id: "user-1", email: "user@example.com" } }
      : {
          error: {
            code: status === 401 ? "unauthorized" : "api_unavailable",
            message: status === 401 ? "authentication is required" : "offline",
          },
        }

  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  })
}
