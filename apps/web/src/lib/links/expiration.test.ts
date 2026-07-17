import { describe, expect, it } from "vitest"

import {
  defaultCustomExpiration,
  expirationStatus,
  minimumCustomExpiration,
  resolveExpiration,
} from "./expiration"

const now = new Date("2026-07-17T08:00:00Z")

describe("resolveExpiration", () => {
  it("omits expiration for permanent links", () => {
    expect(resolveExpiration("never", "", now)).toEqual({})
  })

  it.each([
    ["day", "2026-07-18T08:00:00.000Z"],
    ["week", "2026-07-24T08:00:00.000Z"],
    ["month", "2026-08-16T08:00:00.000Z"],
  ] as const)("converts the %s preset to RFC3339", (preset, expiresAt) => {
    expect(resolveExpiration(preset, "", now)).toEqual({ expiresAt })
  })

  it("converts a future local date-time value", () => {
    const customValue = "2026-07-20T14:30"

    expect(resolveExpiration("custom", customValue, now)).toEqual({
      expiresAt: new Date(customValue).toISOString(),
    })
  })

  it.each(["", "not-a-date", "2026-07-17T08:00:00Z"])(
    "rejects an invalid or non-future custom value",
    (customValue) => {
      expect(resolveExpiration("custom", customValue, now)).toEqual({
        error: "Choose a future expiration date and time.",
      })
    }
  )
})

describe("expiration date-time inputs", () => {
  it("keeps the minimum safely ahead of the current minute", () => {
    const minimum = new Date(minimumCustomExpiration(now))
    expect(minimum.getTime()).toBeGreaterThanOrEqual(now.getTime() + 5 * 60_000)
  })

  it("defaults custom expiration to roughly one day ahead", () => {
    const expiration = new Date(defaultCustomExpiration(now))
    expect(expiration.getTime()).toBeGreaterThanOrEqual(now.getTime() + 24 * 60 * 60 * 1000)
    expect(expiration.getTime()).toBeLessThan(now.getTime() + 24 * 60 * 60 * 1000 + 60_000)
  })
})

describe("expirationStatus", () => {
  it("marks past timestamps as expired", () => {
    expect(expirationStatus("2026-07-17T07:59:00Z", now)).toMatchObject({
      state: "expired",
      label: "Expired",
    })
  })

  it("uses hours and days for nearby expiration", () => {
    expect(expirationStatus("2026-07-17T11:00:00Z", now)).toMatchObject({
      state: "expiring",
      label: "Expires in 3h",
    })
    expect(expirationStatus("2026-07-20T08:00:00Z", now)).toMatchObject({
      state: "expiring",
      label: "Expires in 3d",
    })
  })

  it("uses a calendar label for later expiration", () => {
    expect(expirationStatus("2026-08-16T08:00:00Z", now)).toMatchObject({
      state: "scheduled",
      label: "Expires Aug 16, 2026",
    })
  })
})
