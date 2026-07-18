# URL Shortener

A production-minded URL shortening service with a Go + MongoDB API and a Next.js application.

The API creates short links, manages destinations, returns link statistics and daily analytics, and redirects visitors while recording access counts. The web app provides authentication, link management, expiration controls, and analytics visualization through a same-origin backend-for-frontend layer.

## Project Layout

```text
apps/
  api/    Go HTTP API and MongoDB persistence
  web/    Next.js landing page and authenticated workspace
deploy/
  docker-compose.yml         Local web, API, MongoDB, and Redis stack
  docker-compose.load.yml    Isolated API load-test stack
tests/
  load/   k6 smoke, management, and redirect workloads
```

## Current Features

- Collision-safe Base62 short-code generation backed by a unique MongoDB index
- URL creation, retrieval, update, deletion, and statistics endpoints with optional custom codes
- Canonical public short URLs generated from the deployment's configured base URL
- Optional expiry timestamps with immediate expiry-aware reads and MongoDB TTL cleanup
- Distributed fixed-window request rate limiting with atomic MongoDB counters
- Optional Redis redirect cache with bounded asynchronous access recording
- Daily UTC click analytics backed by atomic MongoDB aggregates and bounded asynchronous recording
- Configurable short-link redirects with atomic access counting
- Strict URL and short-code validation
- Consistent JSON error responses
- Health and MongoDB-backed readiness probes
- Structured request logs, request correlation IDs, and panic recovery
- Prometheus-compatible request metrics grouped by route and status class
- Container-native API readiness checks that gate dependent service startup
- Email/password authentication with Bearer access tokens, rotating refresh sessions, logout, and URL ownership
- Same-origin Next.js backend-for-frontend with secure HTTP-only session cookies
- Authenticated link workspace with cursor pagination, custom codes, expiration, and daily analytics
- Graceful API shutdown with MongoDB, Redis, access-recorder, and analytics-recorder lifecycle management

## Prerequisites

- Go 1.26.5 or newer
- Docker and Docker Compose
- Node.js and npm for the frontend

## Local Development

Create a local environment file:

```bash
cp .env.example .env
```

Start MongoDB and Redis for local Go API development:

```bash
docker compose -f deploy/docker-compose.yml up -d mongodb redis
```

Set `REDIRECT_CACHE_ENABLED=true` in `.env` to exercise Redis-backed redirects when running the Go API directly. The complete containerized stack enables caching automatically.

Run the API:

```bash
cd apps/api
set -a
source ../../.env
set +a
go run ./cmd/api
```

The API listens on `http://localhost:8080` by default.

Run the frontend in a second terminal:

```bash
cd apps/web
npm ci
npm run dev
```

The frontend listens on `http://localhost:3000` and uses `API_BASE_URL=http://localhost:8080` by default.

Alternatively, run the complete containerized stack from the repository root:

```bash
make stack-up
```

This starts the web app on `http://localhost:3000`, the API on `http://localhost:8080`, MongoDB, and Redis. Replace the example `AUTH_TOKEN_SECRET` before using the Compose configuration outside local development.

## Common Commands

Run `make help` from the repository root to list local development commands.

```bash
make api-check
make web-check
make load-smoke
make data-up
make stack-up
```

Load-test scenarios, thresholds, and tuning variables are documented in
[tests/load/README.md](tests/load/README.md).

## API

The machine-readable API contract is available at [docs/openapi.yaml](docs/openapi.yaml).

All API error responses use this shape:

```json
{
  "error": {
    "code": "invalid_url",
    "message": "url must be a valid http or https URL"
  }
}
```

Every API response also includes a server-generated `X-Request-ID` header. Use it to correlate a client request with structured server logs.

When rate limiting is enabled, non-probe requests include `X-RateLimit-Limit`, `X-RateLimit-Remaining`, and `X-RateLimit-Reset`. Exceeded requests return `429 Too Many Requests` with `Retry-After`. Health checks, readiness checks, and CORS preflights do not consume quota.

URL management, statistics, analytics, and listing endpoints require a Bearer access token. Redirects remain public.

