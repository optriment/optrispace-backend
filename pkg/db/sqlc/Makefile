SQLCBIN = ./.bin/sqlc

.phony: all
all: sqlc

$(SQLCBIN):
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) go install github.com/kyleconroy/sqlc/cmd/sqlc@v1.13.0

.phony: sqlc
sqlc: $(SQLCBIN) sqlc.yaml $(./queries/*)
	$(SQLCBIN) generate
