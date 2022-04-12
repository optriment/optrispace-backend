BINDIR = ./.bin
SQLCBIN = $(BINDIR)/sqlc

.phony: all
all: sqlc

$(SQLCBIN):
	mkdir -p $(BINDIR)
	GOBIN=$(abspath $(BINDIR)) go install github.com/kyleconroy/sqlc/cmd/sqlc@latest

.phony: sqlc
sqlc: $(SQLCBIN) sqlc.yaml $(./queries/*)
	$(SQLCBIN) generate
