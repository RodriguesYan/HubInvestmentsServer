# HubInvestments Makefile
# Helper commands for development and testing

.PHONY: help test test-verbose coverage coverage-html coverage-func coverage-open coverage-open-force clean

# Help target that lists all available commands
help: ## Show this help message
	@echo "📋 Available commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "🗄️  Database Migration Commands:"
	@echo "   migrate-help            Show detailed migration usage"
	@echo "   migrate-up              Run all pending migrations"
	@echo "   migrate-version         Show current migration version"
	@echo ""
	@echo "📖 Quick starts:"
	@echo "   make coverage-open      Generate and view test coverage"
	@echo "   make check              Run all quality checks"
	@echo "   make migrate-help       Show migration commands"

# Testing commands
test: ## Run all tests
	go test ./...

test-verbose: ## Run all tests with verbose output
	go test -v ./...

test-race: ## Run tests with race detection
	go test -race ./...

# Coverage commands
coverage: ## Run tests and show basic coverage
	go test -cover ./...

coverage-profile: ## Generate coverage profile
	go test -coverprofile=coverage.out ./...

coverage-profile-force: ## Generate coverage profile even if tests fail
	-go test -coverprofile=coverage.out ./...

coverage-func: coverage-profile ## Show function-level coverage details
	go tool cover -func=coverage.out

coverage-func-sorted: coverage-profile ## Show function-level coverage sorted by percentage
	go tool cover -func=coverage.out | sort -k3 -nr

coverage-html: coverage-profile ## Generate HTML coverage report
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

coverage-html-force: coverage-profile-force ## Generate HTML coverage report even if tests fail
	@if [ -f coverage.out ]; then \
		go tool cover -html=coverage.out -o coverage.html; \
		echo "Coverage report generated: coverage.html"; \
	else \
		echo "No coverage.out file found. Run tests first."; \
		exit 1; \
	fi

coverage-open: coverage-html ## Generate and open HTML coverage report
	@echo "Opening coverage report in browser..."
	@if command -v open >/dev/null 2>&1; then \
		open coverage.html; \
	elif command -v xdg-open >/dev/null 2>&1; then \
		xdg-open coverage.html; \
	else \
		echo "Please open coverage.html manually in your browser"; \
	fi

coverage-open-force: coverage-html-force ## Generate and open HTML coverage report even if tests fail
	@echo "Opening coverage report in browser..."
	@if command -v open >/dev/null 2>&1; then \
		open coverage.html; \
	elif command -v xdg-open >/dev/null 2>&1; then \
		xdg-open coverage.html; \
	else \
		echo "Please open coverage.html manually in your browser"; \
	fi

coverage-summary: coverage-profile ## Show coverage summary with key metrics
	@echo "=== COVERAGE SUMMARY ==="
	@go tool cover -func=coverage.out | tail -1
	@echo ""
	@echo "=== TOP COVERED FUNCTIONS ==="
	@go tool cover -func=coverage.out | sort -k3 -nr | head -10
	@echo ""
	@echo "=== FUNCTIONS NEEDING TESTS (0% coverage) ==="
	@go tool cover -func=coverage.out | grep "0.0%" | head -10

coverage-summary-force: coverage-profile-force ## Show coverage summary even if tests fail
	@if [ -f coverage.out ]; then \
		echo "=== COVERAGE SUMMARY ==="; \
		go tool cover -func=coverage.out | tail -1; \
		echo ""; \
		echo "=== TOP COVERED FUNCTIONS ==="; \
		go tool cover -func=coverage.out | sort -k3 -nr | head -10; \
		echo ""; \
		echo "=== FUNCTIONS NEEDING TESTS (0% coverage) ==="; \
		go tool cover -func=coverage.out | grep "0.0%" | head -10; \
	else \
		echo "No coverage.out file found. Run tests first."; \
		exit 1; \
	fi

# Build commands
build: ## Build the application
	go build -o bin/hubinvestments .

run: ## Run the application
	go run .

# Development commands
fmt: ## Format Go code
	go fmt ./...

lint: ## Run go vet
	go vet ./...

tidy: ## Tidy up module dependencies
	go mod tidy

# Clean up
clean: ## Clean up generated files
	rm -f coverage.out coverage.html
	rm -rf bin/

# All-in-one development check
check: fmt lint test coverage-summary ## Run format, lint, tests, and coverage summary

check-force: fmt lint coverage-summary-force ## Run format, lint, and coverage summary even if tests fail

# Docker commands (if you use Docker in the future)
docker-build: ## Build Docker image
	docker build -t hubinvestments .

# Database commands (add your DB setup commands here)
db-migrate: ## Run database migrations (customize as needed)
	@echo "Add your database migration commands here"

db-seed: ## Seed database with test data (customize as needed)
	@echo "Add your database seeding commands here"

# Migration commands for the entire project
migrate-up: ## Run all pending migrations
	go run cmd/migrate/main.go -command=up

migrate-down: ## Rollback the most recent migration
	go run cmd/migrate/main.go -command=down

migrate-version: ## Show current migration version
	go run cmd/migrate/main.go -command=version

migrate-steps: ## Run specific number of migration steps (use STEPS=n)
	go run cmd/migrate/main.go -command=steps -steps=$(or $(STEPS),1)

migrate-force: ## Force migration to specific version (use VERSION=n)
	go run cmd/migrate/main.go -command=force -version=$(or $(VERSION),0)

migrate-help: ## Show migration usage examples
	@echo "Database Migration Commands:"
	@echo "  make migrate-up             - Run all pending migrations"
	@echo "  make migrate-down           - Rollback the most recent migration"
	@echo "  make migrate-version        - Show current migration version"
	@echo "  make migrate-steps STEPS=2  - Run 2 migration steps forward"
	@echo "  make migrate-steps STEPS=-1 - Run 1 migration step backward"
	@echo "  make migrate-force VERSION=1 - Force migration to version 1 (use with caution)"
	@echo ""
	@echo "Direct CLI usage:"
	@echo "  go run cmd/migrate/main.go -command=up"
	@echo "  go run cmd/migrate/main.go -command=down"
	@echo "  go run cmd/migrate/main.go -command=version"
	@echo ""
	@echo "Custom database URL:"
	@echo "  go run cmd/migrate/main.go -command=up -db='postgres://user:pass@host/db?sslmode=disable'" 