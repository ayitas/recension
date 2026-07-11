COMPOSE := docker compose -f ops/compose.yaml
COMPOSE_ENV := --env-file .env

.DEFAULT_GOAL := help

.PHONY: help \
	env install \
	up down build rebuild restart logs ps clean \
	db minio deps \
	api web \
	example example-blobs \
	test check smoke smoke-hardening smoke-blobs

help:
	@echo "Recension — make targets"
	@echo ""
	@echo "  Full stack (Docker)"
	@echo "    make up          Build & start all services (api, web, postgres, minio, gateway)"
	@echo "    make down        Stop all services"
	@echo "    make build       Build images without starting"
	@echo "    make rebuild     Rebuild images from scratch & start"
	@echo "    make restart     Restart all services"
	@echo "    make logs        Tail logs (SVC=api|web|postgres|minio|gateway)"
	@echo "    make ps          Show running services"
	@echo "    make clean       Stop & remove containers + volumes"
	@echo ""
	@echo "  Local development"
	@echo "    make env         Copy .env.example → .env (if missing)"
	@echo "    make install     Install Go modules + npm deps"
	@echo "    make deps        Start postgres + minio only"
	@echo "    make api         Run API locally (needs deps)"
	@echo "    make web         Run web locally (Vite)"
	@echo ""
	@echo "  Examples / tests"
	@echo "    make example              Run minimal Go example"
	@echo "    make example-blobs        Run blobs example (REV=... BREAK=1)"
	@echo "    make test                 Go unit tests (pkg + sdk)"
	@echo "    make check                Svelte/TS typecheck"
	@echo "    make smoke                All smoke scripts"
	@echo ""
	@echo "  URLs (Docker up)"
	@echo "    Dashboard  http://localhost:3000"
	@echo "    API        http://localhost:8080"
	@echo "    MinIO      http://localhost:9001  (recension / recensionsecret)"

# --- setup ---

env:
	@test -f .env || cp .env.example .env
	@echo ".env ready"

install: env
	cd api && go mod download
	cd pkg && go mod download
	cd sdk/go && go mod download
	cd web && npm ci

# --- Docker full stack ---

up: env
	$(COMPOSE) $(COMPOSE_ENV) up -d --build

down:
	$(COMPOSE) down

build: env
	$(COMPOSE) $(COMPOSE_ENV) build

rebuild: env
	$(COMPOSE) $(COMPOSE_ENV) up -d --build --force-recreate

restart:
	$(COMPOSE) restart

logs:
	$(COMPOSE) logs -f $(SVC)

ps:
	$(COMPOSE) ps

clean:
	$(COMPOSE) down -v --remove-orphans

# --- infra only ---

db:
	$(COMPOSE) up -d postgres

minio:
	$(COMPOSE) up -d minio

deps: db minio

# --- local processes ---

api: deps env
	cd api && set -a && . ../.env && set +a && go run ./cmd/server

web:
	cd web && npm run dev

# --- examples ---

example:
	cd examples/go/minimal && \
	RECENSION_API_KEY=$${RECENSION_API_KEY:-dev-api-key} \
	RECENSION_API_URL=$${RECENSION_API_URL:-http://localhost:8080} \
	RECENSION_TEAM=$${RECENSION_TEAM:-acme} \
	go run . -revision $${REV:-v1.0}

example-blobs:
	cd examples/go/blobs && \
	RECENSION_API_KEY=$${RECENSION_API_KEY:-dev-api-key} \
	RECENSION_API_URL=$${RECENSION_API_URL:-http://localhost:8080} \
	RECENSION_TEAM=$${RECENSION_TEAM:-acme} \
	go run . -revision $${REV:-export-$$(date +%Y%m%d-%H%M%S)} $${BREAK:+-break}

# --- test / smoke ---

test:
	cd pkg && go test ./...
	cd sdk/go && go test ./...
	cd api && go test ./...

check:
	cd web && npm run check

smoke: smoke-hardening smoke-blobs

smoke-hardening:
	bash scripts/smoke-hardening.sh

smoke-blobs:
	bash scripts/smoke-blobs.sh
