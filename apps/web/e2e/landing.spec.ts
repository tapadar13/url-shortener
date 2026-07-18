import { expect, test } from "@playwright/test"

test("opens registration from the landing page", async ({ page }) => {
  await page.goto("/")

  await expect(page).toHaveTitle(/Relay/)
  await expect(
    page.getByRole("heading", {
      level: 1,
      name: "One link. Total clarity.",
    })
  ).toBeVisible()

  const registrationLink = page.getByRole("link", {
    name: "Create your first link",
  })
  await expect(registrationLink).toHaveAttribute("href", "/register")
  await registrationLink.click()

  await expect(page).toHaveURL(/\/register$/)
  await expect(
    page.getByRole("heading", { level: 1, name: "Create your workspace" })
  ).toBeVisible()
  await expect(page.getByRole("form", { name: "Create account" })).toBeVisible()
})
