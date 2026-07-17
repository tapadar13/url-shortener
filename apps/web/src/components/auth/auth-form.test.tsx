import { act, cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { afterEach, describe, expect, it, vi } from "vitest"

import { authSessionQueryKey } from "@/hooks/use-auth"

import { AuthForm } from "./auth-form"

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

describe("AuthForm", () => {
  it("logs in and updates the authenticated session", async () => {
    const fetchMock = successfulAuthFetch(200)
    vi.stubGlobal("fetch", fetchMock)
    const { queryClient } = renderAuthForm("login")

    fillCredentials()
    fireEvent.submit(screen.getByRole("form", { name: "Log in" }))

    await waitFor(() =>
      expect(queryClient.getQueryData(authSessionQueryKey)).toEqual({
        id: "user-1",
        email: "user@example.com",
      })
    )
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/auth/login",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          email: "user@example.com",
          password: "correct horse battery staple",
        }),
      })
    )
  })

  it("preserves the destination when switching auth modes", () => {
    renderAuthForm("login", "/links?page=2")

    expect(
      screen.getByRole("link", { name: "Create an account" }).getAttribute("href")
    ).toBe("/register?returnTo=%2Flinks%3Fpage%3D2")
  })

  it("shows a sanitized API error", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: {
              code: "invalid_credentials",
              message: "email or password is invalid",
            },
          }),
          { status: 401, headers: { "Content-Type": "application/json" } }
        )
      )
    )
    renderAuthForm("login")

    fillCredentials()
    fireEvent.submit(screen.getByRole("form", { name: "Log in" }))

    expect((await screen.findByRole("alert")).textContent).toContain(
      "email or password is invalid"
    )
  })

  it("rejects mismatched registration passwords locally", async () => {
    const fetchMock = vi.fn()
    vi.stubGlobal("fetch", fetchMock)
    renderAuthForm("register")

    fillCredentials()
    fireEvent.change(screen.getByLabelText("Confirm password"), {
      target: { value: "a different secure password" },
    })
    fireEvent.submit(screen.getByRole("form", { name: "Create account" }))

    expect((await screen.findByRole("alert")).textContent).toContain(
      "Passwords do not match"
    )
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it("toggles password visibility", () => {
    renderAuthForm("login")
    const password = screen.getByLabelText("Password")

    expect(password.getAttribute("type")).toBe("password")
    act(() => screen.getByRole("button", { name: "Show password" }).click())
    expect(password.getAttribute("type")).toBe("text")
    expect(screen.getByRole("button", { name: "Hide password" })).toBeDefined()
  })
})

function renderAuthForm(mode: "login" | "register", returnTo?: string) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  const view = render(
    <QueryClientProvider client={queryClient}>
      <AuthForm mode={mode} returnTo={returnTo} />
    </QueryClientProvider>
  )

  return { ...view, queryClient }
}

function fillCredentials() {
  fireEvent.change(screen.getByLabelText("Email"), {
    target: { value: "user@example.com" },
  })
  fireEvent.change(screen.getByLabelText("Password"), {
    target: { value: "correct horse battery staple" },
  })
}

function successfulAuthFetch(status: number) {
  return vi.fn().mockResolvedValue(
    new Response(
      JSON.stringify({ user: { id: "user-1", email: "user@example.com" } }),
      { status, headers: { "Content-Type": "application/json" } }
    )
  )
}
