/**
 * Shapes mirror the Go API's JSON contract (`internal/transport/httpapi`).
 * `accessCount` / `lastAccessedAt` ride along on list items here because the
 * workspace shows them per row; the live list endpoint will need to expose
 * them the same way (they already exist on the domain model).
 */

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
