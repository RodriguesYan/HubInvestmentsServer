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
make coverage                  # Show basic coverage percentages
make check                     # Run format + lint + tests + coverage summary
```

---

## 📊 Scripts Documentation

For detailed information about all available scripts and commands, see [scripts/README.md](scripts/README.md).

## 🎯 Development Workflow

1. **Start Redis**: `redis-server --daemonize yes`
2. **Quick coverage check**: `make coverage-open`
3. **View API documentation**: `go run main.go` → Check console output for Swagger URL
4. **Before committing**: `make check` 
5. **While writing tests**: `./scripts/test.sh watch`
6. **Test cache functionality**: `go test ./market_data/infra/cache/`
