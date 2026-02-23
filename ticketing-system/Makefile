SHELL := /bin/bash

OPENAPI_SPEC := $(CURDIR)/openapi.yaml
DATABASE_URL ?= postgres://ticketing:ticketing@localhost:5433/ticketing?sslmode=disable
DESIGN_MIGRATION_DIR ?= $(CURDIR)/../design-penpot-migration

.PHONY: generate generate-backend generate-frontend migrate e2e-contract e2e-deps e2e-playwright-install e2e-frontend-build e2e design-migration design-migration-artifacts design-migration-starter

generate: generate-backend generate-frontend

# Generate Go server/types from OpenAPI
generate-backend:
	cd backend && go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--package httpapi \
		--generate chi-server,models \
		-o internal/httpapi/generated.gen.go \
		../openapi.yaml

# Generate TypeScript schema from OpenAPI
generate-frontend:
	cd frontend && npm run -s generate:api

# Run database migrations
migrate:
	cd backend && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate

# Generate backend E2E frontend contract (routes + selector keys).
e2e-contract:
	node scripts/generate-e2e-contract.mjs

# Add backend E2E dependencies to go.mod/go.sum.
e2e-deps:
	cd backend && go get \
		github.com/playwright-community/playwright-go@latest \
		github.com/testcontainers/testcontainers-go@latest \
		github.com/testcontainers/testcontainers-go/modules/postgres@latest
	cd backend && go mod tidy

# Install Playwright browser runtime.
e2e-playwright-install:
	cd backend && go run github.com/playwright-community/playwright-go/cmd/playwright@latest install chromium

# Build frontend assets used by backend-served E2E tests.
e2e-frontend-build:
	cd frontend && npm ci && npm run build

# Run backend E2E browser tests.
# -parallel 16 lets Go schedule up to 16 concurrent t.Parallel tests;
# the E2E_MAX_PARALLEL semaphore in the harness (default 4) is the real throttle.
# -timeout is set generously because a single shared container is started in TestMain.
e2e:
	$(MAKE) e2e-contract
	@set -o pipefail; \
	log_file=$$(mktemp -t ticketing-e2e-XXXX.log); \
	if cd backend && go test -tags=e2e -parallel 16 -timeout 20m ./e2e -v 2>&1 | tee "$$log_file"; then \
		rm -f "$$log_file"; \
	else \
		echo ""; \
		echo "Failed tests summary:"; \
		grep -E '^--- FAIL: ' "$$log_file" | sed -E 's/^--- FAIL: ([^ ]+).*/- \1/' | sort -u || true; \
		echo "Raw log: $$log_file"; \
		exit 1; \
	fi

# Run full frontend->Penpot migration extraction pipeline (tokens, inventory, E2E artifact harvest, priorities).
design-migration:
	cd "$(DESIGN_MIGRATION_DIR)" && npm run -s extract:all

# Refresh only E2E artifact snapshots + prioritized screen outputs.
design-migration-artifacts:
	cd "$(DESIGN_MIGRATION_DIR)" && npm run -s extract:e2e-artifacts && npm run -s extract:priorities

# Rebuild Penpot starter package (page-bucketed screenshots + seed files).
design-migration-starter:
	cd "$(DESIGN_MIGRATION_DIR)" && npm run -s extract:starter
