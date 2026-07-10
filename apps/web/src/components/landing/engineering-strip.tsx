import { Activity, Braces, Database, Hash } from "lucide-react"

const points = [
  {
    icon: Hash,
    title: "Collision-safe codes",
    body: "Random Base62 codes backed by a unique index and bounded retries.",
  },
  {
    icon: Database,
    title: "MongoDB persistence",
    body: "Every link is a durable document with created, updated, and last-visit timestamps.",
  },
  {
    icon: Activity,
    title: "Atomic visit counting",
    body: "Designed to record each redirect as a single atomic counter update.",
  },
  {
    icon: Braces,
    title: "API-first design",
    body: "One clean JSON contract shared by the workspace and your own scripts.",
  },
] as const

export function EngineeringStrip() {
  return (
    <section aria-label="Engineering principles" className="border-y bg-muted/40">
      <div className="mx-auto grid max-w-6xl grid-cols-1 gap-x-8 gap-y-6 px-4 py-10 sm:grid-cols-2 sm:px-6 lg:grid-cols-4">
        {points.map((point) => (
          <div key={point.title} className="flex gap-3">
            <point.icon
              className="mt-0.5 size-4 shrink-0 text-brand"
              aria-hidden="true"
            />
            <div>
              <h3 className="text-sm font-medium">{point.title}</h3>
              <p className="mt-1 text-sm text-pretty text-muted-foreground">
                {point.body}
              </p>
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}
