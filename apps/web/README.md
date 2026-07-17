# Web

The Next.js application for the URL shortener. It contains the public landing page, authentication flows, and the protected link-management workspace.

## Stack

- Next.js App Router and React Server Components
- TypeScript (strict)
- Tailwind CSS
- shadcn/ui + Lucide icons
- TanStack Query
- Vitest and Testing Library

## Features

- Login, registration, session restoration, token refresh, and logout
- Same-origin backend-for-frontend routes backed by secure HTTP-only cookies
- Link creation with custom codes and optional expiration
- Cursor-paginated link management with edit and delete actions
- Visit statistics and 7, 30, or 90-day click analytics

## Development

```bash
npm ci
cp .env.example .env.local
npm run dev
```

Then open http://localhost:3000.

`API_BASE_URL` is server-only and defaults to `http://localhost:8080`. Browser requests stay on the web origin and pass through the app's `/api` route handlers.

## Scripts

| Command         | Purpose                  |
| --------------- | ------------------------ |
| `npm run dev`   | Start the dev server     |
| `npm run lint`  | Run ESLint               |
| `npm run test`  | Run the Vitest suite      |
| `npm run build` | Production build         |
| `npm run start` | Serve a production build |

## Container

Build the standalone production image from the repository root:

```bash
make web-image
```

The image runs as a non-root user on port `3000`. The complete Compose stack sets `API_BASE_URL=http://api:8080` for internal service traffic.

## Structure

```text
src/
  app/                Pages and same-origin API route handlers
  components/
    auth/             Authentication forms and screens
    landing/          Landing page sections and product mockups
    links/            Authenticated workspace and link controls
    ui/               shadcn/ui primitives
  hooks/              Authentication and link query hooks
  lib/                BFF clients, route handlers, and domain helpers
  providers/          TanStack Query and theme providers
```

The working product name lives in `src/config/site.ts` and is referenced
everywhere else, so renaming the product is a one-file change.
