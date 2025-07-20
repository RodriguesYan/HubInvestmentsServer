# HubInvestmentsServer

## 📖 API Documentation (Swagger)

**Start the server and access interactive Swagger documentation:**

```bash
# Start the server
go run main.go

# The server will start on 192.168.0.6:8080
# Swagger documentation will be available at:
# http://192.168.0.6:8080/swagger/index.html
```

**Quick access to Swagger UI:**
```bash
# Start server in background and open Swagger in browser
go run main.go &
sleep 3
open http://192.168.0.6:8080/swagger/index.html  # macOS
# or manually open: http://192.168.0.6:8080/swagger/index.html
```

**Available API endpoints documented:**
- `POST /login` - User authentication
- `GET /getBalance` - Get user balance (requires auth)
- `GET /getAucAggregation` - Get position aggregation (requires auth)
- `GET /getPortfolioSummary` - Get complete portfolio summary (requires auth)

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

1. **Quick coverage check**: `make coverage-open`
2. **View API documentation**: `go run main.go` → http://192.168.0.6:8080/swagger/index.html
3. **Before committing**: `make check` 
4. **While writing tests**: `./scripts/test.sh watch`
