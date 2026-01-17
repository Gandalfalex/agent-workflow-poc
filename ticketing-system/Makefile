SHELL := /bin/bash

OPENAPI_SPEC := $(CURDIR)/openapi.yaml
DATABASE_URL ?= postgres://ticketing:ticketing@localhost:5433/ticketing?sslmode=disable

.PHONY: generate generate-backend generate-frontend migrate

generate: generate-backend generate-frontend

# Generate Go server/types from OpenAPI
generate-backend:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.1 \
		--config $(CURDIR)/backend/oapi-codegen.yaml \
		$(OPENAPI_SPEC)

# Generate TypeScript schema from OpenAPI
generate-frontend:
	cd frontend && npx --yes openapi-typescript $(OPENAPI_SPEC) --output src/lib/api.schema.ts

# Run database migrations
migrate:
	cd backend && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate
