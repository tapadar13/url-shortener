# Web

The Next.js frontend for the URL shortener. This phase ships the public
landing page; authentication, API integration, and link management follow
once the Go API contract is stable.

## Stack

- Next.js (App Router, React Server Components)
- TypeScript (strict)
- Tailwind CSS
- shadcn/ui + Lucide icons
- TanStack Query (provider wired, no requests yet)

## Development

```bash
npm install
npm run dev
```

Then open http://localhost:3000.

## Scripts

| Command         | Purpose                  |
| --------------- | ------------------------ |
| `npm run dev`   | Start the dev server     |
| `npm run lint`  | Run ESLint               |
| `npm run build` | Production build         |
| `npm run start` | Serve a production build |

## Structure

```text
src/
  app/(marketing)/   Landing page route
  components/
    landing/         Landing page sections and product mockups
    layout/          Header, footer, brand
    ui/              shadcn/ui primitives
  config/site.ts     Brand name, navigation, repository URL
  providers/         TanStack Query provider
```

The working product name lives in `src/config/site.ts` and is referenced
everywhere else, so renaming the product is a one-file change.
