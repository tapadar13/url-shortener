"use client"

import { useEffect, useRef, useState } from "react"

import { cn } from "@/lib/utils"

interface RevealProps {
  children: React.ReactNode
  className?: string
  /** Stagger offset in milliseconds. */
  delay?: number
}

/**
 * Fades content up as it scrolls into view. Content is fully visible during
 * server render and for reduced-motion users; the hidden state is only
 * applied after mount, and only to elements still below the viewport.
 */
export function Reveal({ children, className, delay = 0 }: RevealProps) {
  const ref = useRef<HTMLDivElement>(null)
  const [state, setState] = useState<"idle" | "hidden" | "shown">("idle")

  useEffect(() => {
    const el = ref.current
    if (!el) return
    if (
      typeof window.matchMedia !== "function" ||
      typeof IntersectionObserver === "undefined"
    ) {
      return
    }
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return
    if (el.getBoundingClientRect().top < window.innerHeight * 0.92) return

    setState("hidden")
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setState("shown")
          observer.disconnect()
        }
      },
      { rootMargin: "0px 0px -8% 0px" }
    )
    observer.observe(el)
    // Safety net: never leave content invisible if the observer misbehaves.
    const fallback = window.setTimeout(() => setState("shown"), 4000)
    return () => {
      observer.disconnect()
      window.clearTimeout(fallback)
    }
  }, [])

  return (
    <div
      ref={ref}
      style={delay ? { transitionDelay: `${delay}ms` } : undefined}
      className={cn(
        "transition-[opacity,translate] duration-700 ease-[cubic-bezier(0.16,1,0.3,1)]",
        state === "hidden" && "translate-y-4 opacity-0",
        className
      )}
    >
      {children}
    </div>
  )
}
