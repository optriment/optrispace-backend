version		=v4.15.2
platform	?=linux

DB_URL	?=postgres://postgres:postgres@localhost:65432/optrwork?sslmode=disable
SCRDIR 	?=./scripts

MBIN	= ./.bin/migrate

.PHONY: all
all: migrate-up

#https://github.com/golang-migrate/migrate
#https://github.com/golang-migrate/migrate/tree/master/cmd/migrate

$(MBIN):
	@mkdir -p $(@D)
	curl -L https://github.com/golang-migrate/migrate/releases/download/$(version)/migrate.$(platform)-amd64.tar.gz | tar -C $(dir $(MBIN)) -xvz

.PHONY: migrate-up
migrate-up: $(MBIN)
	$(MBIN) -database $(DB_URL) -source file://$(SCRDIR) -verbose up

.PHONY: migrate-down
migrate-down: $(MBIN)
	$(MBIN) -database $(DB_URL) -source file://$(SCRDIR) -verbose down 1

.PHONY: migrate-drop
migrate-drop: $(MBIN)
	$(MBIN) -database $(DB_URL) -source file://$(SCRDIR) -verbose drop -f

.PHONY: migrate-new-script
migrate-new-script: $(MBIN) # Add new script. Specify title in form SCRIPT_TITLE=<title>
	$(MBIN) create -dir $(SCRDIR) -ext sql $(SCRIPT_TITLE)
