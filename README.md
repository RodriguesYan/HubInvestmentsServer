# HubInvestmentsServer

## 🚀 Quick Start

### Environment Configuration

Before running the server, set up your environment configuration:

```bash
# Copy the example configuration
cp config.example.env config.env

# Edit config.env with your preferred settings
nano config.env
```

The `config.env` file contains:
```bash
# Server Configuration
HTTP_PORT=192.168.0.3:8080
GRPC_PORT=192.168.0.6:50051
```

**Available configurations:**
- **Production**: Use your actual IP address (default: `192.168.0.3:8080`)
- **Local Development**: Use `localhost:8080` for local testing
- **Custom**: Configure any IP:PORT combination

### Start the Server

```bash
go run main.go
```

The server will:
- Load configuration from `config.env`
- Start HTTP server on the configured `HTTP_PORT`
- Start gRPC server on the configured `GRPC_PORT`
- Display Swagger documentation URL in the startup logs

## 📖 API Documentation (Swagger)

**Access interactive Swagger documentation:**

The Swagger UI URL will be displayed in the startup logs. For the default configuration:
```
http://192.168.0.3:8080/swagger/index.html
```

**Quick access to Swagger UI:**
```bash
# Start server in background and open Swagger in browser
go run main.go &
sleep 3
# The exact URL will be shown in the server logs
```

**Available API endpoints documented:**
- `POST /login` - User authentication
- `GET /getBalance` - Get user balance (requires auth)
- `GET /getAucAggregation` - Get position aggregation (requires auth)
- `GET /getPortfolioSummary` - Get complete portfolio summary (requires auth)
- `GET /getMarketData` - Get market data with caching (requires auth)
- `DELETE /admin/market-data/cache/invalidate` - Admin cache invalidation (requires auth)
- `POST /admin/market-data/cache/warm` - Admin cache warming (requires auth)

**Swagger files generated:**
- `docs/swagger.json` - OpenAPI 2.0 specification (JSON)
- `docs/swagger.yaml` - OpenAPI 2.0 specification (YAML)
- `docs/docs.go` - Generated Go documentation
- `docs/API_DOCUMENTATION.md` - Detailed API usage guide

**Regenerate Swagger documentation:**
```bash
# Install swag CLI (if not already installed)
go install github.com/swaggo/swag/cmd/swag@latest

# Regenerate documentation from code annotations
~/go/bin/swag init
```

---

## 🗄️ Database Migrations

**HubInvestments uses a project-wide database migration system to manage schema changes safely across different environments.**

### What are Database Migrations?

Database migrations are version-controlled scripts that define incremental changes to your database schema. They allow you to:
- **Track Changes**: Every schema change is recorded as a versioned migration file
- **Environment Consistency**: Ensure development, staging, and production databases have identical structures  
- **Team Collaboration**: All developers work with the same database structure
- **Safe Deployments**: Apply schema changes automatically and safely during deployments
- **Rollback Capability**: Revert problematic changes using down migrations

### Migration Structure

All database migrations are centrally located in the shared infrastructure:
```
shared/infra/migration/
├── migration_manager.go          # Migration logic for the entire project
├── migration_manager_test.go     # Tests for migration functionality
└── sql/                          # All migration SQL files in chronological order
    ├── 000001_create_users_table.up.sql       # Users table creation
    ├── 000001_create_users_table.down.sql     # Users table rollback
    ├── 000002_create_balances_table.up.sql    # Balances table (depends on users)
    ├── 000002_create_balances_table.down.sql  # Balances table rollback
    ├── 000003_seed_initial_data.up.sql        # Initial data seeding
    ├── 000003_seed_initial_data.down.sql      # Remove initial data
    └── 000004_your_next_migration.up.sql      # Your future migrations
```

### Available Migration Commands

**Quick Commands (via Makefile):**
```bash
# Project-wide migrations
make migrate-up              # Run all pending migrations
make migrate-down            # Rollback the most recent migration  
make migrate-version         # Show current migration version
make migrate-help            # Show all migration commands and examples

# Advanced usage
make migrate-steps STEPS=2   # Run 2 migration steps forward
make migrate-steps STEPS=-1  # Run 1 migration step backward
make migrate-force VERSION=1 # Force migration to version 1 (use with caution)
```

