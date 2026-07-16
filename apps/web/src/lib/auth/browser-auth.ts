import { requestBFF } from "@/lib/api/browser-client"
import { APIRequestError } from "@/lib/api/errors"

import type { AuthCredentials, AuthUser } from "./types"

interface SessionResponse {
  user: AuthUser
}

export async function getSession(): Promise<AuthUser | null> {
  try {
    const response = await requestBFF<SessionResponse>("/api/auth/session")
    return response.user
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      return null
    }
    throw error
  }
}

export function login(credentials: AuthCredentials): Promise<AuthUser> {
  return authenticate("/api/auth/login", credentials)
}

export function register(credentials: AuthCredentials): Promise<AuthUser> {
  return authenticate("/api/auth/register", credentials)
}

export function logout(): Promise<void> {
  return requestBFF<void>("/api/auth/logout", { method: "POST" })
}

async function authenticate(
  path: "/api/auth/login" | "/api/auth/register",
  credentials: AuthCredentials
): Promise<AuthUser> {
  const response = await requestBFF<SessionResponse>(path, {
    method: "POST",
    body: JSON.stringify(credentials),
  })
  return response.user
}
