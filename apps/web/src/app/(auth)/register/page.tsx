import type { Metadata } from "next"

import { AuthForm } from "@/components/auth/auth-form"
import { safeReturnPath } from "@/lib/navigation/return-path"

export const metadata: Metadata = {
  title: "Create account",
  description: "Create your Relay link workspace.",
}

interface RegisterPageProps {
  searchParams: Promise<{ returnTo?: string | string[] }>
}

export default async function RegisterPage({ searchParams }: RegisterPageProps) {
  const { returnTo } = await searchParams

  return <AuthForm mode="register" returnTo={safeReturnPath(returnTo)} />
}
