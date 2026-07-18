import { expect, test } from "@playwright/test"

test.skip(
  process.env.PLAYWRIGHT_FULL_STACK !== "true",
  "requires the full API and MongoDB stack"
)

test("registers, restores the session, signs out, and logs back in", async ({
  page,
}, testInfo) => {
  const email = `e2e-${Date.now()}-${testInfo.workerIndex}@example.com`
  const password = "correct horse battery staple"

  await page.goto("/dashboard")
  await expect(page).toHaveURL(/\/login\?returnTo=%2Fdashboard$/)

  await page.getByRole("link", { name: "Create an account" }).click()
  await expect(page).toHaveURL(/\/register\?returnTo=%2Fdashboard$/)

  await page.getByLabel("Email").fill(email)
  await page.getByLabel("Password", { exact: true }).fill(password)
  await page.getByLabel("Confirm password").fill(password)
  await page.getByRole("button", { name: "Create account" }).click()

  await expect(page).toHaveURL(/\/dashboard$/)
  await expect(
    page.getByRole("heading", { level: 1, name: "Your links" })
  ).toBeVisible()
  await expect(page.getByTitle(email)).toBeVisible()

  await page.reload()
  await expect(page.getByTitle(email)).toBeVisible()

  await page.getByRole("button", { name: "Sign out" }).click()
  await expect(page).toHaveURL(/\/$/)

  await page.goto("/dashboard")
  await expect(page).toHaveURL(/\/login\?returnTo=%2Fdashboard$/)
  await page.getByLabel("Email").fill(email)
  await page.getByLabel("Password").fill(password)
  await page.getByRole("button", { name: "Log in" }).click()

  await expect(page).toHaveURL(/\/dashboard$/)
  await expect(page.getByTitle(email)).toBeVisible()
})
