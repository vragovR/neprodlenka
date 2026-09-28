.PHONY: install up down logs test lint build

install:
	go mod download
	npm ci

up:
	docker compose up --watch

down:
	docker compose down

logs:
	docker compose logs -f

test:
	go test ./cmd/... ./internal/...

lint:
	go vet ./cmd/... ./internal/...
	test -z "$$(gofmt -l cmd internal)"
	npm run lint

build:
	go build -o bin/neprodlenka ./cmd/neprodlenka
	npm run build