**Direct CLI Usage:**
```bash
# Basic commands
go run cmd/migrate/main.go -command=up
go run cmd/migrate/main.go -command=down
go run cmd/migrate/main.go -command=version

# With custom database URL
go run cmd/migrate/main.go -command=up -db='postgres://user:pass@host/db?sslmode=disable'

# Advanced commands
go run cmd/migrate/main.go -command=steps -steps=2
go run cmd/migrate/main.go -command=force -version=1
```

### When to Use Migrations

**✅ Always use migrations for:**
- **Schema Changes**: Creating, altering, or dropping tables, columns, indexes
- **Cross-Module Relationships**: Foreign keys between users, balances, positions
- **Data Changes**: Inserting, updating, or deleting data that affects application logic
- **Constraint Changes**: Adding or removing foreign keys, unique constraints, check constraints
- **Index Management**: Creating or dropping database indexes for performance
- **Function/Trigger Changes**: Adding or modifying stored procedures, triggers, functions

**❌ Don't use migrations for:**
- **Application Data**: User-generated content (use application logic instead)
- **Environment-Specific Data**: Configuration that varies between environments
- **Temporary Testing Data**: Use test fixtures or seed scripts instead

### Migration Workflow

**1. Development Workflow:**
```bash
# 1. Create a new feature requiring database changes
git checkout -b feature/add-positions-table

# 2. Check current migration status
make migrate-version

# 3. Create new migration files
./scripts/create-migration.sh create_positions_table

# 4. Edit the generated SQL files to add your changes
# Edit: shared/infra/migration/sql/000004_create_positions_table.up.sql
# Edit: shared/infra/migration/sql/000004_create_positions_table.down.sql

# 5. Test migration locally
make migrate-up

# 6. Test rollback works
make migrate-down
make migrate-up

# 7. Commit migration files with your code changes
git add shared/infra/migration/sql/000004_*
git commit -m "Add positions table with user relationships"
```

**2. Team Collaboration:**
```bash
# When pulling changes from teammates
git pull origin main

# Check if new migrations are available
make migrate-version

# Apply new migrations
make migrate-up
```

**3. Production Deployment:**
```bash
# In CI/CD pipeline or deployment script
make migrate-up

# Or with production database URL
go run cmd/migrate/main.go -command=up -db="$PRODUCTION_DB_URL"
```

### Creating New Migrations

**1. File Naming Convention:**
- Format: `NNNNNN_description.up.sql` and `NNNNNN_description.down.sql`
- Use sequential numbering (000001, 000002, etc.)
- Use descriptive names: `create_positions_table`, `add_balance_currency`, `create_user_indexes`

**2. Migration Best Practices:**

**✅ Good Migration Example:**
```sql
-- 000004_add_balance_currency.up.sql
-- Migration: Add currency support to balances
-- Dependencies: 000002_create_balances_table
-- Description: Add currency column with USD default

ALTER TABLE balances 
ADD COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'USD';

CREATE INDEX idx_balances_currency ON balances(currency);

-- Add constraint to validate currency codes
ALTER TABLE balances 
ADD CONSTRAINT valid_currency 
CHECK (currency IN ('USD', 'EUR', 'GBP', 'BRL'));
```

```sql
-- 000004_add_balance_currency.down.sql  
-- Migration: Remove currency support (ROLLBACK)

ALTER TABLE balances DROP CONSTRAINT IF EXISTS valid_currency;
DROP INDEX IF EXISTS idx_balances_currency;
ALTER TABLE balances DROP COLUMN IF EXISTS currency;
```

**❌ Avoid These Patterns:**
```sql
-- Don't modify existing migrations after they've been committed
-- Don't use database-specific features without fallbacks
-- Don't forget to create corresponding down migrations
-- Don't make destructive changes without backups
```

**3. Testing Migrations:**
```bash
# Always test both directions
make migrate-up    # Apply your migration
make migrate-down  # Test rollback works
make migrate-up    # Apply again to ensure idempotency
```

### Cross-Module Relationships

One of the key advantages of project-wide migrations is managing relationships between different domains:

```sql
-- Example: 000005_create_positions_table.up.sql
CREATE TABLE positions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol VARCHAR(10) NOT NULL,
    quantity DECIMAL(15, 4) NOT NULL,
    average_price DECIMAL(15, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Cross-module constraint: ensure user has sufficient balance
    CONSTRAINT positive_quantity CHECK (quantity > 0),
    CONSTRAINT positive_price CHECK (average_price > 0)
);

-- Index for efficient user queries (used by position module)
CREATE INDEX idx_positions_user_id ON positions(user_id);
-- Index for symbol queries (used across modules)
CREATE INDEX idx_positions_symbol ON positions(symbol);
```

