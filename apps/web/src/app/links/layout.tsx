import type { Metadata } from "next"

import { Toaster } from "@/components/ui/sonner"

export const metadata: Metadata = {
  title: "Your links",
  description:
    "Create short links, change their destinations, and watch every visit — the Relay workspace.",
}

export default function LinksLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <>
      {children}
      <Toaster position="bottom-right" />
    </>
  )
}
