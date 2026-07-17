"use client"

import { useEffect } from "react"
import { useRouter } from "next/navigation"
import { LoaderCircle } from "lucide-react"

import { Workspace } from "@/components/links/workspace"
import { Button } from "@/components/ui/button"
import { useAuthSession } from "@/hooks/use-auth"
import { authPath } from "@/lib/navigation/return-path"

export function DashboardShell() {
  const router = useRouter()
  const session = useAuthSession()

  useEffect(() => {
    if (session.data === null) {
      router.replace(authPath("/login", "/dashboard"))
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

  return <Workspace user={session.data} />
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
