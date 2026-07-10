import { Reveal } from "@/components/landing/reveal"
import { siteConfig } from "@/config/site"

const steps = [
  {
    title: "Submit a destination",
    body: "Paste the long URL you want to share. The API checks that it's a real HTTP or HTTPS address before anything is stored.",
    snippet: (
      <>
        <p className="text-muted-foreground">POST /shorten</p>
        <p>{`{ "url": "https://docs.example.com/changelog/…" }`}</p>
      </>
    ),
  },
  {
    title: "Get a unique short code",
    body: "A random six-character code is generated and checked against every existing link. On the rare clash, it simply tries again.",
    snippet: (
      <>
        <p className="text-muted-foreground">code generated</p>
        <p>
          x7Kd2a <span className="text-brand">· unique ✓</span>
        </p>
      </>
    ),
  },
  {
    title: "Share it, follow the visits",
    body: `Anyone opening ${siteConfig.shortHost}/x7Kd2a lands on your destination in one hop — and the visit is counted as it happens.`,
    snippet: (
      <>
        <p className="text-muted-foreground">GET /x7Kd2a → 302</p>
        <p>
          visits <span className="text-brand">1,283 → 1,284</span>
        </p>
      </>
    ),
  },
] as const

export function HowItWorks() {
  return (
    <section id="how-it-works" className="border-y bg-muted/40">
      <div className="mx-auto max-w-6xl px-4 py-20 sm:px-6 sm:py-28">
        <Reveal className="mx-auto max-w-2xl text-center">
          <p className="text-xs font-semibold tracking-[0.14em] text-brand uppercase">
            How it works
          </p>
          <h2 className="mt-3 text-3xl font-semibold text-balance sm:text-4xl">
            From long URL to measured link in three steps
          </h2>
        </Reveal>

        <ol className="mt-14 grid gap-6 sm:mt-16 lg:grid-cols-3">
          {steps.map((step, index) => (
            <Reveal key={step.title} delay={index * 100} className="flex">
              <li className="flex w-full flex-col rounded-lg border bg-background p-5 transition-shadow duration-300 hover:shadow-[0_8px_24px_-8px_rgb(0_0_0/0.10)]">
                <div className="flex items-center gap-3">
                  <span
                    className="flex size-6 shrink-0 items-center justify-center rounded-full bg-foreground font-mono text-xs font-medium text-background"
                    aria-hidden="true"
                  >
                    {index + 1}
                  </span>
                  <h3 className="text-base font-semibold">{step.title}</h3>
                </div>
                <p className="mt-3 flex-1 text-sm text-pretty text-muted-foreground">
                  {step.body}
                </p>
                <div className="mt-5 space-y-1 rounded-md border bg-muted/40 p-3 font-mono text-[11px]">
                  {step.snippet}
                </div>
              </li>
            </Reveal>
          ))}
        </ol>
      </div>
    </section>
  )
}
