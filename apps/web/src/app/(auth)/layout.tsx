import Link from "next/link"
import { BarChart3, Link2, ShieldCheck } from "lucide-react"

import { Brand } from "@/components/layout/brand"

const benefits = [
  { icon: Link2, label: "Change destinations without replacing shared links" },
  { icon: BarChart3, label: "Review daily clicks and lifetime activity" },
  { icon: ShieldCheck, label: "Keep every management action account-scoped" },
]

export default function AuthLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <main className="grid min-h-svh bg-background lg:grid-cols-[minmax(20rem,0.85fr)_minmax(30rem,1.15fr)]">
      <aside className="hidden flex-col justify-between bg-[#17211d] p-10 text-white lg:flex xl:p-14">
        <Link href="/" className="w-fit rounded-md outline-none focus-visible:ring-3 focus-visible:ring-white/40">
          <Brand className="text-white" />
        </Link>

        <div className="max-w-md">
          <p className="text-3xl leading-tight font-semibold tracking-normal text-white">
            Short links that stay useful after you share them.
          </p>
          <p className="mt-4 max-w-sm text-sm leading-6 text-white/65">
            Keep destinations, activity, and link health in one quiet workspace.
          </p>
          <div className="mt-10 space-y-5">
            {benefits.map(({ icon: Icon, label }) => (
              <div key={label} className="flex items-start gap-3 text-sm text-white/80">
                <span className="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-md bg-white/10 text-emerald-300">
                  <Icon className="size-3.5" aria-hidden="true" />
                </span>
                <span className="leading-6">{label}</span>
              </div>
            ))}
          </div>
        </div>

        <p className="text-xs text-white/40">Focused link operations, built on the Relay API.</p>
      </aside>

      <section className="flex min-w-0 items-center justify-center px-5 py-10 sm:px-8 lg:px-12">
        <div className="w-full max-w-sm">
          <Link
            href="/"
            className="mb-10 block w-fit rounded-md outline-none focus-visible:ring-3 focus-visible:ring-ring/50 lg:hidden"
          >
            <Brand />
          </Link>
          {children}
        </div>
      </section>
    </main>
  )
}
