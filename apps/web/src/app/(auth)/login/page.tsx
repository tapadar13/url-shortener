import type { Metadata } from "next"

import { AuthForm } from "@/components/auth/auth-form"
import { safeReturnPath } from "@/lib/navigation/return-path"

export const metadata: Metadata = {
  title: "Log in",
  description: "Log in to your Relay link workspace.",
}

interface LoginPageProps {
  searchParams: Promise<{ returnTo?: string | string[] }>
}

export default async function LoginPage({ searchParams }: LoginPageProps) {
  const { returnTo } = await searchParams

  return <AuthForm mode="login" returnTo={safeReturnPath(returnTo)} />
}