Clients are identified from the direct socket address by default. When `TRUSTED_PROXY_CIDRS` is configured, `X-Forwarded-For` is accepted only from a trusted socket peer and the proxy chain is evaluated from right to left, preventing clients from bypassing limits with spoofed values.

### Authentication

```http
POST /auth/register
POST /auth/login
POST /auth/refresh
POST /auth/logout
GET /auth/me
```

Registration and login accept `email` and a password from 12 characters up to bcrypt's 72-byte limit. Successful responses contain a short-lived Bearer access token, an opaque refresh token, and sanitized user details. Send the refresh token to `/auth/refresh` to rotate it and receive replacement credentials, or to `/auth/logout` to revoke the active session. Use `/auth/me` with the Bearer token to restore the current user when the frontend starts.

```bash
curl -i http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"correct horse battery staple"}'
```

### Create a Short URL

```http
POST /shorten
Content-Type: application/json
```

```json
{
  "url": "https://example.com/articles/123",
  "expiresAt": "2026-08-12T08:00:00Z",
  "shortCode": "summer2026"
}
```

Successful response: `201 Created`

```json
{
  "id": "507f1f77bcf86cd799439011",
  "url": "https://example.com/articles/123",
  "shortCode": "AbC1234",
  "shortUrl": "http://localhost:8080/AbC1234",
  "createdAt": "2026-07-12T08:00:00Z",
  "updatedAt": "2026-07-12T08:00:00Z",
  "expiresAt": "2026-08-12T08:00:00Z"
}
```

```bash
curl -i http://localhost:8080/shorten \
  -X POST \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/articles/123","expiresAt":"2026-08-12T08:00:00Z","shortCode":"summer2026"}'
```

### List My Short URLs

```http
GET /shorten?limit=25
Authorization: Bearer <access-token>
```

Returns the authenticated user’s newest short URLs with their current visit statistics. `limit` defaults to `25` and must be between `1` and `100`.

```json
{
  "items": [
    {
      "id": "507f1f77bcf86cd799439011",
      "url": "https://example.com/articles/123",
      "shortCode": "AbC1234",
      "shortUrl": "http://localhost:8080/AbC1234",
      "accessCount": 42,
      "createdAt": "2026-07-12T08:00:00Z",
      "updatedAt": "2026-07-12T08:00:00Z",
      "lastAccessedAt": "2026-07-17T09:15:00Z"
    }
  ],
  "nextCursor": "eyJ2IjoxLCJjcmVhdGVkQXQiOiIyMDI2LTA3LTEyVDA4OjAwOjAwWiIsImlkIjoiNTA3ZjFmNzdiY2Y4NmNkNzk5NDM5MDExIn0"
}
```

When `nextCursor` is present, pass it back unchanged to retrieve the next page:

```http
GET /shorten?limit=25&cursor=eyJ2IjoxLCJjcmVhdGVkQXQiOiIyMDI2LTA3LTEyVDA4OjAwOjAwWiIsImlkIjoiNTA3ZjFmNzdiY2Y4NmNkNzk5NDM5MDExIn0
Authorization: Bearer <access-token>
```

The cursor is opaque and is omitted when there are no more results.

Omit `shortCode` to have the service generate one. Custom codes must be 4-32 Base62 characters, cannot use reserved route names, and return `409 Conflict` when already taken.

### Retrieve a Short URL

```http
GET /shorten/{shortCode}
```

Successful response: `200 OK`

```bash
curl -i http://localhost:8080/shorten/AbC1234
```

### Update a Destination

```http
PUT /shorten/{shortCode}
Content-Type: application/json
```

```json
{
  "url": "https://example.com/new-destination"
}
```

Successful response: `200 OK`

```bash
curl -i http://localhost:8080/shorten/AbC1234 \
  -X PUT \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/new-destination"}'
```

### Delete a Short URL

```http
DELETE /shorten/{shortCode}
```

Successful response: `204 No Content`

```bash
curl -i http://localhost:8080/shorten/AbC1234 -X DELETE
```

### Retrieve Link Statistics

```http
GET /shorten/{shortCode}/stats
```

Successful response: `200 OK`

