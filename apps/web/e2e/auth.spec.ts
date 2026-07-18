import { expect, test } from "@playwright/test"

import {
  createTestCredentials,
  loginCurrentPage,
  registerCurrentPage,
} from "./support/auth"

test.skip(
  process.env.PLAYWRIGHT_FULL_STACK !== "true",
  "requires the full API and MongoDB stack"
)

test("registers, restores the session, signs out, and logs back in", async ({
  page,
}, testInfo) => {
  const credentials = createTestCredentials(testInfo)

  await page.goto("/dashboard")
  await expect(page).toHaveURL(/\/login\?returnTo=%2Fdashboard$/)

  await page.getByRole("link", { name: "Create an account" }).click()
  await expect(page).toHaveURL(/\/register\?returnTo=%2Fdashboard$/)
  await registerCurrentPage(page, credentials)
  await expect(
    page.getByRole("heading", { level: 1, name: "Your links" })
  ).toBeVisible()

  await page.reload()
  await expect(page.getByTitle(credentials.email)).toBeVisible()

  await page.getByRole("button", { name: "Sign out" }).click()
  await expect(page).toHaveURL(/\/$/)

  await page.goto("/dashboard")
  await expect(page).toHaveURL(/\/login\?returnTo=%2Fdashboard$/)
  await loginCurrentPage(page, credentials)
})
