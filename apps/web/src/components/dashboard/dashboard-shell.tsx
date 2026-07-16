"use client"

import { useEffect } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { CheckCircle2, LoaderCircle, LogOut } from "lucide-react"

import { Brand } from "@/components/layout/brand"
import { Button } from "@/components/ui/button"
import { useAuthSession, useLogout } from "@/hooks/use-auth"

export function DashboardShell() {
  const router = useRouter()
  const session = useAuthSession()
  const logout = useLogout()

  useEffect(() => {
    if (session.data === null) {
      router.replace("/login")
    }
  }, [router, session.data])

  if (session.isPending || session.data === null) {
    return <WorkspaceLoading />
  }

  if (session.isError) {
    return (
      <main className="flex min-h-svh items-center justify-center bg-muted/30 px-5">
        <div className="w-full max-w-sm text-center">
          <h1 className="text-xl font-semibold tracking-normal">
            Couldn&apos;t load your workspace
          </h1>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">
            Relay could not reach the API. Check the service and try again.
          </p>
          <Button className="mt-5" onClick={() => void session.refetch()}>
            Try again
          </Button>
        </div>
      </main>
    )
  }

  const user = session.data

  return (
    <div className="min-h-svh bg-muted/30">
      <header className="border-b bg-background">
        <div className="mx-auto flex h-14 max-w-5xl items-center justify-between px-5 sm:px-8">
          <Link
            href="/"
            className="rounded-md outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
            aria-label="Relay home"
          >
            <Brand />
          </Link>
          <Button
            variant="ghost"
            onClick={() =>
              logout.mutate(undefined, {
                onSettled: () => router.replace("/"),
              })
            }
            disabled={logout.isPending}
          >
            {logout.isPending ? (
              <LoaderCircle className="animate-spin" aria-hidden="true" />
            ) : (
              <LogOut aria-hidden="true" />
            )}
            {logout.isPending ? "Signing out" : "Sign out"}
          </Button>
        </div>
      </header>

      <main className="mx-auto max-w-5xl px-5 py-10 sm:px-8 sm:py-14">
        <div className="max-w-2xl">
          <p className="text-xs font-medium text-muted-foreground">Workspace</p>
          <h1 className="mt-2 text-2xl font-semibold tracking-normal sm:text-3xl">
            Account
          </h1>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">
            Your authenticated Relay session is active.
          </p>
        </div>

        <section className="mt-8 max-w-2xl rounded-lg border bg-background" aria-labelledby="account-heading">
          <div className="border-b px-5 py-4 sm:px-6">
            <h2 id="account-heading" className="text-sm font-semibold">
              Account details
            </h2>
          </div>
          <dl>
            <div className="grid gap-1 border-b px-5 py-4 sm:grid-cols-[10rem_1fr] sm:px-6">
              <dt className="text-sm text-muted-foreground">Email</dt>
              <dd className="min-w-0 break-words text-sm font-medium">{user.email}</dd>
            </div>
            <div className="grid gap-1 px-5 py-4 sm:grid-cols-[10rem_1fr] sm:px-6">
              <dt className="text-sm text-muted-foreground">Session</dt>
              <dd className="flex items-center gap-2 text-sm font-medium">
                <CheckCircle2 className="size-4 text-brand" aria-hidden="true" />
                Active
              </dd>
            </div>
          </dl>
        </section>
      </main>
    </div>
  )
}

function WorkspaceLoading() {
  return (
    <main
      className="flex min-h-svh items-center justify-center bg-muted/30 text-muted-foreground"
      aria-live="polite"
    >
      <div className="flex items-center gap-2 text-sm">
        <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />
        Loading workspace
      </div>
    </main>
  )
}
