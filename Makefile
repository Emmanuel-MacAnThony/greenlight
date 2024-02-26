include .$(PWD)/.env

test_target:
	@echo ${GREENLIGHT_DB_DSN}

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans &&  [ $${ans:-N} = y ]

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api
	
## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${GREENLIGHT_DB_DSN}
	
## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm	
	@echo 'running migrations...'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...
	

.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

# GOOS=linux GOARCH=amd64 go build -ldflags='-s -X main.buildTime=${current_time}' -o=./bin/linux_amd64/api ./cmd/api
current_time = ${shell date --iso-8601=seconds}


.PHONY: build/api
build/api:
	@echo 'Building cmd/api..'
	go build -ldflags='-s -X main.buildTime=${current_time}' -o=./bin/api ./cmd/api
	

	

