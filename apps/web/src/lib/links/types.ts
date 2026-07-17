export interface ShortLink {
  id: string
  url: string
  shortCode: string
  shortUrl: string
  createdAt: string
  updatedAt: string
  expiresAt?: string
}

export interface ShortLinkListPage {
  items: ShortLink[]
  nextCursor?: string
}

export interface LinkStats extends ShortLink {
  accessCount: number
  lastAccessedAt?: string
}

export interface DailyClicks {
  date: string
  clicks: number
}

export interface LinkAnalytics {
  shortCode: string
  from: string
  to: string
  totalClicks: number
  daily: DailyClicks[]
}

/** Temporary workspace shape used by the mock adapter until API migration. */

export interface LinkRecord {
  id: string
  url: string
  shortCode: string
  accessCount: number
  createdAt: string
  updatedAt: string
  lastAccessedAt?: string
  expiresAt?: string
}

export interface LinkListPage {
  items: LinkRecord[]
  /** Opaque cursor for the next page, absent on the last page. */
  nextCursor?: string
  /** Total links in the workspace (mock convenience for the header count). */
  total: number
}

export interface CreateLinkInput {
  url: string
  shortCode?: string
  expiresAt?: string
}

/** Mirrors the API's `{ error: { code, message } }` envelope. */
export class ApiError extends Error {
  readonly code: string

  constructor(code: string, message: string) {
    super(message)
    this.code = code
  }
}
