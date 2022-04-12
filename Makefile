
.phony: all
all:
	echo "Specify target"

.phony: sqlc
sqlc:
	cd pkg/internal/db/sqlc && make sqlc

.phony: migrate-up
migrate-up:
	cd pkg/internal/db/db-migrations && make $@

.phony: migrate-drop
migrate-drop:
	cd pkg/internal/db/db-migrations && make $@

.phony: docker-compose-up
docker-compose-up:
	cd testdata/docker-compose-dev && docker-compose up -d

.phony: docker-compose-down
docker-compose-down:
	cd testdata/docker-compose-dev && docker-compose down

.phony: go-run
run:
	go run . run --config ./testdata/dev.yaml

