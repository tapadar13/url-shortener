import type { Metadata } from "next"

import { AuthPage } from "@/components/auth/auth-page"
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

  return <AuthPage mode="register" returnTo={safeReturnPath(returnTo)} />
}