### Migration Timeline

Current migration history:
1. **000001**: Create users table (foundation for all user-related features)
2. **000002**: Create balances table (depends on users)
3. **000003**: Seed initial data (development/testing data)
4. **000004+**: Your future migrations (positions, market data, etc.)

### Troubleshooting

**Common Issues:**

1. **"Dirty" Migration State:**
   ```bash
   # Check migration status
   make migrate-version
   
   # If dirty, manually fix the issue and force clean state
   make migrate-force VERSION=N
   ```

2. **Migration File Not Found:**
   ```bash
   # Ensure migration files are in correct location:
   ls shared/infra/migration/sql/
   
   # Check file naming convention (must end with .up.sql/.down.sql)
   ```

3. **Database Connection Issues:**
   ```bash
   # Test database connection
   psql -h localhost -U yanrodrigues -d yanrodrigues
   
   # Use custom database URL
   go run cmd/migrate/main.go -command=version -db="your-db-url"
   ```

4. **Foreign Key Constraint Errors:**
   ```bash
   # When creating relationships, ensure parent tables exist
   # Check migration order in shared/infra/migration/sql/
   # Users must be created before balances, balances before positions
   ```

### Migration Testing

**Run migration tests:**
```bash
# Test migration functionality
go test ./shared/infra/migration/

# Test with verbose output
go test -v ./shared/infra/migration/
```

**Integration with CI/CD:**
```bash
# In your CI pipeline, always run migrations before application deployment
make migrate-up
go test ./...  # Ensure migrations don't break existing tests
```

### Advanced Usage

**Multiple Steps:**
```bash
# Move forward 2 migrations
make migrate-steps STEPS=2

# Move backward 1 migration  
make migrate-steps STEPS=-1
```

**Environment-Specific Migrations:**
```bash
# Development
make migrate-up

# Staging
go run cmd/migrate/main.go -command=up -db="$STAGING_DB_URL"

# Production  
go run cmd/migrate/main.go -command=up -db="$PRODUCTION_DB_URL"
```

**Backup Before Major Migrations:**
```bash
# Always backup before running migrations in production
pg_dump -h localhost -U yanrodrigues yanrodrigues > backup_$(date +%Y%m%d_%H%M%S).sql

# Run migrations
make migrate-up

# If something goes wrong, restore from backup
# psql -h localhost -U yanrodrigues yanrodrigues < backup_YYYYMMDD_HHMMSS.sql
```

### Creating Migrations for New Features

**Example: Adding a new positions module**
```bash
# 1. Create migration for positions table
./scripts/create-migration.sh create_positions_table

# 2. Edit the up migration
# shared/infra/migration/sql/000004_create_positions_table.up.sql

# 3. Add foreign key to users, indexes, constraints
# 4. Create corresponding down migration
# 5. Test thoroughly
# 6. Commit and deploy
```

This project-wide approach ensures all your database changes are coordinated, and relationships between users, balances, positions, and future modules are properly managed.

---

## 🗄️ Redis Setup & Configuration

**HubInvestments uses Redis for high-performance caching of market data.**

### Installation

**macOS (using Homebrew):**
```bash
# Install Redis
brew install redis

# Verify installation
redis-server --version
```

**Ubuntu/Debian:**
```bash
# Install Redis
sudo apt update
sudo apt install redis-server

# Verify installation
redis-server --version
```

**Windows:**
```bash
# Download Redis from https://redis.io/download
# Or use WSL with Ubuntu instructions above
```

### Starting Redis

**Start Redis server:**
```bash
# Option 1: Start as daemon (recommended for development)
redis-server --daemonize yes --port 6379

# Option 2: Start in foreground (for debugging)
redis-server --port 6379

# Option 3: Start with Homebrew service (macOS)
brew services start redis
```

**Verify Redis is running:**
```bash
redis-cli ping
# Expected output: PONG
```

### Redis Configuration

**Default configuration used by HubInvestments:**
- **Host:** `localhost`
- **Port:** `6379`
- **Password:** None (default)
- **Database:** `0` (default)

