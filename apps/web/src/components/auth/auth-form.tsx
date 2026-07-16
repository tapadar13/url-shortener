"use client"

import { useState } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { Eye, EyeOff, LoaderCircle } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useLogin, useRegister } from "@/hooks/use-auth"
import { APIConnectionError, APIRequestError } from "@/lib/api/errors"

type AuthMode = "login" | "register"

const copy = {
  login: {
    title: "Welcome back",
    description: "Sign in to manage your links and review their performance.",
    submit: "Log in",
    pending: "Logging in",
    alternate: "New to Relay?",
    alternateAction: "Create an account",
    alternateHref: "/register",
  },
  register: {
    title: "Create your workspace",
    description: "Start managing dependable short links from one focused place.",
    submit: "Create account",
    pending: "Creating account",
    alternate: "Already have an account?",
    alternateAction: "Log in",
    alternateHref: "/login",
  },
} as const

export function AuthForm({ mode }: { mode: AuthMode }) {
  const router = useRouter()
  const loginMutation = useLogin()
  const registerMutation = useRegister()
  const mutation = mode === "login" ? loginMutation : registerMutation
  const [showPassword, setShowPassword] = useState(false)
  const [formError, setFormError] = useState<string>()
  const content = copy[mode]

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setFormError(undefined)

    const form = new FormData(event.currentTarget)
    const email = String(form.get("email") ?? "")
    const password = String(form.get("password") ?? "")
    if (mode === "register" && password !== form.get("confirmPassword")) {
      setFormError("Passwords do not match")
      return
    }

    mutation.mutate(
      { email, password },
      {
        onSuccess: () => router.replace("/dashboard"),
      }
    )
  }

  const error = formError ?? authErrorMessage(mutation.error)
  const passwordType = showPassword ? "text" : "password"

  return (
    <div className="w-full">
      <div className="mb-8">
        <h1 className="text-2xl font-semibold tracking-normal text-foreground">
          {content.title}
        </h1>
        <p className="mt-2 text-sm leading-6 text-muted-foreground">
          {content.description}
        </p>
      </div>

      <form
        aria-label={mode === "login" ? "Log in" : "Create account"}
        className="space-y-5"
        onSubmit={handleSubmit}
      >
        <div className="space-y-2">
          <Label htmlFor="email">Email</Label>
          <Input
            id="email"
            name="email"
            type="email"
            autoComplete="email"
            placeholder="you@example.com"
            maxLength={320}
            disabled={mutation.isPending}
            aria-invalid={Boolean(error)}
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="password">Password</Label>
          <div className="relative">
            <Input
              id="password"
              name="password"
              type={passwordType}
              autoComplete={mode === "login" ? "current-password" : "new-password"}
              className="pr-10"
              minLength={12}
              maxLength={72}
              disabled={mutation.isPending}
              aria-invalid={Boolean(error)}
              aria-describedby={mode === "register" ? "password-requirement" : undefined}
              required
            />
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              className="absolute top-1/2 right-1.5 -translate-y-1/2 text-muted-foreground"
              onClick={() => setShowPassword((visible) => !visible)}
              aria-label={showPassword ? "Hide password" : "Show password"}
              title={showPassword ? "Hide password" : "Show password"}
              disabled={mutation.isPending}
            >
              {showPassword ? (
                <EyeOff aria-hidden="true" />
              ) : (
                <Eye aria-hidden="true" />
              )}
            </Button>
          </div>
          {mode === "register" && (
            <p id="password-requirement" className="text-xs text-muted-foreground">
              Use at least 12 characters.
            </p>
          )}
        </div>

        {mode === "register" && (
          <div className="space-y-2">
            <Label htmlFor="confirmPassword">Confirm password</Label>
            <Input
              id="confirmPassword"
              name="confirmPassword"
              type={passwordType}
              autoComplete="new-password"
              minLength={12}
              maxLength={72}
              disabled={mutation.isPending}
              aria-invalid={Boolean(error)}
              required
            />
          </div>
        )}

        <div className="min-h-5" aria-live="polite">
          {error && (
            <p role="alert" className="text-sm text-destructive">
              {error}
            </p>
          )}
        </div>

        <Button type="submit" size="lg" className="w-full" disabled={mutation.isPending}>
          {mutation.isPending && <LoaderCircle className="animate-spin" aria-hidden="true" />}
          {mutation.isPending ? content.pending : content.submit}
        </Button>
      </form>

      <p className="mt-6 text-center text-sm text-muted-foreground">
        {content.alternate}{" "}
        <Link
          href={content.alternateHref}
          className="font-medium text-foreground underline-offset-4 hover:underline"
        >
          {content.alternateAction}
        </Link>
      </p>
    </div>
  )
}

function authErrorMessage(error: unknown): string | undefined {
  if (error instanceof APIRequestError) {
    return error.message
  }
  if (error instanceof APIConnectionError) {
    return "Unable to connect. Try again in a moment."
  }
  if (error) {
    return "Something went wrong. Try again."
  }
}
