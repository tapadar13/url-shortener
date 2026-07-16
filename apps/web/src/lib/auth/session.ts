import { cookies } from "next/headers"

import type { AuthTokens } from "./types"

export const authCookieNames = {
  accessToken: "relay_access_token",
  refreshToken: "relay_refresh_token",
} as const

interface CookieOptions {
  httpOnly: true
  sameSite: "lax"
  secure: boolean
  path: "/"
  expires?: Date
  maxAge?: number
}

export function getAuthCookieOptions(
  expiresAt: string,
  secure = process.env.NODE_ENV === "production"
): { accessToken: CookieOptions; refreshToken: CookieOptions } {
  const accessTokenExpiry = new Date(expiresAt)
  if (Number.isNaN(accessTokenExpiry.getTime())) {
    throw new TypeError("access token expiry must be a valid timestamp")
  }

  const shared = {
    httpOnly: true,
    sameSite: "lax",
    secure,
    path: "/",
  } as const

  return {
    accessToken: { ...shared, expires: accessTokenExpiry },
    refreshToken: shared,
  }
}

export async function writeAuthSession(tokens: AuthTokens): Promise<void> {
  const cookieStore = await cookies()
  const options = getAuthCookieOptions(tokens.expiresAt)

  cookieStore.set(
    authCookieNames.accessToken,
    tokens.accessToken,
    options.accessToken
  )
  cookieStore.set(
    authCookieNames.refreshToken,
    tokens.refreshToken,
    options.refreshToken
  )
}

export async function clearAuthSession(): Promise<void> {
  const cookieStore = await cookies()
  const expired = {
    httpOnly: true,
    sameSite: "lax" as const,
    secure: process.env.NODE_ENV === "production",
    path: "/" as const,
    maxAge: 0,
  }

  cookieStore.set(authCookieNames.accessToken, "", expired)
  cookieStore.set(authCookieNames.refreshToken, "", expired)
}

export async function readAccessToken(): Promise<string | undefined> {
  return (await cookies()).get(authCookieNames.accessToken)?.value
}

export async function readRefreshToken(): Promise<string | undefined> {
  return (await cookies()).get(authCookieNames.refreshToken)?.value
}
