# Include variables from the .env file
include .env

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run: run the cmd/undr application
.PHONY: run
run:
	go run ./cmd/undr

## test: run all tests
.PHONY: test
test:
	go test ./... -v


## db: connect to the database using psql
.PHONY: db
db:
	psql -U postgres ${DSN}

## migrate/new name=$1: create a new database migration
.PHONY: migrate/new
migrate/new:
	@echo  'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## migrate/seed/new name=$1: create a new database seed migration
.PHONY: migrate/seed/new
migrate/seed/new:
	@echo  'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations/seed ${name}

## migrate/up: apply all up database migrations
.PHONY: migrate/up
migrate/up: confirm
	@echo 'Running up migrations...'
	migrate -source file://./migrations/ -database ${DSN} up

## migrate/seed/up: apply all up database migrations
.PHONY: migrate/seed/up
migrate/seed/up: confirm
	@echo 'Running up seed migrations...'
	migrate -source file://./migrations/seed/ -database ${DSN} up

## migrate/down: apply all down database migrations
.PHONY: migrate/down
migrate/down: confirm
	@echo 'Running down migrations...'
	migrate -source file://./migrations/ -database ${DSN} down

## migrate/seed/down: apply all down database seed migrations
.PHONY: migrate/seed/down
migrate/seed/down: confirm
	@echo 'Running down seed migrations...'
	migrate -source file://./migrations/seed/ -database ${DSN} down