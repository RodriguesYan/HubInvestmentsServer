# Service Startup Guide
## HubInvestments - Step 4.6.6 Testing

**Last Updated**: October 20, 2025

---

## 🚀 Quick Start

### Start All Services
```bash
cd /Users/yanrodrigues/Documents/HubInvestmentsProject
./start_all_services.sh
```

### Stop All Services
```bash
cd /Users/yanrodrigues/Documents/HubInvestmentsProject
./stop_all_services.sh
```

---

## 📋 Services Overview

### 1. HubInvestments Monolith
- **HTTP**: `localhost:8080` (direct access)
- **gRPC**: `localhost:50060`
- **Logs**: `/tmp/monolith.log`
- **Required**: ✅ Yes (core service)

### 2. API Gateway
- **HTTP**: `localhost:8080` (gateway access)
- **Logs**: `/tmp/gateway.log`
- **Required**: ✅ Yes (for Step 4.6.6)

### 3. Hub User Service
- **gRPC**: `localhost:50051`
- **Logs**: `/tmp/user-service.log`
- **Required**: ⚠️ Optional (only for authentication testing)
- **Note**: Requires PostgreSQL database

---

## 🔧 Configuration Files

### API Gateway: `.env`
**Location**: `hub-api-gateway/.env`

```env
JWT_SECRET=HubInv3stm3nts_S3cur3_JWT_K3y_2024_!@#$%^
HTTP_PORT=8080
USER_SERVICE_ADDRESS=localhost:50051
HUB_MONOLITH_ADDRESS=localhost:50060
```

**Note**: This file is auto-created by `start_all_services.sh` if missing.

---

### User Service: `config.env`
**Location**: `hub-user-service/config.env`

```env
MY_JWT_SECRET=HubInv3stm3nts_S3cur3_JWT_K3y_2024_!@#$%^
HTTP_PORT=localhost:8080
GRPC_PORT=localhost:50051
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=hub_investments
DB_SSLMODE=disable
```

---

### Monolith: `config.env`
**Location**: `HubInvestmentsServer/config.env`

```env
HTTP_PORT=localhost:8080
GRPC_PORT=localhost:50060
# ... other config
```

---

## ❌ Common Errors & Solutions

### Error 1: API Gateway - JWT_SECRET Missing
**Error**:
```
❌ Failed to load configuration: configuration validation failed: JWT_SECRET environment variable is required
```

**Solution**:
```bash
# Option 1: Use startup script (recommended)
./start_all_services.sh

# Option 2: Manual fix
cd hub-api-gateway
cat > .env << 'EOF'
JWT_SECRET=HubInv3stm3nts_S3cur3_JWT_K3y_2024_!@#$%^
HTTP_PORT=8080
USER_SERVICE_ADDRESS=localhost:50051
HUB_MONOLITH_ADDRESS=localhost:50060
EOF

# Then start manually
export $(cat .env | grep -v '^#' | xargs)
./bin/gateway
```

---

### Error 2: User Service - Database Connection Failed
**Error**:
```
Failed to connect to database: failed to connect to PostgreSQL: pq: role "postgres" does not exist
```

**Solution**:
This is **OK for Step 4.6.6 testing**. The user service is optional and only needed for authentication testing.

**If you need user service**:
1. Install PostgreSQL
2. Create database:
   ```bash
   createdb hub_investments
   createuser postgres
   ```
3. Update `hub-user-service/config.env` with correct credentials

---

### Error 3: Port Already in Use
**Error**:
```
⚠️  Port 8080 is already in use
```

**Solution**:
```bash
# Stop all services
./stop_all_services.sh

# Check ports
lsof -i :8080
lsof -i :50051
lsof -i :50060

# Kill specific process if needed
kill -9 <PID>

# Restart
./start_all_services.sh
```

---

### Error 4: JWT Secret Mismatch
**Error**:
```
⚠️  WARNING: JWT tokens will NOT be compatible with monolith unless secrets match!
```

**Solution**:
Ensure all services use the **same JWT secret**:
- Monolith: `config.env` → `JWT_SECRET`
- Gateway: `.env` → `JWT_SECRET`
- User Service: `config.env` → `MY_JWT_SECRET`

**Value**: `HubInv3stm3nts_S3cur3_JWT_K3y_2024_!@#$%^`

---

## 🧪 Testing

### Quick Health Check
```bash
# Gateway health
curl http://localhost:8080/health

# Expected: {"status":"healthy",...}
```

### Test Market Data (via Gateway → Monolith)
```bash
curl http://localhost:8080/api/v1/market-data/AAPL

# Expected: 500 error (known proto marshaling limitation)
# But connection is established!
```

### Run Complete Test Suite
```bash
cd /Users/yanrodrigues/Documents/HubInvestmentsProject/HubInvestmentsServer
./test_step_4_6_6_complete.sh
```

