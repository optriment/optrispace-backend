#TOOLDIR = ./.bin
LINTBIN = ./.bin/golangci-lint

DB_URL	?=postgres://postgres:postgres@localhost:65432/optrwork?sslmode=disable
APP_URL	?=http://localhost:8080

.phony: all
all:
	echo "Specify target"

.phony: sqlc
sqlc:
	cd pkg/db/sqlc && make sqlc

.phony: migrate-up
migrate-up:
	cd pkg/db/db-migrations && make $@

.phony: migrate-drop
migrate-drop:
	cd pkg/db/db-migrations && make $@

.phony: docker-compose-up
docker-compose-up:
	cd ops/docker-compose-dev && docker-compose up -d

.phony: docker-compose-down
docker-compose-down:
	cd ops/docker-compose-dev && docker-compose down

.phony: run
run:
	go run . run --config ./testdata/dev.yaml

.phony: run-intest
run-intest:
	env DB_URL=postgres://postgres:postgres@localhost:65432/optrwork?sslmode=disable APP_URL=http://localhost:8080 go test ./intest/

$(LINTBIN):
	@echo "Getting $@"
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.45.2

.phony: lint
lint: $(LINTBIN)
	$(LINTBIN) run ./...
