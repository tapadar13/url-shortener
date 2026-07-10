import { Architecture } from "@/components/landing/architecture"
import { Capabilities } from "@/components/landing/capabilities"
import { EngineeringStrip } from "@/components/landing/engineering-strip"
import { Hero } from "@/components/landing/hero"
import { HowItWorks } from "@/components/landing/how-it-works"

export default function LandingPage() {
  return (
    <main className="flex-1">
      <Hero />
      <EngineeringStrip />
      <Capabilities />
      <HowItWorks />
      <Architecture />
    </main>
  )
}
