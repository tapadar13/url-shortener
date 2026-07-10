import { siteConfig } from "@/config/site"

export default function LandingPage() {
  return (
    <main className="flex flex-1 items-center justify-center">
      <h1 className="text-2xl font-semibold">{siteConfig.name}</h1>
    </main>
  )
}