**Configuration location in code:**
```go
// pck/container.go
redisClient := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",  // Redis server address
    Password: "",                // No password set
    DB:       0,                 // Use default DB
})
```

### Cache Features

**📊 Market Data Caching:**
- **Pattern:** Cache-aside with automatic TTL
- **TTL:** 5 minutes for market data
- **Key format:** `market_data:SYMBOL` (e.g., `market_data:AAPL`)
- **Benefits:** Faster API responses, reduced database load

**🔧 Admin Cache Management:**
```bash
# Invalidate specific symbols (requires JWT auth)
curl -X DELETE "http://[YOUR_HTTP_PORT]/admin/market-data/cache/invalidate?symbols=AAPL,GOOGL" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Warm cache with symbols (requires JWT auth)
curl -X POST "http://[YOUR_HTTP_PORT]/admin/market-data/cache/warm?symbols=AAPL,GOOGL" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Testing Redis Integration

**Run cache-specific tests:**
```bash
# Test Redis cache functionality
go test -v ./market_data/infra/cache/ -run TestMarketDataCacheRepository

# Test with Redis running
redis-cli ping && go test ./market_data/infra/cache/
```

**Manual Redis testing:**
```bash
# Check if cache keys exist
redis-cli keys "market_data:*"

# Monitor Redis operations (for debugging)
redis-cli monitor

# View Redis info
redis-cli info memory
```

### Troubleshooting

**Common issues and solutions:**

1. **Connection refused:**
   ```bash
   # Check if Redis is running
   redis-cli ping
   
   # If not running, start Redis
   redis-server --daemonize yes
   ```

2. **Permission denied:**
   ```bash
   # Check Redis logs
   tail -f /usr/local/var/log/redis.log  # macOS
   tail -f /var/log/redis/redis-server.log  # Ubuntu
   ```

3. **Port already in use:**
   ```bash
   # Check what's using port 6379
   lsof -i :6379
   
   # Kill existing Redis process
   pkill redis-server
   ```

4. **Memory issues:**
   ```bash
   # Check Redis memory usage
   redis-cli info memory
   
   # Clear all cache (if needed)
   redis-cli flushall
   ```

### Production Considerations

**For production deployment:**
- Enable Redis authentication (`requirepass`)
- Configure Redis persistence (RDB/AOF)
- Set up Redis clustering for high availability
- Monitor Redis memory usage and performance
- Configure appropriate TTL values based on data freshness requirements

---

## 🚀 Quick Coverage Commands

**Generate coverage for the ENTIRE project and open HTML report:**

```bash
make coverage-open
```

**If you have failing tests but still want to see coverage:**

```bash
make coverage-open-force
```

**Alternative commands for the same result:**
```bash
# Using bash script (with colored output)
./scripts/coverage.sh open

# Manual step-by-step
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
```

**Other useful coverage commands:**
```bash
make coverage-summary          # Show detailed coverage summary in terminal
make coverage-summary-force    # Show coverage summary even if tests fail
make coverage                  # Show basic coverage percentages
make check                     # Run format + lint + tests + coverage summary
make check-force               # Run format + lint + coverage summary (skip test failures)
```

**Force commands (useful when some tests are failing):**
- `make coverage-open-force` - Generate and open coverage report even with failing tests
- `make coverage-html-force` - Generate HTML coverage report even with failing tests
- `make coverage-summary-force` - Show coverage summary even with failing tests
- `make check-force` - Run all checks but ignore test failures

**Note:** The regular `make coverage-open` command requires all tests to pass. If you have failing tests and want to see coverage for the passing tests, use the `-force` variants.

---

## 📊 Scripts Documentation

For detailed information about all available scripts and commands, see [scripts/README.md](scripts/README.md).

## 🎯 Development Workflow

1. **Start Redis**: `redis-server --daemonize yes`
2. **Quick coverage check**: `make coverage-open` (or `make coverage-open-force` if tests fail)
3. **View API documentation**: `go run main.go` → Check console output for Swagger URL
4. **Before committing**: `make check` (or `make check-force` to ignore test failures)
5. **While writing tests**: `./scripts/test.sh watch`
6. **Test cache functionality**: `go test ./market_data/infra/cache/`

**Quick command reference:**
- `make help` - Show all available make commands
- `make coverage-open-force` - Always works, even with failing tests
- `make check-force` - Run all checks, ignore test failures
