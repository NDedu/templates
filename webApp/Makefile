-include .env

## run: run the server
run:
	go run ./cmd/web

## build: build and run
build:
	go build -o /tmp/webapp ./cmd/web
	/tmp/webapp

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

## test: run all tests verbose with race detection and coverage
test:
	go test -v -race -cover -count=1 ./...

## test/run name=TestName [pkg=./path/to/pkg]: run a single test by name (default: all packages)
test/run:
	go test -v -race -count=1 -run ${name} $(or ${pkg},./...)

## test/cover: run all tests and open coverage in browser
test/cover:
	go test -v -race -coverprofile=/tmp/coverage.out -count=1 ./...
	go tool cover -html=/tmp/coverage.out

## test/cover/run name=TestName [pkg=./path/to/pkg]: run a single test with coverage and open in browser
test/cover/run:
	go test -v -race -coverprofile=/tmp/coverage.out -count=1 -run ${name} $(or ${pkg},./...)
	go tool cover -html=/tmp/coverage.out

## test/file file=path/to/file_test.go: run all tests in a single file
test/file:
	$(eval _pattern := $(shell grep -E '^func Test[a-zA-Z0-9_]+' ${file} | sed 's/[()]/ /g' | awk '{print $$2}' | paste -sd'|'))
	@test -n '${_pattern}' || (echo "no tests found in ${file}" && exit 1)
	go test -v -race -count=1 -run '^(${_pattern})$$' $(shell dirname ./${file})/

## test/cover/file file=path/to/file_test.go: run all tests in a single file with coverage and open in browser
test/cover/file:
	$(eval _pattern := $(shell grep -E '^func Test[a-zA-Z0-9_]+' ${file} | sed 's/[()]/ /g' | awk '{print $$2}' | paste -sd'|'))
	@test -n '${_pattern}' || (echo "no tests found in ${file}" && exit 1)
	go test -v -race -coverprofile=/tmp/coverage.out -count=1 -run '^(${_pattern})$$' $(shell dirname ./${file})/
	go tool cover -html=/tmp/coverage.out

.PHONY: run build test test/run test/cover test/cover/run test/file test/cover/file help
help:
	@grep -E '^## ' Makefile | sed 's/## //'
