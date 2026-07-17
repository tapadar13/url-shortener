import { Toaster } from "@/components/ui/sonner"

export default function DashboardLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <>
      {children}
      <Toaster position="bottom-right" />
    </>
  )
}
