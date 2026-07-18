import { expect, test } from "@playwright/test"

import { registerAndOpenDashboard } from "./support/auth"

test.skip(
  process.env.PLAYWRIGHT_FULL_STACK !== "true",
  "requires the full API and MongoDB stack"
)

test("records redirects in link analytics", async ({
  page,
  request,
}, testInfo) => {
  await registerAndOpenDashboard(page, testInfo)

  const shortCode = `ana${Date.now().toString(36)}${testInfo.workerIndex}`
  const originalURL = `https://example.com/analytics-${shortCode}`
  const apiBaseURL = `http://127.0.0.1:${process.env.E2E_API_PORT ?? "18080"}`
  const displayShortURL = `127.0.0.1:${process.env.E2E_API_PORT ?? "18080"}/${shortCode}`
  const links = page.getByRole("region", { name: "All links" })

  await page.getByLabel("Destination URL").fill(originalURL)
  await page.getByRole("button", { name: "Link options" }).click()
  await page.getByLabel("Custom code").fill(shortCode)
  await page.getByRole("button", { name: "Shorten" }).click()

  const linkButton = links.getByRole("button", {
    name: displayShortURL,
    exact: true,
  })
  await expect(linkButton).toBeVisible()

  for (let click = 0; click < 3; click += 1) {
    const redirect = await request.get(`${apiBaseURL}/${shortCode}`, {
      maxRedirects: 0,
    })
    expect(redirect.status()).toBe(302)
    expect(redirect.headers().location).toBe(originalURL)
  }

  await expect
    .poll(
      async () => {
        const response = await page.request.get(
          `/api/links/${shortCode}/analytics`
        )
        if (!response.ok()) return -1

        const report = (await response.json()) as { totalClicks: number }
        return report.totalClicks
      },
      { message: "wait for asynchronous click analytics", timeout: 10_000 }
    )
    .toBe(3)

  await linkButton.click()

  const statsDialog = page.getByRole("dialog")
  const overview = statsDialog.getByRole("region", { name: "Link overview" })
  await expect(overview.getByText("Total visits").locator("..")).toContainText(
    "3"
  )
  await expect(
    statsDialog.getByText("Clicks in range").locator("..")
  ).toContainText("3")
  await expect(
    statsDialog.getByRole("img", { name: /3 clicks from/ })
  ).toBeVisible()
})
