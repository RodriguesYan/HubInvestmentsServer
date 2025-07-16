# HubInvestments Scripts

This directory contains helper scripts to streamline development and testing workflows for the HubInvestments project.

## 📁 Available Scripts

### 🔧 Makefile Commands
The project includes a comprehensive Makefile with common development tasks:

```bash
# Show all available commands
make help

# Testing commands
make test                    # Run all tests
make test-verbose           # Run tests with verbose output
make test-race              # Run tests with race detection

# Coverage commands  
make coverage               # Show basic coverage
make coverage-func          # Show function-level coverage
make coverage-func-sorted   # Show coverage sorted by percentage
make coverage-html          # Generate HTML coverage report
make coverage-open          # Generate and open HTML coverage report
make coverage-summary       # Show comprehensive coverage summary

# Development commands
make fmt                    # Format Go code
make lint                   # Run go vet
make build                  # Build the application
make run                    # Run the application
make clean                  # Clean up generated files
make check                  # Run format, lint, tests, and coverage
```

### 📊 Coverage Script (`scripts/coverage.sh`)
Dedicated script for test coverage analysis:

```bash
# Make script executable
chmod +x scripts/coverage.sh

# Available commands
./scripts/coverage.sh basic       # Show basic coverage percentages
./scripts/coverage.sh detailed    # Show detailed function-level coverage
./scripts/coverage.sh html        # Generate HTML coverage report
./scripts/coverage.sh open        # Generate and open HTML coverage report
./scripts/coverage.sh summary     # Show comprehensive coverage summary
./scripts/coverage.sh clean       # Clean up generated coverage files
./scripts/coverage.sh help        # Show help message
```

**Features:**
- ✨ Colored output for better readability
- 📊 Comprehensive coverage analysis
- 🌐 Cross-platform browser opening
- 🧹 Easy cleanup of generated files
- 📈 Sorted coverage reports

### 🧪 Test Script (`scripts/test.sh`)
Comprehensive testing script with multiple options:

```bash
# Make script executable
chmod +x scripts/test.sh

# Available commands
./scripts/test.sh all             # Run all tests (default)
./scripts/test.sh unit            # Run only unit tests
./scripts/test.sh package -p ./auth/...  # Run tests for specific package
./scripts/test.sh verbose         # Run tests with verbose output
./scripts/test.sh race            # Run tests with race detection
./scripts/test.sh bench           # Run benchmark tests
./scripts/test.sh watch           # Watch for changes and re-run tests
./scripts/test.sh clean           # Clean test cache and temp files
./scripts/test.sh summary         # Show test summary
./scripts/test.sh help            # Show help message

# Options
./scripts/test.sh all -v          # Verbose mode
./scripts/test.sh all -r          # Race detection
./scripts/test.sh all -c          # Include coverage
```

**Features:**
- 🎯 Multiple test execution modes
- 👀 File watching for continuous testing
- 🏃‍♂️ Race condition detection
- ⚡ Benchmark test support
- 📊 Test summary statistics
- 🧹 Test cache cleaning

## 🚀 Quick Start

1. **Make scripts executable:**
   ```bash
   chmod +x scripts/*.sh
   ```

2. **Run basic coverage:**
   ```bash
   make coverage-open
   # OR
   ./scripts/coverage.sh open
   ```

3. **Run all tests with coverage:**
   ```bash
   make check
   # OR
   ./scripts/test.sh all -c
   ```

4. **Watch tests during development:**
   ```bash
   ./scripts/test.sh watch
   ```

## 📊 Coverage Analysis

Your project currently has **34.8% overall coverage**:

### 🟢 Excellent Coverage (100%)
- `portfolio_summary/application/usecase/`
- `position/application/usecase/`
- `balance/application/usecase/`

### 🟡 Good Coverage (60-87%)
- `portfolio_summary/presentation/http/` (86.7%)
- `auth/auth.go` (66.7%)
- `position/presentation/http/` (60.0%)

### 🔴 Needs Tests (0%)
- Infrastructure layers (`infra/persistence/`, `infra/dto/`)
- Service layers (`application/service/`)
- Token service
- Login functionality

## 💡 Tips

1. **Use Makefile for common tasks** - It's shorter and includes helpful combinations
2. **Use scripts for detailed analysis** - They provide more options and colored output
3. **Watch mode for development** - Use `./scripts/test.sh watch` when writing tests
4. **HTML reports for line-by-line analysis** - Use `make coverage-open` to see exactly what needs testing

## 🛠 Customization

You can easily extend these scripts by:

1. **Adding new Makefile targets** - Edit the `Makefile` to add project-specific commands
2. **Modifying script functions** - Edit the bash scripts to add new functionality
3. **Adding integration tests** - Implement integration test support in `test.sh`
4. **Adding database commands** - Extend Makefile with your database migration/seeding commands

## 📋 Prerequisites

Some features may require additional tools:

- **File watching**: Install `fswatch` (macOS) or `inotify-tools` (Linux)
- **Alternative file watching**: Install `entr` as a cross-platform alternative
- **HTML report opening**: Built-in browser opening works on macOS, Linux, and Windows

## 🔗 Related Files

- `Makefile` - Main development commands
- `scripts/coverage.sh` - Coverage analysis script  
- `scripts/test.sh` - Comprehensive testing script
- `coverage.out` - Generated coverage profile (git-ignored)
- `coverage.html` - Generated HTML coverage report (git-ignored) 