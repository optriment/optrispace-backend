version		=v4.15.1
platform	?=linux

DB_URL	?=postgres://postgres:postgres@localhost:65432/optrwork?sslmode=disable
SCRDIR 	?=file://./scripts

MBIN	= ./.bin/migrate

.phony: all
all: migrate-up

#https://github.com/golang-migrate/migrate
#https://github.com/golang-migrate/migrate/tree/master/cmd/migrate

$(MBIN):
	@mkdir -p $(@D)
	curl -L https://github.com/golang-migrate/migrate/releases/download/$(version)/migrate.$(platform)-amd64.tar.gz | tar -C $(dir $(MBIN)) -xvz

.phony: migrate-up
migrate-up: $(MBIN)
	$(MBIN) -database $(DB_URL) -source $(SCRDIR) -verbose up

.phony: migrate-down
migrate-down: $(MBIN)
	$(MBIN) -database $(DB_URL) -source $(SCRDIR) -verbose down 1

.phony: migrate-drop
migrate-drop: $(MBIN)
	$(MBIN) -database $(DB_URL) -source $(SCRDIR) -verbose drop -f

