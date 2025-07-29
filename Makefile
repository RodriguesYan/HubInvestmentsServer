# HubInvestments Makefile
# Helper commands for development and testing

.PHONY: help test test-verbose coverage coverage-html coverage-func coverage-open coverage-open-force clean

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

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