---

## 📊 Service Status Check

### Check Running Services
```bash
# Check all ports
lsof -i :8080   # Gateway or Monolith HTTP
lsof -i :50051  # User Service gRPC
lsof -i :50060  # Monolith gRPC

# Check processes
ps aux | grep "bin/server"          # Monolith
ps aux | grep "bin/gateway"         # Gateway
ps aux | grep "bin/hub-user-service" # User Service
```

### View Logs
```bash
# Real-time logs
tail -f /tmp/monolith.log
tail -f /tmp/gateway.log
tail -f /tmp/user-service.log

# Last 50 lines
tail -50 /tmp/monolith.log
tail -50 /tmp/gateway.log
tail -50 /tmp/user-service.log
```

---

## 🔄 Manual Service Management

### Start Services Individually

#### 1. Monolith
```bash
cd /Users/yanrodrigues/Documents/HubInvestmentsProject/HubInvestmentsServer
./bin/server > /tmp/monolith.log 2>&1 &
```

#### 2. API Gateway
```bash
cd /Users/yanrodrigues/Documents/HubInvestmentsProject/hub-api-gateway
export JWT_SECRET='HubInv3stm3nts_S3cur3_JWT_K3y_2024_!@#$%^'
export HTTP_PORT=8080
export USER_SERVICE_ADDRESS=localhost:50051
export HUB_MONOLITH_ADDRESS=localhost:50060
./bin/gateway > /tmp/gateway.log 2>&1 &
```

#### 3. User Service (Optional)
```bash
cd /Users/yanrodrigues/Documents/HubInvestmentsProject/hub-user-service
./bin/hub-user-service > /tmp/user-service.log 2>&1 &
```

---

## 🏗️ Build Services

### Build All
```bash
# Monolith
cd HubInvestmentsServer
go build -o bin/server .

# Gateway
cd ../hub-api-gateway
go build -o bin/gateway ./cmd/server

# User Service
cd ../hub-user-service
go build -o bin/hub-user-service ./cmd/server
```

---

## 🎯 Step 4.6.6 Requirements

### Minimum Required Services
For Step 4.6.6 testing, you need:
1. ✅ **Monolith** (gRPC on :50060)
2. ✅ **API Gateway** (HTTP on :8080)
3. ⚠️ **User Service** (optional, for auth testing)

### Service Dependencies
```
API Gateway (8080)
    ├── → User Service (50051) [optional]
    └── → Monolith gRPC (50060) [required]
```

---

## 📝 Configuration Checklist

Before starting services, verify:

- [ ] `HubInvestmentsServer/config.env` exists
  - [ ] `GRPC_PORT=localhost:50060`
  
- [ ] `hub-api-gateway/.env` exists (auto-created if missing)
  - [ ] `JWT_SECRET` matches monolith
  - [ ] `HUB_MONOLITH_ADDRESS=localhost:50060`
  
- [ ] `hub-user-service/config.env` exists
  - [ ] `MY_JWT_SECRET` matches monolith
  - [ ] `GRPC_PORT=localhost:50051`
  - [ ] Database credentials correct (if using)

---

## 🚨 Troubleshooting

### Gateway Won't Start
1. Check JWT_SECRET is set
2. Check port 8080 is free
3. View logs: `cat /tmp/gateway.log`

### Monolith Won't Start
1. Check port 50060 is free
2. Check config.env exists
3. View logs: `cat /tmp/monolith.log`

### User Service Won't Start
1. Check database is running
2. Check database credentials
3. This is OK for Step 4.6.6 (optional)

### Services Start But Don't Communicate
1. Check all JWT secrets match
2. Check service addresses in gateway config
3. Test connectivity: `curl http://localhost:8080/health`

---

## 📞 Support

### Test Commands
```bash
# Health check
curl http://localhost:8080/health

# Market data
curl http://localhost:8080/api/v1/market-data/AAPL

# Full test suite
./test_step_4_6_6_complete.sh
```

### Log Locations
- Monolith: `/tmp/monolith.log`
- Gateway: `/tmp/gateway.log`
- User Service: `/tmp/user-service.log`
- Test Results: `/tmp/step_4_6_6_test_results.txt`

---

## ✅ Success Indicators

When services are running correctly, you should see:

1. **Monolith**:
   ```
   ✅ Monolith started successfully
   - HTTP: localhost:8080
   - gRPC: localhost:50060
   ```

2. **API Gateway**:
   ```
   ✅ API Gateway started successfully
   - HTTP: localhost:8080
   ```

3. **Health Check**:
   ```bash
   curl http://localhost:8080/health
   # Returns: {"status":"healthy",...}
   ```

---

**Document Version**: 1.0  
**Last Updated**: October 20, 2025  
**For**: Step 4.6.6 - API Gateway Integration Testing

