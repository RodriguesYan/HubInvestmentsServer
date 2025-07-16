#!/bin/bash

# HubInvestments Test Script
# Comprehensive testing helper script

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show help
show_help() {
    echo "HubInvestments Test Script"
    echo ""
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  all         Run all tests (default)"
    echo "  unit        Run only unit tests"
    echo "  integration Run only integration tests"
    echo "  package     Run tests for specific package"
    echo "  verbose     Run tests with verbose output"
    echo "  race        Run tests with race detection"
    echo "  bench       Run benchmark tests"
    echo "  watch       Watch for changes and re-run tests"
    echo "  clean       Clean test cache and temp files"
    echo "  help        Show this help message"
    echo ""
    echo "Options:"
    echo "  -v, --verbose    Verbose output"
    echo "  -r, --race       Enable race detection"
    echo "  -c, --coverage   Include coverage"
    echo "  -p, --package    Specify package (e.g., ./auth/...)"
    echo ""
    echo "Examples:"
    echo "  $0 all"
    echo "  $0 unit -v"
    echo "  $0 package -p ./auth/..."
    echo "  $0 race"
}

# Function to run all tests
run_all_tests() {
    local verbose=""
    local race=""
    local coverage=""
    
    # Parse options
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--verbose)
                verbose="-v"
                shift
                ;;
            -r|--race)
                race="-race"
                shift
                ;;
            -c|--coverage)
                coverage="-cover"
                shift
                ;;
            *)
                shift
                ;;
        esac
    done
    
    print_status "Running all tests..."
    
    local cmd="go test $verbose $race $coverage ./..."
    print_status "Command: $cmd"
    
    if eval $cmd; then
        print_success "All tests passed! ✅"
    else
        print_error "Some tests failed! ❌"
        exit 1
    fi
}

# Function to run unit tests (excluding integration tests)
run_unit_tests() {
    local verbose=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--verbose)
                verbose="-v"
                shift
                ;;
            *)
                shift
                ;;
        esac
    done
    
    print_status "Running unit tests (excluding integration tests)..."
    
    # Run tests but exclude any integration test files/packages
    local cmd="go test $verbose -short ./..."
    print_status "Command: $cmd"
    
    if eval $cmd; then
        print_success "Unit tests passed! ✅"
    else
        print_error "Some unit tests failed! ❌"
        exit 1
    fi
}

# Function to run tests for specific package
run_package_tests() {
    local package="./..."
    local verbose=""
    local race=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -p|--package)
                package="$2"
                shift 2
                ;;
            -v|--verbose)
                verbose="-v"
                shift
                ;;
            -r|--race)
                race="-race"
                shift
                ;;
            *)
                shift
                ;;
        esac
    done
    
    print_status "Running tests for package: $package"
    
    local cmd="go test $verbose $race $package"
    print_status "Command: $cmd"
    
    if eval $cmd; then
        print_success "Package tests passed! ✅"
    else
        print_error "Some package tests failed! ❌"
        exit 1
    fi
}

# Function to run tests with verbose output
run_verbose_tests() {
    print_status "Running tests with verbose output..."
    
    if go test -v ./...; then
        print_success "Verbose tests completed! ✅"
    else
        print_error "Some tests failed! ❌"
        exit 1
    fi
}

# Function to run tests with race detection
run_race_tests() {
    print_status "Running tests with race detection..."
    print_warning "This may take longer than usual..."
    
    if go test -race ./...; then
        print_success "Race detection tests passed! ✅"
    else
        print_error "Race conditions detected or tests failed! ❌"
        exit 1
    fi
}

# Function to run benchmark tests
run_benchmark_tests() {
    print_status "Running benchmark tests..."
    
    # Check if there are any benchmark functions
    if grep -r "func Benchmark" . --include="*_test.go" > /dev/null 2>&1; then
        go test -bench=. ./...
        print_success "Benchmark tests completed! ✅"
    else
        print_warning "No benchmark tests found"
        print_status "💡 Tip: Create benchmark functions starting with 'Benchmark' in your test files"
    fi
}

