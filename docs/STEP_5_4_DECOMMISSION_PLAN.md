# Step 5.4: Decommission Monolith Module - Execution Plan

## Overview

Remove `internal/market_data/` and `internal/realtime_quotes/` from the monolith, keeping only the gRPC client adapter for inter-service communication.

## What Will Be Removed

### 1. Market Data Module (`internal/market_data/`)

**To Remove**:
- `internal/market_data/infra/dto/` - DTOs (migrated to microservice)
- `internal/market_data/infra/cache/` - Cache implementation (migrated)
- `internal/market_data/infra/persistence/` - Repository (migrated)
- `internal/market_data/application/usecase/` - Use cases (migrated)
- `internal/market_data/domain/` - Domain models and interfaces (migrated)
- `internal/market_data/presentation/grpc/grpc_server.go` - gRPC server (migrated)
- `internal/market_data/presentation/grpc/market_data_grpc_handler.go` - Handler (migrated)
- `internal/market_data/presentation/grpc/market_data_grpc_server.go` - Server implementation (migrated)
- `internal/market_data/presentation/grpc/proto/` - Proto files (moved to hub-proto-contracts)
- `internal/market_data/presentation/grpc/interceptors/` - Interceptors (not needed)

**To Keep**:
- `internal/market_data/presentation/grpc/client/` - gRPC client adapter
  - `market_data_grpc_client.go` - Client implementation
  - `market_data_grpc_client_test.go` - Client tests

### 2. Real-time Quotes Module (`internal/realtime_quotes/`)

**To Remove** (entire module):
- `internal/realtime_quotes/infra/websocket/` - WebSocket handler (migrated to API Gateway)
- `internal/realtime_quotes/application/service/` - Price oscillation service (migrated)
- `internal/realtime_quotes/domain/model/` - Asset quote model (migrated)
- `internal/realtime_quotes/domain/service/` - Asset data service (migrated)
- `internal/realtime_quotes/presentation/http/` - HTTP handlers (migrated)

## Dependencies to Check

### Services Using Market Data Client

1. **Order Management System**
   - File: `internal/order_mngmt_system/infra/external/market_data_client.go`
   - Uses: `MarketDataGRPCClient` for symbol validation
   - Action: ✅ Keep - client will remain

2. **Position Service**
   - File: `internal/position/application/usecase/get_position_aggregation_usecase.go`
   - Uses: `MarketDataGRPCClient` for current prices
   - Action: ✅ Keep - client will remain

### Routes to Remove from Main

Check `cmd/server/main.go` for:
- Market data HTTP routes (if any)
- Real-time quotes WebSocket routes
- gRPC server registration for market data

## Execution Steps

### Step 1: Create Backup

```bash
cd /Users/yanrodrigues/Documents/HubInvestmentsProject/HubInvestmentsServer
mkdir -p backups/market_data_decommission_$(date +%Y%m%d)
cp -r internal/market_data backups/market_data_decommission_$(date +%Y%m%d)/
cp -r internal/realtime_quotes backups/market_data_decommission_$(date +%Y%m%d)/
```

### Step 2: Remove Market Data Modules (Except Client)

```bash
# Remove everything except the client
rm -rf internal/market_data/infra/
rm -rf internal/market_data/application/
rm -rf internal/market_data/domain/
rm -rf internal/market_data/presentation/grpc/grpc_server.go
rm -rf internal/market_data/presentation/grpc/market_data_grpc_handler.go
rm -rf internal/market_data/presentation/grpc/market_data_grpc_server.go
rm -rf internal/market_data/presentation/grpc/proto/
rm -rf internal/market_data/presentation/grpc/interceptors/
```

### Step 3: Remove Real-time Quotes Module

```bash
# Remove entire module
rm -rf internal/realtime_quotes/
```

### Step 4: Update Main Server

Remove from `cmd/server/main.go`:
- Market data gRPC server initialization
- Real-time quotes route registration
- WebSocket handler setup

### Step 5: Clean Up Imports

Search and remove unused imports:
```bash
# Find files importing removed packages
grep -r "internal/market_data" --include="*.go" --exclude-dir="client" .
grep -r "internal/realtime_quotes" --include="*.go" .
```

### Step 6: Update Documentation

- Update README.md
- Update architecture diagrams
- Add migration notes

## Verification

### 1. Verify Client Still Works

```bash
# Run tests for the gRPC client
go test ./internal/market_data/presentation/grpc/client/... -v
```

### 2. Verify Dependent Services

```bash
# Test order management system
go test ./internal/order_mngmt_system/... -v

# Test position service
go test ./internal/position/... -v
```

### 3. Build Monolith

```bash
# Ensure monolith still builds
go build -o bin/monolith cmd/server/main.go
```

### 4. Run Integration Tests

```bash
# Test that services can communicate via gRPC
make test-integration
```

## Rollback Plan

If issues arise:

```bash
# Restore from backup
cp -r backups/market_data_decommission_YYYYMMDD/market_data internal/
cp -r backups/market_data_decommission_YYYYMMDD/realtime_quotes internal/
```

## Expected Outcome

### Before Decommission
```
internal/
├── market_data/
│   ├── application/
│   ├── domain/
│   ├── infra/
│   └── presentation/
│       └── grpc/
│           ├── client/          ← Keep
│           ├── interceptors/    ← Remove
│           ├── proto/           ← Remove
│           ├── grpc_server.go   ← Remove
│           ├── market_data_grpc_handler.go  ← Remove
│           └── market_data_grpc_server.go   ← Remove
└── realtime_quotes/             ← Remove entirely
```

### After Decommission
```
internal/
└── market_data/
    └── presentation/
        └── grpc/
            └── client/          ← Only this remains
                ├── market_data_grpc_client.go
                └── market_data_grpc_client_test.go
```

## Benefits

1. **Reduced Monolith Size**: ~2,000 lines of code removed
2. **Clear Separation**: Market data logic fully in microservice
3. **Simplified Maintenance**: One source of truth for market data
4. **Better Testability**: Each service tested independently
5. **Improved Scalability**: Market data service scales independently

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Breaking dependent services | Keep gRPC client, test thoroughly |
| Lost functionality | All functionality migrated to microservice |
| Configuration issues | Update environment variables and configs |
| Deployment issues | Deploy microservice first, then decommission |

## Post-Decommission Tasks

1. **Update CI/CD**
   - Remove market data tests from monolith pipeline
   - Ensure microservice is deployed before monolith

2. **Update Documentation**
   - Architecture diagrams
   - API documentation
   - Deployment guides

3. **Monitor**
   - Check logs for errors
   - Monitor gRPC client metrics
   - Verify no 404s or connection errors

4. **Clean Up**
   - Remove unused dependencies from go.mod
   - Remove unused configuration
   - Archive old documentation

---

**Date**: November 2, 2025  
**Status**: READY FOR EXECUTION

