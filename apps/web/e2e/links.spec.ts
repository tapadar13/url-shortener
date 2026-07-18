import { expect, test } from "@playwright/test"

import { registerAndOpenDashboard } from "./support/auth"

test.skip(
  process.env.PLAYWRIGHT_FULL_STACK !== "true",
  "requires the full API and MongoDB stack"
)

test("creates, redirects, updates, and deletes a short link", async ({
  page,
  request,
}, testInfo) => {
  await registerAndOpenDashboard(page, testInfo)

  const shortCode = `e2e${Date.now().toString(36)}${testInfo.workerIndex}`
  const originalURL = `https://example.com/original-${shortCode}`
  const updatedURL = `https://example.org/updated-${shortCode}`
  const apiBaseURL = `http://127.0.0.1:${process.env.E2E_API_PORT ?? "18080"}`
  const displayShortURL = `127.0.0.1:${process.env.E2E_API_PORT ?? "18080"}/${shortCode}`
  const links = page.getByRole("region", { name: "All links" })

  await expect(links.getByText("No links yet")).toBeVisible()
  await page.getByLabel("Destination URL").fill(originalURL)
  await page.getByRole("button", { name: "Link options" }).click()
  await page.getByLabel("Custom code").fill(shortCode)
  await page.getByRole("button", { name: "Shorten" }).click()

  const linkButton = links.getByRole("button", {
    name: displayShortURL,
    exact: true,
  })
  await expect(linkButton).toBeVisible()
  await expect(links).toContainText(`example.com/original-${shortCode}`)

  const firstRedirect = await request.get(`${apiBaseURL}/${shortCode}`, {
    maxRedirects: 0,
  })
  expect(firstRedirect.status()).toBe(302)
  expect(firstRedirect.headers().location).toBe(originalURL)

  await linkButton.click()
  const statsDialog = page.getByRole("dialog")
  await expect(statsDialog).toBeVisible()
  const overview = statsDialog.getByRole("region", { name: "Link overview" })
  await expect(overview.getByText("Total visits").locator("..")).toContainText(
    "1"
  )
  await statsDialog.getByRole("button", { name: "Close" }).click()

  const actionsButton = links.getByRole("button", {
    name: `Actions for ${displayShortURL}`,
  })
  await actionsButton.click()
  await page.getByRole("menuitem", { name: "Edit destination" }).click()
  await page.getByLabel("Destination", { exact: true }).fill(updatedURL)
  await page.getByRole("button", { name: "Save destination" }).click()
  await expect(links).toContainText(`example.org/updated-${shortCode}`)

  const updatedRedirect = await request.get(`${apiBaseURL}/${shortCode}`, {
    maxRedirects: 0,
  })
  expect(updatedRedirect.status()).toBe(302)
  expect(updatedRedirect.headers().location).toBe(updatedURL)

  await actionsButton.click()
  await page.getByRole("menuitem", { name: "Delete link" }).click()
  const deleteDialog = page.getByRole("alertdialog")
  await expect(deleteDialog).toContainText(`Delete ${displayShortURL}?`)
  await deleteDialog.getByRole("button", { name: "Delete link" }).click()

  await expect(linkButton).toHaveCount(0)
  await expect(links.getByText("No links yet")).toBeVisible()
})
