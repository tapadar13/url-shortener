export interface AuthUser {
  id: string
  email: string
}

export interface AuthCredentials {
  email: string
  password: string
}

export interface AuthTokens {
  accessToken: string
  refreshToken: string
  tokenType: "Bearer"
  expiresAt: string
}

export interface AuthResponse extends AuthTokens {
  user: AuthUser
}

export type RefreshResponse = AuthTokens
