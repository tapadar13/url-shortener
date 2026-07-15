"use client"

import { useEffect } from "react"

/**
 * Scopes the dark olive workspace theme to this route. Radix portals render
 * outside the route subtree, so the class has to live on <html>.
 */
export function WorkspaceTheme({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    document.documentElement.classList.add("dark")
    return () => document.documentElement.classList.remove("dark")
  }, [])

  return <>{children}</>
}
