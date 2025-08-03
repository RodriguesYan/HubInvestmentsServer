#!/bin/bash

# Script to create new database migration files
# Usage: ./scripts/create-migration.sh <module> <description>
# Example: ./scripts/create-migration.sh balance add_currency_column

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
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

# Check arguments
if [ $# -ne 2 ]; then
    print_error "Usage: $0 <module> <description>"
    print_error "Example: $0 balance add_currency_column"
    print_error "Available modules: balance"
    exit 1
fi

MODULE=$1
DESCRIPTION=$2

# Validate module
case $MODULE in
    "balance")
        ;;
    *)
        print_error "Unsupported module: $MODULE"
        print_error "Available modules: balance"
        exit 1
        ;;
esac

# Validate description
if [[ ! $DESCRIPTION =~ ^[a-z0-9_]+$ ]]; then
    print_error "Description must contain only lowercase letters, numbers, and underscores"
    print_error "Good: add_currency_column, create_index_on_user_id"
    print_error "Bad: Add Currency Column, add-currency-column"
    exit 1
fi

# Migration directory
MIGRATION_DIR="internal/$MODULE/infra/migration/sql"

# Create migration directory if it doesn't exist
if [ ! -d "$MIGRATION_DIR" ]; then
    print_warning "Migration directory doesn't exist, creating: $MIGRATION_DIR"
    mkdir -p "$MIGRATION_DIR"
fi

# Find the next migration number
LAST_MIGRATION=$(ls -1 "$MIGRATION_DIR" 2>/dev/null | grep -E '^[0-9]{6}_.*\.up\.sql$' | sort | tail -1 | cut -d'_' -f1 || echo "000000")
NEXT_NUMBER=$(printf "%06d" $((10#$LAST_MIGRATION + 1)))

# File names
UP_FILE="${MIGRATION_DIR}/${NEXT_NUMBER}_${DESCRIPTION}.up.sql"
DOWN_FILE="${MIGRATION_DIR}/${NEXT_NUMBER}_${DESCRIPTION}.down.sql"

print_info "Creating migration files for module: $MODULE"
print_info "Migration number: $NEXT_NUMBER"
print_info "Description: $DESCRIPTION"

# Create UP migration file
cat > "$UP_FILE" << EOF
-- Migration: $DESCRIPTION
-- Module: $MODULE
-- Created: $(date '+%Y-%m-%d %H:%M:%S')
-- Description: Add your migration description here

-- TODO: Add your UP migration SQL here
-- Example:
-- CREATE TABLE new_table (
--     id SERIAL PRIMARY KEY,
--     name VARCHAR(255) NOT NULL
-- );
-- 
-- CREATE INDEX idx_new_table_name ON new_table(name);

EOF

# Create DOWN migration file
cat > "$DOWN_FILE" << EOF
-- Migration: $DESCRIPTION (ROLLBACK)
-- Module: $MODULE
-- Created: $(date '+%Y-%m-%d %H:%M:%S')
-- Description: Rollback changes from $DESCRIPTION migration

-- TODO: Add your DOWN migration SQL here (reverse of the UP migration)
-- Example:
-- DROP INDEX IF EXISTS idx_new_table_name;
-- DROP TABLE IF EXISTS new_table;

EOF

print_success "Migration files created:"
print_success "  UP:   $UP_FILE"
print_success "  DOWN: $DOWN_FILE"

print_info ""
print_info "Next steps:"
print_info "1. Edit the SQL files to add your migration logic"
print_info "2. Test the migration: make migrate-${MODULE}-up"
print_info "3. Test the rollback: make migrate-${MODULE}-down"
print_info "4. Apply again: make migrate-${MODULE}-up"
print_info "5. Commit the files: git add $MIGRATION_DIR/${NEXT_NUMBER}_*"

print_info ""
print_info "Quick commands for this migration:"
print_info "  make migrate-${MODULE}-up      # Apply migration"
print_info "  make migrate-${MODULE}-down    # Rollback migration"
print_info "  make migrate-${MODULE}-version # Check current version" 