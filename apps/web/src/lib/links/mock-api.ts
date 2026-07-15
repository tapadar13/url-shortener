import {
  ApiError,
  type CreateLinkInput,
  type LinkListPage,
  type LinkRecord,
} from "@/lib/links/types"

/**
 * Temporary in-memory stand-in for the Go API so the workspace UI can be
 * exercised end to end. Behavior intentionally mirrors the backend:
 * validated http/https destinations, unique Base62 codes with retries,
 * cursor pagination ordered by (createdAt, id) descending, and the
 * `{ code, message }` error envelope. Swap the functions here for real
 * `fetch` calls when the API is wired up.
 */

const BASE62 =
  "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const RESERVED_CODES = new Set([
  "shorten",
  "auth",
  "healthz",
  "readyz",
  "metrics",
  "links",
  "app",
  "www",
])
const DEFAULT_PAGE_SIZE = 8

const MINUTE = 60_000
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR

interface SeedSpec {
  code: string
  url: string
  visits: number
  ageMs: number
  lastVisitMs?: number
}

const seeds: SeedSpec[] = [
  { code: "launch", url: "https://relay.so/summer-campaign/launch?utm_source=newsletter&utm_medium=email", visits: 1284, ageMs: 26 * DAY, lastVisitMs: 2 * MINUTE },
  { code: "x7Kd2a", url: "https://docs.example.com/changelog/2026-june-release-notes", visits: 962, ageMs: 21 * DAY, lastVisitMs: 18 * MINUTE },
  { code: "aT5nWc", url: "https://github.com/tapadar13/url-shortener", visits: 417, ageMs: 19 * DAY, lastVisitMs: HOUR },
  { code: "hiring", url: "https://jobs.example.com/openings/senior-product-engineer?ref=social", visits: 388, ageMs: 17 * DAY, lastVisitMs: 3 * HOUR },
  { code: "Q9mBf3", url: "https://blog.example.com/posts/how-we-cut-our-redirect-latency-in-half", visits: 341, ageMs: 15 * DAY, lastVisitMs: 5 * HOUR },
  { code: "deck26", url: "https://drive.example.com/file/d/1x8Kq2mNfR7/view?usp=sharing", visits: 296, ageMs: 14 * DAY, lastVisitMs: 8 * HOUR },
  { code: "pF4wZn", url: "https://calendar.example.com/booking/maya/30min-intro-call", visits: 233, ageMs: 12 * DAY, lastVisitMs: 11 * HOUR },
  { code: "notes", url: "https://notes.example.com/shared/quarterly-planning-agenda-and-minutes", visits: 187, ageMs: 11 * DAY, lastVisitMs: DAY },
  { code: "mV2sHq", url: "https://status.example.com/incidents/2026-06-30-elevated-error-rates", visits: 154, ageMs: 9 * DAY, lastVisitMs: DAY + 6 * HOUR },
  { code: "shop", url: "https://store.example.com/collections/limited-run/products/field-notebook", visits: 129, ageMs: 8 * DAY, lastVisitMs: 2 * DAY },
  { code: "rN8cVy", url: "https://maps.example.com/place/Third+Wave+Coffee+Indiranagar/@12.97,77.64", visits: 96, ageMs: 6 * DAY, lastVisitMs: 2 * DAY + 9 * HOUR },
  { code: "wJ3tKp", url: "https://forms.example.com/community-meetup-july-rsvp", visits: 88, ageMs: 6 * DAY, lastVisitMs: 3 * DAY },
  { code: "gH6xLd", url: "https://www.example.com/press/relay-featured-in-makers-weekly-issue-142", visits: 61, ageMs: 4 * DAY, lastVisitMs: 3 * DAY + 4 * HOUR },
  { code: "docs", url: "https://docs.example.com/guides/getting-started-with-the-relay-api", visits: 47, ageMs: 3 * DAY, lastVisitMs: 4 * DAY },
  { code: "zB9qRt", url: "https://video.example.com/watch?v=product-walkthrough-v2", visits: 29, ageMs: 2 * DAY, lastVisitMs: 5 * DAY },
  { code: "kC5mWv", url: "https://newsletter.example.com/issues/what-we-shipped-in-june", visits: 18, ageMs: DAY + 3 * HOUR, lastVisitMs: 6 * HOUR },
  { code: "sD7pXe", url: "https://research.example.com/papers/url-entropy-and-collision-bounds.pdf", visits: 7, ageMs: 22 * HOUR },
  { code: "beta", url: "https://app.example.com/waitlist/beta-invite?cohort=july", visits: 3, ageMs: 9 * HOUR, lastVisitMs: 40 * MINUTE },
  { code: "yF1nGs", url: "https://community.example.com/t/show-and-tell-what-are-you-linking/482", visits: 0, ageMs: 2 * HOUR },
]

let sequence = 0

function record(spec: SeedSpec): LinkRecord {
  const now = Date.now()
  const createdAt = new Date(now - spec.ageMs)
  sequence += 1
  return {
    id: `url_${String(sequence).padStart(4, "0")}`,
    url: spec.url,
    shortCode: spec.code,
    accessCount: spec.visits,
    createdAt: createdAt.toISOString(),
    updatedAt: createdAt.toISOString(),
    lastAccessedAt:
      spec.lastVisitMs === undefined
        ? undefined
        : new Date(now - spec.lastVisitMs).toISOString(),
  }
}

const store: LinkRecord[] = seeds.map(record)

