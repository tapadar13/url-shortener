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
    title: "Accounts are almost here",
    description: `${siteConfig.name} is being built in the open, and sign-in ships in the next phase alongside the link workspace. Nothing to log into just yet — but the Go API and data model behind it are already taking shape.`,
  },
  "get-started": {
    title: "You're early — in a good way",
    description: `${siteConfig.name} is being built in the open. Account creation and the link workspace connect to the API in the next phase. Until then, the engineering is public: follow the repository to watch it come together.`,
  },
} as const

export function AuthDialog({ intent, children }: AuthDialogProps) {
  const { title, description } = copy[intent]

  return (
    <Dialog>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-md">
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
