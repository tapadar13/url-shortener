COMPOSE = docker compose -f deploy/docker-compose.yml
GOVULNCHECK_VERSION = v1.6.0
GO_VERSION = $(shell awk '/^go / { print $$2 }' apps/api/go.mod)
GITLEAKS_IMAGE = ghcr.io/gitleaks/gitleaks:v8.30.1

.PHONY: help api-run api-build api-test api-integration api-vet api-audit api-check api-image web-dev web-lint web-test web-e2e web-e2e-full web-build web-audit web-check web-image secrets-audit security-check mongo-up mongo-down redis-up redis-down data-up data-down stack-up stack-down
.PHONY: load-smoke load-management load-redirects load-down

help:
	@printf '%s\n' \
		'api-run      Run the Go API with values from .env when present' \
		'api-build    Build the Go API binary into bin/' \
		'api-test     Run Go unit tests' \
		'api-integration Run MongoDB and Redis integration tests' \
		'api-vet      Run Go static analysis' \
		'api-audit    Scan the Go API for reachable vulnerabilities' \
		'api-check    Run Go tests and static analysis' \
		'api-image    Build the API container image' \
		'web-dev      Start the Next.js frontend' \
		'web-lint     Lint the frontend' \
		'web-test     Run frontend tests' \
		'web-e2e      Run fast frontend browser tests' \
		'web-e2e-full Run full-stack browser tests' \
		'web-build    Build the frontend' \
		'web-audit    Scan frontend dependencies for high-severity vulnerabilities' \
		'web-check    Run frontend lint, tests, and build' \
		'web-image    Build the frontend container image' \
		'secrets-audit Scan Git history for committed secrets' \
		'security-check Run vulnerability and secret scans' \
		'load-smoke   Run the load-test probe smoke scenario' \
		'load-management Run the authenticated management load scenario' \
		'load-redirects Run the redirect throughput load scenario' \
		'load-down    Remove the disposable load-test stack' \
		'mongo-up     Start MongoDB for local API development' \
		'mongo-down   Stop the local MongoDB service' \
		'redis-up     Start Redis for local API development' \
		'redis-down   Stop the local Redis service' \
		'data-up      Start MongoDB and Redis for local API development' \
		'data-down    Stop the local MongoDB and Redis services' \
		'stack-up     Start the containerized API, MongoDB, and Redis stack' \
		'stack-down   Stop the complete containerized stack'

api-run:
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; cd apps/api && go run ./cmd/api

api-build:
	@mkdir -p bin
	@cd apps/api && go build -o ../../bin/url-shortener-api ./cmd/api

api-test:
	@cd apps/api && go test ./...

api-integration:
	@cd apps/api && go test -tags=integration ./integration

api-vet:
	@cd apps/api && go vet ./...

api-audit:
	@cd apps/api && GOTOOLCHAIN=go$(GO_VERSION) go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

api-check: api-test api-vet

api-image:
	@docker build --tag url-shortener-api:local apps/api

web-dev:
	@cd apps/web && npm run dev

web-lint:
	@cd apps/web && npm run lint

web-test:
	@cd apps/web && npm run test

web-e2e:
	@cd apps/web && npm run test:e2e

web-e2e-full:
	@cd apps/web && npm run test:e2e:full

web-build:
	@cd apps/web && npm run build

web-audit:
	@cd apps/web && npm audit --audit-level=high

web-check: web-lint web-test web-build

web-image:
	@docker build --tag url-shortener-web:local apps/web

secrets-audit:
	@docker run --rm --volume "$(CURDIR):/repo:ro" $(GITLEAKS_IMAGE) git --redact --verbose /repo

security-check: api-audit web-audit secrets-audit

load-smoke:
	@./tests/load/run.sh smoke

load-management:
	@./tests/load/run.sh management

load-redirects:
	@./tests/load/run.sh redirects

load-down:
	@docker compose -f deploy/docker-compose.load.yml --profile load down --volumes --remove-orphans

mongo-up:
	@$(COMPOSE) up -d mongodb

mongo-down:
	@$(COMPOSE) stop mongodb

redis-up:
	@$(COMPOSE) up -d redis

redis-down:
	@$(COMPOSE) stop redis

data-up:
	@$(COMPOSE) up -d mongodb redis

data-down:
	@$(COMPOSE) stop mongodb redis

stack-up:
	@$(COMPOSE) up --build

stack-down:
	@$(COMPOSE) down
