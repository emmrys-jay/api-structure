# Makefile for Dev Environment
# TODO: Find a way to parse configurations in yml

# Path to the YAML configuration file
CONFIG_FILE=config-sample.yml

# Variables extracted from YAML using yq
APP_NAME := $(shell yq '.app.name' $(CONFIG_FILE))
APP_VERSION := $(shell yq '.app.env' $(CONFIG_FILE))

DB_PROTOCOL := $(shell yq '.database.protocol' $(CONFIG_FILE))
DB_HOST := $(shell yq '.database.host' $(CONFIG_FILE))
DB_PORT := $(shell yq '.database.port' $(CONFIG_FILE))
DB_NAME := $(shell yq '.database.name' $(CONFIG_FILE))
DB_USER := $(shell yq '.database.user' $(CONFIG_FILE))
DB_PASSWORD := $(shell yq '.database.password' $(CONFIG_FILE))

DSN := $(DB_PROTOCOL)://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

ARG ?= 

.PHONY: default install service-up service-down db-docs db-create db-drop db-cli \
        migrate-up migrate-down redis-cli dev lint build start swag test sqlc-gen

default: install ## Getting started

install: ## Install dependencies
	go mod download
	go install github.com/air-verse/air@latest
	brew install yq
# go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
# go install github.com/swaggo/swag/cmd/swag@latest
# go install go.uber.org/mock/mockgen@latest

service-build: ## Rebuild image and containers
	DB_PASSWORD=$(DB_PASSWORD) DB_NAME=$(DB_NAME) DB_USER=$(DB_USER) docker-compose up --build -d

service-up: ## Start docker services
	DB_PASSWORD=$(DB_PASSWORD) DB_NAME=$(DB_NAME) DB_USER=$(DB_USER) docker-compose up -d

service-down: ## Stop services
	docker-compose down

service-down-add: ## Stop services, volumes and networks
	docker-compose down -v

# db-docs: ## Generate database documentation from DBML file
# 	dbdocs build $(DBML_FILE)

# db-create: ## Create database if not exists
# 	docker exec -it api-structure_postgres sh -c "psql -U $(DB_USER) -c 'SELECT 1' -d $(DB_NAME) &>/dev/null || psql -U $(DB_USER) -c 'CREATE DATABASE $(DB_NAME);'"

# db-drop: ## Drop database
# 	docker exec -it api-structure_postgres sh -c "psql -U $(DB_USER) -c 'DROP DATABASE $(DB_NAME);'"

# db-cli: ## Connect to database using command line interface
# 	docker exec -it api-structure_postgres sh -c "psql -U $(DB_USER) -d $(DB_NAME)"

create-migration:
	migrate create -ext sql -dir internal/adapter/storage/postgres/migrations -seq $(NAME)

migrate-up: ## Run database migrations
	migrate -path ./internal/adapter/storage/postgres/migrations -database $(DSN) -verbose up $(ARG)

migrate-down: ## Rollback database migrations
	migrate -path ./internal/adapter/storage/postgres/migrations -database $(DSN) -verbose down $(ARG)

redis-cli: ## Connect to redis using command line interface
	docker exec -it api-structure_redis redis-cli

dev: ## Start development server
	air

lint: ## Run linter
	golangci-lint run ./...

print_dsn:
	echo $(DSN)

build: ## Build binary
	go build -o ./bin/$(APP_NAME) ./cmd/http/main.go

start: build ## Start binary
	./bin/$(APP_NAME)

swag: ## Generate swagger documentation
	swag fmt
	swag init -g ./cmd/http/main.go -o ./docs --parseInternal true

test: ## Run tests
	go test -v ./... -race -cover -timeout 30s -count 1 -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html