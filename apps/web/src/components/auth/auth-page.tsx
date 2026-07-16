"use client"

import { useEffect } from "react"
import { useRouter } from "next/navigation"
import { LoaderCircle } from "lucide-react"

import { useAuthSession } from "@/hooks/use-auth"

import { AuthForm } from "./auth-form"

type AuthMode = "login" | "register"

interface AuthPageProps {
  mode: AuthMode
  returnTo: string
}

export function AuthPage({ mode, returnTo }: AuthPageProps) {
  const router = useRouter()
  const session = useAuthSession()

  useEffect(() => {
    if (session.data) {
      router.replace(returnTo)
    }
  }, [returnTo, router, session.data])

  if (session.isPending || session.data) {
    return (
      <div
        role="status"
        className="flex min-h-80 items-center justify-center gap-2 text-sm text-muted-foreground"
      >
        <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />
        Checking your session
      </div>
    )
  }

  return <AuthForm mode={mode} returnTo={returnTo} />
}
