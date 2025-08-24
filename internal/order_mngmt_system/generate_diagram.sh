#!/bin/bash

# Order Worker Interaction Diagram Generator
# This script generates PNG images from Mermaid diagram files

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# Check if Mermaid CLI is installed
check_mermaid_cli() {
    if ! command -v mmdc &> /dev/null; then
        print_error "Mermaid CLI (mmdc) is not installed."
        echo
        echo "To install Mermaid CLI, run one of the following commands:"
        echo "  npm install -g @mermaid-js/mermaid-cli"
        echo "  yarn global add @mermaid-js/mermaid-cli"
        echo
        echo "Or using Docker:"
        echo "  docker pull minlag/mermaid-cli"
        echo
        exit 1
    fi
}

# Generate diagram from .mmd file
generate_diagram() {
    local mmd_file="$1"
    local output_file="${mmd_file%.mmd}.png"
    
    print_status "Generating diagram: $mmd_file -> $output_file"
    
    # Generate PNG with high quality settings
    mmdc -i "$mmd_file" -o "$output_file" \
        --width 1920 \
        --height 1080 \
        --backgroundColor white
    
    if [ $? -eq 0 ]; then
        print_success "Generated: $output_file"
    else
        print_error "Failed to generate: $output_file"
        return 1
    fi
}

# Main execution
main() {
    print_status "Order Worker Interaction Diagram Generator"
    echo
    
    # Check prerequisites
    check_mermaid_cli
    
    # Find and process .mmd files in current directory
    mmd_files=(*.mmd)
    
    if [ ${#mmd_files[@]} -eq 0 ] || [ ! -f "${mmd_files[0]}" ]; then
        print_warning "No .mmd files found in current directory"
        exit 0
    fi
    
    print_status "Found ${#mmd_files[@]} Mermaid diagram file(s)"
    echo
    
    # Generate diagrams
    success_count=0
    for mmd_file in "${mmd_files[@]}"; do
        if [ -f "$mmd_file" ]; then
            if generate_diagram "$mmd_file"; then
                ((success_count++))
            fi
        fi
    done
    
    echo
    print_success "Successfully generated $success_count diagram(s)"
    
    # List generated files
    if [ $success_count -gt 0 ]; then
        echo
        print_status "Generated files:"
        for png_file in *.png; do
            if [ -f "$png_file" ]; then
                echo "  - $png_file"
            fi
        done
    fi
}

# Run main function
main "$@"
