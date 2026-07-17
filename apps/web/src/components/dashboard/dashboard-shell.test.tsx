import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { afterEach, describe, expect, it, vi } from "vitest"

import { DashboardShell } from "./dashboard-shell"

const navigation = vi.hoisted(() => ({ replace: vi.fn() }))

vi.mock("next/navigation", () => ({
  useRouter: () => navigation,
}))

afterEach(() => {
  cleanup()
  navigation.replace.mockReset()
  vi.unstubAllGlobals()
})

describe("DashboardShell", () => {
  it("renders the authenticated links workspace", async () => {
    vi.stubGlobal("fetch", dashboardFetch(200))
    renderDashboard()

    expect(await screen.findByText("user@example.com")).toBeDefined()
    expect(screen.getByRole("heading", { name: "Your links" })).toBeDefined()
    expect(navigation.replace).not.toHaveBeenCalled()
  })

  it("redirects anonymous visitors to login", async () => {
    vi.stubGlobal("fetch", dashboardFetch(401))
    renderDashboard()

    await waitFor(() =>
      expect(navigation.replace).toHaveBeenCalledWith(
        "/login?returnTo=%2Fdashboard"
      )
    )
  })

  it("shows a recoverable API outage state", async () => {
    vi.stubGlobal("fetch", dashboardFetch(502))
    renderDashboard()

    expect(
      await screen.findByRole("heading", { name: "Couldn't load your workspace" })
    ).toBeDefined()
    expect(screen.getByRole("button", { name: "Try again" })).toBeDefined()
  })

  it("logs out and returns to the landing page", async () => {
    const fetchMock = dashboardFetch(200)
    vi.stubGlobal("fetch", fetchMock)
    renderDashboard()

    fireEvent.click(await screen.findByRole("button", { name: "Sign out" }))

    await waitFor(() => expect(navigation.replace).toHaveBeenCalledWith("/"))
    expect(fetchMock).toHaveBeenLastCalledWith(
      "/api/auth/logout",
      expect.objectContaining({ method: "POST" })
    )
  })
})

function renderDashboard() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <DashboardShell />
    </QueryClientProvider>
  )
}

function dashboardFetch(sessionStatus: number) {
  return vi.fn((input: RequestInfo | URL) => {
    const path = String(input)
    if (path === "/api/auth/session") {
      return Promise.resolve(authResponse(sessionStatus))
    }
    if (path.startsWith("/api/links")) {
      return Promise.resolve(jsonResponse({ items: [] }))
    }
    if (path === "/api/auth/logout") {
      return Promise.resolve(new Response(null, { status: 204 }))
    }
    return Promise.reject(new Error(`Unexpected request: ${path}`))
  })
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

  return jsonResponse(body, status)
}

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  })
}
