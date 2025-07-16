#!/bin/bash

# HubInvestments Coverage Script
# Helper script for generating test coverage reports

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
    echo "HubInvestments Coverage Script"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  basic       Show basic coverage percentages"
    echo "  detailed    Show detailed function-level coverage"
    echo "  html        Generate HTML coverage report"
    echo "  open        Generate and open HTML coverage report"
    echo "  summary     Show comprehensive coverage summary"
    echo "  clean       Clean up generated coverage files"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 basic"
    echo "  $0 html"
    echo "  $0 open"
}

# Function to run basic coverage
run_basic_coverage() {
    print_status "Running basic test coverage..."
    go test -cover ./...
    print_success "Basic coverage complete!"
}

# Function to generate coverage profile
generate_coverage_profile() {
    print_status "Generating coverage profile..."
    go test -coverprofile=coverage.out ./...
    if [ -f coverage.out ]; then
        print_success "Coverage profile generated: coverage.out"
    else
        print_error "Failed to generate coverage profile"
        exit 1
    fi
}

# Function to show detailed coverage
run_detailed_coverage() {
    print_status "Generating detailed function-level coverage..."
    generate_coverage_profile
    
    echo ""
    print_status "Function-level coverage (sorted by percentage):"
    go tool cover -func=coverage.out | sort -k3 -nr
    
    print_success "Detailed coverage analysis complete!"
}

# Function to generate HTML report
generate_html_report() {
    print_status "Generating HTML coverage report..."
    generate_coverage_profile
    
    go tool cover -html=coverage.out -o coverage.html
    
    if [ -f coverage.html ]; then
        print_success "HTML coverage report generated: coverage.html"
    else
        print_error "Failed to generate HTML report"
        exit 1
    fi
}

# Function to open HTML report
open_html_report() {
    generate_html_report
    
    print_status "Opening coverage report in browser..."
    
    # Cross-platform open command
    if command -v open >/dev/null 2>&1; then
        # macOS
        open coverage.html
    elif command -v xdg-open >/dev/null 2>&1; then
        # Linux
        xdg-open coverage.html
    elif command -v start >/dev/null 2>&1; then
        # Windows
        start coverage.html
    else
        print_warning "Could not automatically open browser."
        print_status "Please open coverage.html manually in your browser."
    fi
    
    print_success "Coverage report opened!"
}

# Function to show comprehensive summary
show_coverage_summary() {
    print_status "Generating comprehensive coverage summary..."
    generate_coverage_profile
    
    echo ""
    echo "================================================================"
    echo "                    COVERAGE SUMMARY"
    echo "================================================================"
    
    # Overall coverage
    echo ""
    print_status "Overall Project Coverage:"
    go tool cover -func=coverage.out | tail -1
    
    # Top covered functions
    echo ""
    print_status "Top 10 Best Covered Functions:"
    go tool cover -func=coverage.out | sort -k3 -nr | head -10
    
    # Functions needing tests
    echo ""
    print_status "Functions Needing Tests (0% coverage):"
    uncovered_count=$(go tool cover -func=coverage.out | grep "0.0%" | wc -l)
    if [ "$uncovered_count" -gt 0 ]; then
        go tool cover -func=coverage.out | grep "0.0%" | head -10
        if [ "$uncovered_count" -gt 10 ]; then
            print_warning "...and $((uncovered_count - 10)) more functions with 0% coverage"
        fi
    else
        print_success "No functions with 0% coverage! 🎉"
    fi
    
    # Coverage by module
    echo ""
    print_status "Coverage by Module:"
    go tool cover -func=coverage.out | grep -E "\.go:" | awk -F: '{print $1}' | \
        sed 's|HubInvestments/||' | awk -F/ '{print $1"/"$2}' | sort | uniq -c | \
        sort -nr
    
    echo ""
    echo "================================================================"
    print_success "Coverage analysis complete!"
    print_status "💡 Tip: Run '$0 open' to view detailed line-by-line coverage"
}

# Function to clean up generated files
clean_coverage_files() {
    print_status "Cleaning up coverage files..."
    
    files_to_remove=("coverage.out" "coverage.html")
    removed_count=0
    
    for file in "${files_to_remove[@]}"; do
        if [ -f "$file" ]; then
            rm -f "$file"
            print_status "Removed: $file"
            ((removed_count++))
        fi
    done
    
    if [ $removed_count -eq 0 ]; then
        print_status "No coverage files to clean"
    else
        print_success "Cleaned up $removed_count coverage file(s)"
    fi
}

# Main script logic
main() {
    case "${1:-help}" in
        "basic")
            run_basic_coverage
            ;;
        "detailed")
            run_detailed_coverage
            ;;
        "html")
            generate_html_report
            ;;
        "open")
            open_html_report
            ;;
        "summary")
            show_coverage_summary
            ;;
        "clean")
            clean_coverage_files
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