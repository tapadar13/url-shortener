import { defineConfig, devices } from "@playwright/test"
import path from "node:path"

const externalBaseURL = process.env.PLAYWRIGHT_BASE_URL
const baseURL = externalBaseURL ?? "http://127.0.0.1:3100"
const fullStack = process.env.PLAYWRIGHT_FULL_STACK === "true"
const apiPort = process.env.E2E_API_PORT ?? "18080"
const mongodbPort = process.env.E2E_MONGODB_PORT ?? "27018"
const apiBaseURL = `http://127.0.0.1:${apiPort}`
const apiDirectory = path.resolve(__dirname, "../api")

const webServer = {
  command: "npm run dev -- --hostname 127.0.0.1 --port 3100",
  env: {
    NEXT_DIST_DIR: ".next-e2e",
    ...(fullStack ? { API_BASE_URL: apiBaseURL } : {}),
  },
  url: baseURL,
  reuseExistingServer: false,
  timeout: 120_000,
}

const apiServer = {
  command: "go run ./cmd/api",
  cwd: apiDirectory,
  env: {
    APP_ENV: "test",
    HTTP_ADDR: `127.0.0.1:${apiPort}`,
    BASE_URL: apiBaseURL,
    MONGODB_URI: `mongodb://127.0.0.1:${mongodbPort}`,
    MONGODB_DATABASE: "url_shortener_e2e",
    REDIRECT_CACHE_ENABLED: "false",
    RATE_LIMIT_REQUESTS: "0",
    AUTH_TOKEN_SECRET: "e2e-auth-token-secret-at-least-32-characters",
    LOG_LEVEL: "error",
    LOG_FORMAT: "text",
  },
  url: `${apiBaseURL}/readyz`,
  reuseExistingServer: false,
  timeout: 120_000,
}

export default defineConfig({
  testDir: "./e2e",
  outputDir: "test-results",
  fullyParallel: true,
  forbidOnly: Boolean(process.env.CI),
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [["list"], ["html", { open: "never" }]],
  use: {
    baseURL,
    trace: "on-first-retry",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  webServer: externalBaseURL
    ? undefined
    : fullStack
      ? [apiServer, webServer]
      : webServer,
})
