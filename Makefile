.PHONY: install up down logs test lint build

GOLANGCI_LINT_VERSION := v2.15.0
GOLANGCI_LINT := $(CURDIR)/backend/bin/golangci-lint/$(GOLANGCI_LINT_VERSION)/golangci-lint

install: $(GOLANGCI_LINT)
	cd backend && go mod download
	cd frontend && npm ci
	cd frontend && npx playwright install chromium

up:
	docker compose up --watch

down:
	docker compose down

logs:
	docker compose logs -f

test:
	cd backend && go test ./cmd/... ./internal/...
	cd frontend && npm run test:unit -- --run
	cd frontend && npm run test:e2e

$(GOLANGCI_LINT):
	curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(dir $(GOLANGCI_LINT)) $(GOLANGCI_LINT_VERSION)

lint: $(GOLANGCI_LINT)
	cd backend && go fix ./...
	cd backend && go vet ./cmd/... ./internal/...
	cd backend && test -z "$$(gofmt -l cmd internal)"
	cd backend && $(GOLANGCI_LINT) run ./...
	cd frontend && npm run lint

build:
	cd backend && go build -o bin/neprodlenka ./cmd/neprodlenka
	cd frontend && npm run build
