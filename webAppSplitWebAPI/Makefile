-include .env

## start: run both web and api
run/all:
	@sh -c 'lsof -ti:${WEB_PORT} -ti:${API_PORT} 2>/dev/null | xargs kill 2>/dev/null || true'
	trap 'kill 0' SIGINT; go run ./cmd/web & go run ./cmd/api & wait

## startBuild: build and run both
build/all:
	@sh -c 'lsof -ti:${WEB_PORT} -ti:${API_PORT} 2>/dev/null | xargs kill 2>/dev/null || true'
	go build -o /tmp/web ./cmd/web
	go build -o /tmp/api ./cmd/api
	trap 'kill 0' SIGINT; /tmp/web & /tmp/api & wait

## run/web: run web
run/web:
	go run ./cmd/web

## run/api: run api
run/api:
	go run ./cmd/api

## db/migrations/new name=$1: create a new migration
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up migrations
db/migrations/up:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${DB_DSN} up

## db/migrations/down: revert the last migration
db/migrations/down:
	@echo 'Reverting database...'
	migrate -path ./migrations -database ${DB_DSN} down 1

.PHONY: start startBuild runWeb runApi help
help:
	@grep -E '^## ' Makefile | sed 's/## //'
