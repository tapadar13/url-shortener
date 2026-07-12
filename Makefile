COMPOSE = docker compose -f deploy/docker-compose.yml

.PHONY: help api-run api-build api-test api-vet api-check api-image web-dev web-lint web-test web-build web-check mongo-up mongo-down stack-up stack-down

help:
	@printf '%s\n' \
		'api-run      Run the Go API with values from .env when present' \
		'api-build    Build the Go API binary into bin/' \
		'api-test     Run Go unit tests' \
		'api-vet      Run Go static analysis' \
		'api-check    Run Go tests and static analysis' \
		'api-image    Build the API container image' \
		'web-dev      Start the Next.js frontend' \
		'web-lint     Lint the frontend' \
		'web-test     Run frontend tests' \
		'web-build    Build the frontend' \
		'web-check    Run frontend lint, tests, and build' \
		'mongo-up     Start MongoDB for local API development' \
		'mongo-down   Stop the local MongoDB service' \
		'stack-up     Start the containerized API and MongoDB stack' \
		'stack-down   Stop the containerized API and MongoDB stack'

api-run:
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; cd apps/api && go run ./cmd/api

api-build:
	@mkdir -p bin
	@cd apps/api && go build -o ../../bin/url-shortener-api ./cmd/api

api-test:
	@cd apps/api && go test ./...

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

web-build:
	@cd apps/web && npm run build

web-check: web-lint web-test web-build

mongo-up:
	@$(COMPOSE) up -d mongodb

mongo-down:
	@$(COMPOSE) stop mongodb

stack-up:
	@$(COMPOSE) up --build

stack-down:
	@$(COMPOSE) down
