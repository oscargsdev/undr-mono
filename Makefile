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

## run/undr: run the cmd/undr application
.PHONY: run/undr
run/undr:
	go run ./cmd/undr

## db/psql: connect to the database using psql
.PHONY: db/sql
db/psql:
	psql -U postgres ${DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo  'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/seed/new name=$1: create a new database seed migration
.PHONY: db/migrations/seed/new
db/migrations/seed/new:
	@echo  'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations/seed ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -source file://./migrations/ -database ${DSN} up

## db/migrations/seed/up: apply all up database migrations
.PHONY: db/migrations/seed/up
db/migrations/seed/up: confirm
	@echo 'Running up seed migrations...'
	migrate -source file://./migrations/seed/ -database ${DSN} up

## db/migrations/down: apply all down database migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	@echo 'Running down migrations...'
	migrate -source file://./migrations/ -database ${DSN} down

## db/migrations/seed/down: apply all down database seed migrations
.PHONY: db/migrations/seed/down
db/migrations/seed/down: confirm
	@echo 'Running down seed migrations...'
	migrate -source file://./migrations/seed/ -database ${DSN} down