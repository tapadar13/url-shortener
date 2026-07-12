# URL Shortener

A production-minded URL shortening service with a Go + MongoDB API and a Next.js marketing frontend.

The API creates short links, manages destinations, returns link statistics, and redirects visitors while atomically recording access counts.

## Project Layout

```text
apps/
  api/    Go HTTP API and MongoDB persistence
  web/    Next.js marketing frontend
deploy/
  docker-compose.yml    Local MongoDB service
```

## Current Features

- Collision-safe Base62 short-code generation backed by a unique MongoDB index
- URL creation, retrieval, update, deletion, and statistics endpoints
- Optional expiry timestamps with immediate expiry-aware reads and MongoDB TTL cleanup
- Configurable short-link redirects with atomic access counting
- Strict URL and short-code validation
- Consistent JSON error responses
- Health and MongoDB-backed readiness probes
- Structured request logs, request correlation IDs, and panic recovery
- Graceful API shutdown and MongoDB connection lifecycle management

## Prerequisites

- Go 1.23.5 or newer
- Docker and Docker Compose
- Node.js and npm for the frontend

## Local Development

Create a local environment file:

```bash
cp .env.example .env
```

Start only MongoDB for local Go API development:

```bash
docker compose -f deploy/docker-compose.yml up -d mongodb
```

Run the API:

```bash
cd apps/api
set -a
source ../../.env
set +a
go run ./cmd/api
```

The API listens on `http://localhost:8080` by default.

Run the complete containerized API stack:

```bash
docker compose -f deploy/docker-compose.yml up --build
```

Run the frontend in a second terminal:

```bash
cd apps/web
npm install
npm run dev
```

## Common Commands

Run `make help` from the repository root to list local development commands.

```bash
make api-check
make web-check
make mongo-up
make stack-up
```

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

### Create a Short URL

```http
POST /shorten
Content-Type: application/json
```

```json
{
  "url": "https://example.com/articles/123",
  "expiresAt": "2026-08-12T08:00:00Z"
}
```

Successful response: `201 Created`

```json
{
  "id": "507f1f77bcf86cd799439011",
  "url": "https://example.com/articles/123",
  "shortCode": "AbC1234",
  "createdAt": "2026-07-12T08:00:00Z",
  "updatedAt": "2026-07-12T08:00:00Z",
  "expiresAt": "2026-08-12T08:00:00Z"
}
```

```bash
curl -i http://localhost:8080/shorten \
  -X POST \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/articles/123","expiresAt":"2026-08-12T08:00:00Z"}'
```

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

### Redirect from a Short URL

```http
GET /{shortCode}
```

The API responds with the configured redirect status (`302` by default), sets `Location` to the original destination, and atomically increments `accessCount` before returning the redirect.

```bash
curl -i http://localhost:8080/AbC1234
```

### Probes

```http
GET /healthz
GET /readyz
```

- `/healthz` confirms that the HTTP process is running.
- `/readyz` additionally confirms that MongoDB can be reached. It returns `503 Service Unavailable` with `not_ready` when the dependency is unavailable.

```bash
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/readyz
```

### Status Codes

| Status | Meaning |
| --- | --- |
| `200` | Successful retrieval, update, or statistics response |
| `201` | Short URL created |
| `204` | Short URL deleted |
| `302`, `301`, `307`, `308` | Redirect response, controlled by `REDIRECT_STATUS` |
| `400` | Invalid JSON, URL, short code, or expiration |
| `404` | Missing short URL or unknown route |
| `405` | Unsupported HTTP method |
| `503` | Service dependency is not ready or unique code generation retries were exhausted |
| `504` | A downstream operation exceeded the configured request deadline |
| `500` | Unexpected server failure |

## Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `APP_ENV` | `development` | Application environment: `development`, `test`, or `production` |
| `HTTP_ADDR` | `:8080` | HTTP bind address |
| `BASE_URL` | `http://localhost:8080` | Validated public base URL reserved for future generated-link presentation |
| `CORS_ALLOWED_ORIGINS` | empty | Comma-separated HTTP(S) origins allowed to call the API from browsers |
| `MONGODB_URI` | `mongodb://localhost:27017` | MongoDB connection URI |
| `MONGODB_DATABASE` | `url_shortener` | MongoDB database name |
| `MONGODB_URLS_COLLECTION` | `urls` | Collection containing short URLs |
| `SHORT_CODE_LENGTH` | `7` | Generated Base62 short-code length, from 4 to 32 |
| `SHORT_CODE_MAX_RETRIES` | `5` | Maximum attempts after a unique-index collision |
| `REDIRECT_STATUS` | `302` | Redirect status: `301`, `302`, `307`, or `308` |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |
| `LOG_FORMAT` | `text` | `text` or `json` |
| `REQUEST_TIMEOUT` | `10s` | HTTP request deadline, socket timeout, and MongoDB connection timeout |
| `SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown deadline |

## Verification

Run the API checks:

```bash
cd apps/api
go test ./...
go vet ./...
```

Run MongoDB integration tests against an explicitly configured local MongoDB instance:

```bash
make mongo-up
MONGODB_INTEGRATION_URI=mongodb://localhost:27017 make api-integration
```

Integration tests create a unique temporary database and remove it after each run.

Run the frontend checks:

```bash
cd apps/web
npm run lint
npm run build
```