```json
{
  "id": "507f1f77bcf86cd799439011",
  "url": "https://example.com/articles/123",
  "shortCode": "AbC1234",
  "shortUrl": "http://localhost:8080/AbC1234",
  "accessCount": 42,
  "createdAt": "2026-07-12T08:00:00Z",
  "updatedAt": "2026-07-12T08:00:00Z",
  "lastAccessedAt": "2026-07-12T09:15:00Z",
  "expiresAt": "2026-08-12T08:00:00Z"
}
```

```bash
curl -i http://localhost:8080/shorten/AbC1234/stats
```

### Retrieve Daily Click Analytics

```http
GET /shorten/{shortCode}/analytics?from=2026-07-10&to=2026-07-12
```

Successful response: `200 OK`

```json
{
  "shortCode": "AbC1234",
  "from": "2026-07-10",
  "to": "2026-07-12",
  "totalClicks": 12,
  "daily": [
    {"date": "2026-07-10", "clicks": 5},
    {"date": "2026-07-11", "clicks": 0},
    {"date": "2026-07-12", "clicks": 7}
  ]
}
```

```bash
curl -i 'http://localhost:8080/shorten/AbC1234/analytics?from=2026-07-10&to=2026-07-12'
```

`from` and `to` are inclusive UTC dates. Omit them to request the latest 30 days; custom ranges cannot exceed 90 days or include future dates. Missing days are returned with zero clicks for direct charting. Analytics and cached access-count writes use independent queues, so daily totals may briefly differ from the all-time `accessCount` returned by the statistics endpoint.

### Redirect from a Short URL

```http
GET /{shortCode}
```

The API responds with the configured redirect status (`302` by default), sets `Location` to the original destination, and records `accessCount` with an atomic MongoDB update. Redis cache hits queue that update asynchronously to keep redirect latency low.

```bash
curl -i http://localhost:8080/AbC1234
```

### Probes

```http
GET /healthz
GET /readyz
```

- `/healthz` confirms that the HTTP process is running.
- `/readyz` additionally confirms that MongoDB can be reached. It returns `503 Service Unavailable` with `not_ready` when the dependency is unavailable. The API container uses this endpoint for its built-in health check, and Compose waits for it before starting the frontend.

```bash
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/readyz
```

### Metrics

```http
GET /metrics
```

The endpoint returns Prometheus text format with request totals and cumulative
request duration grouped by HTTP method, registered route pattern, and status
class. Keep this endpoint on an internal network or protect it at the reverse
proxy when deploying publicly.

```bash
curl -i http://localhost:8080/metrics
```

### Status Codes

| Status | Meaning |
| --- | --- |
| `200` | Successful retrieval, update, statistics, or analytics response |
| `201` | Short URL created |
| `204` | Short URL deleted |
| `302`, `301`, `307`, `308` | Redirect response, controlled by `REDIRECT_STATUS` |
| `400` | Invalid JSON, URL, short code, expiration, or analytics date range |
| `413` | Request body exceeds `MAX_REQUEST_BODY_BYTES` |
| `404` | Missing short URL or unknown route |
| `405` | Unsupported HTTP method |
| `409` | Requested custom short code is already taken |
| `429` | Client request quota was exceeded |
| `500` | Unexpected server failure |
| `503` | Required dependency is unavailable or unique code generation retries were exhausted |
| `504` | A downstream operation exceeded the configured request deadline |

## Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `APP_ENV` | `development` | Application environment: `development`, `test`, or `production` |
| `HTTP_ADDR` | `:8080` | HTTP bind address |
| `BASE_URL` | `http://localhost:8080` | Externally reachable redirect base used to construct canonical `shortUrl` values |
| `CORS_ALLOWED_ORIGINS` | empty | Comma-separated HTTP(S) origins allowed to call the API from browsers |
| `TRUSTED_PROXY_CIDRS` | empty | Comma-separated proxy CIDRs allowed to supply `X-Forwarded-For` client addresses |
| `MONGODB_URI` | `mongodb://localhost:27017` | MongoDB connection URI |
| `MONGODB_DATABASE` | `url_shortener` | MongoDB database name |
| `MONGODB_URLS_COLLECTION` | `urls` | Collection containing short URLs |
| `MONGODB_USERS_COLLECTION` | `users` | Collection containing registered users |
| `MONGODB_SESSIONS_COLLECTION` | `sessions` | Collection containing hashed refresh sessions |
| `MONGODB_RATE_LIMITS_COLLECTION` | `rate_limits` | Collection containing distributed rate-limit counters |
| `MONGODB_ANALYTICS_COLLECTION` | `click_analytics` | Collection containing per-link UTC daily click aggregates |
| `REDIS_URL` | `redis://localhost:6379/0` | Redis connection URL; `rediss://` enables TLS |
| `REDIS_KEY_PREFIX` | `url-shortener` | Namespace prefix used for this service's Redis keys |
| `REDIS_CONNECT_TIMEOUT` | `5s` | Maximum duration allowed for the initial Redis connection check |
| `SHORT_CODE_LENGTH` | `7` | Generated Base62 short-code length, from 4 to 32 |
| `SHORT_CODE_MAX_RETRIES` | `5` | Maximum attempts after a unique-index collision |
| `REDIRECT_STATUS` | `302` | Redirect status: `301`, `302`, `307`, or `308` |
| `REDIRECT_CACHE_ENABLED` | `false` | Enables Redis-backed redirect caching |
| `REDIRECT_CACHE_TTL` | `10m` | Maximum lifetime of a cached redirect destination |
| `REDIRECT_CACHE_ACCESS_WORKERS` | `2` | Background workers that persist access counts from cache hits |
| `REDIRECT_CACHE_ACCESS_QUEUE_SIZE` | `1024` | Maximum cache-hit access events buffered in memory |
| `REDIRECT_CACHE_ACCESS_TIMEOUT` | `5s` | MongoDB deadline for each queued access update |
| `ANALYTICS_WORKERS` | `2` | Background workers that persist daily click analytics |
| `ANALYTICS_QUEUE_SIZE` | `4096` | Maximum click events buffered for asynchronous analytics writes |
| `ANALYTICS_WRITE_TIMEOUT` | `5s` | MongoDB deadline for each queued analytics update |
| `RATE_LIMIT_REQUESTS` | `60` | Maximum requests from one client in a rate-limit window; `0` disables limiting |
| `RATE_LIMIT_WINDOW` | `1m` | Fixed window used for request rate limiting |
| `AUTH_TOKEN_SECRET` | development-only value | HMAC secret; production requires a non-default value of at least 32 characters |
| `AUTH_TOKEN_ISSUER` | `url-shortener` | Access-token issuer claim |
| `AUTH_TOKEN_AUDIENCE` | `url-shortener-api` | Access-token audience claim |
| `AUTH_TOKEN_TTL` | `15m` | Access-token lifetime |
| `AUTH_REFRESH_TOKEN_TTL` | `720h` | Refresh-session lifetime |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |
| `LOG_FORMAT` | `text` | `text` or `json` |
| `REQUEST_TIMEOUT` | `10s` | HTTP request deadline, socket timeout, and MongoDB connection timeout |
| `MAX_REQUEST_BODY_BYTES` | `1048576` | Maximum request body size, from 1 byte to 10 MiB |
| `SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown deadline |

The frontend uses one server-only environment variable:

| Variable | Default | Purpose |
| --- | --- | --- |
| `API_BASE_URL` | `http://localhost:8080` | Internal API origin used by Next.js route handlers; Compose uses `http://api:8080` |

## Verification

Run the API checks:

```bash
cd apps/api
go test ./...
go vet ./...
```

Run MongoDB and Redis integration tests against explicitly configured local services:

```bash
make data-up
MONGODB_INTEGRATION_URI=mongodb://localhost:27017 \
REDIS_INTEGRATION_URL=redis://localhost:6379/0 \
make api-integration
```

Integration tests create isolated MongoDB databases and Redis key namespaces, exercise concurrent analytics aggregation and reporting, then remove their data after each run.

Run the frontend checks:

```bash
cd apps/web
npm run lint
npm run test
npm run build
```

Run browser tests after installing Playwright's Chromium build:

```bash
cd apps/web
npx playwright install chromium
npm run test:e2e
npm run test:e2e:full
```

The fast suite starts only Next.js. The full suite requires Docker and exercises the real Next.js BFF, Go API, and an isolated temporary MongoDB instance. Playwright failure reports are retained by CI for seven days.
