# Simple Makefile for a Go project

# # Load env file helper
# ifneq (,$(wildcard .env.development))
#     include .env.development
#     export
# endif

APP_ENV ?= development
ENV_FILE := .env.$(APP_ENV)

ifneq (,$(wildcard $(ENV_FILE)))
    include $(ENV_FILE)
    export
endif

# Build the application
all: build test

build:
	@echo "Building..."
	
	
	@go build -o main cmd/api/main.go

# Run the application
runner:
	@go run cmd/api/main.go
	@echo "Running in development mode..."

run-dev:
	@APP_ENV=development go run cmd/api/main.go
	@echo "Running in production mode..."

run-prod:
	@APP_ENV=production go run cmd/api/main.go
	@echo "Running in production mode..."
# Create DB container
docker-run:
	@if docker compose --env-file .env.development up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose --env-file .env.development up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v
# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi
swagger-generate:
	@echo "Generating Swagger documentation..."
	swag init -g internal/router/router.go -o ./docs

setup:
	make docker-run
	make migrate-up


migrate-up:
	@echo "Running migrations up..."
	migrate -path db/migrations \
	-database "postgresql://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?sslmode=$(DB_SSLMODE)&search_path=$(DB_SCHEMA)" \
	up

migrate-down:
	@echo "Rolling back last migration..."
	migrate -path db/migrations \
	-database "postgresql://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?sslmode=$(DB_SSLMODE)&search_path=$(DB_SCHEMA)" \
	down 1

migrate-fix:
	@echo "Fixing dirty database..."
	migrate -path db/migrations \
	-database "postgresql://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?sslmode=$(DB_SSLMODE)&search_path=$(DB_SCHEMA)" \
	force 0



migrate-create:
	@echo "Creating migration: $(name)"
	migrate create -ext sql -dir db/migrations -seq $(name)


sqlc-generate:
	@echo "Generating sqlc code..."
	sqlc generate

sqlc-verify:
	@echo "Verifying sqlc queries..."
	sqlc vet


.PHONY: all build run test clean watch docker-run docker-down itest
