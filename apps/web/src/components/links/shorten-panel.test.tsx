import { createRef } from "react"
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { afterEach, describe, expect, it, vi } from "vitest"

import { ShortenPanel } from "./shorten-panel"

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

describe("ShortenPanel", () => {
  it("omits blank optional link settings", async () => {
    const fetchMock = createLinkFetch()
    vi.stubGlobal("fetch", fetchMock)
    renderPanel()

    fireEvent.change(screen.getByRole("textbox", { name: "Destination URL" }), {
      target: { value: "https://example.com/articles/123" },
    })
    fireEvent.click(screen.getByRole("button", { name: "Link options" }))

    const customCode = screen.getByLabelText("Custom code")
    expect(customCode.getAttribute("maxlength")).toBe("32")
    fireEvent.click(screen.getByRole("button", { name: "Shorten" }))

    await waitFor(() => expect(fetchMock).toHaveBeenCalled())
    expect(requestBody(fetchMock)).toEqual({
      url: "https://example.com/articles/123",
    })
  })

  it("submits a seven-day RFC3339 expiration", async () => {
    const fetchMock = createLinkFetch()
    vi.stubGlobal("fetch", fetchMock)
    renderPanel()

    fireEvent.change(screen.getByRole("textbox", { name: "Destination URL" }), {
      target: { value: "https://example.com/articles/123" },
    })
    fireEvent.click(screen.getByRole("button", { name: "Link options" }))
    fireEvent.pointerDown(screen.getByRole("button", { name: "Never" }), {
      button: 0,
      ctrlKey: false,
    })
    fireEvent.click(
      await screen.findByRole("menuitemradio", { name: "In 7 days" })
    )

    const submittedAt = Date.now()
    fireEvent.click(screen.getByRole("button", { name: "Shorten" }))

    await waitFor(() => expect(fetchMock).toHaveBeenCalled())
    const body = requestBody(fetchMock)
    const expiration = Date.parse(body.expiresAt as string)
    expect(expiration - submittedAt).toBeGreaterThanOrEqual(
      7 * 24 * 60 * 60 * 1000
    )
    expect(expiration - submittedAt).toBeLessThan(
      7 * 24 * 60 * 60 * 1000 + 2_000
    )
  })
})

function renderPanel() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <ShortenPanel inputRef={createRef<HTMLInputElement>()} />
    </QueryClientProvider>
  )
}

function createLinkFetch() {
  return vi.fn((_input: RequestInfo | URL, init?: RequestInit) => {
    const input = JSON.parse(String(init?.body)) as {
      url: string
      shortCode?: string
      expiresAt?: string
    }
    return Promise.resolve(
      Response.json(
        {
          id: "link-1",
          url: input.url,
          shortCode: input.shortCode ?? "AbC1234",
          shortUrl: "https://rly.to/AbC1234",
          createdAt: "2026-07-17T08:00:00Z",
          updatedAt: "2026-07-17T08:00:00Z",
          ...(input.expiresAt ? { expiresAt: input.expiresAt } : {}),
        },
        { status: 201 }
      )
    )
  })
}

function requestBody(fetchMock: ReturnType<typeof createLinkFetch>) {
  const init = fetchMock.mock.calls[0]?.[1]
  return JSON.parse(String(init?.body)) as Record<string, unknown>
}
