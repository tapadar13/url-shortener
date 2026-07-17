import { createAuthRoute } from "@/lib/auth/auth-route"

export const POST = createAuthRoute("/auth/register", 201)
