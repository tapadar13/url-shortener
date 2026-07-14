import { cleanup, render, screen } from "@testing-library/react"
import { afterEach, describe, expect, it } from "vitest"

import { SiteFooter } from "@/components/layout/site-footer"
import { SiteHeader } from "@/components/layout/site-header"
import { siteConfig } from "@/config/site"

import LandingPage from "./page"

afterEach(cleanup)

describe("landing page", () => {
  it("renders the hero headline and calls to action", () => {
    render(<LandingPage />)

    expect(
      screen.getByRole("heading", { level: 1, name: siteConfig.tagline })
    ).toBeDefined()
    expect(
      screen.getAllByRole("button", { name: "Create your first link" }).length
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
  it("renders navigation and account entry points", () => {
    render(<SiteHeader />)

    const nav = screen.getByRole("navigation", { name: "Main" })
    expect(nav).toBeDefined()
    for (const item of siteConfig.nav) {
      expect(
        screen.getAllByRole("link", { name: item.label }).length
      ).toBeGreaterThan(0)
    }
    expect(
      screen.getAllByRole("button", { name: "Log in" }).length
    ).toBeGreaterThan(0)
    expect(
      screen.getAllByRole("button", { name: "Get started" }).length
    ).toBeGreaterThan(0)
  })
})

describe("site footer", () => {
  it("links to the GitHub repository", () => {
    render(<SiteFooter />)

    const repoLink = screen.getByRole("link", { name: /GitHub repository/i })
    expect(repoLink.getAttribute("href")).toBe(siteConfig.repoUrl)
  })
})
