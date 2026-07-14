"use client"

import { ArrowUpRight } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { siteConfig } from "@/config/site"

interface AuthDialogProps {
  intent: "log-in" | "get-started"
  children: React.ReactNode
}

const copy = {
  "log-in": {
    title: "Your workspace is almost ready",
    description: `${siteConfig.name} is being built in the open, with account access arriving alongside the live link workspace. The Go service and data model underneath are already in place; the calm, focused experience comes next.`,
  },
  "get-started": {
    title: "You found Relay early",
    description: `${siteConfig.name} is being built in the open. Account creation and the live workspace connect to the API in the next phase. Until then, follow the public repository and see every thoughtful layer come together.`,
  },
} as const

export function AuthDialog({ intent, children }: AuthDialogProps) {
  const { title, description } = copy[intent]

  return (
    <Dialog>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="overflow-hidden rounded-[1.5rem] border-foreground/10 bg-background/95 p-6 shadow-[0_30px_100px_-35px_rgb(20_24_16/0.7)] backdrop-blur-xl sm:max-w-md">
        <div aria-hidden="true" className="absolute inset-x-0 top-0 h-1 bg-brand" />
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription className="text-pretty">
            {description}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="gap-2 sm:gap-2">
          <DialogClose asChild>
            <Button variant="outline">Close</Button>
          </DialogClose>
          <Button asChild>
            <a href={siteConfig.repoUrl} target="_blank" rel="noreferrer">
              View the repository
              <ArrowUpRight data-icon="inline-end" aria-hidden="true" />
            </a>
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