function sleep(): Promise<void> {
  const ms = 240 + Math.random() * 260
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/**
 * Ambient traffic so the workspace feels alive: between calls, busy links
 * occasionally pick up a visit or two.
 */
let lastTrafficAt = Date.now()

function simulateTraffic() {
  const elapsed = Date.now() - lastTrafficAt
  if (elapsed < 4_000) return
  lastTrafficAt = Date.now()

  for (const link of store) {
    const popularity = Math.min(link.accessCount / 400, 1)
    if (Math.random() < 0.12 + popularity * 0.25) {
      link.accessCount += 1 + Math.floor(Math.random() * 2)
      link.lastAccessedAt = new Date(
        Date.now() - Math.random() * 90_000
      ).toISOString()
    }
  }
}

function sortForList(): LinkRecord[] {
  return [...store].sort((a, b) => {
    if (a.createdAt !== b.createdAt) return a.createdAt < b.createdAt ? 1 : -1
    return a.id < b.id ? 1 : -1
  })
}

/** Cursor format mirrors the backend's ListCursor: base64url({v, createdAt, id}). */
function encodeCursor(item: LinkRecord): string {
  return btoa(
    JSON.stringify({ v: 1, createdAt: item.createdAt, id: item.id })
  ).replaceAll("+", "-").replaceAll("/", "_").replace(/=+$/, "")
}

function decodeCursor(cursor: string): { createdAt: string; id: string } {
  try {
    const payload = JSON.parse(
      atob(cursor.replaceAll("-", "+").replaceAll("_", "/"))
    )
    if (payload.v !== 1 || !payload.createdAt || !payload.id) throw new Error()
    return payload
  } catch {
    throw new ApiError("invalid_cursor", "pagination cursor is invalid")
  }
}

function validateDestination(raw: string): string {
  const value = raw.trim()
  if (!value) {
    throw new ApiError("invalid_url", "Paste a link first — the field is empty.")
  }
  let parsed: URL
  try {
    parsed = new URL(value)
  } catch {
    throw new ApiError(
      "invalid_url",
      "That doesn't look like a full URL. Include https:// at the start."
    )
  }
  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
    throw new ApiError(
      "invalid_url",
      "Only http and https destinations are supported."
    )
  }
  if (!parsed.hostname.includes(".")) {
    throw new ApiError("invalid_url", "The destination needs a real hostname.")
  }
  return parsed.toString()
}

function normalizeCustomCode(raw: string): string {
  const value = raw.trim()
  if (!/^[0-9A-Za-z]{4,20}$/.test(value)) {
    throw new ApiError(
      "invalid_short_code",
      "Custom codes are 4–20 letters or numbers."
    )
  }
  if (RESERVED_CODES.has(value.toLowerCase())) {
    throw new ApiError("invalid_short_code", `“${value}” is reserved.`)
  }
  return value
}

function generateCode(): string {
  for (let attempt = 0; attempt < 5; attempt++) {
    let code = ""
    for (let i = 0; i < 6; i++) {
      code += BASE62[Math.floor(Math.random() * BASE62.length)]
    }
    if (!store.some((link) => link.shortCode === code)) return code
  }
  throw new ApiError(
    "short_code_unavailable",
    "Couldn't generate a unique code. Try again."
  )
}

export async function listLinks(params: {
  cursor?: string
  limit?: number
}): Promise<LinkListPage> {
  await sleep()
  simulateTraffic()

  const limit = params.limit ?? DEFAULT_PAGE_SIZE
  const sorted = sortForList()

  let startIndex = 0
  if (params.cursor) {
    const decoded = decodeCursor(params.cursor)
    const cursorIndex = sorted.findIndex(
      (item) => item.createdAt === decoded.createdAt && item.id === decoded.id
    )
    startIndex = cursorIndex === -1 ? 0 : cursorIndex + 1
  }

  const items = sorted.slice(startIndex, startIndex + limit)
  const last = items[items.length - 1]
  const hasMore = startIndex + limit < sorted.length

  return {
    items: structuredClone(items),
    nextCursor: hasMore && last ? encodeCursor(last) : undefined,
    total: sorted.length,
  }
}

export async function createLink(input: CreateLinkInput): Promise<LinkRecord> {
  await sleep()

  const url = validateDestination(input.url)
  let shortCode: string
  if (input.shortCode && input.shortCode.trim() !== "") {
    shortCode = normalizeCustomCode(input.shortCode)
    if (store.some((link) => link.shortCode === shortCode)) {
      throw new ApiError(
        "short_code_taken",
        `“${shortCode}” is already in use. Pick another.`
      )
    }
  } else {
    shortCode = generateCode()
  }

  sequence += 1
  const now = new Date().toISOString()
  const created: LinkRecord = {
    id: `url_${String(sequence).padStart(4, "0")}`,
    url,
    shortCode,
    accessCount: 0,
    createdAt: now,
    updatedAt: now,
    expiresAt: input.expiresAt,
  }
  store.unshift(created)
  return structuredClone(created)
}

export async function updateLink(
  shortCode: string,
  url: string
): Promise<LinkRecord> {
  await sleep()

  const destination = validateDestination(url)
  const existing = store.find((link) => link.shortCode === shortCode)
  if (!existing) {
    throw new ApiError("not_found", "That short link no longer exists.")
  }
  existing.url = destination
  existing.updatedAt = new Date().toISOString()
  return structuredClone(existing)
}

export async function deleteLink(shortCode: string): Promise<void> {
  await sleep()

  const index = store.findIndex((link) => link.shortCode === shortCode)
  if (index === -1) {
    throw new ApiError("not_found", "That short link no longer exists.")
  }
  store.splice(index, 1)
}

export async function getLinkStats(shortCode: string): Promise<LinkRecord> {
  await sleep()
  simulateTraffic()

  const existing = store.find((link) => link.shortCode === shortCode)
  if (!existing) {
    throw new ApiError("not_found", "That short link no longer exists.")
  }
  return structuredClone(existing)
}
