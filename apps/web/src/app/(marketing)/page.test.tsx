import { cleanup, render, screen } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { afterEach, describe, expect, it, vi } from "vitest"

import { SiteFooter } from "@/components/layout/site-footer"
import { SiteHeader } from "@/components/layout/site-header"
import { siteConfig } from "@/config/site"

import LandingPage from "./page"

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

describe("landing page", () => {
  it("renders the hero headline and calls to action", () => {
    render(<LandingPage />)

    expect(
      screen.getByRole("heading", { level: 1, name: siteConfig.tagline })
    ).toBeDefined()
    expect(
      screen.getAllByRole("link", { name: "Create your first link" }).length
    ).toBeGreaterThan(0)
    expect(
      screen.getByRole("link", { name: "See how it flows" })
    ).toBeDefined()
  })

  it("renders every anchored section", () => {
    const { container } = render(<LandingPage />)

    for (const item of siteConfig.nav) {
      const id = item.href.replace("#", "")
      expect(container.querySelector(`section[id="${id}"]`)).not.toBeNull()
    }
  })

  it("does not claim live functionality in the product preview", () => {
    render(<LandingPage />)

    expect(screen.getByText(/not a live workspace yet/i)).toBeDefined()
  })
})

describe("site header", () => {
  it("renders navigation and anonymous account entry points", async () => {
    vi.stubGlobal("fetch", sessionFetch(401))
    renderHeader()

    const nav = screen.getByRole("navigation", { name: "Main" })
    expect(nav).toBeDefined()
    for (const item of siteConfig.nav) {
      expect(
        screen.getAllByRole("link", { name: item.label }).length
      ).toBeGreaterThan(0)
    }
    expect((await screen.findAllByRole("link", { name: "Log in" })).length)
      .toBeGreaterThan(0)
    expect(
      screen.getAllByRole("link", { name: "Get started" }).length
    ).toBeGreaterThan(0)
  })

  it("links authenticated visitors to their dashboard", async () => {
    vi.stubGlobal("fetch", sessionFetch(200))
    renderHeader()

    const dashboardLinks = await screen.findAllByRole("link", {
      name: "Dashboard",
    })

    expect(dashboardLinks.length).toBeGreaterThan(0)
    for (const link of dashboardLinks) {
      expect(link.getAttribute("href")).toBe("/dashboard")
    }
    expect(screen.queryByRole("link", { name: "Log in" })).toBeNull()
  })

  it("keeps account entry points available when session loading fails", async () => {
    vi.stubGlobal("fetch", sessionFetch(502))
    renderHeader()

    expect((await screen.findAllByRole("link", { name: "Log in" })).length)
      .toBeGreaterThan(0)
  })
})

describe("site footer", () => {
  it("links to the GitHub repository", () => {
    render(<SiteFooter />)

    const repoLink = screen.getByRole("link", { name: /GitHub repository/i })
    expect(repoLink.getAttribute("href")).toBe(siteConfig.repoUrl)
  })
})

function renderHeader() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <SiteHeader />
    </QueryClientProvider>
  )
}

function sessionFetch(status: number) {
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
