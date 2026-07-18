import { expect, type Page, type TestInfo } from "@playwright/test"

export interface TestCredentials {
  email: string
  password: string
}

export function createTestCredentials(testInfo: TestInfo): TestCredentials {
  const uniqueID = `${Date.now().toString(36)}-${testInfo.workerIndex}`
  return {
    email: `e2e-${uniqueID}@example.com`,
    password: "correct horse battery staple",
  }
}

export async function registerCurrentPage(
  page: Page,
  credentials: TestCredentials
): Promise<void> {
  await page.getByLabel("Email").fill(credentials.email)
  await page
    .getByLabel("Password", { exact: true })
    .fill(credentials.password)
  await page.getByLabel("Confirm password").fill(credentials.password)
  await page.getByRole("button", { name: "Create account" }).click()

  await expect(page).toHaveURL(/\/dashboard$/)
  await expect(page.getByTitle(credentials.email)).toBeVisible()
}

export async function registerAndOpenDashboard(
  page: Page,
  testInfo: TestInfo
): Promise<TestCredentials> {
  const credentials = createTestCredentials(testInfo)
  await page.goto("/register?returnTo=%2Fdashboard")
  await registerCurrentPage(page, credentials)
  return credentials
}

export async function loginCurrentPage(
  page: Page,
  credentials: TestCredentials
): Promise<void> {
  await page.getByLabel("Email").fill(credentials.email)
  await page.getByLabel("Password").fill(credentials.password)
  await page.getByRole("button", { name: "Log in" }).click()

  await expect(page).toHaveURL(/\/dashboard$/)
  await expect(page.getByTitle(credentials.email)).toBeVisible()
}
