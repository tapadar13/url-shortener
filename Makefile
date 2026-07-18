COMPOSE = docker compose -f deploy/docker-compose.yml

.PHONY: help api-run api-build api-test api-integration api-vet api-check api-image web-dev web-lint web-test web-e2e web-e2e-full web-build web-check web-image mongo-up mongo-down redis-up redis-down data-up data-down stack-up stack-down

help:
	@printf '%s\n' \
		'api-run      Run the Go API with values from .env when present' \
		'api-build    Build the Go API binary into bin/' \
		'api-test     Run Go unit tests' \
		'api-integration Run MongoDB and Redis integration tests' \
		'api-vet      Run Go static analysis' \
		'api-check    Run Go tests and static analysis' \
		'api-image    Build the API container image' \
		'web-dev      Start the Next.js frontend' \
		'web-lint     Lint the frontend' \
		'web-test     Run frontend tests' \
		'web-e2e      Run fast frontend browser tests' \
		'web-e2e-full Run full-stack browser tests' \
		'web-build    Build the frontend' \
		'web-check    Run frontend lint, tests, and build' \
		'web-image    Build the frontend container image' \
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

web-check: web-lint web-test web-build

web-image:
	@docker build --tag url-shortener-web:local apps/web

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
