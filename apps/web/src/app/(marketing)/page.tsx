import { Capabilities } from "@/components/landing/capabilities"
import { EngineeringStrip } from "@/components/landing/engineering-strip"
import { Hero } from "@/components/landing/hero"

export default function LandingPage() {
  return (
    <main className="flex-1">
      <Hero />
      <EngineeringStrip />
      <Capabilities />
    </main>
  )
}