# Function to watch for changes and re-run tests
run_watch_tests() {
    print_status "Starting test watcher..."
    print_status "Press Ctrl+C to stop watching"
    
    # Check if fswatch is available (macOS) or inotifywait (Linux)
    if command -v fswatch >/dev/null 2>&1; then
        print_status "Using fswatch for file monitoring..."
        fswatch -o . --exclude=".*\.git.*" --exclude=".*coverage\..*" --exclude=".*\.html" | while read num; do
            clear
            echo "===================================="
            echo "Files changed, re-running tests..."
            echo "===================================="
            go test ./... || true
            echo "===================================="
            echo "Waiting for changes..."
        done
    elif command -v inotifywait >/dev/null 2>&1; then
        print_status "Using inotifywait for file monitoring..."
        while inotifywait -r -e modify,create,delete --exclude=".*\.git.*" --exclude=".*coverage\..*" .; do
            clear
            echo "===================================="
            echo "Files changed, re-running tests..."
            echo "===================================="
            go test ./... || true
            echo "===================================="
            echo "Waiting for changes..."
        done
    else
        print_error "File watcher not available. Please install 'fswatch' (macOS) or 'inotify-tools' (Linux)"
        print_status "Alternative: Use 'entr' command if available"
        if command -v entr >/dev/null 2>&1; then
            print_status "Found 'entr', using it instead..."
            find . -name "*.go" | entr -c go test ./...
        else
            exit 1
        fi
    fi
}

# Function to clean test cache and temporary files
clean_test_files() {
    print_status "Cleaning test cache and temporary files..."
    
    # Clean Go test cache
    go clean -testcache
    print_status "Cleaned Go test cache"
    
    # Remove coverage files
    if [ -f coverage.out ]; then
        rm -f coverage.out
        print_status "Removed coverage.out"
    fi
    
    if [ -f coverage.html ]; then
        rm -f coverage.html
        print_status "Removed coverage.html"
    fi
    
    # Remove any test binary files
    find . -name "*.test" -type f -delete 2>/dev/null || true
    
    print_success "Test cleanup completed! ✅"
}

# Function to show test summary
show_test_summary() {
    print_status "Generating test summary..."
    
    echo ""
    echo "================================================================"
    echo "                    TEST SUMMARY"
    echo "================================================================"
    
    # Count test files
    test_files=$(find . -name "*_test.go" -type f | wc -l)
    echo "📁 Total test files: $test_files"
    
    # Count test functions
    test_functions=$(grep -r "func Test" . --include="*_test.go" | wc -l)
    echo "🧪 Total test functions: $test_functions"
    
    # Count benchmark functions
    benchmark_functions=$(grep -r "func Benchmark" . --include="*_test.go" 2>/dev/null | wc -l)
    echo "⚡ Total benchmark functions: $benchmark_functions"
    
    # Show test files by package
    echo ""
    print_status "Test files by package:"
    find . -name "*_test.go" -type f | sed 's|/[^/]*$||' | sort | uniq -c | sort -nr
    
    echo ""
    echo "================================================================"
    print_success "Test summary complete!"
}

# Main script logic
main() {
    case "${1:-all}" in
        "all")
            shift
            run_all_tests "$@"
            ;;
        "unit")
            shift
            run_unit_tests "$@"
            ;;
        "integration")
            print_warning "Integration tests not yet implemented"
            print_status "💡 Tip: Add integration tests with build tag '// +build integration'"
            ;;
        "package")
            shift
            run_package_tests "$@"
            ;;
        "verbose")
            run_verbose_tests
            ;;
        "race")
            run_race_tests
            ;;
        "bench")
            run_benchmark_tests
            ;;
        "watch")
            run_watch_tests
            ;;
        "clean")
            clean_test_files
            ;;
        "summary")
            show_test_summary
            ;;
        "help"|"--help"|"-h")
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Run the main function with all arguments
main "$@" 