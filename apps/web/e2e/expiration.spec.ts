import { expect, test } from "@playwright/test"

import { registerAndOpenDashboard } from "./support/auth"

test.skip(
  process.env.PLAYWRIGHT_FULL_STACK !== "true",
  "requires the full API and MongoDB stack"
)

test("creates an active link with a preset expiration", async ({
  page,
  request,
}, testInfo) => {
  await registerAndOpenDashboard(page, testInfo)

  const shortCode = `exp${Date.now().toString(36)}${testInfo.workerIndex}`
  const originalURL = `https://example.com/expiring-${shortCode}`
  const apiBaseURL = `http://127.0.0.1:${process.env.E2E_API_PORT ?? "18080"}`
  const displayShortURL = `127.0.0.1:${process.env.E2E_API_PORT ?? "18080"}/${shortCode}`
  const links = page.getByRole("region", { name: "All links" })

  await page.getByLabel("Destination URL").fill(originalURL)
  await page.getByRole("button", { name: "Link options" }).click()
  await page.getByLabel("Custom code").fill(shortCode)

  const expirationButton = page.getByRole("button", {
    name: "Never",
    exact: true,
  })
  await expirationButton.click()
  await page
    .getByRole("menuitemradio", { name: "In 24 hours" })
    .click()
  await expect(expirationButton).toHaveAccessibleName("In 24 hours")

  await page.getByRole("button", { name: "Shorten" }).click()

  await expect(
    links.getByRole("button", { name: displayShortURL, exact: true })
  ).toBeVisible()
  await expect(links.getByText("Expires in 1d", { exact: true })).toBeVisible()

  const redirect = await request.get(`${apiBaseURL}/${shortCode}`, {
    maxRedirects: 0,
  })
  expect(redirect.status()).toBe(302)
  expect(redirect.headers().location).toBe(originalURL)
})
