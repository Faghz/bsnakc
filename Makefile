.PHONY: help run build test migrate-up migrate-down migrate-create swagger clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the application
	@go run cmd/server/main.go

dev:  ## Run the application with hot reload
	@go run github.com/air-verse/air@latest -c ./configs/.air.toml

build: ## Build the application
	@echo "Building..."
	@go build -o bin/server cmd/server/main.go
	@echo "Build complete: bin/server"

test: ## Run tests
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

test-coverage: test ## Run tests with coverage report
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report: coverage.html"

migrate-up: ## Run database migrations
	@docker run --rm --network host \
		-v "$(PWD)/db:/db" \
		-e DATABASE_URL="postgres://postgres:postgres@localhost:5432/marshal?sslmode=disable" \
		amacneil/dbmate:latest up

migrate-down: ## Rollback database migrations
	@docker run --rm --network host \
		-v "$(PWD)/db:/db" \
		-e DATABASE_URL="postgres://postgres:postgres@localhost:5432/marshal?sslmode=disable" \
		amacneil/dbmate:latest down

migrate-create: ## Create a new migration (usage: make migrate-create NAME=your_migration_name)
	@read -p "migration name (do not use space): " NAME \
	@docker run --rm \
		-v "$(PWD)/db:/db" \
		amacneil/dbmate:latest new $$(NAME)

migrate-status:
	@docker run --rm --network host \
		-v "$(PWD)/db:/db" \
		-e DATABASE_URL="postgres://postgres:postgres@localhost:5432/marshal?sslmode=disable" \
		amacneil/dbmate:latest status

gen-docs: ## Generate OpenAPI documentation
	@echo "Generating OpenAPI docs..."
	@go run github.com/swaggo/swag/cmd/swag@latest init -g ../cmd/server/main.go -d ./internal -o docs
	@echo "OpenAPI docs generated in docs/"

docker-up: ## Start Docker containers
	@docker-compose up -d
	@echo "Docker containers started"

docker-down: ## Stop Docker containers
	@docker-compose down
	@echo "Docker containers stopped"

docker-logs: ## Show Docker container logs
	@docker-compose logs -f

install-tools: ## Install development tools
	@echo "All tools run automatically via go run or Docker"
	@echo "No manual installation required!"
	@echo "  - swag: runs via go run"
	@echo "  - air: runs via go run"
	@echo "  - dbmate: runs via Docker"

deps: ## Download dependencies
	@go mod download
	@go mod tidy

clean: ## Clean build artifacts
	@rm -rf bin/ coverage.txt coverage.html tmp/ build-errors.log

dbmate-help: ## Show dbmate help
	@docker run --rm amacneil/dbmate:latest --help

.DEFAULT_GOAL := help
