import { spawn } from "node:child_process"
import { fileURLToPath } from "node:url"

const composeFile = fileURLToPath(
  new URL("../../../deploy/docker-compose.e2e.yml", import.meta.url)
)
const composeArguments = [
  "compose",
  "--project-name",
  "url-shortener-e2e",
  "--file",
  composeFile,
]

let activeChild
let interrupted = false

for (const signal of ["SIGINT", "SIGTERM"]) {
  process.once(signal, () => {
    interrupted = true
    activeChild?.kill(signal)
  })
}

try {
  await run("docker", [
    ...composeArguments,
    "up",
    "--detach",
    "--wait",
    "--wait-timeout",
    "60",
  ])
  await run("npm", ["exec", "--", "playwright", "test"], {
    ...process.env,
    PLAYWRIGHT_FULL_STACK: "true",
  })
} catch (error) {
  console.error(error instanceof Error ? error.message : error)
  process.exitCode = interrupted ? 130 : 1
} finally {
  try {
    await run("docker", [
      ...composeArguments,
      "down",
      "--volumes",
      "--remove-orphans",
    ])
  } catch (error) {
    console.error("E2E cleanup failed:", error instanceof Error ? error.message : error)
    process.exitCode ||= 1
  }
}

function run(command, args, env = process.env) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      env,
      stdio: "inherit",
    })
    activeChild = child

    child.once("error", reject)
    child.once("exit", (code, signal) => {
      if (activeChild === child) {
        activeChild = undefined
      }
      if (code === 0) {
        resolve()
        return
      }

      const outcome = signal ? `signal ${signal}` : `exit code ${code}`
      reject(new Error(`${command} failed with ${outcome}`))
    })
  })
}
