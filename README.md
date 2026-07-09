# URL Shortener

A production-grade URL shortening service built with Go, MongoDB, and a future Next.js frontend.

## Project Layout

```text
apps/
  api/    Go backend service
  web/    Future Next.js frontend
```

The backend will be implemented first. The frontend will be added after the API is stable.

## Local Development

Create a local environment file:

```bash
cp .env.example .env
```

Start MongoDB:

```bash
docker compose -f deploy/docker-compose.yml up -d
```

Run the API:

```bash
cd apps/api
set -a
source ../../.env
set +a
go run ./cmd/api
```
