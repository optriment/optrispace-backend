.DEFAULT_GOAL := help

LINTBIN = ./.bin/golangci-lint
STRINGER = ./.bin/stringer
SHELL := bash

SWAGBIN = ./.bin/swag

DB_URL	?=postgres://postgres:postgres@localhost:65432/optrwork?sslmode=disable
APP_URL	?=http://localhost:8080

help: # Show this help
	@egrep -h '\s#\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?# "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: sqlc
sqlc:
	cd pkg/db/sqlc && make sqlc

.PHONY: migrate-up
migrate-up: # Apply database migrations
	cd pkg/db/db-migrations && make $@

.PHONY: migrate-down
migrate-down: # Downgrade DB one step down
	cd pkg/db/db-migrations && make $@

.PHONY: migrate-drop
migrate-drop: # Drop database migrations
	cd pkg/db/db-migrations && make $@

.PHONY: migrate-new-script
migrate-new-script: # Add a new script. Specify title in form SCRIPT_TITLE=<title>
	cd pkg/db/db-migrations && make $@


.PHONY: docker-compose-build
docker-compose-build: # Build Docker image
	cd ops/docker-compose-dev && docker-compose build --no-cache

.PHONY: docker-compose-up
docker-compose-up: docker-compose-build # Start dockerized application with database in background
	cd ops/docker-compose-dev && docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down: # Stop application
	cd ops/docker-compose-dev && docker-compose down

.PHONY: run
run: # Run application in development mode
	source ./testdata/.env && go run . --config ./testdata/dev.yaml run

.PHONY: seed
seed: # Populates database with sample data
	@env DB_URL=$(DB_URL) go run ./testdata/seed.go

$(LINTBIN):
	@echo "Getting $@"
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.46.2

.PHONY: lint
lint: $(LINTBIN) # Run linter
	$(LINTBIN) run ./...

# NOTE: All lines below used for testing purpose only
TEST_DB_URL ?=postgres://postgres:postgres@localhost:65433/optrwork_test?sslmode=disable

.PHONY: test-start-database
test-start-database: # Start dockerized database for testing purpose
	cd ops/docker-compose-test && docker-compose up postgres_test

.PHONY: test-start-backend
test-start-backend: # Run application in test mode
	go run . --config ./testdata/test.yaml run

.PHONY: test-integration
test-integration: test-migrate-drop test-migrate-up # Run integration tests
	env DB_URL="${TEST_DB_URL}" APP_URL=http://localhost:8081 go test -count=1 -v ./intest/ -test.failfast

.PHONY: test-migrate-up
test-migrate-up: # Apply migrations in test database
	cd pkg/db/db-migrations && env DB_URL="${TEST_DB_URL}" make migrate-up

.PHONY: test-migrate-drop
test-migrate-drop: # Drop migrations in test database
	cd pkg/db/db-migrations && env DB_URL="${TEST_DB_URL}" make migrate-drop

$(STRINGER): # Installs stringer tool
	@echo "Getting $@"
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) go install golang.org/x/tools/cmd/stringer@v0.1.10


.PHONY: generate
generate: $(STRINGER) # code generate with go tools
	go generate ./...

$(SWAGBIN): # Installs swag tool (Swagger 2.0 description generator)
	@echo "Getting $@"
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) go install github.com/swaggo/swag/cmd/swag@v1.8.4

.PHONY: swag-init
swag-init: $(SWAGBIN) # (re)generate swagger specification with swag
	$(SWAGBIN) init --dir ./ --generalInfo ./pkg/controller/api.go --parseDependency --output ./pkg/docs

.PHONY: swag-fmt
swag-fmt: $(SWAGBIN) # format swag comment annotations 
	$(SWAGBIN) fmt --dir ./pkg/controller --generalInfo api.go 
